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

func (glc APIClient) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Create")
	opt.DebugLog(logger).Msg("GitLab:CreateOption")

	projectID, err := glc.projectService.Create(ctx, cfg, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create github project: %w", err)
	}

	return projectID, nil
}

func (glc APIClient) DefaultBranch(ctx context.Context, owner, name, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:DefaultBranch")
	logger.Debug().Str("branch", branch).Str("owner", owner).Str("name", name).Msg("GitLab:DefaultBranch")

	err := glc.projectService.setDefaultBranch(ctx, owner, name, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (glc APIClient) Protect(ctx context.Context, _ string, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Protect")
	logger.Debug().Str("projectID", projectIDStr).Str("branch", branch).Msg("GitLab:Protect")

	err := glc.protectionService.protect(ctx, branch, projectIDStr)
	if err != nil {
		return fmt.Errorf("failed to to protect  %s: %w", projectIDStr, err)
	}

	return nil
}

func (glc APIClient) Unprotect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Unprotect")
	logger.Debug().Str("projectID", projectIDStr).Str("branch", branch).Msg("GitLab:Unprotect")

	err := glc.protectionService.unprotect(ctx, branch, projectIDStr)
	if err != nil {
		return fmt.Errorf("failed to to unprotect  %s: %w", projectIDStr, err)
	}

	return nil
}

func (glc APIClient) ProjectInfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("GitLab:ProjectInfos")

	projectinfos, err := glc.projectService.getRepositoryProjectInfos(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository projectinfos: %w", err)
	}

	if filtering {
		return glc.filterService.FilterProjectinfos(ctx, cfg, projectinfos, targetfilter.FilterIncludedExcludedGen(), targetfilter.IsInInterval)
	}

	return projectinfos, nil
}

// IsValidRepositoryName checks if the given name is a valid GitLab repository name.
func (glc APIClient) IsValidRepositoryName(ctx context.Context, name string) bool {
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

func (glc APIClient) Name() string {
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
