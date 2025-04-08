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

	"github.com/google/go-github/v71/github"
)

type APIClient struct {
	raw               *github.Client
	projectService    *ProjectService
	protectionService *ProtectionService
	filterService     *filterService
}

func (api APIClient) CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:CreateProject")
	opt.DebugLog(logger).Msg("GitHub:CreateOption")

	projectID, err := api.projectService.createProject(ctx, opt)
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub project. err: %w", err)
	}

	return projectID, nil
}

func (api APIClient) ProjectExists(ctx context.Context, owner, repo string) (bool, string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:ProjectExists")

	exists, projectID, err := api.projectService.Exists(ctx, owner, repo)
	if err != nil {
		logger.Error().Msg("failed to see if project existed. err:" + err.Error())

		return false, "", err
	}

	return exists, projectID, nil
}

func (api APIClient) IsValidProjectName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:IsValidProjectName")
	logger.Debug().Str("name", name).Msg("GitHub:IsValidProjectName")

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

func (api APIClient) GetProjectInfos(ctx context.Context, providerOpt model.ProviderOption, filtering bool) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:ProjectInfos")
	logger.Debug().Bool("filtering", filtering).Msg("GitHub:ProjectInfos")

	projectinfos, err := api.projectService.getProjectInfos(ctx, providerOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository infos. err: %w", err)
	}

	if filtering {
		return api.filterService.FilterProjectInfos(ctx, providerOpt, projectinfos)
	}

	return projectinfos, nil
}

func (api APIClient) Protect(ctx context.Context, owner string, _, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Protect")
	logger.Debug().Str("projectIDStr", projectIDStr).Msg("GitHub:Protect")

	_, projectName, err := splitProjectPath(projectIDStr)
	if err != nil {
		return err
	}

	err = api.protectionService.protect(ctx, owner, projectName)
	if err != nil {
		return fmt.Errorf("failed to to protect. projectIDStr: %s, err: %w", projectIDStr, err)
	}

	return nil
}

func (api APIClient) SetDefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:SetDefaultBranch")
	logger.Debug().Str("branch", branch).Str("owner", owner).Str("projectName", projectName).Msg("GitHub:SetDefaultBranch")

	err := api.projectService.setDefaultBranch(ctx, owner, projectName, branch)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (api APIClient) Unprotect(ctx context.Context, defaultBranch, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Unprotect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("defaultBranch", defaultBranch).Msg("GitHub:Unprotect:")

	owner, project, err := splitProjectPath(projectIDStr)
	if err != nil {
		return err
	}

	err = api.protectionService.unprotect(ctx, defaultBranch, owner, project)
	if err != nil {
		return fmt.Errorf("failed to to unprotect project. projectIDStr: %s, err: %w", projectIDStr, err)
	}

	return nil
}

func NewGitHubAPIClient(ctx context.Context, httpClient *http.Client, opt model.GitProviderClientOption) (APIClient, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:NewGitHubClient")

	defaultBaseURL := "https://api.github.com/"
	uploadBaseURL := "https://uploads.github.com/"

	rawClient := github.NewClient(httpClient)

	if opt.AuthCfg.Token != "" {
		rawClient = rawClient.WithAuthToken(opt.AuthCfg.Token)
	}

	if opt.Domain == "" {
		rawClient.BaseURL, _ = url.Parse(defaultBaseURL)
	}

	if opt.UploadURL == "" {
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
