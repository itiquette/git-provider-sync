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
	"fmt"
	"net/http"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/google/go-github/v65/github"
)

// Client represents a GitHub client with associated operations.
// It wraps the go-git-providers GitHub client and provides additional
// functionality specific to this application.
type Client struct {
	rawClient *github.Client
	filter    Filter
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
func (ghc Client) Create(ctx context.Context, config config.ProviderConfig, option model.CreateOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Create:")
	config.DebugLog(logger).Msg("GitHub:Create:")

	var (
		err error
	)

	isPrivate := false
	if strings.EqualFold(option.Visibility, "private") {
		isPrivate = true
	}

	groupName := ""
	if config.IsGroup() {
		groupName = config.Group
	}

	rep := &github.Repository{Name: &option.RepositoryName, Private: &isPrivate, DefaultBranch: &option.DefaultBranch, Description: &option.Description}
	_, _, err = ghc.rawClient.Repositories.Create(ctx, groupName, rep)

	if err != nil {
		return fmt.Errorf("create: failed to create %s: %w", option.RepositoryName, err)
	}

	logger.Trace().Msg("User repository created successfully")

	return nil
}

// Name returns the name of the client, which is always "github".
// This method is used to identify the type of git provider being used.
func (ghc Client) Name() string {
	return config.GITHUB
}

// Metainfos retrieves metadata for repositories.
// It can list repositories for both users and organizations, and optionally apply filtering.
func (ghc Client) Metainfos(ctx context.Context, config config.ProviderConfig, filtering bool) ([]model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:Metainfos:")

	var metainfos []model.RepositoryMetainfo

	var repos []*github.Repository

	var err error

	listType := "all"
	if config.Git.IncludeForks {
		listType = "public,private,forks"
	}

	if config.IsGroup() {
		opt := &github.RepositoryListByOrgOptions{Type: listType, Sort: "full_name", ListOptions: github.ListOptions{PerPage: 300}}

		repos, _, err = ghc.rawClient.Repositories.ListByOrg(ctx, config.Group, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch org repositories: %w", err)
		}
	} else {
		opt := &github.RepositoryListByAuthenticatedUserOptions{
			Visibility:  "all",
			Affiliation: "owner",
			Sort:        "full_name",
			ListOptions: github.ListOptions{PerPage: 300}}

		repos, _, err = ghc.rawClient.Repositories.ListByAuthenticatedUser(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user repositories: %w", err)
		}
	}

	metainfos = ghc.processRepositories(ctx, config, repos)

	if filtering {
		return ghc.filter.FilterMetainfo(ctx, config, metainfos)
	}

	return metainfos, nil
}

// processRepositories is a helper function to process a list of repositories (either Org or User)
// and convert them into RepositoryMetainfo slices.
func (ghc *Client) processRepositories(ctx context.Context, config config.ProviderConfig, repos []*github.Repository) []model.RepositoryMetainfo {
	var metainfos []model.RepositoryMetainfo //nolint:prealloc

	logger := log.Logger(ctx)

	for _, repo := range repos {
		if !config.Git.IncludeForks && repo.Fork != nil && *repo.Fork {
			continue
		}

		name := repo.GetName()
		metainfo, err := newRepositoryMeta(ctx, config, ghc.rawClient, name)

		if err != nil {
			logger.Warn().Err(err).Str("repo", name).Msg("Failed to create organization repository metadata")

			continue
		}

		metainfos = append(metainfos, metainfo)
	}

	return metainfos
}

// Validate checks if the given repository name is valid.
// This is a convenience method that calls IsValidRepositoryName.
//
// Parameters:
//   - ctx: The context for the operation.
//   - name: The repository name to validate.
//
// Returns true if the repository name is valid, false otherwise.
func (ghc *Client) Validate(ctx context.Context, name string) bool {
	return ghc.IsValidRepositoryName(ctx, name)
}

func (ghc Client) DefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:DefaultBranch:")
	logger.Debug().Str("branch", branch).Msg("GitHub:DefaultBranch:")

	_, _, err := ghc.rawClient.Repositories.Edit(ctx, owner, projectName, &github.Repository{
		DefaultBranch: github.String(branch),
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
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

// NewGitHubClient creates a new GitHub client.
// It sets up the client with the provided options, including authentication if a token is provided.
//
// Parameters:
//   - ctx: The context for the operation.
//   - option: Options for creating the client, including the domain and authentication token.
//
// Returns a new Client and an error if the client creation fails.
func NewGitHubClient(ctx context.Context, option model.GitProviderClientOption, httpClient *http.Client) (Client, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering NewGitHubClient")

	client := github.NewClient(httpClient)

	if option.HTTPClient.Token != "" {
		client = client.WithAuthToken(option.HTTPClient.Token)
	}

	// TODO: Implement custom domain support for GitHub Enterprise

	return Client{rawClient: client}, nil
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
func newRepositoryMeta(ctx context.Context, config config.ProviderConfig, gitClient *github.Client, name string) (model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering newRepositoryMeta:")
	logger.Debug().Str("usr", config.User).Str("name", name).Str("provider", config.ProviderType).Str("domain", config.Domain).Msg("newRepositoryMeta:")

	owner := config.Group
	if !config.IsGroup() {
		owner = config.User
	}

	gitHubProject, _, err := gitClient.Repositories.Get(ctx, owner, name)
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

	return "N/A"
}

// getTimeOrNil is a helper function that converts a GitHub Timestamp to a standard time.Time pointer,
// or returns nil if the input is nil.
func getTimeOrNil(t *github.Timestamp) *time.Time {
	if t != nil {
		return &t.Time
	}

	return nil
}
