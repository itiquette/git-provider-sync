// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"net/http"

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

func (p ProjectService) createProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:creatNeProject")
	opt.DebugLog(logger).Msg("gitea:CreateOption")

	p.optBuilder.BasicOpts(opt.Visibility, opt.RepositoryName, opt.Description, opt.DefaultBranch)

	var createdRepo *gitea.Repository

	var err error

	if opt.IsGroup {
		createdRepo, _, err = p.client.CreateOrgRepo(opt.Owner, *p.optBuilder.opts)
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

func (p ProjectService) newProjectInfo(ctx context.Context, rawClient *gitea.Client, repositoryName string, opt model.ProviderOption) (model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:newProjectInfo")

	giteaProject, _, err := rawClient.GetRepo(opt.Owner, repositoryName)
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

func (p ProjectService) getProjectInfos(ctx context.Context, providerOpt model.ProviderOption) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:getProjectInfos")

	var (
		repositories []*gitea.Repository
		err          error
	)

	if providerOpt.IsGroup() {
		opt := gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     -1, // Set to -1 to get all items
				PageSize: -1,
			},
		}
		repositories, _, err = p.client.ListOrgRepos(providerOpt.Owner, opt)
	} else {
		opt := gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     -1,
				PageSize: -1,
			},
		}

		repositories, _, err = p.client.ListUserRepos(providerOpt.User, opt)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	logger.Debug().Int("total_repositories", len(repositories)).Msg("Found repositories")

	var projectinfos []model.ProjectInfo //nolint:prealloc

	for _, repo := range repositories {
		if !providerOpt.IncludeForks && repo.Fork {
			continue
		}

		rm, _ := p.newProjectInfo(ctx, p.client, repo.Name, providerOpt)
		projectinfos = append(projectinfos, rm)
	}

	return projectinfos, nil
}

func (p ProjectService) Exists(ctx context.Context, owner, repo string) (bool, string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:Exists")

	repository, resp, err := p.client.GetRepo(owner, repo)

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil // Repository doesn't exist
		}

		return false, "", err //nolint
	}

	return repository != nil, repository.FullName, nil
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
