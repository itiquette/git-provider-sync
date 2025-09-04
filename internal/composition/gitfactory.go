// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

//nolint:funcorder // Factory with many create methods
package composition

import (
	"fmt"

	"itiquette/git-provider-sync/internal/adapters/repository/archive"
	"itiquette/git-provider-sync/internal/adapters/repository/directory"
	"itiquette/git-provider-sync/internal/adapters/repository/gitbinary"
	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// GitFactory creates git operations instances based on configuration.
type GitFactory struct{}

// NewGitFactory creates a new git operations factory.
func NewGitFactory(_ ports.GitConfig) *GitFactory {
	return &GitFactory{}
}

// CreateOperations creates git operations based on configuration.
func (f *GitFactory) CreateOperations(config ports.GitConfig) (ports.GitOperations, error) { //nolint:ireturn // Factory method returning interface
	implementation := config.PreferredImplementation
	if implementation == "" {
		implementation = ProviderTypeGoGit
	}

	switch implementation {
	case ProviderTypeGoGit:
		return gogit.New(config), nil
	case ProviderTypeGitBinary:
		return gitbinary.New(config), nil
	case ProviderTypeDirectory:
		return directory.New(config), nil
	case ProviderTypeArchive:
		return archive.New(config), nil
	default:
		return nil, fmt.Errorf("%w: %s", domain.ErrUnsupportedGitImplementation, implementation)
	}
}

// AvailableImplementations returns the list of available git implementations.
func AvailableImplementations() []string {
	return []string{ProviderTypeGoGit, ProviderTypeGitBinary, ProviderTypeDirectory, ProviderTypeArchive}
}

// IsImplementationAvailable checks if a specific implementation is available.
func IsImplementationAvailable(name string) bool {
	available := AvailableImplementations()
	for _, impl := range available {
		if impl == name {
			return true
		}
	}

	return false
}

// GetDefaultConfig returns a default git configuration.
func GetDefaultConfig() ports.GitConfig {
	return ports.GitConfig{
		PreferredImplementation: ProviderTypeGoGit,
		UserName:                "git-provider-sync",
		UserEmail:               "sync@git-provider-sync.local",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}
}
