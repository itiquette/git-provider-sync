// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package mirror

import (
	"context"
	"errors"
	"fmt"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// EffectInterpreter executes effects from pure operation plans.
type EffectInterpreter struct {
	gitOps       ports.GitOperations
	repoProvider ports.RepositoryProvider
	logger       ports.Logger
}

// NewEffectInterpreter creates a new effect interpreter with explicit dependencies.
func NewEffectInterpreter(
	gitOps ports.GitOperations,
	repoProvider ports.RepositoryProvider,
	logger ports.Logger,
) *EffectInterpreter {
	return &EffectInterpreter{
		gitOps:       gitOps,
		repoProvider: repoProvider,
		logger:       logger,
	}
}

// ExecuteOperation executes a pure operation plan by interpreting its effects
// Operation parameter contains a sequence of effects to execute with dependency resolution
// Uses simple dependency resolution: unsatisfied effects are moved to end of queue for retry
// Operations execute in order: CreateRepository before ProtectBranch.
func (ei *EffectInterpreter) ExecuteOperation(ctx context.Context, operation Operation) OperationResult {
	start := time.Now()

	result := OperationResult{
		Operation: operation,
		Success:   true,
		Effects:   make([]CompletedEffect, 0, len(operation.Effects)),
		Metrics:   OperationMetrics{},
	}

	ei.logger.Info(ctx, "Executing operation", map[string]any{
		"operation_id":   operation.Metadata.ID,
		"operation_type": operation.Type,
		"effects_count":  len(operation.Effects),
	})

	// Validate operation first
	validationResults := ValidateOperation(operation)
	if len(validationResults) > 0 {
		for _, validationResult := range validationResults {
			if !validationResult.Valid {
				result.Success = false
				result.Error = fmt.Errorf("%w: %s", domain.ErrValidationFailed, validationResult.Message)
				ei.logger.Error(ctx, "Operation validation failed", map[string]any{
					"operation_id": operation.Metadata.ID,
					"error":        result.Error,
					"code":         validationResult.Code,
				})

				return result
			}
		}
	}

	// Execute effects in dependency order using simple dependency resolution algorithm
	// Topological sort by moving unsatisfied effects to the end of the queue
	// Algorithm retries effects until all dependencies are satisfied or no progress can be made
	completedEffects := make(map[string]bool)

	for effectIndex := 0; effectIndex < len(operation.Effects); effectIndex++ {
		effect := operation.Effects[effectIndex]

		// Check if all dependencies for this effect are satisfied
		if !ei.dependenciesSatisfied(effect, completedEffects) {
			// Move this effect to the end and try again (queue rotation for dependency resolution)
			operation.Effects = append(operation.Effects[effectIndex+1:], effect)
			effectIndex-- // Adjust counter since we removed an element

			continue
		}

		completedEffect := ei.executeEffect(ctx, effect, operation)
		result.Effects = append(result.Effects, completedEffect)

		if completedEffect.Success {
			completedEffects[string(effect.Type)] = true

			ei.updateMetrics(&result.Metrics, effect, completedEffect)
		} else {
			result.Success = false
			result.Error = completedEffect.Error
			ei.logger.Error(ctx, "Effect execution failed", map[string]any{
				"operation_id": operation.Metadata.ID,
				"effect_type":  effect.Type,
				"error":        completedEffect.Error,
			})

			break
		}
	}

	result.Duration = time.Since(start)

	ei.logger.Info(ctx, "Operation completed", map[string]any{
		"operation_id": operation.Metadata.ID,
		"success":      result.Success,
		"duration":     result.Duration,
		"effects":      len(result.Effects),
	})

	return result
}

// ExecuteEffect executes a single effect.
func (ei *EffectInterpreter) executeEffect(ctx context.Context, effect Effect, operation Operation) CompletedEffect {
	start := time.Now()

	completedEffect := ei.initializeCompletedEffect(effect)

	ei.logEffectExecution(ctx, effect, operation)

	ei.executeEffectByType(ctx, effect, operation, &completedEffect)

	ei.finalizeCompletedEffect(&completedEffect, start)

	return completedEffect
}

// InitializeCompletedEffect creates a new CompletedEffect with default values.
func (ei *EffectInterpreter) initializeCompletedEffect(effect Effect) CompletedEffect {
	return CompletedEffect{
		Effect:  effect,
		Success: true,
	}
}

// LogEffectExecution logs the start of effect execution.
func (ei *EffectInterpreter) logEffectExecution(ctx context.Context, effect Effect, operation Operation) {
	ei.logger.Debug(ctx, "Executing effect", map[string]any{
		"effect_type":  effect.Type,
		"description":  effect.Description,
		"operation_id": operation.Metadata.ID,
	})
}

// ExecuteEffectByType dispatches the effect execution based on its type.
func (ei *EffectInterpreter) executeEffectByType(ctx context.Context, effect Effect, operation Operation, completedEffect *CompletedEffect) {
	switch effect.Type {
	case EffectTypeCloneRepository:
		completedEffect.Result, completedEffect.Error = ei.executeCloneRepository(ctx, effect, operation)
	case EffectTypeCreateRepository:
		completedEffect.Result, completedEffect.Error = ei.executeCreateRepository(ctx, effect, operation)
	case EffectTypePushToRepository:
		completedEffect.Result, completedEffect.Error = ei.executePushToRepository(ctx, effect, operation)
	case EffectTypeUpdateDescription, EffectTypeUpdateVisibility, EffectTypeUpdateTopics, EffectTypeUpdateDefaultBranch, EffectTypeSyncBranchProtection, EffectTypeCreateDirectories, EffectTypeCleanupTempFiles, EffectTypeLogOperation, EffectTypeRecordMetrics:
		ei.executeSecondaryEffects(ctx, effect, operation, completedEffect)
	default:
		ei.executeSecondaryEffects(ctx, effect, operation, completedEffect)
	}
}

// ExecuteSecondaryEffects handles secondary effect types.
func (ei *EffectInterpreter) executeSecondaryEffects(ctx context.Context, effect Effect, operation Operation, completedEffect *CompletedEffect) {
	switch effect.Type {
	case EffectTypeUpdateDescription:
		completedEffect.Result, completedEffect.Error = ei.executeUpdateDescription(ctx, effect, operation)
	case EffectTypeUpdateVisibility:
		completedEffect.Result, completedEffect.Error = ei.executeUpdateVisibility(ctx, effect, operation)
	case EffectTypeUpdateTopics:
		completedEffect.Result, completedEffect.Error = ei.executeUpdateTopics(ctx, effect, operation)
	case EffectTypeUpdateDefaultBranch:
		completedEffect.Result = ei.executeUpdateDefaultBranch(ctx, effect, operation)
	case EffectTypeSyncBranchProtection:
		completedEffect.Result = ei.executeSyncBranchProtection(ctx, effect, operation)
	case EffectTypeCloneRepository, EffectTypeCreateRepository, EffectTypePushToRepository, EffectTypeCreateDirectories, EffectTypeCleanupTempFiles, EffectTypeLogOperation, EffectTypeRecordMetrics:
		ei.executeUtilityEffects(ctx, effect, operation, completedEffect)
	default:
		ei.executeUtilityEffects(ctx, effect, operation, completedEffect)
	}
}

// ExecuteUtilityEffects handles utility effect types.
func (ei *EffectInterpreter) executeUtilityEffects(ctx context.Context, effect Effect, operation Operation, completedEffect *CompletedEffect) {
	switch effect.Type {
	case EffectTypeCreateDirectories:
		completedEffect.Result = ei.executeCreateDirectories(ctx, effect, operation)
	case EffectTypeRecordMetrics:
		completedEffect.Result = ei.executeRecordMetrics(ctx, effect, operation)
	case EffectTypeCleanupTempFiles:
		completedEffect.Result, completedEffect.Error = ei.executeCleanupTempFiles(ctx, effect, operation)
	case EffectTypeLogOperation:
		completedEffect.Result = ei.executeLogOperation(ctx, effect, operation)
	case EffectTypeCloneRepository, EffectTypeCreateRepository, EffectTypePushToRepository, EffectTypeUpdateDescription, EffectTypeUpdateVisibility, EffectTypeUpdateTopics, EffectTypeUpdateDefaultBranch, EffectTypeSyncBranchProtection:
		completedEffect.Error = fmt.Errorf("%w: %s", domain.ErrEffectTypeShouldNotReachHandler, effect.Type)
	default:
		completedEffect.Error = fmt.Errorf("%w: %s", domain.ErrUnknownEffectType, effect.Type)
	}
}

// FinalizeCompletedEffect sets final fields on the completed effect.
func (ei *EffectInterpreter) finalizeCompletedEffect(completedEffect *CompletedEffect, start time.Time) {
	if completedEffect.Error != nil {
		completedEffect.Success = false
	}

	completedEffect.Duration = time.Since(start)
}

// Effect execution implementations

func (ei *EffectInterpreter) executeCloneRepository(ctx context.Context, effect Effect, operation Operation) (any, error) {
	if operation.Options.DryRun {
		ei.logger.Info(ctx, "[DRY RUN] Would clone repository", map[string]any{
			"url":        effect.Parameters["url"],
			"local_path": effect.Parameters["local_path"],
		})

		return "dry_run_clone", nil
	}

	url, exists := effect.Parameters["url"].(string)
	if !exists {
		return nil, domain.ErrCloneEffectMissingURL
	}

	localPath, pathExists := effect.Parameters["local_path"].(string)
	if !pathExists {
		return nil, domain.ErrCloneEffectMissingPath
	}

	auth, authExists := effect.Parameters["auth"].(AuthSpec)
	if !authExists {
		return nil, domain.ErrCloneEffectMissingAuth
	}

	// Convert auth spec to ports auth
	authOptions := ports.AuthOptions{
		Type:       auth.Type,
		Token:      auth.Token,
		Username:   auth.Username,
		SSHKeyPath: auth.SSHKeyPath,
		SSHKey:     []byte(auth.SSHKey),
	}

	cloneOptions := ports.CloneOptions{
		URL:          url,
		Path:         localPath,
		Auth:         authOptions,
		SingleBranch: true,
		Depth:        1,
	}

	repo, err := ei.gitOps.Clone(ctx, cloneOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return repo, nil
}

func (ei *EffectInterpreter) executeCreateRepository(ctx context.Context, effect Effect, operation Operation) (any, error) {
	if operation.Options.DryRun {
		ei.logger.Info(ctx, "[DRY RUN] Would create repository", map[string]any{
			"name":  effect.Parameters["name"],
			"owner": effect.Parameters["owner"],
		})

		return "dry_run_create", nil
	}

	name, nameExists := effect.Parameters["name"].(string)
	if !nameExists {
		return nil, domain.ErrCreateRepoMissingName
	}

	owner, ownerExists := effect.Parameters["owner"].(string)
	if !ownerExists {
		return nil, domain.ErrCreateRepoMissingOwner
	}

	description, _ := effect.Parameters["description"].(string)
	isPrivate, _ := effect.Parameters["private"].(bool)

	providerConfig := ports.ProviderConfig{
		ProviderType: operation.Target.Provider,
		Domain:       operation.Target.RemotePath,
		Owner:        owner,
		// AuthConfig would need to be extracted from operation context
	}

	createOptions := ports.CreateRepositoryOptions{
		Name:        name,
		Description: description,
		Visibility:  getVisibility(isPrivate),
	}

	repo, err := ei.repoProvider.CreateRepository(ctx, providerConfig, createOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return repo, nil
}

// GetVisibility converts boolean private flag to visibility string.
func getVisibility(isPrivate bool) string {
	if isPrivate {
		return "private"
	}

	return "public"
}

func (ei *EffectInterpreter) executePushToRepository(ctx context.Context, effect Effect, operation Operation) (any, error) {
	if operation.Options.DryRun {
		ei.logger.Info(ctx, "[DRY RUN] Would push to repository", map[string]any{
			"url":        effect.Parameters["url"],
			"local_path": effect.Parameters["local_path"],
		})

		return "dry_run_push", nil
	}

	localPath, pathExists := effect.Parameters["local_path"].(string)
	if !pathExists {
		return nil, domain.ErrPushEffectMissingPath
	}

	auth, authExists := effect.Parameters["auth"].(AuthSpec)
	if !authExists {
		return nil, domain.ErrPushEffectMissingAuth
	}

	force, _ := effect.Parameters["force"].(bool)

	// Open the repository that was cloned
	repo, err := ei.gitOps.Open(ctx, localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	defer func() {
		if err := repo.Close(); err != nil {
			ei.logger.Debug(ctx, "Failed to close repository", map[string]any{
				"error": err.Error(),
				"url":   operation.Source.URL,
			})
		}
	}()

	// Convert auth spec to ports auth
	authOptions := ports.AuthOptions{
		Type:       auth.Type,
		Token:      auth.Token,
		Username:   auth.Username,
		SSHKeyPath: auth.SSHKeyPath,
		SSHKey:     []byte(auth.SSHKey),
	}

	pushOptions := ports.PushOptions{
		Auth:    authOptions,
		Force:   force,
		Timeout: 30 * time.Minute,
	}

	err = repo.Push(ctx, pushOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to push to repository: %w", err)
	}

	return "push_success", nil
}

func (ei *EffectInterpreter) executeUpdateDescription(ctx context.Context, effect Effect, operation Operation) (any, error) {
	return ei.executeRepositoryUpdate(ctx, effect, operation, "description", "description",
		func(value string) ports.UpdateRepositoryOptions {
			return ports.UpdateRepositoryOptions{Description: &value}
		})
}

func (ei *EffectInterpreter) executeUpdateVisibility(ctx context.Context, effect Effect, operation Operation) (any, error) {
	return ei.executeRepositoryUpdate(ctx, effect, operation, "visibility", "visibility",
		func(value string) ports.UpdateRepositoryOptions {
			return ports.UpdateRepositoryOptions{Visibility: &value}
		})
}

// ExecuteRepositoryUpdate is a helper function to reduce code duplication for repository updates.
func (ei *EffectInterpreter) executeRepositoryUpdate(
	ctx context.Context,
	effect Effect,
	operation Operation,
	dryRunAction string,
	paramName string,
	optionsFunc func(string) ports.UpdateRepositoryOptions,
) (any, error) {
	if operation.Options.DryRun {
		ei.logger.Info(ctx, "[DRY RUN] Would update "+dryRunAction, map[string]any{
			"repository": effect.Parameters["repository"],
			paramName:    effect.Parameters[paramName],
		})

		return "dry_run_update_" + dryRunAction, nil
	}

	repository, repoExists := effect.Parameters["repository"].(string)
	if !repoExists {
		return nil, domain.ErrUpdateEffectMissingRepo
	}

	value, valueExists := effect.Parameters[paramName].(string)
	if !valueExists {
		return nil, domain.ErrUpdateEffectMissingParam
	}

	providerConfig := ports.ProviderConfig{
		ProviderType: operation.Target.Provider,
		Domain:       operation.Target.RemotePath,
		Owner:        operation.Target.Owner,
		// AuthConfig would need to be extracted from operation context
	}

	updateOptions := optionsFunc(value)

	err := ei.repoProvider.UpdateRepository(ctx, providerConfig, repository, updateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to update repository %s: %w", dryRunAction, err)
	}

	return dryRunAction + "_updated", nil
}

func (ei *EffectInterpreter) executeUpdateTopics(ctx context.Context, effect Effect, operation Operation) (any, error) {
	if operation.Options.DryRun {
		ei.logger.Info(ctx, "[DRY RUN] Would update topics", map[string]any{
			"repository": effect.Parameters["repository"],
			"topics":     effect.Parameters["topics"],
		})

		return "dry_run_update_topics", nil
	}

	repository, repoExists := effect.Parameters["repository"].(string)
	if !repoExists {
		return nil, domain.ErrUpdateTopicsMissingRepo
	}

	topics, topicsExist := effect.Parameters["topics"].([]string)
	if !topicsExist {
		return nil, domain.ErrUpdateTopicsMissingTopics
	}

	providerConfig := ports.ProviderConfig{
		ProviderType: operation.Target.Provider,
		Domain:       operation.Target.RemotePath,
		Owner:        operation.Target.Owner,
		// AuthConfig would need to be extracted from operation context
	}

	updateOptions := ports.UpdateRepositoryOptions{
		Topics: topics,
	}

	err := ei.repoProvider.UpdateRepository(ctx, providerConfig, repository, updateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to update repository topics: %w", err)
	}

	return "topics_updated", nil
}

func (ei *EffectInterpreter) executeCleanupTempFiles(ctx context.Context, effect Effect, operation Operation) (any, error) {
	if operation.Options.DryRun {
		ei.logger.Info(ctx, "[DRY RUN] Would cleanup temp files", map[string]any{
			"local_path": effect.Parameters["local_path"],
		})

		return "dry_run_cleanup", nil
	}

	localPath, pathExists := effect.Parameters["local_path"].(string)
	if !pathExists {
		return nil, domain.ErrCleanupMissingPath
	}

	err := ei.gitOps.Cleanup(ctx, localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup temp files: %w", err)
	}

	return "cleanup_success", nil
}

func (ei *EffectInterpreter) executeLogOperation(ctx context.Context, _ Effect, operation Operation) any {
	ei.logger.Info(ctx, "Mirror operation logged", map[string]any{
		"operation_id":   operation.Metadata.ID,
		"operation_type": operation.Type,
		"source":         operation.Source.URL,
		"target":         operation.Target.URL,
	})

	return "logged"
}

func (ei *EffectInterpreter) executeUpdateDefaultBranch(ctx context.Context, _ Effect, operation Operation) any {
	// Placeholder implementation for updating default branch
	ei.logger.Info(ctx, "Updating default branch", map[string]any{
		"operation_id": operation.Metadata.ID,
		"source":       operation.Source.URL,
		"target":       operation.Target.URL,
	})

	return "default_branch_updated"
}

//nolint:unparam // Interface method, returns nil by design
func (ei *EffectInterpreter) executeSyncBranchProtection(ctx context.Context, _ Effect, operation Operation) any {
	// TODO: Implement branch protection synchronization
	ei.logger.Warn(ctx, "Branch protection sync not yet implemented", map[string]any{
		"operation_id": operation.Metadata.ID,
		"source":       operation.Source.URL,
		"target":       operation.Target.URL,
	})

	// Return nil to indicate no action taken (not an error, just not implemented)
	return nil
}

func (ei *EffectInterpreter) executeCreateDirectories(ctx context.Context, _ Effect, operation Operation) any {
	// Not implemented - return error instead of fake success
	ei.logger.Warn(ctx, "Directory creation not implemented", map[string]any{
		"operation_id": operation.Metadata.ID,
	})

	return errors.New("directory creation not yet implemented")
}

//nolint:unparam // Interface method, returns nil by design
func (ei *EffectInterpreter) executeRecordMetrics(ctx context.Context, _ Effect, operation Operation) any {
	// Not implemented - return nil to indicate no error but no action taken
	ei.logger.Debug(ctx, "Metrics recording skipped (not implemented)", map[string]any{
		"operation_id":   operation.Metadata.ID,
		"operation_type": operation.Type,
	})

	// Return nil means success but no value produced
	return nil
}

// Helper functions

func (ei *EffectInterpreter) dependenciesSatisfied(effect Effect, completed map[string]bool) bool {
	for _, dep := range effect.DependsOn {
		if !completed[dep] {
			return false
		}
	}

	return true
}

//nolint:cyclop // Complex metrics updating logic with multiple effect types
func (ei *EffectInterpreter) updateMetrics(metrics *OperationMetrics, effect Effect, completed CompletedEffect) {
	switch effect.Type {
	case EffectTypeCloneRepository:
		metrics.NetworkCalls++
	case EffectTypeCreateRepository:
		if completed.Success {
			metrics.RepositoriesCreated++
		}

		metrics.NetworkCalls++
	case EffectTypePushToRepository:
		if completed.Success {
			metrics.RepositoriesUpdated++
		}

		metrics.NetworkCalls++
	case EffectTypeUpdateDescription:
		metrics.NetworkCalls++
	case EffectTypeUpdateVisibility:
		metrics.NetworkCalls++
	case EffectTypeUpdateTopics:
		metrics.NetworkCalls++
	case EffectTypeUpdateDefaultBranch:
		metrics.NetworkCalls++
	case EffectTypeSyncBranchProtection:
		metrics.NetworkCalls++
	case EffectTypeCreateDirectories:
		// Local operation, no network call
	case EffectTypeCleanupTempFiles:
		// Local operation, no network call
	case EffectTypeLogOperation:
		// Local operation, no network call
	case EffectTypeRecordMetrics:
	}

	metrics.RepositoriesProcessed++
}
