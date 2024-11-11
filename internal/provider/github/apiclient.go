// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

func (ghc APIClient) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Create")
	opt.DebugLog(logger).Msg("GitHub:CreateOption")

	projectID, err := ghc.projectService.Create(ctx, cfg, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create github project: %w", err)
	}

	return projectID, nil
}

func (ghc APIClient) DefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:DefaultBranch")
	logger.Debug().Str("branch", branch).Msg("DefaultBranch")

	err := ghc.projectService.setDefaultBranch(ctx, owner, projectName, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (ghc APIClient) Protect(ctx context.Context, owner string, _, ownerRepoStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Protect")
	logger.Debug().Str("projectID", ownerRepoStr).Msg("GitLab:Protect")

	parts := strings.Split(ownerRepoStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected owner/repo, got %s", ownerRepoStr)
	}

	repo := parts[1]

	err := ghc.protectionService.protect(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to to protect  %s: %w", ownerRepoStr, err)
	}

	return nil
}

func (ghc APIClient) Unprotect(ctx context.Context, branch, ownerrepoStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Unprotect")
	logger.Debug().Str("ownerrepo", ownerrepoStr).Str("branch", branch).Msg("GitLab:Unprotect")

	parts := strings.Split(ownerrepoStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected owner/repo, got %s", ownerrepoStr)
	}

	owner := parts[0]
	repo := parts[1]

	err := ghc.protectionService.unprotect(ctx, branch, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to to unprotect  %s: %w", ownerrepoStr, err)
	}

	return nil
}

func (ghc APIClient) ProjectInfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("GitHub:ProjectInfos")

	projectinfos, err := ghc.projectService.getRepositoryProjectInfos(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository projectinfos: %w", err)
	}

	if filtering {
		return ghc.filterService.FilterProjectInfos(ctx, cfg, projectinfos)
	}

	return projectinfos, nil
}

func (ghc APIClient) IsValidRepositoryName(ctx context.Context, name string) bool {
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

func (ghc APIClient) Name() string {
	return config.GITHUB
}

func (ghc *APIClient) Validate(ctx context.Context, name string) bool {
	return ghc.IsValidRepositoryName(ctx, name)
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
