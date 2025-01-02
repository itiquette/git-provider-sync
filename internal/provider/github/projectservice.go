// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"net/http"
	"time"

	"github.com/google/go-github/v68/github"
)

type ProjectService struct {
	client            *github.Client
	optBuilder        *ProjectOptionsBuilder
	protectionService *ProtectionService
}

func NewProjectService(client *github.Client) *ProjectService {
	return &ProjectService{client: client, optBuilder: NewProjectOptionsBuilder(), protectionService: NewProtectionService(client)}
}

func (p ProjectService) createProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:createProject")
	opt.DebugLog(logger).Msg("GitHub:CreateOption")

	p.optBuilder.basicOpts(opt.Visibility, opt.RepositoryName, opt.Description, opt.DefaultBranch)

	if opt.Disabled {
		p.optBuilder.disableFeatures()
	}

	groupName := ""
	if opt.IsGroup {
		groupName = opt.Owner
	}

	createdRepo, _, err := p.client.Repositories.Create(ctx, groupName, p.optBuilder.opts)
	if err != nil {
		return "", fmt.Errorf("create: failed to create project. name: %s, err: %w", opt.RepositoryName, err)
	}

	logger.Trace().Str("name", opt.RepositoryName).Msg("Project created successfully")

	return *createdRepo.FullName, nil
}

func (p ProjectService) newProjectInfo(ctx context.Context, opt model.ProviderOption, name string) (model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:newProjectInfo")
	logger.Debug().Str("name", name).Str("providerOption", opt.String()).Msg("newProjectInfo")

	gitHubProject, _, err := p.client.Repositories.Get(ctx, opt.Owner, name)
	if err != nil {
		return model.ProjectInfo{}, fmt.Errorf("failed to get projectInfo. name: %s, err: %w", name, err)
	}

	return model.ProjectInfo{
		OriginalName:   name,
		Description:    getValueOrEmpty(gitHubProject.Description),
		HTTPSURL:       getValueOrEmpty(gitHubProject.CloneURL),
		SSHURL:         getValueOrEmpty(gitHubProject.SSHURL),
		DefaultBranch:  getValueOrEmpty(gitHubProject.DefaultBranch),
		LastActivityAt: getTimeOrNil(gitHubProject.UpdatedAt),
		Visibility:     getValueOrEmpty(gitHubProject.Visibility),
		ProjectID:      getValueOrEmpty(gitHubProject.FullName),
	}, nil
}

func (p ProjectService) Exists(ctx context.Context, owner, repo string) (bool, string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Exists")

	project, resp, err := p.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil // Repository doesn't exist
		}

		return false, "", err //nolint
	}

	projectID := getValueOrEmpty(project.FullName)

	return true, projectID, nil
}

func (p ProjectService) getProjectInfos(ctx context.Context, providerOpt model.ProviderOption) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:getProjectinfos")

	var allRepos []*github.Repository

	listType := "sources"
	if providerOpt.IncludeForks {
		listType = "all"
	}

	if providerOpt.IsGroup() {
		opt := &github.RepositoryListByOrgOptions{
			Type:        listType,
			Sort:        "full_name",
			ListOptions: github.ListOptions{PerPage: 100}, // GitHub's max is 100
		}

		for {
			repos, resp, err := p.client.Repositories.ListByOrg(ctx, providerOpt.Owner, opt)
			if err != nil {
				return nil, fmt.Errorf("failed to list org repositories. page: %d, err: %w", opt.Page, err)
			}

			allRepos = append(allRepos, repos...)

			if resp.NextPage == 0 {
				break
			}

			opt.Page = resp.NextPage
		}
	} else {
		opt := &github.RepositoryListByAuthenticatedUserOptions{
			Visibility:  "all",
			Affiliation: "owner",
			Sort:        "full_name",
			ListOptions: github.ListOptions{PerPage: 100}, // GitHub's max is 100
		}

		for {
			repos, resp, err := p.client.Repositories.ListByAuthenticatedUser(ctx, opt)
			if err != nil {
				return nil, fmt.Errorf("failed to list user repositories. page: %d, err: %w", opt.Page, err)
			}

			allRepos = append(allRepos, repos...)

			if resp.NextPage == 0 {
				break
			}

			opt.Page = resp.NextPage
		}
	}

	logger.Debug().Int("total_repositories", len(allRepos)).Msg("Total fetched repositories projectinfo")

	var projectinfos []model.ProjectInfo //nolint:prealloc

	for _, repo := range allRepos {
		if !providerOpt.IncludeForks && repo.Fork != nil && *repo.Fork {
			continue
		}

		name := repo.GetName()
		metainfo, err := p.newProjectInfo(ctx, providerOpt, name)

		if err != nil {
			logger.Warn().Err(err).Str("repo", name).Msg("failed to create projectinfo")

			continue
		}

		projectinfos = append(projectinfos, metainfo)
	}

	return projectinfos, nil
}

func (p ProjectService) setDefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:setDefaultBranch")

	_, _, err := p.client.Repositories.Edit(ctx, owner, projectName, &github.Repository{
		DefaultBranch: github.Ptr(branch),
	})
	if err != nil {
		return fmt.Errorf("failed to set default branch. err: %w", err)
	}

	return nil
}

// getValueOrEmpty is a helper function that returns the value of a string pointer if it's not nil,
// or an empty string otherwise.
func getValueOrEmpty(s *string) string {
	if s != nil {
		return *s
	}

	return "N/A"
}

// getTimeOrNil is a helper function that converts a GitHub Timestamp to a standard time.Time pointer,
// or returns nil if the input is nil.
func getTimeOrNil(t *github.Timestamp) *time.Time {
	if t != nil {
		return &t.Time
	}

	return nil
}
