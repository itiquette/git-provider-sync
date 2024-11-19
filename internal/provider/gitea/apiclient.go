// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"fmt"
	"net/http"

	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	"code.gitea.io/sdk/gitea"
)

type APIClient struct {
	raw               *gitea.Client
	projectService    *ProjectService
	protectionService *ProtectionService
	filterService     *FilterService
}

func (api APIClient) CreateProject(ctx context.Context, cfg config.ProviderConfig, opt model.CreateProjectOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:CreateProject")
	opt.DebugLog(logger).Msg("Gitea:CreateOption")

	projectID, err := api.projectService.createProject(ctx, cfg, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create gitea project: err: %w", err)
	}

	return projectID, nil
}

func (api APIClient) IsValidProjectName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:IsValidProjectName")
	logger.Debug().Str("name", name).Msg("Gitea:Validate")

	return IsValidGiteaRepositoryName(name)
}

func (APIClient) Name() string {
	return config.GITEA
}

func (api APIClient) ProjectInfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("Gitea:ProjectInfos")

	projectinfos, err := api.projectService.getProjectInfos(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository infos. err: %w", err)
	}

	if filtering {
		return api.filterService.FilterProjectinfos(ctx, cfg, projectinfos)
	}

	return projectinfos, nil
}

func (api APIClient) ProtectProject(ctx context.Context, _ string, branch string, projectIDstr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:Protect")
	logger.Debug().Str("projectIDStr", projectIDstr).Str("branch", branch).Msg("Gitea:Protect")

	err := api.protectionService.protect(ctx, branch, projectIDstr)
	if err != nil {
		return fmt.Errorf("failed to to protect project. projectIDStr: %s, err: %w", projectIDstr, err)
	}

	return nil
}

func (api APIClient) SetDefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:SetDefaultBranch")
	logger.Debug().Str("branch", branch).Str("owner", owner).Str("projectName", projectName).Msg("Gitea:SetDefaultBranch")

	err := api.projectService.setDefaultBranch(ctx, owner, projectName, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (api APIClient) UnprotectProject(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:Unprotect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("branch", branch).Msg("Gitea:Unprotect")

	err := api.protectionService.unprotect(ctx, branch, projectIDStr)
	if err != nil {
		return fmt.Errorf("failed to to unprotect %s: %w", projectIDStr, err)
	}

	return nil
}

func NewGiteaAPIClient(ctx context.Context, option model.GitProviderClientOption, httpClient *http.Client) (APIClient, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:NewGiteaClient")

	clientOptions := []gitea.ClientOption{
		gitea.SetToken(option.HTTPClient.Token),
	}

	clientOptions = append(clientOptions, gitea.SetHTTPClient(httpClient))

	defaultBaseURL := "https://gitea.com"

	if option.Domain != "" {
		defaultBaseURL = option.DomainWithScheme(option.HTTPClient.Scheme)
	}

	rawClient, err := gitea.NewClient(
		defaultBaseURL,
		clientOptions...,
	)
	if err != nil {
		return APIClient{}, fmt.Errorf("failed to create a new Gitea client: %w", err)
	}

	return APIClient{
		raw:               rawClient,
		projectService:    NewProjectService(rawClient),
		protectionService: NewProtectionService(rawClient),
		filterService:     NewFilter(),
	}, nil
}