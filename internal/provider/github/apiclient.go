// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/google/go-github/v66/github"
)

type APIClient struct {
	raw               *github.Client
	projectService    *ProjectService
	protectionService *ProtectionService
	filterService     *filterService
}

func (api APIClient) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Create")
	opt.DebugLog(logger).Msg("GitHub:CreateOption")

	projectID, err := api.projectService.Create(ctx, cfg, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create github project: %w", err)
	}

	return projectID, nil
}

func (api APIClient) DefaultBranch(ctx context.Context, owner string, repoName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:DefaultBranch")
	logger.Debug().Str("branch", branch).Str("owner", owner).Str("repoName", repoName).Msg("GitHub:DefaultBranch")

	err := api.projectService.setDefaultBranch(ctx, owner, repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (api APIClient) Protect(ctx context.Context, owner string, _, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Protect")
	logger.Debug().Str("projectIDStr", projectIDStr).Msg("GitHub:Protect")

	_, repo := splitProjectPath(projectIDStr)

	err := api.protectionService.protect(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to to protect  %s: %w", projectIDStr, err)
	}

	return nil
}

func (api APIClient) Unprotect(ctx context.Context, branch, ownerRepoStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Unprotect")
	logger.Debug().Str("ownerrepo", ownerRepoStr).Str("branch", branch).Msg("GitHub:Unprotect")

	owner, repo := splitProjectPath(ownerRepoStr)

	err := api.protectionService.unprotect(ctx, branch, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to to unprotect  %s: %w", ownerRepoStr, err)
	}

	return nil
}

func (api APIClient) ProjectInfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("GitHub:ProjectInfos")

	projectinfos, err := api.projectService.getRepositoryProjectInfos(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository infos: %w", err)
	}

	if filtering {
		return api.filterService.FilterProjectInfos(ctx, cfg, projectinfos)
	}

	return projectinfos, nil
}

func (api APIClient) IsValidRepositoryName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:IsValidRepositoryName")
	logger.Debug().Str("name", name).Msg("IsValidRepositoryName")

	if !IsValidGitHubRepositoryName(name) {
		logger.Debug().Str("name", name).Msg("Invalid GitHub repository name")
		logger.Debug().Msg("See https://github.com/dead-claudia/github-limits?tab=readme-ov-file#repository-names")

		return false
	}

	return true
}

func (api APIClient) Name() string {
	return config.GITHUB
}

func (api APIClient) Validate(ctx context.Context, name string) bool {
	return api.IsValidRepositoryName(ctx, name)
}

func NewGitHubAPIClient(ctx context.Context, option model.GitProviderClientOption, httpClient *http.Client) (APIClient, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:NewGitHubClient")

	defaultBaseURL := "https://api.github.com/"
	uploadBaseURL := "https://uploads.github.com/"

	rawClient := github.NewClient(httpClient)

	if option.HTTPClient.Token != "" {
		rawClient = rawClient.WithAuthToken(option.HTTPClient.Token)
	}

	if option.Domain == "" {
		rawClient.BaseURL, _ = url.Parse(defaultBaseURL)
	}

	if option.UploadURL == "" {
		rawClient.UploadURL, _ = url.Parse(uploadBaseURL)
	}

	// TODO: secondary rate limiting check

	return APIClient{
		raw:               rawClient,
		projectService:    NewProjectService(rawClient),
		protectionService: NewProtectionService(rawClient),
		filterService:     NewFilter(),
	}, nil
}
