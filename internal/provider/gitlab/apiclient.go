// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/targetfilter"

	"github.com/xanzy/go-gitlab"
)

// APIClient represents GitLab APIClient Facade operations.
type APIClient struct {
	raw               *gitlab.Client
	projectService    *ProjectService
	protectionService *ProtectionService
	filterService     *filterService
}

func (api APIClient) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Create")
	opt.DebugLog(logger).Msg("GitLab:CreateOption")

	projectID, err := api.projectService.Create(ctx, cfg, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create github project: %w", err)
	}

	return projectID, nil
}

func (api APIClient) DefaultBranch(ctx context.Context, owner, repoName, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:DefaultBranch")
	logger.Debug().Str("branch", branch).Str("owner", owner).Str("repoName", repoName).Msg("GitLab:DefaultBranch")

	err := api.projectService.setDefaultBranch(ctx, owner, repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (api APIClient) Protect(ctx context.Context, _ string, branch string, projectIDstr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Protect")
	logger.Debug().Str("projectIDStr", projectIDstr).Str("branch", branch).Msg("GitLab:Protect")

	err := api.protectionService.protect(ctx, branch, projectIDstr)
	if err != nil {
		return fmt.Errorf("failed to to protect  %s: %w", projectIDstr, err)
	}

	return nil
}

func (api APIClient) Unprotect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Unprotect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("branch", branch).Msg("GitLab:Unprotect")

	err := api.protectionService.unprotect(ctx, branch, projectIDStr)
	if err != nil {
		return fmt.Errorf("failed to to unprotect %s: %w", projectIDStr, err)
	}

	return nil
}

func (api APIClient) ProjectInfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("GitLab:ProjectInfos")

	projectinfos, err := api.projectService.getRepositoryProjectInfos(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository infos: %w", err)
	}

	if filtering {
		return api.filterService.FilterProjectinfos(ctx, cfg, projectinfos, targetfilter.FilterIncludedExcludedGen(), targetfilter.IsInInterval)
	}

	return projectinfos, nil
}

func (api APIClient) IsValidRepositoryName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:IsValidRepositoryName")
	logger.Debug().Str("name", name).Msg("IsValidRepositoryName")

	if !IsValidGitLabRepositoryName(name) || !isValidGitLabRepositoryNameCharacters(name) {
		logger.Debug().Str("name", name).Msg("Invalid GitLab repository name")
		logger.Debug().Msg("See https://docs.gitlab.com/ee/user/reserved_names.html")

		return false
	}

	return true
}

func (APIClient) Name() string {
	return config.GITLAB
}

func NewGitLabAPIClient(ctx context.Context, option model.GitProviderClientOption, httpClient *http.Client) (APIClient, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:NewGitLabClient")

	defaultBaseURL := "https://gitlab.com/"
	if option.Domain != "" {
		defaultBaseURL = option.DomainWithScheme(option.HTTPClient.Scheme)
	}

	rawClient, err := gitlab.NewClient(option.HTTPClient.Token,
		gitlab.WithBaseURL(defaultBaseURL),
		gitlab.WithHTTPClient(httpClient),
	)
	if err != nil {
		return APIClient{}, fmt.Errorf("create new GitLab client: %w", err)
	}

	return APIClient{
		raw:               rawClient,
		projectService:    NewProjectService(rawClient),
		protectionService: NewProtectionService(rawClient),
		filterService:     NewFilter(nil),
	}, nil
}
