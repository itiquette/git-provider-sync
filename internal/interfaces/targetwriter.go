// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

// TargetWriter defines the interface for pushing data to a target git provider.
// Implementations of this interface are responsible for handling the specifics
// of pushing data to different git hosting services or local repositories.
type TargetWriter interface {
	// Push performs the operation of pushing data to a target.
	//
	// Parameters:
	//   - ctx: A context.Context for handling cancellation and timeouts.
	//   - option: A model.PushOption containing configuration for the push operation.
	//   - protocol: A model.GitProtocol containing Git Protocol configuration.
	//
	// Returns:
	//   - error: An error if the push operation fails, or nil if successful.
	//
	// The exact behavior of Push may vary depending on the specific implementation,
	// but generally it should:
	//   1. Authenticate with the target git provider if necessary.
	//   2. Prepare the data to be pushed based on the PushOption.
	//   3. Perform the actual push operation.
	//   4. Handle any errors or conflicts that may arise during the push.
	Push(ctx context.Context, repository GitRepository, option model.PushOption, sourceGitOption config.ProviderConfig, targetGitOption config.GitOption) error
}

// Example usage:
//
//	type GitHubWriter struct {
//		// fields for GitHub-specific configuration
//	}
//
//	func (gw *GitHubWriter) Push(ctx context.Context, option model.PushOption) error {
//		// Implementation for pushing to GitHub
//		// ...
//	}
//
//	func SomeFunction(ctx context.Context, writer TargetWriter) error {
//		option := model.PushOption{
//			// Configure push options
//		}
//		return writer.Push(ctx, option)
//	}
