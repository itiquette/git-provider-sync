// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package github provides a client for interacting with GitHub repositories
// using the go-git-providers library. It offers a range of functionalities including:
//   - Creating and listing repositories
//   - Filtering repository metadata based on various criteria
//   - Validating repository names according to GitHub's rules
//   - Performing common operations on repositories
//
// This package aims to simplify GitHub interactions in Go applications, providing
// a interface for repository management and metadata handling.
package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	gogithub "github.com/google/go-github/v64/github"
)

// Client represents a GitHub client with associated operations.
// It wraps the go-git-providers GitHub client and provides additional
// functionality specific to this application.
type Client struct {
	gitProviderClient gitprovider.Client
	filter            Filter
}

// Create creates a new repository on GitHub.
// It supports creating both user and organization repositories.
//
// Parameters:
//   - ctx: The context for the operation, which may include deadlines or cancellation signals.
//   - config: The configuration for the provider, which contains information about the user or organization.
//   - option: Options for creating the repository, including name, visibility, and description.
//
// Returns an error if the repository creation fails.
func (ghc Client) Create(ctx context.Context, config configuration.ProviderConfig, option model.CreateOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Create:")
	config.DebugLog(logger).Msg("GitHub:Create:")

	repoInfo := gitprovider.RepositoryInfo{
		Visibility:  gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibility(option.Visibility)),
		Description: &option.Description,
	}

	var (
		resp any
		err  error
	)

	if config.IsGroup() {
		resp, err = ghc.gitProviderClient.OrgRepositories().Create(
			ctx,
			model.NewOrgRepositoryRef(config.Domain, config.Group, option.RepositoryName),
			repoInfo,
			&gitprovider.RepositoryCreateOptions{},
		)
	} else {
		resp, err = ghc.gitProviderClient.UserRepositories().Create(
			ctx,
			model.NewUserRepositoryRef(config.Domain, config.User, option.RepositoryName),
			repoInfo,
			&gitprovider.RepositoryCreateOptions{},
		)
	}

	if err != nil {
		return fmt.Errorf("create: failed to create %s: %w", option.RepositoryName, err)
	}

	switch resp.(type) {
	case gitprovider.OrgRepository:
		logger.Trace().Msg("Organization repository created successfully")
	case gitprovider.UserRepository:
		logger.Trace().Msg("User repository created successfully")
	default:
		logger.Trace().Msg("Unknown repository type created")
	}

	return nil
}

// Name returns the name of the client, which is always "github".
// This method is used to identify the type of git provider being used.
func (ghc Client) Name() string {
	return configuration.GITHUB
}

// Metainfos retrieves metadata for repositories.
// It can list repositories for both users and organizations, and optionally apply filtering.
func (ghc Client) Metainfos(ctx context.Context, config configuration.ProviderConfig, filtering bool) ([]model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Metainfos:")

	var metainfos []model.RepositoryMetainfo

	var err error

	if config.IsGroup() {
		orgRef := model.NewOrgRef(config.Domain, config.Group)

		orgRepos, err := ghc.gitProviderClient.OrgRepositories().List(ctx, orgRef)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch organization repositories: %w", err)
		}

		metainfos, err = ghc.processRepositories(ctx, config, orgRepos)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch organization repositories: %w", err)
		}
	} else {
		userRef := model.NewUserRef(config.Domain, config.User)

		userRepos, err := ghc.gitProviderClient.UserRepositories().List(ctx, userRef)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user repositories: %w", err)
		}

		metainfos, err = ghc.processRepositories(ctx, config, userRepos)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user repositories: %w", err)
		}
	}

	if err != nil {
		return nil, err
	}

	if filtering {
		return ghc.filter.FilterMetainfo(ctx, config, metainfos)
	}

	return metainfos, nil
}

// processRepositories is a helper function to process a list of repositories (either Org or User)
// and convert them into RepositoryMetainfo slices.
func (ghc Client) processRepositories(ctx context.Context, config configuration.ProviderConfig, repos interface{}) ([]model.RepositoryMetainfo, error) {
	var metainfos []model.RepositoryMetainfo

	logger := log.Logger(ctx)

	switch rep := repos.(type) {
	case []gitprovider.OrgRepository:
		for _, repo := range rep {
			name := repo.Repository().GetRepository()
			metainfo, err := newRepositoryMeta(ctx, config, ghc, name)

			if err != nil {
				logger.Warn().Err(err).Str("repo", name).Msg("Failed to create organization repository metadata")

				continue
			}

			metainfos = append(metainfos, metainfo)
		}
	case []gitprovider.UserRepository:
		for _, repo := range rep {
			name := repo.Repository().GetRepository()
			metainfo, err := newRepositoryMeta(ctx, config, ghc, name)

			if err != nil {
				logger.Warn().Err(err).Str("repo", name).Msg("Failed to create user repository metadata")

				continue
			}

			metainfos = append(metainfos, metainfo)
		}
	default:
		return nil, fmt.Errorf("unknown repository type: %T", repos)
	}

	return metainfos, nil
}

// Validate checks if the given repository name is valid.
// This is a convenience method that calls IsValidRepositoryName.
//
// Parameters:
//   - ctx: The context for the operation.
//   - name: The repository name to validate.
//
// Returns true if the repository name is valid, false otherwise.
func (ghc Client) Validate(ctx context.Context, name string) bool {
	return ghc.IsValidRepositoryName(ctx, name)
}

// IsValidRepositoryName checks if the given repository name is valid for GitHub.
// It applies GitHub-specific rules for repository names.
//
// Parameters:
//   - ctx: The context for the operation.
//   - name: The repository name to validate.
//
// Returns true if the repository name is valid, false otherwise.
func (ghc Client) IsValidRepositoryName(ctx context.Context, name string) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Validate:")
	logger.Debug().Str("name", name).Msg("GitHub:Validate:")

	if !IsValidGitHubRepositoryName(name) {
		logger.Debug().Str("name", name).Msg("Invalid GitHub repository name")
		logger.Debug().Msg("See https://github.com/dead-claudia/github-limits?tab=readme-ov-file#repository-names")

		return false
	}

	return true
}

// Client returns the underlying gitprovider.Client.
// This method allows access to the raw GitHub client for operations
// not covered by this wrapper.
func (ghc Client) Client() gitprovider.Client {
	return ghc.gitProviderClient
}

// NewGitHubClient creates a new GitHub client.
// It sets up the client with the provided options, including authentication if a token is provided.
//
// Parameters:
//   - ctx: The context for the operation.
//   - option: Options for creating the client, including the domain and authentication token.
//
// Returns a new Client and an error if the client creation fails.
func NewGitHubClient(ctx context.Context, option model.GitProviderClientOption) (Client, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering NewGitHubClient:")

	clientOpts := &gitprovider.ClientOptions{
		CommonClientOptions: gitprovider.CommonClientOptions{Domain: &option.Domain},
	}

	var client gitprovider.Client

	var err error

	if option.Token != "" {
		client, err = github.NewClient(gitprovider.WithOAuth2Token(option.Token), clientOpts)
	} else {
		client, err = github.NewClient(clientOpts)
	}

	if err != nil {
		return Client{}, fmt.Errorf("failed to create a new GitHub client: %w", err)
	}

	return Client{gitProviderClient: client}, nil
}

// newRepositoryMeta creates a new RepositoryMetainfo struct from a GitHub repository.
// This is an internal function used to convert GitHub-specific repository data
// into the application's generic RepositoryMetainfo format.
//
// Parameters:
//   - ctx: The context for the operation.
//   - config: The configuration for the provider.
//   - gitClient: The GitHub client interface.
//   - name: The name of the repository.
//
// Returns a RepositoryMetainfo and an error if the operation fails.
func newRepositoryMeta(ctx context.Context, config configuration.ProviderConfig, gitClient interfaces.GitClient, name string) (model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newRepositoryMeta:")
	logger.Debug().Str("usr", config.User).Str("name", name).Str("provider", config.Provider).Str("domain", config.Domain).Msg("newRepositoryMeta:")

	rawClient, ok := gitClient.Client().Raw().(*gogithub.Client)
	if !ok {
		return model.RepositoryMetainfo{}, errors.New("failed getting raw client include by activity time")
	}

	owner := config.Group
	if !config.IsGroup() {
		owner = config.User
	}

	gitHubProject, _, err := rawClient.Repositories.Get(ctx, owner, name)
	if err != nil {
		return model.RepositoryMetainfo{}, fmt.Errorf("failed to get projectinfo for %s: %w", name, err)
	}

	return model.RepositoryMetainfo{
		OriginalName:   name,
		Description:    getValueOrEmpty(gitHubProject.Description),
		HTTPSURL:       getValueOrEmpty(gitHubProject.CloneURL),
		SSHURL:         getValueOrEmpty(gitHubProject.SSHURL),
		DefaultBranch:  getValueOrEmpty(gitHubProject.DefaultBranch),
		LastActivityAt: getTimeOrNil(gitHubProject.UpdatedAt),
		Visibility:     getValueOrEmpty(gitHubProject.Visibility),
	}, nil
}

// getValueOrEmpty is a helper function that returns the value of a string pointer if it's not nil,
// or an empty string otherwise.
func getValueOrEmpty(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

// getTimeOrNil is a helper function that converts a GitHub Timestamp to a standard time.Time pointer,
// or returns nil if the input is nil.
func getTimeOrNil(t *gogithub.Timestamp) *time.Time {
	if t != nil {
		return &t.Time
	}

	return nil
}
