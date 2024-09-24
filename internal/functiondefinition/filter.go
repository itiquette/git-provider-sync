// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package functiondefinition provides type definitions for function signatures
// used throughout git-provider-sync. These definitions allow for
// flexible and modular implementations of key operations.
package functiondefinition

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

// FilterIncludedExcludedFunc defines a function type for filtering repository metadata
// based on inclusion and exclusion criteria specified in the provider configuration.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation, timeouts, and request-scoped values.
//   - config: A model.ProviderConfig containing provider-specific settings,
//     including inclusion and exclusion patterns for repositories.
//   - metainfos: A slice of model.RepositoryMetainfo representing the repositories to be filtered.
//
// Returns:
//   - []model.RepositoryMetainfo: A slice of filtered repository metadata that matches the inclusion
//     criteria and does not match the exclusion criteria.
//   - error: An error if the filtering process encounters any issues, or nil if successful.
//
// This function type is designed to be implemented for flexible filtering strategies.
// Implementations should consider:
//   - Applying inclusion patterns specified in the config to select repositories.
//   - Applying exclusion patterns to remove repositories from the selection.
//   - Handling edge cases such as conflicting inclusion and exclusion patterns.
//   - Efficient processing of potentially large sets of repository metadata.
type FilterIncludedExcludedFunc func(ctx context.Context, config config.ProviderConfig, metainfos []model.RepositoryMetainfo) ([]model.RepositoryMetainfo, error)

// Example usage:
//
//	func SimpleFilterImplementation(ctx context.Context, config model.ProviderConfig, metainfos []model.RepositoryMetainfo) ([]model.RepositoryMetainfo, error) {
//		var filtered []model.RepositoryMetainfo
//		for _, repo := range metainfos {
//			// Example: Include repositories with names starting with "project-"
//			if strings.HasPrefix(repo.Name, "project-") {
//				// Example: Exclude repositories with names ending in "-archive"
//				if !strings.HasSuffix(repo.Name, "-archive") {
//					filtered = append(filtered, repo)
//				}
//			}
//		}
//		return filtered, nil
//	}
//
//	func UseFilter(filter FilterIncludedExcludedFunc) {
//		ctx := context.Background()
//		config := model.ProviderConfig{
//			// Configure provider settings
//		}
//		allRepos := []model.RepositoryMetainfo{
//			{Name: "project-a"},
//			{Name: "project-b-archive"},
//			{Name: "other-repo"},
//		}
//
//		filteredRepos, err := filter(ctx, config, allRepos)
//		if err != nil {
//			log.Fatalf("Filtering failed: %v", err)
//		}
//
//		for _, repo := range filteredRepos {
//			fmt.Printf("Filtered repo: %s\n", repo.Name)
//		}
//	}
