// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/targetfilter"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// APIClient represents a facade to GitLab API operations.
type APIClient struct {
	raw               *gitlab.Client
	projectService    interfaces.ProjectServicer
	protectionService interfaces.ProtectionServicer
	filterService     interfaces.FilterServicer
}

func (api APIClient) CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:CreateProject")
	opt.DebugLog(logger).Msg("GitLab:CreateOption")

	projectIDStr, err := api.projectService.CreateProject(ctx, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create a GitLab project. err: %w", err)
	}

	return projectIDStr, nil
}

func (api APIClient) ProjectExists(ctx context.Context, owner, repo string) (bool, string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:ProjectExists")

	exists, projectID, err := api.projectService.ProjectExists(ctx, owner, repo)
	if err != nil {
		return false, "", fmt.Errorf("failed to see if program existed. err: %w", err)
	}

	return exists, projectID, nil
}

func (api APIClient) IsValidProjectName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:IsValidProjectName")
	logger.Debug().Str("name", name).Msg("GitLab:IsValidProjectName")

	if !IsValidGitLabName(name) || !isValidGitLabNameCharacters(name) {
		logger.Debug().Str("name", name).Msg("Invalid GitLab repository name")
		logger.Debug().Msg("See https://docs.gitlab.com/ee/user/reserved_names.html")

		return false
	}

	return true
}

func (APIClient) Name() string {
	return config.GITLAB
}

func (api APIClient) GetProjectInfos(ctx context.Context, providerOpt model.ProviderOption, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("GitLab:ProjectInfos")

	projectInfos, err := api.projectService.GetProjectInfos(ctx, providerOpt, filtering)
	if err != nil {
		return nil, fmt.Errorf("failed to get project infos. err: %w", err)
	}

	if filtering {
		return api.filterService.FilterProjectinfos(ctx, providerOpt, projectInfos, targetfilter.FilterIncludedExcludedGen(), targetfilter.IsInInterval) //nolint
	}

	return projectInfos, nil
}

func (api APIClient) Protect(ctx context.Context, _ string, defaultBranch string, projectIDstr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Protect")
	logger.Debug().Str("defaultBranch", defaultBranch).Str("projectIDStr", projectIDstr).Msg("GitLab:Protect")

	err := api.protectionService.Protect(ctx, "", defaultBranch, projectIDstr)
	if err != nil {
		return fmt.Errorf("failed to to protect project. defaultBranch: %s, projectIDstr: %s, err: %w", defaultBranch, projectIDstr, err)
	}

	return nil
}

func (api APIClient) SetDefaultBranch(ctx context.Context, owner, projectName, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:SetDefaultBranch")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Str("branch", branch).Msg("GitLab:SetDefaultBranch")

	err := api.projectService.SetDefaultBranch(ctx, owner, projectName, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %s, projectName: %s, owner: %s, err: %w", branch, projectName, owner, err)
	}

	return nil
}

func (api APIClient) Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Unprotect")
	logger.Debug().Str("defaultBranch", defaultBranch).Str("projectIDStr", projectIDStr).Msg("GitLab:Unprotect")

	err := api.protectionService.Unprotect(ctx, defaultBranch, projectIDStr)
	if err != nil {
		return fmt.Errorf("failed to unprotect project. projectIDStr: %s, err: %w", projectIDStr, err)
	}

	return nil
}

func NewGitLabAPIClient(ctx context.Context, httpClient *http.Client, opt model.GitProviderClientOption) (APIClient, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:NewGitLabClient")

	defaultBaseURL := "https://gitlab.com/"
	if opt.Domain != "" {
		defaultBaseURL = opt.DomainWithScheme(opt.AuthCfg.HTTPScheme)
	}

	rawClient, err := gitlab.NewClient(opt.AuthCfg.Token,
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
		filterService:     NewFilter(),
	}, nil
}
