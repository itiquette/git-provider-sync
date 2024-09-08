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
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/targetfilter"

	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	gogitlab "github.com/xanzy/go-gitlab"
)

// Client represents a GitLab client.
type Client struct {
	providerClient gitprovider.Client
	filter         Filter
}

// Create creates a new repository in GitLab.
func (glc Client) Create(ctx context.Context, config configuration.ProviderConfig, option model.CreateOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Create:")
	config.DebugLog(logger).Msg("GitLab:Create:")

	repoInfo := gitprovider.RepositoryInfo{
		Visibility:  gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibility(option.Visibility)),
		Description: &option.Description,
	}

	var err error

	if config.IsGroup() {
		_, err = glc.providerClient.OrgRepositories().Create(
			ctx,
			model.NewOrgRepositoryRef(config.Domain, config.Group, option.RepositoryName),
			repoInfo,
			&gitprovider.RepositoryCreateOptions{},
		)
	} else {
		_, err = glc.providerClient.UserRepositories().Create(
			ctx,
			model.NewUserRepositoryRef(config.Domain, config.User, option.RepositoryName),
			repoInfo,
			&gitprovider.RepositoryCreateOptions{},
		)
	}

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

	if config.IsGroup() {
		metainfos, err = glc.getGroupRepositories(ctx, config)
	} else {
		metainfos, err = glc.getUserRepositories(ctx, config)
	}

	if err != nil {
		return nil, err
	}

	if filtering {
		return glc.filter.FilterMetainfo(ctx, config, metainfos, targetfilter.FilterIncludedExcludedGen(), targetfilter.IsInInterval)
	}

	return metainfos, nil
}

func (glc Client) getGroupRepositories(ctx context.Context, config configuration.ProviderConfig) ([]model.RepositoryMetainfo, error) {
	orgRef := model.NewOrgRef(config.Domain, config.Group)

	repositories, err := glc.providerClient.OrgRepositories().List(ctx, orgRef)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group repository URLs: %w", err)
	}

	metainfos := make([]model.RepositoryMetainfo, 0, len(repositories))

	for _, repo := range repositories {
		name := repo.Repository().GetRepository()

		rm, err := newRepositoryMetainfo(ctx, config, glc, name)
		if err != nil {
			return nil, fmt.Errorf("failed to init repository meta: %w", err)
		}

		metainfos = append(metainfos, rm)
	}

	return metainfos, nil
}

func (glc Client) getUserRepositories(ctx context.Context, config configuration.ProviderConfig) ([]model.RepositoryMetainfo, error) {
	userRef := model.NewUserRef(config.Domain, config.User)

	repositories, err := glc.providerClient.UserRepositories().List(ctx, userRef)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user repository URLs: %w", err)
	}

	metainfos := make([]model.RepositoryMetainfo, 0, len(repositories))

	for _, repo := range repositories {
		// a project name and clone url might differ. A bug in flux cd always give the clone url with NAME instead of the real url
		// https://github.com/fluxcd/go-git-providers/issues/267 - which means it can handle these currently
		cloneURL := repo.Repository().GetCloneURL(gitprovider.TransportTypeHTTPS)
		name := strings.TrimSuffix(filepath.Base(cloneURL), ".git")

		rm, err := newRepositoryMetainfo(ctx, config, glc, name)
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
func newRepositoryMetainfo(ctx context.Context, config configuration.ProviderConfig, gitClient interfaces.GitClient, name string) (model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newRepositoryMeta:")
	logger.Debug().Str("name", name).Msg("newRepositoryMeta:")

	rawClient, ok := gitClient.Client().Raw().(*gogitlab.Client)
	if !ok {
		return model.RepositoryMetainfo{}, errors.New("failed getting the raw GitLab client")
	}

	projectPath := getProjectPath(config, name)

	gitlabProject, _, err := rawClient.Projects.GetProject(projectPath, nil)
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
func getVisibility(vis gogitlab.VisibilityValue) string {
	switch vis {
	case gogitlab.PublicVisibility:
		return "public"
	case gogitlab.PrivateVisibility:
		return "private"
	case gogitlab.InternalVisibility:
		return "internal"
	default:
		return "public" // TODO: Handle this case better
	}
}

//nolint:ireturn
func (glc Client) Client() gitprovider.Client {
	return glc.providerClient
}

// NewGitLabClient creates a new GitLab client.
func NewGitLabClient(ctx context.Context, option model.GitProviderClientOption) (Client, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering NewGitLabClient:")

	client, err := gitlab.NewClient(option.Token, "any", &gitprovider.ClientOptions{
		CommonClientOptions: gitprovider.CommonClientOptions{Domain: &option.Domain},
	})
	if err != nil {
		return Client{}, fmt.Errorf("failed to create a new gitlab client: %w", err)
	}

	return Client{providerClient: client}, nil
}

// Helper function

func getProjectPath(config configuration.ProviderConfig, name string) string {
	if config.IsGroup() {
		return config.Group + "/" + name
	}

	return config.User + "/" + name
}
