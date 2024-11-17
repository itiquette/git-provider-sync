// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	"code.gitea.io/sdk/gitea"
)

type ProjectService struct {
	client            *gitea.Client
	optBuilder        *ProjectOptionsBuilder
	protectionService *ProtectionService
}

func NewProjectService(client *gitea.Client) *ProjectService {
	return &ProjectService{client: client, optBuilder: NewProjectOptionsBuilder(), protectionService: NewProtectionService(client)}
}

func (p ProjectService) createProject(ctx context.Context, cfg config.ProviderConfig, opt model.CreateProjectOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:creatNeProject")
	opt.DebugLog(logger).Msg("gitea:CreateOption")

	p.optBuilder.BasicOpts(opt.Visibility, opt.RepositoryName, opt.Description, opt.DefaultBranch)

	var createdRepo *gitea.Repository

	var err error

	if cfg.IsGroup() {
		createdRepo, _, err = p.client.CreateOrgRepo(cfg.Group, *p.optBuilder.opts)
	} else {
		createdRepo, _, err = p.client.CreateRepo(*p.optBuilder.opts)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create project. name: %s, err: %w", opt.RepositoryName, err)
	}

	if opt.Disabled {
		err = p.ApplyDisabledSettings(ctx, createdRepo.Owner.UserName, opt.RepositoryName)
		if err != nil {
			return "", fmt.Errorf("failed to apply disabled settings for repo %s: %w", opt.RepositoryName, err)
		}
	}

	logger.Trace().Msg("Repository created successfully")

	return createdRepo.FullName, nil
}

func (p ProjectService) newProjectInfo(ctx context.Context, config config.ProviderConfig, rawClient *gitea.Client, repositoryName string) (model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:newProjectInfo")

	owner := config.Group
	if !config.IsGroup() {
		owner = config.User
	}

	giteaProject, _, err := rawClient.GetRepo(owner, repositoryName)
	if err != nil {
		return model.ProjectInfo{}, fmt.Errorf("failed to get project info for %s: %w", repositoryName, err)
	}

	return model.ProjectInfo{
		OriginalName:   repositoryName,
		HTTPSURL:       giteaProject.CloneURL,
		SSHURL:         giteaProject.SSHURL,
		Description:    giteaProject.Description,
		DefaultBranch:  giteaProject.DefaultBranch,
		LastActivityAt: &giteaProject.Updated,
		Visibility:     string(giteaProject.Owner.Visibility),
		ProjectID:      giteaProject.FullName,
	}, nil
}

func (p ProjectService) getProjectInfos(ctx context.Context, cfg config.ProviderConfig) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:getProjectInfos")

	var (
		repositories []*gitea.Repository
		err          error
	)

	if cfg.IsGroup() {
		opt := gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     -1, // Set to -1 to get all items
				PageSize: -1,
			},
		}
		repositories, _, err = p.client.ListOrgRepos(cfg.Group, opt)
	} else {
		opt := gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     -1,
				PageSize: -1,
			},
		}

		repositories, _, err = p.client.ListUserRepos(cfg.User, opt)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	logger.Debug().Int("total_repositories", len(repositories)).Msg("Found repositories")

	var projectinfos []model.ProjectInfo //nolint:prealloc

	for _, repo := range repositories {
		if !cfg.Git.IncludeForks && repo.Fork {
			continue
		}

		rm, _ := p.newProjectInfo(ctx, cfg, p.client, repo.Name)
		projectinfos = append(projectinfos, rm)
	}

	return projectinfos, nil
}

func (p ProjectService) setDefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:setDefaultBranch")

	editOptions := gitea.EditRepoOption{
		DefaultBranch: &branch,
	}

	_, _, err := p.client.EditRepo(owner, projectName, editOptions)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}
