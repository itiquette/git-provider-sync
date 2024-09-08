// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
)

// GitInterface defines a comprehensive interface for Git operations.
// It combines the capabilities of SourceReader and TargetWriter,
// along with additional methods for pulling and fetching.
type GitInterface interface {
	// SourceReader embeds all methods from the SourceReader interface.
	// This typically includes operations for reading from a Git source.
	SourceReader

	// TargetWriter embeds all methods from the TargetWriter interface.
	// This typically includes operations for writing to a Git target.
	TargetWriter

	// Pull performs a Git pull operation.
	//
	// Parameters:
	//   - ctx: A context.Context for handling cancellation, timeouts, and request-scoped values.
	//   - option: A model.PullOption containing configuration for the pull operation.
	//
	// Returns:
	//   - error: An error if the pull operation fails, or nil if successful.
	//
	// This method should handle:
	//   - Authenticating with the remote repository if necessary.
	//   - Merging or rebasing the pulled changes as specified in the options.
	//   - Handling merge conflicts if they occur.
	//   - Updating the local repository state after the pull.
	Pull(ctx context.Context, option model.PullOption) error

	// Fetch retrieves updates from a remote repository without merging them.
	//
	// Parameters:
	//   - ctx: A context.Context for handling cancellation, timeouts, and request-scoped values.
	//   - repository: A model.Repository representing the local repository to fetch updates for.
	//
	// Returns:
	//   - error: An error if the fetch operation fails, or nil if successful.
	//
	// This method should handle:
	//   - Authenticating with the remote repository if necessary.
	//   - Retrieving updates for all tracked branches.
	//   - Updating remote-tracking branches in the local repository.
	//   - Not modifying the working directory or current branch.
	Fetch(ctx context.Context, repository model.Repository) error
}

// Note: The SourceReader and TargetWriter interfaces should be defined
// elsewhere in the package, containing methods for reading from a source
// and writing to a target, respectively.

// Example usage:
//
//	type MyGitImplementation struct {
//		// fields for managing Git operations
//	}
//
//	// Implement SourceReader methods...
//	// Implement TargetWriter methods...
//
//	func (g *MyGitImplementation) Pull(ctx context.Context, option model.PullOption) error {
//		// Implementation of Pull
//		// Example:
//		// return g.repo.Pull(&git.PullOptions{
//		//     RemoteName: option.RemoteName,
//		//     ReferenceName: plumbing.ReferenceName(option.Branch),
//		// })
//	}
//
//	func (g *MyGitImplementation) Fetch(ctx context.Context, repository model.Repository) error {
//		// Implementation of Fetch
//		// Example:
//		// return g.repo.Fetch(&git.FetchOptions{
//		//     RemoteName: "origin",
//		//     RefSpecs:   []config.RefSpec{"refs/*:refs/*"},
//		// })
//	}
//
//	func UseGitInterface(gi GitInterface) error {
//		ctx := context.Background()
//
//		// Use SourceReader methods
//		repo, err := gi.Clone(ctx, model.CloneOption{URL: "https://github.com/example/repo.git"})
//		if err != nil {
//			return err
//		}
//
//		// Fetch updates
//		if err := gi.Fetch(ctx, repo); err != nil {
//			return err
//		}
//
//		// Pull changes
//		pullOpt := model.PullOption{RemoteName: "origin", Branch: "main"}
//		if err := gi.Pull(ctx, pullOpt); err != nil {
//			return err
//		}
//
//		// Use TargetWriter methods
//		pushOpt := model.PushOption{RemoteName: "origin", RefSpecs: []string{"refs/heads/main"}}
//		return gi.Push(ctx, pushOpt)
//	}
