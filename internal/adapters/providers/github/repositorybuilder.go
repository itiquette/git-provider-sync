// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"fmt"

	"github.com/google/go-github/v71/github"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// setRepositoryStringFields sets all string-based repository fields that can return errors.
// This helper eliminates code duplication between adapter.go and projectservice.go.
func setRepositoryStringFields(repo *github.Repository, builder *entities.RepositoryBuilder) error {
	if repo.Name != nil {
		var err error

		*builder, err = builder.WithName(*repo.Name)
		if err != nil {
			return fmt.Errorf("failed to set repository name: %w", err)
		}
	}

	if repo.CloneURL != nil {
		var err error

		*builder, err = builder.WithHTTPSURL(*repo.CloneURL)
		if err != nil {
			return fmt.Errorf("failed to set HTTPS URL: %w", err)
		}
	}

	if repo.SSHURL != nil {
		var err error

		*builder, err = builder.WithSSHURL(*repo.SSHURL)
		if err != nil {
			return fmt.Errorf("failed to set SSH URL: %w", err)
		}
	}

	if repo.DefaultBranch != nil {
		var err error

		*builder, err = builder.WithDefaultBranch(*repo.DefaultBranch)
		if err != nil {
			return fmt.Errorf("failed to set default branch: %w", err)
		}
	}

	return nil
}
