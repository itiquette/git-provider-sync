// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package gitlab provides a client for interacting with GitLab repositories
// using the go-git-providers library. It offers a range of functionalities including:
//   - Creating and listing repositories
//   - Filtering repository metadata based on various criteria
//   - Validating repository names according to GitLab's rules
//   - Performing common operations on repositories
//
// This package aims to simplify GitLab interactions in Go applications, providing
// a interface for repository management and metadata handling.
package gitlab

import (
	"context"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/targetfilter"

	"github.com/xanzy/go-gitlab"
)

// Client represents a GitLab client.
type Client struct {
	rawClient *gitlab.Client
	filter    Filter
}

// Create creates a new repository in GitLab.
func (glc Client) Create(ctx context.Context, config configuration.ProviderConfig, option model.CreateOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Create:")
	config.DebugLog(logger).Msg("GitLab:Create:")

	ops := &gitlab.CreateProjectOptions{
		Name:        gitlab.Ptr(option.RepositoryName),
		Description: gitlab.Ptr(option.Description),
		Visibility:  gitlab.Ptr(toVisibility(option.Visibility)),
	}

	_, _, err := glc.rawClient.Projects.CreateProject(ops)

	if err != nil {
		return fmt.Errorf("create: failed to create %s: %w", option.RepositoryName, err)
	}

	logger.Trace().Msg("Repository created successfully")

	return nil
}

// Name returns the name of the client.
func (glc Client) Name() string {
	return configuration.GITLAB
}

// Metainfos retrieves metadata information for repositories.
func (glc Client) Metainfos(ctx context.Context, config configuration.ProviderConfig, filtering bool) ([]model.RepositoryMetainfo, error) {
	var metainfos []model.RepositoryMetainfo

	var err error

	metainfos, err = glc.getRepositoryMetaInfos(ctx, config)
	if err != nil {
		return nil, err
	}

	if filtering {
		return glc.filter.FilterMetainfo(ctx, config, metainfos, targetfilter.FilterIncludedExcludedGen(), targetfilter.IsInInterval)
	}

	return metainfos, nil
}

func (glc Client) getRepositoryMetaInfos(ctx context.Context, config configuration.ProviderConfig) ([]model.RepositoryMetainfo, error) {
	var repositories []*gitlab.Project

	var err error

	if config.IsGroup() {
		opt := &gitlab.ListGroupProjectsOptions{
			OrderBy: gitlab.Ptr("name"),
			Sort:    gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
		}
		repositories, _, err = glc.rawClient.Groups.ListGroupProjects(config.Group, opt)
	} else {
		opt := &gitlab.ListProjectsOptions{
			Owned:   gitlab.Ptr(true),
			OrderBy: gitlab.Ptr("name"),
			Sort:    gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			}}
		repositories, _, err = glc.rawClient.Projects.ListUserProjects(config.User, opt)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user repository URLs: %w", err)
	}

	metainfos := make([]model.RepositoryMetainfo, 0, len(repositories))

	for _, repo := range repositories {
		if !config.GitInfo.IncludeForks && repo.ForkedFromProject != nil {
			continue
		}

		rm, err := newRepositoryMetainfo(ctx, config, glc.rawClient, repo.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to init repository meta: %w", err)
		}

		metainfos = append(metainfos, rm)
	}

	return metainfos, nil
}

// IsValidRepositoryName checks if the given name is a valid GitLab repository name.
func (glc Client) IsValidRepositoryName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Validate:")
	logger.Debug().Str("name", name).Msg("GitLab:Validate:")

	if !IsValidGitLabRepositoryName(name) || !isValidGitLabRepositoryNameCharacters(name) {
		logger.Debug().Str("name", name).Msg("Invalid GitLab repository name")
		logger.Debug().Msg("See https://docs.gitlab.com/ee/user/reserved_names.html")

		return false
	}

	return true
}

// newRepositoryMetainfo creates a new RepositoryMetainfo instance.
func newRepositoryMetainfo(ctx context.Context, config configuration.ProviderConfig, gitClient *gitlab.Client, name string) (model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newRepositoryMeta:")
	logger.Debug().Str("name", name).Msg("newRepositoryMeta:")

	projectPath := getProjectPath(config, name)

	gitlabProject, _, err := gitClient.Projects.GetProject(projectPath, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			logger.Warn().Str("name", name).Msg("404 - repository not found. Ignoring.")

			return model.RepositoryMetainfo{}, nil
		}

		return model.RepositoryMetainfo{}, fmt.Errorf("failed to get gitlab project: %w", err)
	}

	return model.RepositoryMetainfo{
		OriginalName:   name,
		Description:    gitlabProject.Description,
		HTTPSURL:       gitlabProject.HTTPURLToRepo,
		SSHURL:         gitlabProject.SSHURLToRepo,
		DefaultBranch:  gitlabProject.DefaultBranch,
		LastActivityAt: gitlabProject.LastActivityAt,
		Visibility:     getVisibility(gitlabProject.Visibility),
	}, nil
}

// getVisibility returns the visibility of a GitLab project.
// If a public project is fetched from GitLab API without token, this field will be empty
// and is therefore set to public.
func getVisibility(vis gitlab.VisibilityValue) string {
	switch vis {
	case gitlab.PublicVisibility:
		return "public" //nolint:goconst
	case gitlab.PrivateVisibility:
		return "private"
	case gitlab.InternalVisibility:
		return "internal"
	default:
		return "public"
	}
}

// toVisibilty returns the visibility of a GitLab project.
// If a public project is fetched from GitLab API without token, this field will be empty
// and is therefore set to public.
func toVisibility(vis string) gitlab.VisibilityValue {
	switch vis {
	case "public":
		return gitlab.PublicVisibility
	case "private":
		return gitlab.PrivateVisibility
	case "internal":
		return gitlab.InternalVisibility
	default:
		return gitlab.PublicVisibility
	}
}

// NewGitLabClient creates a new GitLab client.
func NewGitLabClient(ctx context.Context, option model.GitProviderClientOption) (Client, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering NewGitLabClient:")

	client, err := gitlab.NewClient(option.Token, gitlab.WithBaseURL(option.DomainWithScheme(option.Scheme)))
	if err != nil {
		return Client{}, fmt.Errorf("failed to create a new gitlab client: %w", err)
	}

	return Client{rawClient: client}, nil
}

func getProjectPath(config configuration.ProviderConfig, name string) string {
	if config.IsGroup() {
		return config.Group + "/" + name
	}

	return config.User + "/" + name
}
