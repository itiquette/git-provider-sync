// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/adapters/repository/archive"
	"itiquette/git-provider-sync/internal/adapters/repository/directory"
	"itiquette/git-provider-sync/internal/adapters/repository/gitbinary"
	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// GitFactory creates git operations instances based on configuration.
// It implements the factory pattern for git operations (go-git, directory).
type GitFactory struct {
	config ports.GitConfig
}

// NewGitFactory creates a new git operations factory.
func NewGitFactory(config ports.GitConfig) *GitFactory {
	return &GitFactory{
		config: config,
	}
}

// CreateOperations creates git operations based on the factory configuration.
func (f *GitFactory) CreateOperations(config ports.GitConfig) (ports.GitOperations, error) {
	// Merge factory config with provided config
	mergedConfig := f.mergeConfigs(config)

	// Determine which implementation to use
	implementation := f.selectImplementation(mergedConfig)

	switch implementation {
	case ProviderTypeGoGit:
		return f.createGoGitOperations(mergedConfig)
	case ProviderTypeGitBinary:
		return f.createGitBinaryOperations(mergedConfig)
	case ProviderTypeDirectory:
		return f.createDirectoryOperations(mergedConfig)
	case ProviderTypeArchive:
		return f.createArchiveOperations(mergedConfig)
	default:
		return nil, fmt.Errorf("unsupported git implementation: %s", implementation)
	}
}

// AvailableImplementations returns the list of available git implementations.
func (f *GitFactory) AvailableImplementations() []string {
	return []string{ProviderTypeGoGit, ProviderTypeGitBinary, ProviderTypeDirectory, ProviderTypeArchive}
}

// IsImplementationAvailable checks if a specific implementation is available.
func (f *GitFactory) IsImplementationAvailable(name string) bool {
	available := f.AvailableImplementations()
	for _, impl := range available {
		if impl == name {
			return true
		}
	}

	return false
}

// GetDefaultConfig returns a default git configuration.
func (f *GitFactory) GetDefaultConfig() ports.GitConfig {
	return ports.GitConfig{
		PreferredImplementation: ProviderTypeGoGit,
		UserName:                "git-provider-sync",
		UserEmail:               "sync@git-provider-sync.local",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}
}

// CreateOperationsForURL creates git operations suitable for a specific URL.
func (f *GitFactory) CreateOperationsForURL(url string, config ports.GitConfig) (ports.GitOperations, error) {
	// Merge configs
	mergedConfig := f.mergeConfigs(config)

	// Select implementation based on URL
	if f.isArchiveURL(url) {
		return f.createArchiveOperations(mergedConfig)
	} else if f.isFileURL(url) {
		return f.createDirectoryOperations(mergedConfig)
	}

	// Default to go-git for remote URLs
	return f.createGoGitOperations(mergedConfig)
}

// ValidateConfig validates a git configuration.
func (f *GitFactory) ValidateConfig(config ports.GitConfig) error {
	if config.UserName == "" {
		return errors.New("user name is required")
	}

	if config.UserEmail == "" {
		return errors.New("user email is required")
	}

	if config.MaxConcurrent <= 0 {
		return errors.New("max concurrent must be positive")
	}

	if config.PreferredImplementation != "" {
		if !f.IsImplementationAvailable(config.PreferredImplementation) {
			return fmt.Errorf("implementation not available: %s", config.PreferredImplementation)
		}
	}

	return nil
}

// Implementation creation methods

// createGoGitOperations creates a go-git implementation.
func (f *GitFactory) createGoGitOperations(config ports.GitConfig) (ports.GitOperations, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config for go-git: %w", err)
	}

	adapter := gogit.New(config)

	return adapter, nil
}

// createGitBinaryOperations creates a git binary implementation.
func (f *GitFactory) createGitBinaryOperations(config ports.GitConfig) (ports.GitOperations, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config for git binary: %w", err)
	}

	adapter := gitbinary.New(config)

	return adapter, nil
}

// createDirectoryOperations creates a directory-based implementation.
func (f *GitFactory) createDirectoryOperations(config ports.GitConfig) (ports.GitOperations, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config for directory: %w", err)
	}

	adapter := directory.New(config)

	return adapter, nil
}

// createArchiveOperations creates an archive-based implementation.
func (f *GitFactory) createArchiveOperations(config ports.GitConfig) (ports.GitOperations, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config for archive: %w", err)
	}

	adapter := archive.New(config)

	return adapter, nil
}

// Helper methods

// selectImplementation selects the best implementation based on configuration.
func (f *GitFactory) selectImplementation(config ports.GitConfig) string {
	// Use preferred implementation if specified and available
	if config.PreferredImplementation != "" {
		if f.IsImplementationAvailable(config.PreferredImplementation) {
			return config.PreferredImplementation
		}
	}

	// Default selection logic
	return ProviderTypeGoGit
}

// mergeConfigs merges the factory config with the provided config.
func (f *GitFactory) mergeConfigs(config ports.GitConfig) ports.GitConfig {
	result := f.config

	// Override with provided config values
	if config.PreferredImplementation != "" {
		result.PreferredImplementation = config.PreferredImplementation
	}

	if config.UserName != "" {
		result.UserName = config.UserName
	}

	if config.UserEmail != "" {
		result.UserEmail = config.UserEmail
	}

	if config.MaxConcurrent > 0 {
		result.MaxConcurrent = config.MaxConcurrent
	}

	if config.CacheSize > 0 {
		result.CacheSize = config.CacheSize
	}

	if config.Timeout > 0 {
		result.Timeout = config.Timeout
	}

	if len(config.TrustDomains) > 0 {
		result.TrustDomains = config.TrustDomains
	}

	if config.LogFile != "" {
		result.LogFile = config.LogFile
	}

	// Boolean fields - only override if explicitly set
	result.VerifySSL = config.VerifySSL || result.VerifySSL
	result.Debug = config.Debug || result.Debug

	return result
}

// isFileURL checks if a URL is a file-based URL.
func (f *GitFactory) isFileURL(url string) bool {
	return strings.HasPrefix(url, "file://") ||
		strings.HasPrefix(url, "/") ||
		strings.Contains(url, ":\\") // Windows paths
}

// isArchiveURL checks if a URL points to an archive file.
func (f *GitFactory) isArchiveURL(url string) bool {
	if !strings.HasPrefix(url, "file://") {
		return false
	}

	archivePath := strings.TrimPrefix(url, "file://")

	return strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz")
}

// GetImplementationInfo returns information about a specific implementation.
func (f *GitFactory) GetImplementationInfo(name string) (GitImplementationInfo, error) {
	switch name {
	case ProviderTypeGoGit:
		return GitImplementationInfo{
			Name:         "go-git",
			Description:  "Pure Go git implementation",
			Capabilities: []string{"clone", "fetch", "push", "pull", "branches", "tags", "commits"},
			Limitations:  []string{"no git hooks support", "limited git lfs support"},
			Recommended:  true,
		}, nil
	case ProviderTypeGitBinary:
		return GitImplementationInfo{
			Name:         "git-binary",
			Description:  "Native git binary implementation",
			Capabilities: []string{"clone", "fetch", "push", "pull", "branches", "tags", "commits", "hooks", "lfs"},
			Limitations:  []string{"requires git binary installed", "external dependency"},
			Recommended:  true,
		}, nil
	case ProviderTypeDirectory:
		return GitImplementationInfo{
			Name:         "directory",
			Description:  "Directory-based file operations",
			Capabilities: []string{"clone", "basic file operations"},
			Limitations:  []string{"no git operations", "no version control"},
			Recommended:  false,
		}, nil
	case ProviderTypeArchive:
		return GitImplementationInfo{
			Name:         "archive",
			Description:  "Tar.gz archive-based operations",
			Capabilities: []string{"clone", "push", "archive creation", "archive extraction"},
			Limitations:  []string{"no git operations", "no version control", "only tar.gz format"},
			Recommended:  false,
		}, nil
	default:
		return GitImplementationInfo{}, fmt.Errorf("unknown implementation: %s", name)
	}
}

// GitImplementationInfo contains information about a git implementation.
type GitImplementationInfo struct {
	Name         string
	Description  string
	Capabilities []string
	Limitations  []string
	Recommended  bool
}

// CreateOperationsWithAuth creates git operations with authentication configured.
func (f *GitFactory) CreateOperationsWithAuth(authConfig ports.AuthenticationConfig, gitConfig ports.GitConfig) (ports.GitOperations, error) {
	// Merge the authentication into git config
	// (This would typically involve configuring git credentials)
	return f.CreateOperations(gitConfig)
}

// CreateBatchOperations creates multiple git operations instances for parallel processing.
func (f *GitFactory) CreateBatchOperations(count int, config ports.GitConfig) ([]ports.GitOperations, error) {
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}

	operations := make([]ports.GitOperations, count)

	for index := 0; index < count; index++ {
		ops, err := f.CreateOperations(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create operations instance %d: %w", index, err)
		}

		operations[index] = ops
	}

	return operations, nil
}

// GetRecommendedImplementation returns the recommended implementation for a use case.
func (f *GitFactory) GetRecommendedImplementation(useCase GitUseCase) string {
	switch useCase {
	case GitUseCaseFullSync:
		return ProviderTypeGoGit
	case GitUseCaseFileBackup:
		return ProviderTypeDirectory
	case GitUseCaseQuickSync:
		return ProviderTypeGoGit
	default:
		return ProviderTypeGoGit
	}
}

// GitUseCase represents different use cases for git operations.
type GitUseCase string

const (
	// GitUseCaseFullSync represents full repository synchronization.
	GitUseCaseFullSync GitUseCase = "full_sync"
	// GitUseCaseFileBackup represents file-based backup operations.
	GitUseCaseFileBackup GitUseCase = "file_backup"
	// GitUseCaseQuickSync represents quick synchronization operations.
	GitUseCaseQuickSync GitUseCase = "quick_sync"
)

// CreateOperationsForUseCase creates git operations optimized for a specific use case.
func (f *GitFactory) CreateOperationsForUseCase(useCase GitUseCase, config ports.GitConfig) (ports.GitOperations, error) {
	// Configure git config based on use case
	optimizedConfig := f.optimizeConfigForUseCase(useCase, config)

	// Get recommended implementation
	implementation := f.GetRecommendedImplementation(useCase)
	optimizedConfig.PreferredImplementation = implementation

	return f.CreateOperations(optimizedConfig)
}

// optimizeConfigForUseCase optimizes git configuration for a specific use case.
func (f *GitFactory) optimizeConfigForUseCase(useCase GitUseCase, config ports.GitConfig) ports.GitConfig {
	optimized := config

	switch useCase {
	case GitUseCaseFullSync:
		// Full sync needs higher concurrency and larger cache
		if optimized.MaxConcurrent == 0 {
			optimized.MaxConcurrent = 10
		}

		if optimized.CacheSize == 0 {
			optimized.CacheSize = 100 * 1024 * 1024 // 100MB
		}
	case GitUseCaseFileBackup:
		// File backup can use lower resource settings
		if optimized.MaxConcurrent == 0 {
			optimized.MaxConcurrent = 3
		}

		if optimized.CacheSize == 0 {
			optimized.CacheSize = 10 * 1024 * 1024 // 10MB
		}
	case GitUseCaseQuickSync:
		// Quick sync prioritizes speed
		if optimized.MaxConcurrent == 0 {
			optimized.MaxConcurrent = 5
		}
	}

	return optimized
}
