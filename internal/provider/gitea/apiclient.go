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

func (api APIClient) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:Create")
	opt.DebugLog(logger).Msg("Gitea:CreateOption")

	projectID, err := api.projectService.Create(ctx, cfg, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create gitea project: %w", err)
	}

	return projectID, nil
}

func (api APIClient) DefaultBranch(ctx context.Context, owner string, repoName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:DefaultBranch")
	logger.Debug().Str("branch", branch).Str("owner", owner).Str("repoName", repoName).Msg("Gitea:DefaultBranch")

	err := api.projectService.setDefaultBranch(ctx, owner, repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (api APIClient) Protect(ctx context.Context, _ string, branch string, projectIDstr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:Protect")
	logger.Debug().Str("projectIDStr", projectIDstr).Str("branch", branch).Msg("Gitea:Protect")

	err := api.protectionService.protect(ctx, branch, projectIDstr)
	if err != nil {
		return fmt.Errorf("failed to to protect  %s: %w", projectIDstr, err)
	}

	return nil
}

func (api APIClient) Unprotect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:Unprotect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("branch", branch).Msg("Gitea:Unprotect")

	err := api.protectionService.unprotect(ctx, branch, projectIDStr)
	if err != nil {
		return fmt.Errorf("failed to to unprotect %s: %w", projectIDStr, err)
	}

	return nil
}

func (api APIClient) ProjectInfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("Gitea:ProjectInfos")

	projectinfos, err := api.projectService.getRepositoryProjectInfos(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository infos: %w", err)
	}

	if filtering {
		return api.filterService.FilterProjectinfos(ctx, cfg, projectinfos)
	}

	return projectinfos, nil
}

func (api APIClient) IsValidRepositoryName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:IsValidRepositoryName")
	logger.Debug().Str("name", name).Msg("Gitea:Validate")

	return IsValidGiteaRepositoryName(name)
}

func (APIClient) Name() string {
	return config.GITEA
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
