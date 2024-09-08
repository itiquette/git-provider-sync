// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
)

// SourceReader defines the contract for components that can clone repositories
// from a source git provider. This interface is crucial for implementing
// different cloning strategies or supporting various git providers.
type SourceReader interface {
	// Clone performs the operation of cloning a repository from a source git provider.
	//
	// Parameters:
	//   - ctx: A context.Context for handling cancellation, timeouts, and passing request-scoped values.
	//     It's the caller's responsibility to cancel this context when appropriate.
	//   - option: A model.CloneOption containing configuration details for the clone operation.
	//     This includes information such as the source repository URL, target path, and any
	//     special flags (e.g., shallow clone, specific branch).
	//
	// Returns:
	//   - model.Repository: A representation of the cloned repository, which may include
	//     metadata and a reference to the local clone.
	//   - error: An error if the clone operation fails, or nil if successful.
	//     Implementations should provide detailed error messages to aid in troubleshooting.
	//
	// The Clone method should handle the following responsibilities:
	//   1. Authenticate with the source git provider if necessary.
	//   2. Prepare the local filesystem for the clone operation.
	//   3. Execute the clone operation according to the provided options.
	//   4. Handle network issues or access permission problems gracefully.
	//   5. Set up the local repository with the correct configuration (e.g., remotes, branches).
	//   6. Provide appropriate logging or telemetry for monitoring the clone process.
	//
	// Implementations of this method should be concurrency-safe and respect the
	// provided context for cancellation and timeouts.
	Clone(ctx context.Context, option model.CloneOption) (model.Repository, error)
}

// Example usage:
//
//	type GitHubReader struct {
//		client *github.Client
//	}
//
//	func (gr *GitHubReader) Clone(ctx context.Context, option model.CloneOption) (model.Repository, error) {
//		// Implementation for cloning from GitHub
//		repo, _, err := gr.client.Repositories.Get(ctx, option.Owner, option.Repo)
//		if err != nil {
//			return model.Repository{}, fmt.Errorf("failed to get repository info: %w", err)
//		}
//
//		// Perform the actual clone operation
//		cmd := exec.CommandContext(ctx, "git", "clone", repo.GetCloneURL(), option.TargetPath)
//		if err := cmd.Run(); err != nil {
//			return model.Repository{}, fmt.Errorf("failed to clone repository: %w", err)
//		}
//
//		return model.Repository{
//			Name: repo.GetName(),
//			Path: option.TargetPath,
//			// Set other relevant fields
//		}, nil
//	}
//
//	// Using the SourceReader
//	func CloneFromGitHub(ctx context.Context, reader SourceReader, repoURL, targetPath string) (model.Repository, error) {
//		option := model.CloneOption{
//			URL:        repoURL,
//			TargetPath: targetPath,
//		}
//		return reader.Clone(ctx, option)
//	}
