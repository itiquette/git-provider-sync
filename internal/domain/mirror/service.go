// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Service provides pure functional mirror operations.
type Service struct {
	interpreter *EffectInterpreter
	config      Config
}

// Config contains configuration for mirror operations.
type Config struct {
	TempDirectory   string
	DefaultTimeout  time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	EnableMetrics   bool
	EnableLogging   bool
	DryRunByDefault bool
	ForceByDefault  bool
}

// NewService creates a new pure functional mirror service.
func NewService(
	gitOps ports.GitOperations,
	repoProvider ports.RepositoryProvider,
	logger ports.Logger,
	config Config,
) *Service {
	return &Service{
		interpreter: NewEffectInterpreter(gitOps, repoProvider, logger),
		config:      config,
	}
}

// NewMirrorService creates a mirror service with default configuration.
func NewMirrorService(
	gitOps ports.GitOperations,
	repoProvider ports.RepositoryProvider,
	logger ports.Logger,
) *Service {
	config := Config{
		TempDirectory:   "/tmp/git-provider-sync",
		DefaultTimeout:  30 * time.Minute,
		MaxRetries:      3,
		RetryDelay:      time.Second,
		EnableMetrics:   true,
		EnableLogging:   true,
		DryRunByDefault: false,
		ForceByDefault:  false,
	}

	return NewService(gitOps, repoProvider, logger, config)
}

// NewDryRunMirrorService creates a mirror service configured for dry run mode.
func NewDryRunMirrorService(
	gitOps ports.GitOperations,
	repoProvider ports.RepositoryProvider,
	logger ports.Logger,
) *Service {
	config := Config{
		TempDirectory:   "/tmp/git-provider-sync-dryrun",
		DefaultTimeout:  30 * time.Minute,
		MaxRetries:      1,
		RetryDelay:      time.Second,
		EnableMetrics:   true,
		EnableLogging:   true,
		DryRunByDefault: true,
		ForceByDefault:  false,
	}

	return NewService(gitOps, repoProvider, logger, config)
}

// MirrorRepository performs a complete mirror operation using pure functions.
func (s *Service) MirrorRepository(
	ctx context.Context,
	source entities.Repository,
	sourceAuth AuthSpec,
	target entities.Repository,
	targetAuth AuthSpec,
	options ...OperationOptionFunc,
) OperationResult {
	// Build operation specifications from repository info
	sourceSpec := s.buildRepositorySpecWithAuth(source, sourceAuth)
	targetSpec := s.buildRepositorySpecWithAuth(target, targetAuth)

	// Build operation options with defaults
	opOptions := s.buildOperationOptions(options...)

	// Plan the mirror operation (pure function)
	operation := PlanCloneAndMirror(sourceSpec, targetSpec, opOptions)

	// Execute the operation plan
	return s.interpreter.ExecuteOperation(ctx, operation)
}

// SyncRepository performs a sync operation (for existing repositories).
func (s *Service) SyncRepository(
	ctx context.Context,
	source entities.Repository,
	sourceAuth AuthSpec,
	target entities.Repository,
	targetAuth AuthSpec,
	options ...OperationOptionFunc,
) OperationResult {
	sourceSpec := s.buildRepositorySpecWithAuth(source, sourceAuth)
	targetSpec := s.buildRepositorySpecWithAuth(target, targetAuth)
	opOptions := s.buildOperationOptions(options...)

	// Plan the sync operation (pure function)
	operation := PlanSync(sourceSpec, targetSpec, opOptions)

	// Execute the operation plan
	return s.interpreter.ExecuteOperation(ctx, operation)
}

// MirrorRepositories performs batch mirroring of multiple repositories.
func (s *Service) MirrorRepositories(
	ctx context.Context,
	sourceMirrorPairs []SourceMirrorPair,
	options ...OperationOptionFunc,
) BatchOperationResult {
	start := time.Now()

	result := BatchOperationResult{
		TotalOperations: len(sourceMirrorPairs),
		Results:         make([]OperationResult, 0, len(sourceMirrorPairs)),
		Success:         true,
		Metrics:         BatchOperationMetrics{},
	}

	for _, pair := range sourceMirrorPairs {
		opResult := s.MirrorRepository(ctx, pair.Source, pair.SourceAuth, pair.Target, pair.TargetAuth, options...)
		result.Results = append(result.Results, opResult)

		// Update batch metrics
		s.updateBatchMetrics(&result.Metrics, opResult)

		if !opResult.Success {
			result.Success = false
			result.FailedOperations++
		} else {
			result.SuccessfulOperations++
		}

		// Check context cancellation
		if ctx.Err() != nil {
			result.Success = false
			result.Error = ctx.Err()

			break
		}
	}

	result.Duration = time.Since(start)

	return result
}

// ValidateRepositoryPair validates that a source and target can be mirrored.
func (s *Service) ValidateRepositoryPair(
	_ context.Context,
	source entities.Repository,
	sourceAuth AuthSpec,
	target entities.Repository,
	targetAuth AuthSpec,
) ValidationResults {
	sourceSpec := s.buildRepositorySpecWithAuth(source, sourceAuth)
	targetSpec := s.buildRepositorySpecWithAuth(target, targetAuth)
	opOptions := s.buildOperationOptions()

	// Create a validation-only operation
	operation := Operation{
		Type:    OperationTypeValidate,
		Source:  sourceSpec,
		Target:  targetSpec,
		Options: opOptions,
		Validations: []ValidationRule{
			{Name: "ValidSourceURL", Predicate: validateSourceURL},
			{Name: "ValidTargetURL", Predicate: validateTargetURL},
			{Name: "ValidAuth", Predicate: validateAuth},
			{Name: "ValidRepositoryNames", Predicate: validateRepositoryNames},
			{Name: "ValidProviderCompatibility", Predicate: validateProviderCompatibility},
		},
	}

	validationResults := ValidateOperation(operation)

	return ValidationResults{
		Valid:   len(validationResults) == 0,
		Results: validationResults,
		Source:  sourceSpec,
		Target:  targetSpec,
	}
}

// PlanMirrorOperation creates a pure operation plan without executing it.
func (s *Service) PlanMirrorOperation(
	source entities.Repository,
	sourceAuth AuthSpec,
	target entities.Repository,
	targetAuth AuthSpec,
	options ...OperationOptionFunc,
) Operation {
	sourceSpec := s.buildRepositorySpecWithAuth(source, sourceAuth)
	targetSpec := s.buildRepositorySpecWithAuth(target, targetAuth)
	opOptions := s.buildOperationOptions(options...)

	return PlanCloneAndMirror(sourceSpec, targetSpec, opOptions)
}

// Helper methods

func (s *Service) buildRepositorySpecWithAuth(repo entities.Repository, auth AuthSpec) RepositorySpec {
	// Generate a unique local path for this operation
	localPath := filepath.Join(s.config.TempDirectory, fmt.Sprintf("%s-%s-%d",
		"repo", repo.Name(), time.Now().Unix()))

	return RepositorySpec{
		URL:         repo.PreferredCloneURL(),
		Name:        repo.Name(),
		Owner:       s.extractOwnerFromURL(repo.PreferredCloneURL()),
		Provider:    repo.ProviderType(),
		Branch:      repo.DefaultBranch(),
		LocalPath:   localPath,
		IsPrivate:   repo.IsPrivate(),
		Topics:      []string{}, // Repository topics not implemented
		Description: repo.Description(),
		Visibility:  repo.Visibility(),
		Auth:        auth,
	}
}

func (s *Service) extractOwnerFromURL(url string) string {
	// Simple extraction - could be improved
	// Example: https://github.com/owner/repo.git -> owner
	// This is a placeholder implementation
	parts := strings.Split(url, "/")
	if len(parts) >= 4 {
		return parts[len(parts)-2]
	}

	return "unknown"
}

func (s *Service) buildOperationOptions(options ...OperationOptionFunc) OperationOptions {
	base := OperationOptions{
		DryRun:               s.config.DryRunByDefault,
		CreateIfNotExists:    true,
		UpdateDescription:    true,
		SyncVisibility:       true,
		SyncTopics:           true,
		SyncDefaultBranch:    true,
		SyncBranchProtection: false, // Security default
		PreservePullRequests: true,
		PreserveIssues:       true,
		EnableLFS:            false,
		Force:                s.config.ForceByDefault,
		Timeout:              s.config.DefaultTimeout,
		RetryPolicy: RetryPolicy{
			MaxAttempts: s.config.MaxRetries,
			Delay:       s.config.RetryDelay,
			Backoff:     BackoffStrategyExponential,
		},
	}

	return ApplyOperationOptions(base, options...)
}

func (s *Service) updateBatchMetrics(metrics *BatchOperationMetrics, result OperationResult) {
	metrics.TotalRepositoriesProcessed += result.Metrics.RepositoriesProcessed
	metrics.TotalRepositoriesCreated += result.Metrics.RepositoriesCreated
	metrics.TotalRepositoriesUpdated += result.Metrics.RepositoriesUpdated
	metrics.TotalRepositoriesSkipped += result.Metrics.RepositoriesSkipped
	metrics.TotalRepositoriesFailed += result.Metrics.RepositoriesFailed
	metrics.TotalBytesTransferred += result.Metrics.BytesTransferred
	metrics.TotalNetworkCalls += result.Metrics.NetworkCalls
	metrics.TotalDuration += result.Duration
}

// Additional pure validation functions

func validateRepositoryNames(operation Operation) ValidationResult {
	if operation.Source.Name == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Source repository name cannot be empty",
			Code:    "EMPTY_SOURCE_NAME",
		}
	}

	if operation.Target.Name == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Target repository name cannot be empty",
			Code:    "EMPTY_TARGET_NAME",
		}
	}

	return ValidationResult{Valid: true}
}

func validateProviderCompatibility(operation Operation) ValidationResult {
	// Add provider-specific validation logic here
	supportedProviders := []string{"github", "gitlab", "gitea", "directory", "archive"}

	sourceSupported := false
	targetSupported := false

	for _, provider := range supportedProviders {
		if operation.Source.Provider == provider {
			sourceSupported = true
		}

		if operation.Target.Provider == provider {
			targetSupported = true
		}
	}

	if !sourceSupported {
		return ValidationResult{
			Valid:   false,
			Message: "Unsupported source provider: " + operation.Source.Provider,
			Code:    "UNSUPPORTED_SOURCE_PROVIDER",
		}
	}

	if !targetSupported {
		return ValidationResult{
			Valid:   false,
			Message: "Unsupported target provider: " + operation.Target.Provider,
			Code:    "UNSUPPORTED_TARGET_PROVIDER",
		}
	}

	return ValidationResult{Valid: true}
}

// Supporting types

// SourceMirrorPair represents a source-target pair for batch operations.
type SourceMirrorPair struct {
	Source     entities.Repository
	SourceAuth AuthSpec
	Target     entities.Repository
	TargetAuth AuthSpec
}

// BatchOperationResult represents the result of a batch mirror operation.
type BatchOperationResult struct {
	TotalOperations      int
	SuccessfulOperations int
	FailedOperations     int
	Results              []OperationResult
	Success              bool
	Error                error
	Duration             time.Duration
	Metrics              BatchOperationMetrics
}

// BatchOperationMetrics contains aggregated metrics for batch operations.
type BatchOperationMetrics struct {
	TotalRepositoriesProcessed int
	TotalRepositoriesCreated   int
	TotalRepositoriesUpdated   int
	TotalRepositoriesSkipped   int
	TotalRepositoriesFailed    int
	TotalBytesTransferred      int64
	TotalNetworkCalls          int
	TotalDuration              time.Duration
}

// ValidationResults represents the results of repository pair validation.
type ValidationResults struct {
	Valid   bool
	Results []ValidationResult
	Source  RepositorySpec
	Target  RepositorySpec
}
