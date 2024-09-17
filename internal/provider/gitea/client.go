// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package gitea provides a client for interacting with Gitea repositories
// using the go-git-providers library. It offers a range of functionalities including:
//   - Creating and listing repositories
//   - Filtering repository metadata based on various criteria
//   - Validating repository names according to GitHub's rules
//   - Performing common operations on repositories
//
// This package aims to simplify Gitea interactions in Go applications, providing
// a interface for repository management and metadata handling.
package gitea

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"

	"code.gitea.io/sdk/gitea"
)

// Client represents a Gitea client that can perform various operations
// on Gitea repositories.
type Client struct {
	giteaClient *gitea.Client
	filter      Filter
}

// Create creates a new repository in Gitea.
// It supports creating repositories for both users and organizations.
//
// Parameters:
// - ctx: The context for the operation.
// - config: Configuration for the provider, including domain and user/group information.
// - option: Options for creating the repository, including name, visibility, and description.
//
// Returns an error if the creation fails.
func (c Client) Create(ctx context.Context, config configuration.ProviderConfig, option model.CreateOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:Create:")
	config.DebugLog(logger).Msg("Gitea:Create:")

	_, _, err := c.giteaClient.CreateRepo(gitea.CreateRepoOption{
		Name:          option.RepositoryName,
		Description:   option.Description,
		DefaultBranch: option.DefaultBranch,
	})

	if err != nil {
		return fmt.Errorf("failed to create repository %s: %w", option.RepositoryName, err)
	}

	logger.Trace().Msg("Repository created successfully")

	return nil
}

// Name returns the name of the provider, which is "GITEA".
func (c Client) Name() string {
	return configuration.GITEA
}

// Metainfos retrieves metadata information for repositories.
// It can list repositories for both users and organizations.
//
// Parameters:
// - ctx: The context for the operation.
// - config: Configuration for the provider, including domain and user/group information.
// - filtering: If true, applies additional filtering to the results.
//
// Returns a slice of RepositoryMetainfo and an error if the operation fails.
func (c Client) Metainfos(ctx context.Context, config configuration.ProviderConfig, filtering bool) ([]model.RepositoryMetainfo, error) {
	var (
		repositories []*gitea.Repository
		err          error
	)

	if config.IsGroup() {
		opt := gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     -1, // Set to -1 to get all items
				PageSize: -1,
			},
		}
		repositories, _, err = c.giteaClient.ListOrgRepos(config.Group, opt)
	} else {
		opt := gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     -1,
				PageSize: -1,
			},
		}

		repositories, _, err = c.giteaClient.ListUserRepos(config.User, opt)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	var metainfos []model.RepositoryMetainfo //nolint:prealloc

	for _, repo := range repositories {
		rm, _ := newRepositoryMeta(ctx, config, c.giteaClient, repo.Name)
		metainfos = append(metainfos, rm)
	}

	if filtering {
		return c.filter.FilterMetainfo(ctx, config, metainfos)
	}

	return metainfos, nil
}

// IsValidRepositoryName checks if the given repository name is valid for Gitea.
// It applies Gitea-specific naming rules.
//
// Parameters:
// - ctx: The context for the operation.
// - name: The repository name to validate.
//
// Returns true if the name is valid, false otherwise.
func (c Client) IsValidRepositoryName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:Validate:")
	logger.Debug().Str("name", name).Msg("Gitea:Validate:")

	return IsValidGiteaRepositoryName(name)
}

// NewGiteaClient creates a new Gitea client.
//
// Parameters:
// - ctx: The context for the operation.
// - option: Options for creating the client, including domain and authentication token.
//
// Returns a new Client and an error if the creation fails.
func NewGiteaClient(ctx context.Context, option model.GitProviderClientOption) (Client, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering NewGiteaClient:")

	var (
		client *gitea.Client
		err    error
	)

	if option.Token == "" {
		client, err = gitea.NewClient(option.DomainWithScheme(option.Scheme), gitea.SetToken(option.Token))
	} else {
		client, err = gitea.NewClient(option.DomainWithScheme(option.Scheme))
	}

	if err != nil {
		return Client{}, fmt.Errorf("failed to create a new Gitea client: %w", err)
	}

	return Client{giteaClient: client}, nil
}

// newRepositoryMeta creates a new RepositoryMetainfo struct for a given repository.
// It fetches detailed information about the repository from Gitea.
//
// Parameters:
// - ctx: The context for the operation.
// - config: Configuration for the provider.
// - gitClient: The Gitea client to use for fetching repository information.
// - repositoryName: The name of the repository to fetch information for.
//
// Returns a RepositoryMetainfo and an error if the operation fails.
func newRepositoryMeta(ctx context.Context, config configuration.ProviderConfig, rawClient *gitea.Client, repositoryName string) (model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newRepositoryMeta:")

	owner := config.Group
	if !config.IsGroup() {
		owner = config.User
	}

	giteaProject, _, err := rawClient.GetRepo(owner, repositoryName)
	if err != nil {
		return model.RepositoryMetainfo{}, fmt.Errorf("failed to get project info for %s: %w", repositoryName, err)
	}

	return model.RepositoryMetainfo{
		OriginalName:   repositoryName,
		HTTPSURL:       giteaProject.CloneURL,
		SSHURL:         giteaProject.SSHURL,
		Description:    giteaProject.Description,
		DefaultBranch:  giteaProject.DefaultBranch,
		LastActivityAt: &giteaProject.Updated,
		Visibility:     string(giteaProject.Owner.Visibility),
	}, nil
}

// TODO: Implement isValidGiteaRepositoryName and isValidGiteaRepositoryNameCharacters functions
// These functions should contain the logic for validating Gitea repository names
// according to Gitea's specific naming rules and allowed characters.
