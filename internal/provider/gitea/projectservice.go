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
	opts              *ProjectOptionsBuilder
	protectionService *ProtectionService
}

func NewProjectService(client *gitea.Client) *ProjectService {
	return &ProjectService{client: client, opts: NewProjectOptionsBuilder(), protectionService: NewProtectionService(client)}
}

func (p ProjectService) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:Create")

	builder := p.opts
	builder = builder.BasicOpts(builder, opt.Visibility, opt.RepositoryName, opt.Description, opt.DefaultBranch)

	var createdRepo *gitea.Repository

	var err error

	if cfg.IsGroup() {
		createdRepo, _, err = p.client.CreateOrgRepo(cfg.Group, *builder.opts)
	} else {
		createdRepo, _, err = p.client.CreateRepo(*builder.opts)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create %s: %w", opt.RepositoryName, err)
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

func (p ProjectService) getRepositoryProjectInfos(ctx context.Context, cfg config.ProviderConfig) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:getRepositoryProjectInfos")

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

func (p ProjectService) setDefaultBranch(ctx context.Context, owner string, repoName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:setDefaultBranch")

	editOptions := gitea.EditRepoOption{
		DefaultBranch: &branch,
	}

	_, _, err := p.client.EditRepo(owner, repoName, editOptions)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
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
