// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"errors"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"net/http"
	"strconv"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type ProjectService struct {
	client            *gitlab.Client
	opts              *ProjectOptionsBuilder
	protectionService *ProtectionService
}

func NewProjectService(client *gitlab.Client) *ProjectService {
	return &ProjectService{client: client, opts: NewProjectOptionsBuilder(), protectionService: NewProtectionService(client)}
}

func (p ProjectService) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Create")

	namespaceID, err := p.getNamespaceID(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("get namespace ID: %w", err)
	}

	createdRepo, err := p.createProject(ctx, opt, namespaceID)
	if err != nil {
		return "", fmt.Errorf("failed to create %s: %w", opt.RepositoryName, err)
	}

	logger.Debug().Msg("Repository created successfully")

	return strconv.Itoa(createdRepo.ID), nil
}

func (p ProjectService) getNamespaceID(ctx context.Context, cfg config.ProviderConfig) (int, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering getNamespaceID")

	if !cfg.IsGroup() {
		return 0, nil
	}

	groups, resp, err := p.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		Search: gitlab.Ptr(cfg.Group),
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return 0, errors.New("authentication failed: please check your token permissions")
		}

		return 0, fmt.Errorf("search for group: %w", err)
	}

	if len(groups) == 0 {
		return 0, fmt.Errorf("no group found with name: %s", cfg.Group)
	}

	return groups[0].ID, nil
}

func (p ProjectService) newProjectInfo(ctx context.Context, cfg config.ProviderConfig, name string) (model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:newProjectInfo")
	logger.Debug().
		Str("usr/grp", cfg.User+cfg.Group).
		Str("name", name).
		Str("provider", cfg.ProviderType).
		Str("domain", cfg.GetDomain()).
		Msg("GitLab:newProjectInfo")

	projectPath := getProjectPath(cfg, name)

	gitlabProject, _, err := p.client.Projects.GetProject(projectPath, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			logger.Warn().Str("name", name).Msg("Repository not found. Ignoring.")

			return model.ProjectInfo{}, nil
		}

		return model.ProjectInfo{}, fmt.Errorf("get gitlab project: %w", err)
	}

	return model.ProjectInfo{
		OriginalName:   name,
		Description:    gitlabProject.Description,
		HTTPSURL:       gitlabProject.HTTPURLToRepo,
		SSHURL:         gitlabProject.SSHURLToRepo,
		DefaultBranch:  gitlabProject.DefaultBranch,
		LastActivityAt: gitlabProject.LastActivityAt,
		ProjectID:      strconv.Itoa(gitlabProject.ID),
		Visibility:     getVisibility(gitlabProject.Visibility),
	}, nil
}

func (p ProjectService) getRepositoryProjectInfos(ctx context.Context, cfg config.ProviderConfig) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:getRepositoryProjectInfos")

	var allRepositories []*gitlab.Project

	if cfg.IsGroup() {
		opt := &gitlab.ListGroupProjectsOptions{
			OrderBy:     gitlab.Ptr("name"),
			Sort:        gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{PerPage: 100}, //TODO: add archived support, consider projectinfo struct
		}

		for {
			repositories, resp, err := p.client.Groups.ListGroupProjects(cfg.Group, opt)
			if err != nil {
				return nil, fmt.Errorf("fetch group repositories page %d: %w", opt.Page, err)
			}

			allRepositories = append(allRepositories, repositories...)

			if resp.CurrentPage >= resp.TotalPages {
				break
			}

			opt.Page = resp.NextPage
		}
	} else {
		opt := &gitlab.ListProjectsOptions{
			Owned:       gitlab.Ptr(true),
			OrderBy:     gitlab.Ptr("name"),
			Sort:        gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{PerPage: 100},
		}

		for {
			repositories, resp, err := p.client.Projects.ListUserProjects(cfg.User, opt)
			if err != nil {
				return nil, fmt.Errorf("fetch user repositories page %d: %w", opt.Page, err)
			}

			allRepositories = append(allRepositories, repositories...)

			if resp.CurrentPage >= resp.TotalPages {
				break
			}

			opt.Page = resp.NextPage
		}
	}

	logger.Debug().Int("total_repositories", len(allRepositories)).Msg("Found repositories")

	projectinfos := make([]model.ProjectInfo, 0, len(allRepositories))

	for _, repo := range allRepositories {
		if !cfg.Git.IncludeForks && repo.ForkedFromProject != nil {
			continue
		}

		rm, err := p.newProjectInfo(ctx, cfg, repo.Path)
		if err != nil {
			return nil, fmt.Errorf("init repository meta for %s: %w", repo.Path, err)
		}

		projectinfos = append(projectinfos, rm)
	}

	return projectinfos, nil
}

func (p ProjectService) setDefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering setDefaultBranch")

	_, _, err := p.client.Projects.EditProject(owner+"/"+projectName, &gitlab.EditProjectOptions{
		DefaultBranch: gitlab.Ptr(branch),
	})
	if err != nil {
		return fmt.Errorf("edit project default branch: %w", err)
	}

	return nil
}

func (p ProjectService) createProject(ctx context.Context, opt model.CreateOption, namespaceID int) (*gitlab.Project, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering createProject")
	logger.Debug().Int("namespaceID", namespaceID).Msg("Entering createProject")

	builder := p.opts
	builder = builder.BasicOpts(builder, opt.Visibility, opt.RepositoryName, opt.Description, opt.DefaultBranch, namespaceID)

	if opt.Disabled {
		builder = p.opts.DisableOpts(builder)
	}

	createdRepo, _, err := p.client.Projects.CreateProject(builder.opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", opt.RepositoryName, err)
	}

	return createdRepo, nil
}

func getProjectPath(cfg config.ProviderConfig, name string) string {
	if cfg.IsGroup() {
		return cfg.Group + "/" + name
	}

	return cfg.User + "/" + name
}
