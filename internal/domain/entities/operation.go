// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain"
)

// Domain errors for sync operations.
var (
	ErrNoRepositoriesFound  = errors.New("no repositories found to sync")
	ErrNoMirrorsConfigured  = errors.New("no mirrors configured for sync")
	ErrInvalidSyncConfig    = errors.New("invalid sync configuration")
	ErrSyncOperationFailed  = errors.New("sync operation failed")
	ErrRepositoryValidation = errors.New("repository validation failed")
)

// SyncOperation represents a complete synchronization operation between source and targets.
// This is an immutable value object that encapsulates sync logic and validation.
type SyncOperation struct {
	sourceConfig  SourceConfig
	mirrorConfigs []MirrorConfig
	repositories  []Repository
	options       SyncOptions
	metadata      SyncMetadata
}

// SourceConfig represents the source provider configuration.
type SourceConfig struct {
	providerType    string
	domain          string
	owner           string
	includeForks    bool
	includePatterns []string
	excludePatterns []string
}

// MirrorConfig represents a mirror target configuration.
type MirrorConfig struct {
	name         string
	providerType string
	domain       string
	owner        string
	path         string
	useGitBinary bool
}

// SyncOptions contains options that control sync behavior.
type SyncOptions struct {
	alphaNumericName bool
	includeForks     bool
	includeArchived  bool
}

// SyncMetadata tracks sync operation metadata.
type SyncMetadata struct {
	startTime   time.Time
	totalRepos  int
	environment string
}

// SyncResult represents the result of a sync operation.
type SyncResult struct {
	// Currently unused - placeholder for future sync result tracking
}

// SyncOperationBuilder provides a functional approach to building sync operations.
type SyncOperationBuilder struct {
	operation SyncOperation
}

// NewSyncOperationBuilder creates a new sync operation builder.
func NewSyncOperationBuilder() SyncOperationBuilder {
	return SyncOperationBuilder{
		operation: SyncOperation{
			options: SyncOptions{
				includeForks:    false, // Secure default
				includeArchived: false, // Secure default
			},
			metadata: SyncMetadata{
				startTime: time.Now(),
			},
		},
	}
}

// WithSourceConfig sets the source configuration.
func (b SyncOperationBuilder) WithSourceConfig(config SourceConfig) (SyncOperationBuilder, error) {
	if err := config.Validate(); err != nil {
		return b, fmt.Errorf("invalid source config: %w", err)
	}

	b.operation.sourceConfig = config

	return b, nil
}

// WithMirrorConfigs sets the mirror configurations.
func (b SyncOperationBuilder) WithMirrorConfigs(configs []MirrorConfig) (SyncOperationBuilder, error) {
	if len(configs) == 0 {
		return b, ErrNoMirrorsConfigured
	}

	for i, config := range configs {
		if err := config.Validate(); err != nil {
			return b, fmt.Errorf("invalid mirror config %d: %w", i, err)
		}
	}

	b.operation.mirrorConfigs = configs

	return b, nil
}

// WithRepositories sets the repositories to sync.
func (b SyncOperationBuilder) WithRepositories(repos []Repository) (SyncOperationBuilder, error) {
	if len(repos) == 0 {
		return b, ErrNoRepositoriesFound
	}

	b.operation.repositories = repos
	b.operation.metadata.totalRepos = len(repos)

	return b, nil
}

// WithOptions sets the sync options.
func (b SyncOperationBuilder) WithOptions(options SyncOptions) SyncOperationBuilder {
	b.operation.options = options

	return b
}

// WithEnvironment sets the environment name for metadata.
func (b SyncOperationBuilder) WithEnvironment(env string) SyncOperationBuilder {
	b.operation.metadata.environment = strings.TrimSpace(env)

	return b
}

// Build creates the final sync operation after validation.
func (b SyncOperationBuilder) Build() (SyncOperation, error) {
	if err := b.operation.Validate(); err != nil {
		return SyncOperation{}, err
	}

	return b.operation, nil
}

// SyncOperation accessor methods

// SourceConfig returns the source configuration.
func (s SyncOperation) SourceConfig() SourceConfig {
	return s.sourceConfig
}

// MirrorConfigs returns the mirror configurations.
func (s SyncOperation) MirrorConfigs() []MirrorConfig {
	return append([]MirrorConfig(nil), s.mirrorConfigs...) // Return copy
}

// Repositories returns the repositories to sync.
func (s SyncOperation) Repositories() []Repository {
	return append([]Repository(nil), s.repositories...) // Return copy
}

// Options returns the sync options.
func (s SyncOperation) Options() SyncOptions {
	return s.options
}

// Metadata returns the sync metadata.
func (s SyncOperation) Metadata() SyncMetadata {
	return s.metadata
}

// SyncOperation behavior methods

// Validate validates the entire sync operation.
func (s SyncOperation) Validate() error {
	if err := s.sourceConfig.Validate(); err != nil {
		return fmt.Errorf("source config invalid: %w", err)
	}

	if len(s.mirrorConfigs) == 0 {
		return ErrNoMirrorsConfigured
	}

	for i, config := range s.mirrorConfigs {
		if err := config.Validate(); err != nil {
			return fmt.Errorf("mirror config %d invalid: %w", i, err)
		}
	}

	if len(s.repositories) == 0 {
		return ErrNoRepositoriesFound
	}

	return nil
}

// FilterRepositories filters repositories based on sync options and patterns.
func (s SyncOperation) FilterRepositories() []Repository {
	var filtered []Repository

	for _, repo := range s.repositories {
		if s.shouldIncludeRepository(repo) {
			filtered = append(filtered, repo)
		}
	}

	return filtered
}

// ValidateRepositoriesForMirrors validates all repositories against all mirror targets.
func (s SyncOperation) ValidateRepositoriesForMirrors() []RepositoryValidationError {
	var errors []RepositoryValidationError

	for _, repo := range s.repositories {
		for _, mirror := range s.mirrorConfigs {
			if err := repo.ValidateForProvider(mirror.providerType); err != nil {
				errors = append(errors, RepositoryValidationError{
					Repository: repo.Name(),
					Mirror:     mirror.name,
					Err:        err,
				})
			}
		}
	}

	return errors
}

// GetRepositoryNameForMirror gets the appropriate repository name for a specific mirror.
func (s SyncOperation) GetRepositoryNameForMirror(repo Repository, _ MirrorConfig) string {
	if s.options.alphaNumericName {
		return repo.CleanName()
	}

	return repo.Name()
}

// EstimateDuration estimates the sync duration based on repository count and mirrors.
func (s SyncOperation) EstimateDuration() time.Duration {
	// Simple estimation: 30 seconds per repository per mirror
	baseTimePerRepo := 30 * time.Second
	totalOperations := len(s.repositories) * len(s.mirrorConfigs)

	return time.Duration(totalOperations) * baseTimePerRepo
}

// CreateSyncPlan creates a detailed execution plan for the sync operation.
func (s SyncOperation) CreateSyncPlan() SyncPlan {
	var steps []SyncStep

	for _, mirror := range s.mirrorConfigs {
		for _, repo := range s.FilterRepositories() {
			step := SyncStep{
				Repository:     repo,
				Mirror:         mirror,
				RepositoryName: s.GetRepositoryNameForMirror(repo, mirror),
				EstimatedTime:  30 * time.Second,
			}
			steps = append(steps, step)
		}
	}

	return SyncPlan{
		Operation:         s,
		Steps:             steps,
		EstimatedDuration: s.EstimateDuration(),
	}
}

// shouldIncludeRepository determines if a repository should be included in sync.
//
//nolint:cyclop // Complex filtering logic with multiple criteria
func (s SyncOperation) shouldIncludeRepository(repo Repository) bool {
	// Check fork inclusion
	if repo.IsFork() && !s.options.includeForks {
		return false
	}

	// Check archived inclusion
	if repo.IsArchived() && !s.options.includeArchived {
		return false
	}

	// Check include patterns
	if len(s.sourceConfig.includePatterns) > 0 {
		included := false

		for _, pattern := range s.sourceConfig.includePatterns {
			if matchesPattern(repo.Name(), pattern) {
				included = true

				break
			}
		}

		if !included {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range s.sourceConfig.excludePatterns {
		if matchesPattern(repo.Name(), pattern) {
			return false
		}
	}

	return true
}

// SourceConfig methods

// NewSourceConfig creates a new source configuration.
func NewSourceConfig(providerType, domain, owner string) SourceConfig {
	return SourceConfig{
		providerType: providerType,
		domain:       domain,
		owner:        owner,
	}
}

// Validate validates the source configuration.
func (sc SourceConfig) Validate() error {
	if strings.TrimSpace(sc.providerType) == "" {
		return domain.ErrProviderTypeRequired
	}

	if strings.TrimSpace(sc.owner) == "" {
		return domain.ErrOwnerRequired
	}

	validProviders := map[string]bool{
		"github": true, "gitlab": true, "gitea": true,
	}

	if !validProviders[strings.ToLower(sc.providerType)] {
		return fmt.Errorf("%w: %s", domain.ErrUnsupportedProviderType, sc.providerType)
	}

	return nil
}

// MirrorConfig methods

// NewMirrorConfig creates a new mirror configuration.
func NewMirrorConfig(name, providerType, domain, owner string) MirrorConfig {
	return MirrorConfig{
		name:         strings.TrimSpace(name),
		providerType: strings.ToLower(strings.TrimSpace(providerType)),
		domain:       strings.TrimSpace(domain),
		owner:        strings.TrimSpace(owner),
	}
}

// Validate validates the mirror configuration.
func (mc MirrorConfig) Validate() error {
	if strings.TrimSpace(mc.name) == "" {
		return domain.ErrMirrorNameRequired
	}

	if strings.TrimSpace(mc.providerType) == "" {
		return domain.ErrProviderTypeRequired
	}

	validProviders := map[string]bool{
		"github": true, "gitlab": true, "gitea": true,
		"directory": true, "archive": true,
	}

	if !validProviders[strings.ToLower(mc.providerType)] {
		return fmt.Errorf("%w: %s", domain.ErrUnsupportedProviderType, mc.providerType)
	}

	// Directory and archive providers require a path
	if (mc.providerType == "directory" || mc.providerType == "archive") && mc.path == "" {
		return fmt.Errorf("%w: %s", domain.ErrPathRequiredForProvider, mc.providerType)
	}

	return nil
}

// Supporting types

// RepositoryValidationError represents a repository validation error for a specific mirror.
type RepositoryValidationError struct {
	Repository string
	Mirror     string
	Err        error
}

// Error implements the error interface.
func (e RepositoryValidationError) Error() string {
	return fmt.Sprintf("repository '%s' invalid for mirror '%s': %v",
		e.Repository, e.Mirror, e.Err)
}

// SyncStep represents a single step in the sync operation.
type SyncStep struct {
	Repository     Repository
	Mirror         MirrorConfig
	RepositoryName string
	EstimatedTime  time.Duration
}

// SyncPlan represents a complete execution plan for a sync operation.
type SyncPlan struct {
	Operation         SyncOperation
	Steps             []SyncStep
	EstimatedDuration time.Duration
}

// TotalSteps returns the total number of sync steps.
func (sp SyncPlan) TotalSteps() int {
	return len(sp.Steps)
}

// StepsForMirror returns steps for a specific mirror.
func (sp SyncPlan) StepsForMirror(mirrorName string) []SyncStep {
	var steps []SyncStep

	for _, step := range sp.Steps {
		if step.Mirror.name == mirrorName {
			steps = append(steps, step)
		}
	}

	return steps
}

// Builder functions for configurations

// WithIncludePatterns adds include patterns to source config.
func (sc SourceConfig) WithIncludePatterns(patterns []string) SourceConfig {
	sc.includePatterns = patterns

	return sc
}

// WithExcludePatterns adds exclude patterns to source config.
func (sc SourceConfig) WithExcludePatterns(patterns []string) SourceConfig {
	sc.excludePatterns = patterns

	return sc
}

// WithForkInclusion sets fork inclusion for source config.
func (sc SourceConfig) WithForkInclusion(include bool) SourceConfig {
	sc.includeForks = include

	return sc
}

// WithPath sets the path for directory/archive providers.
func (mc MirrorConfig) WithPath(path string) MirrorConfig {
	mc.path = strings.TrimSpace(path)

	return mc
}

// WithGitBinary sets whether to use git binary.
func (mc MirrorConfig) WithGitBinary(use bool) MirrorConfig {
	mc.useGitBinary = use

	return mc
}

// Helper functions

// matchesPattern checks if a name matches a glob-like pattern.
func matchesPattern(name, pattern string) bool {
	// Simple glob matching - can be enhanced with proper glob library
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// Basic wildcard matching
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(name, parts[0]) && strings.HasSuffix(name, parts[1])
		}
	}

	return strings.Contains(strings.ToLower(name), strings.ToLower(pattern))
}
