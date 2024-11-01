// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

// GitProvider defines the interface for interacting with a Git hosting service.
// This interface encapsulates operations such as creating repositories,
// fetching repository metadata, and validating repository names.
type GitProvider interface {
	// Create initiates the creation of a new repository on the Git provider.
	//
	// Parameters:
	//   - ctx: A context.Context for handling cancellation, timeouts, and request-scoped values.
	//   - config: Configuration specific to the provider, containing authentication and other settings.
	//   - option: Options for creating the repository, such as name, visibility, and description.
	//
	// Returns:
	//   - error: An error if the creation fails, or nil if successful.
	//
	// This method should handle:
	//   - Authenticating with the Git provider.
	//   - Validating the creation options.
	//   - Creating the repository with the specified settings.
	//   - Handling any provider-specific requirements or limitations.
	Create(ctx context.Context, config config.ProviderConfig, option model.CreateOption) error

	DefaultBranch(ctx context.Context, owner string, projectname string, branch string) error

	// ProjectInfos retrieves metadata for repositories from the Git provider.
	//
	// Parameters:
	//   - ctx: A context.Context for handling cancellation, timeouts, and request-scoped values.
	//   - config: Configuration specific to the provider, containing authentication and other settings.
	//   - filtering: A boolean indicating whether to apply additional filtering to the results.
	//
	// Returns:
	//   - []model.RepositoryMetainfo: A slice of metadata for the repositories.
	//   - error: An error if the retrieval fails, or nil if successful.
	//
	// This method should handle:
	//   - Authenticating with the Git provider.
	//   - Fetching repository information, potentially paginating through results.
	//   - Applying any specified filtering.
	//   - Translating provider-specific data into the RepositoryMetainfo model.
	ProjectInfos(ctx context.Context, config config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error)

	// Name returns the name of the Git provider.
	//
	// Returns:
	//   - string: A string identifier for the Git provider (e.g., "github", "gitlab").
	//
	// This method should return a consistent, unique identifier for the provider.
	Name() string

	// IsValidRepositoryName checks if a given repository name is valid for this Git provider.
	//
	// Parameters:
	//   - ctx: A context.Context for handling cancellation, timeouts, and request-scoped values.
	//   - name: The repository name to validate.
	//
	// Returns:
	//   - bool: true if the name is valid, false otherwise.
	//
	// This method should check the name against the provider's specific naming rules and restrictions.
	IsValidRepositoryName(ctx context.Context, name string) bool
}

// Example usage:
//
//	type GitHubProvider struct {
//		client *github.Client
//	}
//
//	func (g *GitHubProvider) Create(ctx context.Context, config model.ProviderConfig, option model.CreateOption) error {
//		// Implementation for creating a repository on GitHub
//		_, _, err := g.client.Repositories.Create(ctx, "", &github.Repository{
//			Name:        github.String(option.RepositoryName),
//			Description: github.String(option.Description),
//			Private:     github.Bool(option.Visibility == "private"),
//		})
//		return err
//	}
//
//	func (g *GitHubProvider) Metainfos(ctx context.Context, config model.ProviderConfig, filtering bool) ([]model.RepositoryMetainfo, error) {
//		// Implementation for fetching repository metadata from GitHub
//		// ...
//	}
//
//	func (g *GitHubProvider) Name() string {
//		return "github"
//	}
//
//	func (g *GitHubProvider) IsValidRepositoryName(ctx context.Context, name string) bool {
//		// Implementation for checking if a repository name is valid on GitHub
//		// ...
//	}
//
//	func SyncRepositories(ctx context.Context, provider GitProvider, config model.ProviderConfig) error {
//		// Create a new repository
//		err := provider.Create(ctx, config, model.CreateOption{
//			RepositoryName: "new-repo",
//			Description:    "A new repository",
//			Visibility:     "public",
//		})
//		if err != nil {
//			return err
//		}
//
//		// Fetch and print metadata for all repositories
//		metainfos, err := provider.Metainfos(ctx, config, true)
//		if err != nil {
//			return err
//		}
//		for _, info := range metainfos {
//			fmt.Printf("Repository: %s, URL: %s\n", info.Name, info.URL)
//		}
//
//		return nil
//	}
