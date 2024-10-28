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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/targetfilter"

	"github.com/xanzy/go-gitlab"
)

// Client represents a GitLab client.
type Client struct {
	rawClient *gitlab.Client
	filter    Filter
}

// Create creates a new repository in GitLab.
func (c Client) Create(ctx context.Context, cfg config.ProviderConfig, opt model.CreateOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Create:")
	logger.Debug().Str("CreateOption", opt.String()).Msg("GitLab:Create")

	namespaceID, err := c.getNamespaceID(ctx, cfg)
	if err != nil {
		return fmt.Errorf("get namespace ID: %w", err)
	}

	projectOpts := &gitlab.CreateProjectOptions{
		Name:          gitlab.Ptr(opt.RepositoryName),
		Description:   gitlab.Ptr(opt.Description),
		DefaultBranch: gitlab.Ptr(opt.DefaultBranch),
		Visibility:    gitlab.Ptr(toVisibility(opt.Visibility)),
	}

	if namespaceID != 0 {
		projectOpts.NamespaceID = gitlab.Ptr(namespaceID)
	}

	_, _, err = c.rawClient.Projects.CreateProject(projectOpts)
	if err != nil {
		return fmt.Errorf("create: failed to create %s: %w", opt.RepositoryName, err)
	}

	logger.Debug().Msg("Repository created successfully")

	return nil
}

func (c Client) DefaultBranch(ctx context.Context, owner, projectName, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:DefaultBranch:")
	logger.Debug().Str("name", branch).Msg("GitLab:DefaultBranch:")

	_, _, err := c.rawClient.Projects.EditProject(owner+"/"+projectName, &gitlab.EditProjectOptions{
		DefaultBranch: gitlab.Ptr(branch),
	})
	if err != nil {
		return fmt.Errorf("edit project default branch: %w", err)
	}

	return nil
}

// Name returns the name of the client.
func (c Client) Name() string {
	return config.GITLAB
}

// Metainfos retrieves metadata information for repositories.
func (c Client) Metainfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.RepositoryMetainfo, error) {
	metainfos, err := c.getRepositoryMetaInfos(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("get repository metainfos: %w", err)
	}

	if filtering {
		return c.filter.FilterMetainfo(ctx, cfg, metainfos, targetfilter.FilterIncludedExcludedGen(), targetfilter.IsInInterval)
	}

	return metainfos, nil
}

func (c Client) getRepositoryMetaInfos(ctx context.Context, cfg config.ProviderConfig) ([]model.RepositoryMetainfo, error) {
	var (
		repositories []*gitlab.Project
		err          error
	)

	if cfg.IsGroup() {
		repositories, _, err = c.rawClient.Groups.ListGroupProjects(cfg.Group, &gitlab.ListGroupProjectsOptions{
			OrderBy:     gitlab.Ptr("name"),
			Sort:        gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
		})
	} else {
		repositories, _, err = c.rawClient.Projects.ListUserProjects(cfg.User, &gitlab.ListProjectsOptions{
			Owned:       gitlab.Ptr(true),
			OrderBy:     gitlab.Ptr("name"),
			Sort:        gitlab.Ptr("asc"),
			ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("fetch repositories: %w", err)
	}

	metainfos := make([]model.RepositoryMetainfo, 0, len(repositories))

	for _, repo := range repositories {
		if !cfg.Git.IncludeForks && repo.ForkedFromProject != nil {
			continue
		}

		rm, err := newRepositoryMetainfo(ctx, cfg, c.rawClient, repo.Path)
		if err != nil {
			return nil, fmt.Errorf("init repository meta for %s: %w", repo.Path, err)
		}

		metainfos = append(metainfos, rm)
	}

	return metainfos, nil
}

// IsValidRepositoryName checks if the given name is a valid GitLab repository name.
func (c Client) IsValidRepositoryName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Validate:")
	logger.Debug().Str("name", name).Msg("GitLab:Validate")

	if !IsValidGitLabRepositoryName(name) || !isValidGitLabRepositoryNameCharacters(name) {
		logger.Debug().Str("name", name).Msg("Invalid GitLab repository name")
		logger.Debug().Msg("See https://docs.gitlab.com/ee/user/reserved_names.html")

		return false
	}

	return true
}

func newRepositoryMetainfo(ctx context.Context, cfg config.ProviderConfig, gitClient *gitlab.Client, name string) (model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newRepositoryMeta:")
	logger.Debug().Str("name", name).Msg("newRepositoryMeta:")

	projectPath := getProjectPath(cfg, name)

	gitlabProject, _, err := gitClient.Projects.GetProject(projectPath, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			logger.Warn().Str("name", name).Msg("Repository not found. Ignoring.")

			return model.RepositoryMetainfo{}, nil
		}

		return model.RepositoryMetainfo{}, fmt.Errorf("get gitlab project: %w", err)
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

func getVisibility(vis gitlab.VisibilityValue) string {
	switch vis {
	case gitlab.PublicVisibility:
		return "public"
	case gitlab.PrivateVisibility:
		return "private"
	case gitlab.InternalVisibility:
		return "internal"
	default:
		return "public"
	}
}

func toVisibility(vis string) gitlab.VisibilityValue {
	switch vis {
	case "private":
		return gitlab.PrivateVisibility
	case "internal":
		return gitlab.InternalVisibility
	default:
		return gitlab.PublicVisibility
	}
}

func (c Client) getNamespaceID(ctx context.Context, cfg config.ProviderConfig) (int, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering getNamespaceID")

	if !cfg.IsGroup() {
		return 0, nil
	}

	groups, resp, err := c.rawClient.Groups.ListGroups(&gitlab.ListGroupsOptions{
		Search: gitlab.Ptr(cfg.Group),
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return 0, errors.New("authentication failed: please check your token permissions")
		}

		return 0, fmt.Errorf("search for group: %w", err)
	}

	if len(groups) == 0 {
		return 0, fmt.Errorf("no group found with name: %s", cfg.Group)
	}

	return groups[0].ID, nil
}

func getProjectPath(cfg config.ProviderConfig, name string) string {
	if cfg.IsGroup() {
		return cfg.Group + "/" + name
	}

	return cfg.User + "/" + name
}

// NewGitLabClient creates a new GitLab client.
func NewGitLabClient(ctx context.Context, option model.GitProviderClientOption, httpClient *http.Client) (Client, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering NewGitLabClient")

	client, err := gitlab.NewClient(option.HTTPClient.Token,
		gitlab.WithBaseURL(option.DomainWithScheme(option.HTTPClient.Scheme)),
		gitlab.WithHTTPClient(httpClient),
	)
	if err != nil {
		return Client{}, fmt.Errorf("create new GitLab client: %w", err)
	}

	return Client{rawClient: client}, nil
}
