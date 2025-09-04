// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package mirror provides pure functional mirror operations
package mirror

import (
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Operation represents a pure description of a mirror operation to be performed.
type Operation struct {
	Type        OperationType
	Source      RepositorySpec
	Target      RepositorySpec
	Options     OperationOptions
	Metadata    OperationMetadata
	Effects     []Effect
	Validations []ValidationRule
}

// OperationType defines the type of mirror operation.
type OperationType string

const (
	// OperationTypeCloneAndMirror represents clone and mirror operation.
	OperationTypeCloneAndMirror OperationType = "clone_and_mirror"
	// OperationTypeSync represents sync operation.
	OperationTypeSync OperationType = "sync"
	// OperationTypeValidate represents validate operation.
	OperationTypeValidate OperationType = "validate"
	// OperationTypeCleanup represents cleanup operation.
	OperationTypeCleanup OperationType = "cleanup"
)

// RepositorySpec describes a repository in a pure, immutable way.
type RepositorySpec struct {
	URL         string
	Name        string
	Owner       string
	Provider    string
	Branch      string
	Auth        AuthSpec
	LocalPath   string
	RemotePath  string
	IsPrivate   bool
	Topics      []string
	Description string
	Visibility  string
}

// AuthSpec describes authentication in a pure way.
type AuthSpec struct {
	Type       ports.AuthType
	Token      string
	Username   string
	SSHKeyPath string
	SSHKey     string
}

// OperationOptions contains pure configuration for operations.
type OperationOptions struct {
	DryRun               bool
	CreateIfNotExists    bool
	UpdateDescription    bool
	SyncVisibility       bool
	SyncTopics           bool
	SyncDefaultBranch    bool
	SyncBranchProtection bool
	PreservePullRequests bool
	PreserveIssues       bool
	EnableLFS            bool
	Force                bool
	Timeout              time.Duration
	RetryPolicy          RetryPolicy
}

// RetryPolicy describes retry behavior in a pure way.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     BackoffStrategy
}

// BackoffStrategy defines retry backoff strategies.
type BackoffStrategy string

const (
	// BackoffStrategyLinear represents linear backoff.
	BackoffStrategyLinear BackoffStrategy = "linear"
	// BackoffStrategyExponential represents exponential backoff.
	BackoffStrategyExponential BackoffStrategy = "exponential"
	// BackoffStrategyFixed represents fixed backoff.
	BackoffStrategyFixed BackoffStrategy = "fixed"
)

// OperationMetadata contains metadata about the operation.
type OperationMetadata struct {
	ID          string
	CreatedAt   time.Time
	RequestedBy string
	Priority    Priority
	Tags        map[string]string
}

// Priority defines operation priority levels.
type Priority string

const (
	// PriorityLow represents low priority.
	PriorityLow Priority = "low"
	// PriorityNormal represents normal priority.
	PriorityNormal Priority = "normal"
	// PriorityHigh represents high priority.
	PriorityHigh Priority = "high"
)

// Effect represents a side effect that needs to be performed.
type Effect struct {
	Type        EffectType
	Description string
	Parameters  map[string]any
	DependsOn   []string
}

// EffectType defines types of side effects.
type EffectType string

const (
	// EffectTypeCloneRepository represents clone repository effect.
	EffectTypeCloneRepository EffectType = "clone_repository"
	// EffectTypeCreateRepository represents create repository effect.
	EffectTypeCreateRepository EffectType = "create_repository"
	// EffectTypePushToRepository represents push to repository effect.
	EffectTypePushToRepository EffectType = "push_to_repository"
	// EffectTypeUpdateDescription represents update description effect.
	EffectTypeUpdateDescription EffectType = "update_description"
	// EffectTypeUpdateVisibility represents update visibility effect.
	EffectTypeUpdateVisibility EffectType = "update_visibility"
	// EffectTypeUpdateTopics represents update topics effect.
	EffectTypeUpdateTopics EffectType = "update_topics"
	// EffectTypeUpdateDefaultBranch represents update default branch effect.
	EffectTypeUpdateDefaultBranch EffectType = "update_default_branch"
	// EffectTypeSyncBranchProtection represents sync branch protection effect.
	EffectTypeSyncBranchProtection EffectType = "sync_branch_protection"
	// EffectTypeCreateDirectories represents create directories effect.
	EffectTypeCreateDirectories EffectType = "create_directories"
	// EffectTypeCleanupTempFiles represents cleanup temp files effect.
	EffectTypeCleanupTempFiles EffectType = "cleanup_temp_files"
	// EffectTypeLogOperation represents log operation effect.
	EffectTypeLogOperation EffectType = "log_operation"
	// EffectTypeRecordMetrics represents record metrics effect.
	EffectTypeRecordMetrics EffectType = "record_metrics"
)

// ValidationRule represents a validation rule for mirror operations.
type ValidationRule struct {
	Name        string
	Description string
	Predicate   func(Operation) ValidationResult
}

// ValidationResult represents the result of a validation rule.
type ValidationResult struct {
	Valid   bool
	Message string
	Code    string
}

// OperationResult represents the result of a mirror operation plan.
type OperationResult struct {
	Operation Operation
	Success   bool
	Error     error
	Effects   []CompletedEffect
	Metrics   OperationMetrics
	Duration  time.Duration
}

// CompletedEffect represents an effect that has been executed.
type CompletedEffect struct {
	Effect   Effect
	Success  bool
	Error    error
	Result   any
	Duration time.Duration
}

// OperationMetrics contains metrics about the operation.
type OperationMetrics struct {
	RepositoriesProcessed int
	RepositoriesCreated   int
	RepositoriesUpdated   int
	RepositoriesSkipped   int
	RepositoriesFailed    int
	BytesTransferred      int64
	NetworkCalls          int
	CacheHits             int
	CacheMisses           int
}

// Pure Functions for Mirror Operations

// PlanCloneAndMirror creates a pure operation plan for cloning and mirroring a repository.
func PlanCloneAndMirror(source, target RepositorySpec, options OperationOptions) Operation {
	operation := Operation{
		Type:    OperationTypeCloneAndMirror,
		Source:  source,
		Target:  target,
		Options: options,
		Metadata: OperationMetadata{
			ID:        generateOperationID(source, target),
			CreatedAt: time.Now(),
			Priority:  PriorityNormal,
			Tags:      make(map[string]string),
		},
	}

	// Add effects based on options
	effects := []Effect{}

	// Always need to clone the source
	effects = append(effects, Effect{
		Type:        EffectTypeCloneRepository,
		Description: "Clone source repository",
		Parameters: map[string]any{
			"url":        source.URL,
			"local_path": source.LocalPath,
			"auth":       source.Auth,
			"branch":     source.Branch,
		},
	})

	// Create target repository if it doesn't exist
	if options.CreateIfNotExists {
		effects = append(effects, Effect{
			Type:        EffectTypeCreateRepository,
			Description: "Create target repository if not exists",
			Parameters: map[string]any{
				"name":        target.Name,
				"owner":       target.Owner,
				"description": target.Description,
				"private":     target.IsPrivate,
			},
		})
	}

	// Push to target
	effects = append(effects, Effect{
		Type:        EffectTypePushToRepository,
		Description: "Push mirrored content to target",
		Parameters: map[string]any{
			"url":        target.URL,
			"local_path": source.LocalPath,
			"auth":       target.Auth,
			"force":      options.Force,
		},
		DependsOn: []string{"clone_repository"},
	})

	// Update repository metadata if requested
	if options.UpdateDescription {
		effects = append(effects, Effect{
			Type:        EffectTypeUpdateDescription,
			Description: "Update repository description",
			Parameters: map[string]any{
				"repository":  target.Name,
				"description": source.Description,
			},
		})
	}

	if options.SyncVisibility {
		effects = append(effects, Effect{
			Type:        EffectTypeUpdateVisibility,
			Description: "Sync repository visibility",
			Parameters: map[string]any{
				"repository": target.Name,
				"visibility": source.Visibility,
			},
		})
	}

	if options.SyncTopics {
		effects = append(effects, Effect{
			Type:        EffectTypeUpdateTopics,
			Description: "Sync repository topics",
			Parameters: map[string]any{
				"repository": target.Name,
				"topics":     source.Topics,
			},
		})
	}

	// Cleanup temporary files
	if !options.DryRun {
		effects = append(effects, Effect{
			Type:        EffectTypeCleanupTempFiles,
			Description: "Clean up temporary files",
			Parameters: map[string]any{
				"local_path": source.LocalPath,
			},
		})
	}

	operation.Effects = effects

	// Add validations
	operation.Validations = []ValidationRule{
		{
			Name:        "ValidSourceURL",
			Description: "Source URL must be valid",
			Predicate:   validateSourceURL,
		},
		{
			Name:        "ValidTargetURL",
			Description: "Target URL must be valid",
			Predicate:   validateTargetURL,
		},
		{
			Name:        "ValidAuth",
			Description: "Authentication must be valid",
			Predicate:   validateAuth,
		},
	}

	return operation
}

// PlanSync creates a pure operation plan for syncing repositories.
func PlanSync(source, target RepositorySpec, options OperationOptions) Operation {
	operation := Operation{
		Type:    OperationTypeSync,
		Source:  source,
		Target:  target,
		Options: options,
		Metadata: OperationMetadata{
			ID:        generateOperationID(source, target),
			CreatedAt: time.Now(),
			Priority:  PriorityNormal,
			Tags:      make(map[string]string),
		},
	}

	// For sync, we assume repositories already exist
	effects := []Effect{
		{
			Type:        EffectTypeCloneRepository,
			Description: "Clone source repository",
			Parameters: map[string]any{
				"url":        source.URL,
				"local_path": source.LocalPath,
				"auth":       source.Auth,
			},
		},
		{
			Type:        EffectTypePushToRepository,
			Description: "Push changes to target",
			Parameters: map[string]any{
				"url":        target.URL,
				"local_path": source.LocalPath,
				"auth":       target.Auth,
				"force":      options.Force,
			},
			DependsOn: []string{"clone_repository"},
		},
	}

	if !options.DryRun {
		effects = append(effects, Effect{
			Type:        EffectTypeCleanupTempFiles,
			Description: "Clean up temporary files",
			Parameters: map[string]any{
				"local_path": source.LocalPath,
			},
		})
	}

	operation.Effects = effects
	operation.Validations = []ValidationRule{
		{
			Name:        "ValidSourceURL",
			Description: "Source URL must be valid",
			Predicate:   validateSourceURL,
		},
		{
			Name:        "ValidTargetURL",
			Description: "Target URL must be valid",
			Predicate:   validateTargetURL,
		},
	}

	return operation
}

// ValidateOperation validates a mirror operation using pure functions.
func ValidateOperation(operation Operation) []ValidationResult {
	results := []ValidationResult{}

	for _, rule := range operation.Validations {
		result := rule.Predicate(operation)
		if !result.Valid {
			results = append(results, result)
		}
	}

	return results
}

// Pure validation functions

func validateSourceURL(op Operation) ValidationResult {
	if op.Source.URL == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Source URL cannot be empty",
			Code:    "EMPTY_SOURCE_URL",
		}
	}

	return ValidationResult{Valid: true}
}

func validateTargetURL(op Operation) ValidationResult {
	if op.Target.URL == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Target URL cannot be empty",
			Code:    "EMPTY_TARGET_URL",
		}
	}

	return ValidationResult{Valid: true}
}

func validateAuth(operation Operation) ValidationResult {
	// Check source auth
	if operation.Source.Auth.Type != ports.AuthTypeNone &&
		operation.Source.Auth.Token == "" &&
		operation.Source.Auth.SSHKey == "" &&
		operation.Source.Auth.SSHKeyPath == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Source authentication is required but not provided",
			Code:    "MISSING_SOURCE_AUTH",
		}
	}

	// Check target auth
	if operation.Target.Auth.Type != ports.AuthTypeNone &&
		operation.Target.Auth.Token == "" &&
		operation.Target.Auth.SSHKey == "" &&
		operation.Target.Auth.SSHKeyPath == "" {
		return ValidationResult{
			Valid:   false,
			Message: "Target authentication is required but not provided",
			Code:    "MISSING_TARGET_AUTH",
		}
	}

	return ValidationResult{Valid: true}
}

// Helper functions

func generateOperationID(source, target RepositorySpec) string {
	return source.Owner + "/" + source.Name + "->" + target.Owner + "/" + target.Name
}

// BuildRepositorySpec creates a repository spec from configuration.
func BuildRepositorySpec(provider, owner, name, url, branch string, auth AuthSpec) RepositorySpec {
	return RepositorySpec{
		URL:      url,
		Name:     name,
		Owner:    owner,
		Provider: provider,
		Branch:   branch,
		Auth:     auth,
	}
}

// BuildOperationOptions creates operation options with defaults.
func BuildOperationOptions() OperationOptions {
	return OperationOptions{
		DryRun:               false,
		CreateIfNotExists:    true,
		UpdateDescription:    true,
		SyncVisibility:       true,
		SyncTopics:           true,
		SyncDefaultBranch:    true,
		SyncBranchProtection: false, // Security default
		PreservePullRequests: true,
		PreserveIssues:       true,
		EnableLFS:            false,
		Force:                false,
		Timeout:              30 * time.Minute,
		RetryPolicy: RetryPolicy{
			MaxAttempts: 3,
			Delay:       time.Second,
			Backoff:     BackoffStrategyExponential,
		},
	}
}

// OperationOptionFunc applies functional options to operation options.
type OperationOptionFunc func(OperationOptions) OperationOptions

// WithDryRun sets dry run mode.
func WithDryRun(dryRun bool) OperationOptionFunc {
	return func(opts OperationOptions) OperationOptions {
		opts.DryRun = dryRun

		return opts
	}
}

// WithForce sets force mode.
func WithForce(force bool) OperationOptionFunc {
	return func(opts OperationOptions) OperationOptions {
		opts.Force = force

		return opts
	}
}

// WithTimeout sets operation timeout.
func WithTimeout(timeout time.Duration) OperationOptionFunc {
	return func(opts OperationOptions) OperationOptions {
		opts.Timeout = timeout

		return opts
	}
}

// ApplyOperationOptions applies functional options to build operation options.
func ApplyOperationOptions(base OperationOptions, options ...OperationOptionFunc) OperationOptions {
	result := base
	for _, option := range options {
		result = option(result)
	}

	return result
}
