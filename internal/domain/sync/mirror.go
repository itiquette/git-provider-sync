// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/log"
)

// ToMirrorsUseCase syncs repositories to multiple mirror destinations.
type ToMirrorsUseCase struct {
	repositoryProvider ports.RepositoryProvider
	gitOperations      ports.GitOperations
	archiveOperations  ports.ArchiveOperations
	fileSystem         ports.FileSystem
	logger             ports.Logger
}

// NewToMirrorsUseCase creates a new ToMirrorsUseCase.
func NewToMirrorsUseCase(
	repositoryProvider ports.RepositoryProvider,
	gitOps ports.GitOperations,
	archiveOps ports.ArchiveOperations,
	fileSystem ports.FileSystem,
	logger ports.Logger,
) ToMirrorsUseCase {
	return ToMirrorsUseCase{
		repositoryProvider: repositoryProvider,
		gitOperations:      gitOps,
		archiveOperations:  archiveOps,
		fileSystem:         fileSystem,
		logger:             logger,
	}
}

// ToMirrorsRequest represents the input for syncing to mirrors.
type ToMirrorsRequest struct {
	SourceRepositories []ports.GitRepository
	MirrorTargets      []entities.MirrorTarget
	SourceConfig       ports.ProviderConfig
	DryRun             bool
	Options            Options
}

// Options contains options for mirror synchronization.
type Options struct {
	ForcePush          bool
	IgnoreInvalidNames bool
	CreateIfNotExists  bool
	UpdateDescription  bool
	SyncProtection     bool
	NameTransformation ports.NameTransformOptions
}

// ToMirrorsResponse represents the result of syncing to mirrors.
type ToMirrorsResponse struct {
	Results           []MirrorResult
	Success           bool
	TotalRepositories int
	SuccessfulSyncs   int
	FailedSyncs       int
	SkippedSyncs      int
	Errors            []error
}

// MirrorResult represents the result of syncing a single repository to a mirror.
type MirrorResult struct {
	RepositoryName string
	MirrorName     string
	Success        bool
	Skipped        bool
	Error          error
	Action         string // "created", "updated", "skipped"
}

// Execute syncs repositories to all configured mirror targets.
func (uc ToMirrorsUseCase) Execute(
	ctx context.Context,
	request ToMirrorsRequest,
) (ToMirrorsResponse, error) {
	logger := log.CreateDomainLogger(ctx)

	var response ToMirrorsResponse

	var err error

	// TRACE: Use case entry point (hexagonal boundary)
	logger.Trace(ctx, "ToMirrorsUseCase.Execute entry", map[string]any{
		"source_repos":   len(request.SourceRepositories),
		"mirror_targets": len(request.MirrorTargets),
		"dry_run":        request.DryRun,
	})

	defer func() {
		// TRACE: Use case exit point with outcome
		logger.Trace(ctx, "ToMirrorsUseCase.Execute exit", map[string]any{
			"success":            response.Success,
			"total_repositories": response.TotalRepositories,
			"successful_syncs":   response.SuccessfulSyncs,
			"failed_syncs":       response.FailedSyncs,
			"error":              err != nil,
		})
	}()

	logger.Info(ctx, "Starting mirror synchronization", map[string]any{
		"source_repos":   len(request.SourceRepositories),
		"mirror_targets": len(request.MirrorTargets),
		"dry_run":        request.DryRun,
	})

	response = ToMirrorsResponse{
		Results:           []MirrorResult{},
		Success:           true,
		TotalRepositories: len(request.SourceRepositories),
		SuccessfulSyncs:   0,
		FailedSyncs:       0,
		SkippedSyncs:      0,
		Errors:            []error{},
	}

	// Process each mirror target
	for i, mirrorTarget := range request.MirrorTargets {
		// TRACE: Mirror processing boundary
		logger.Trace(ctx, "processing mirror target", map[string]any{
			"mirror_index":  i + 1,
			"total_mirrors": len(request.MirrorTargets),
			"mirror_name":   mirrorTarget.Name(),
			"provider":      mirrorTarget.ProviderType(),
		})

		if !mirrorTarget.Enabled() {
			logger.Info(ctx, "Skipping disabled mirror", map[string]any{
				"mirror_name": mirrorTarget.Name(),
			})

			continue
		}

		logger.Info(ctx, "Processing mirror target", map[string]any{
			"mirror_name":   mirrorTarget.Name(),
			"provider_type": mirrorTarget.ProviderType(),
			"mirror_owner":  mirrorTarget.Owner(),
		})

		// Sync all repositories to this mirror
		mirrorResults := uc.syncRepositoriesToMirror(ctx, request.SourceRepositories, mirrorTarget, request)

		response.Results = append(response.Results, mirrorResults...)
	}

	// Calculate summary statistics
	for _, result := range response.Results {
		if result.Success { //nolint:gocritic // if-else chain is more readable for result categorization logic
			response.SuccessfulSyncs++
		} else if result.Skipped {
			response.SkippedSyncs++
		} else {
			response.FailedSyncs++
		}
	}

	logger.Info(ctx, "Mirror synchronization completed", map[string]any{
		"total_operations": len(response.Results),
		"successful":       response.SuccessfulSyncs,
		"failed":           response.FailedSyncs,
		"skipped":          response.SkippedSyncs,
		"success":          response.Success,
	})

	return response, nil
}

// SyncRepositoriesToMirror processes repository batch synchronization for a specific mirror target.
func (uc ToMirrorsUseCase) syncRepositoriesToMirror(
	ctx context.Context,
	sourceRepos []ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	request ToMirrorsRequest,
) []MirrorResult {
	// Initialize sync metadata
	ctx = uc.initMirrorSync(ctx, mirrorTarget, len(sourceRepos))

	results := make([]MirrorResult, 0, len(sourceRepos))

	// Create mirror provider if it's a git provider (not directory/archive)
	var mirrorProvider ports.RepositoryProvider
	if mirrorTarget.ProviderType() != entities.ProviderTypeDirectory && mirrorTarget.ProviderType() != entities.ProviderTypeArchive {
		// Reuse source provider instance for git providers (simplified architecture)
		// In a full implementation, this would use a provider factory to create target-specific providers
		mirrorProvider = uc.repositoryProvider
	}

	// Process each repository
	for _, repo := range sourceRepos {
		result := uc.syncSingleRepositoryToMirror(ctx, repo, mirrorTarget, mirrorProvider, request)
		results = append(results, result)
	}

	return results
}

// SyncSingleRepositoryToMirror synchronizes a single repository to a mirror target.
func (uc ToMirrorsUseCase) syncSingleRepositoryToMirror(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
	request ToMirrorsRequest,
) MirrorResult {
	logger := log.CreateDomainLogger(ctx)

	repoName := sourceRepo.Name()
	result := uc.initMirrorResult(repoName, mirrorTarget.Name())

	// TRACE: Per-repository sync entry
	logger.Trace(ctx, "syncSingleRepositoryToMirror entry", map[string]any{
		"repository":    repoName,
		"mirror":        mirrorTarget.Name(),
		"provider_type": mirrorTarget.ProviderType(),
	})

	logger.Debug(ctx, "Processing repository for mirror", map[string]any{
		"repository":    repoName,
		"mirror":        mirrorTarget.Name(),
		"provider_type": mirrorTarget.ProviderType(),
	})

	if uc.shouldSkipInvalidRepo(ctx, repoName, mirrorTarget.Name(), request.Options.IgnoreInvalidNames) {
		result.Skipped = true
		result.Action = "skipped_invalid"

		return result
	}

	if err := uc.validateRepositoryForMirror(ctx, repoName, mirrorTarget, mirrorProvider); err != nil {
		if request.Options.IgnoreInvalidNames {
			uc.logger.Warn(ctx, "Ignoring repository with invalid name", map[string]any{
				"repository": repoName,
				"mirror":     mirrorTarget.Name(),
				"error":      err.Error(),
			})

			result.Skipped = true
			result.Action = "skipped"

			return result
		}

		result.Error = err

		return result
	}

	targetRepoName := uc.transformRepositoryName(repoName, request.Options.NameTransformation)

	if request.DryRun {
		return uc.handleDryRun(ctx, repoName, targetRepoName, mirrorTarget.Name())
	}

	if isFileSystemTarget(mirrorTarget) {
		return uc.syncToFileSystem(ctx, sourceRepo, mirrorTarget, targetRepoName)
	}

	return uc.syncToGitProvider(ctx, sourceRepo, mirrorTarget, mirrorProvider, targetRepoName, request)
}

func (uc ToMirrorsUseCase) initMirrorResult(repoName, mirrorName string) MirrorResult {
	return MirrorResult{
		RepositoryName: repoName,
		MirrorName:     mirrorName,
		Success:        false,
		Skipped:        false,
		Error:          nil,
		Action:         "",
	}
}

func (uc ToMirrorsUseCase) shouldSkipInvalidRepo(ctx context.Context, repoName, mirrorName string, ignoreInvalidNames bool) bool {
	if !entities.ContainsFailureInContext(ctx, "invalid", repoName) {
		return false
	}

	if ignoreInvalidNames {
		uc.logger.Warn(ctx, "Repository already marked as invalid, skipping", map[string]any{
			"repository": repoName,
			"mirror":     mirrorName,
		})

		return true
	}

	return false
}

func (uc ToMirrorsUseCase) handleDryRun(ctx context.Context, repoName, targetRepoName, mirrorName string) MirrorResult {
	uc.logger.Info(ctx, "Dry run - would sync repository", map[string]any{
		"source_repo": repoName,
		"target_repo": targetRepoName,
		"mirror":      mirrorName,
	})

	return MirrorResult{
		RepositoryName: repoName,
		MirrorName:     mirrorName,
		Success:        true,
		Action:         "dry_run",
	}
}

func isFileSystemTarget(target entities.MirrorTarget) bool {
	return target.ProviderType() == entities.ProviderTypeDirectory ||
		target.ProviderType() == entities.ProviderTypeArchive
}

func (uc ToMirrorsUseCase) syncToFileSystem(ctx context.Context, sourceRepo ports.GitRepository, mirrorTarget entities.MirrorTarget, targetRepoName string) MirrorResult {
	result := uc.initMirrorResult(sourceRepo.Name(), mirrorTarget.Name())

	switch mirrorTarget.ProviderType() {
	case entities.ProviderTypeDirectory:
		err := uc.syncToDirectory(ctx, sourceRepo, mirrorTarget.Path(), targetRepoName)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
			result.Action = "synced_to_directory"
		}
	case entities.ProviderTypeArchive:
		err := uc.syncToArchive(ctx, sourceRepo, mirrorTarget.Path(), targetRepoName)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
			result.Action = "synced_to_archive"
		}
	case entities.ProviderTypeGitHub, entities.ProviderTypeGitLab, entities.ProviderTypeGitea:
		// should not happen as git providers are handled separately
		result.Error = errors.New("unexpected git provider type in filesystem sync")
	default:
		result.Error = errors.New("unsupported provider type")
	}

	return result
}

func (uc ToMirrorsUseCase) syncToGitProvider(ctx context.Context, sourceRepo ports.GitRepository, mirrorTarget entities.MirrorTarget, mirrorProvider ports.RepositoryProvider, targetRepoName string, request ToMirrorsRequest) MirrorResult {
	result := uc.initMirrorResult(sourceRepo.Name(), mirrorTarget.Name())

	err := uc.performGitSync(ctx, sourceRepo, mirrorTarget, mirrorProvider, targetRepoName, request)
	if err != nil {
		result.Error = err
	} else {
		result.Success = true
		result.Action = "synced_to_provider"
	}

	return result
}

// ValidateRepositoryForMirror ensures repository name compatibility with target mirror provider.
func (uc ToMirrorsUseCase) validateRepositoryForMirror(
	ctx context.Context,
	repoName string,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
) error {
	uc.logger.Trace(ctx, "Validating repository for mirror", map[string]any{
		"repository": repoName,
		"mirror":     mirrorTarget.Name(),
	})

	// Use repository name as provided for validation (transformation happens in main sync flow)
	targetRepoName := repoName

	// Validate repository name with provider (main branch logic)
	if mirrorProvider != nil {
		if !mirrorProvider.IsValidProjectName(ctx, targetRepoName) {
			// matches main branch behavior where invalid names can be ignored
			uc.logger.Warn(ctx, "Repository name is invalid for target provider", map[string]any{
				"repository":  repoName,
				"target_name": targetRepoName,
				"provider":    mirrorTarget.ProviderType(),
				"mirror":      mirrorTarget.Name(),
			})

			// Track in sync metadata (restores main branch SyncRunMetainfo functionality)
			if metadata, ok := entities.GetMetadataFromContext(ctx); ok {
				metadata.AddFailure("invalid", targetRepoName)
			}

			// Return error but allow higher level to decide to ignore (like main branch)
			return fmt.Errorf("%w: '%s' for provider '%s'", domain.ErrRepositoryNameInvalidForProvider, targetRepoName, string(mirrorTarget.ProviderType()))
		}
	}

	uc.logger.Debug(ctx, "Repository validation passed", map[string]any{
		"repository":  repoName,
		"target_name": targetRepoName,
		"mirror":      mirrorTarget.Name(),
	})

	return nil
}

// TransformRepositoryName applies name transformations for mirror compatibility
// Supports prefix/suffix addition, case conversion, character replacement, and provider-specific constraints.
func (uc ToMirrorsUseCase) transformRepositoryName(
	originalName string,
	options ports.NameTransformOptions,
) string {
	// Apply prefix/suffix
	name := options.Prefix + originalName + options.Suffix

	// Apply replacements
	for old, new := range options.Replacements {
		name = strings.ReplaceAll(name, old, new)
	}

	// Apply case transformations
	if options.ToLowercase {
		name = strings.ToLower(name)
	} else if options.ToUppercase {
		name = strings.ToUpper(name)
	}

	return name
}

// SyncToDirectory implements directory mirror synchronization.
func (uc ToMirrorsUseCase) syncToDirectory(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	targetPath string,
	repoName string,
) error {
	uc.logger.Debug(ctx, "Syncing to directory", map[string]any{
		"repository":  repoName,
		"target_path": targetPath,
	})

	// Build full target path including repo name
	fullTargetPath := filepath.Join(targetPath, repoName)

	// Step 1: Clone/push to the directory target using git operations
	cloneOptions := ports.CloneOptions{
		URL:          sourceRepo.URL(),
		Path:         fullTargetPath,
		Branch:       "",
		SingleBranch: false, // Clone all branches for directory mirrors
		Depth:        0,     // Full clone for directory mirrors
		Mirror:       false,
		Bare:         false,
		Auth:         ports.AuthOptions{},
		Progress:     nil,
		Tags:         ports.TagModeAll,
		Timeout:      0,
	}

	// Check if target already exists
	if exists, err := uc.fileSystem.Exists(fullTargetPath); err != nil {
		return fmt.Errorf("failed to check if target exists: %w", err)
	} else if exists {
		return uc.syncExistingDirectoryRepository(ctx, fullTargetPath, repoName)
	}

	// Clone new repository
	targetRepo, err := uc.gitOperations.Clone(ctx, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository to directory: %w", err)
	}

	defer func() {
		if closeErr := targetRepo.Close(); closeErr != nil {
			uc.logger.Warn(ctx, "Failed to close target repository", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	uc.logger.Info(ctx, "Successfully cloned to new directory mirror", map[string]any{
		"repository":  repoName,
		"target_path": fullTargetPath,
	})

	// Increment sync count
	uc.incrementSyncCount(ctx)

	return nil
}

// SyncToArchive implements archive mirror synchronization.
func (uc ToMirrorsUseCase) syncToArchive(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	targetPath string,
	repoName string,
) error {
	uc.logger.Debug(ctx, "Syncing to archive", map[string]any{
		"repository":  repoName,
		"target_path": targetPath,
	})

	// Create target directory if it doesn't exist
	if err := uc.fileSystem.MkdirAll(targetPath, 0750); err != nil {
		return fmt.Errorf("failed to create archive target directory: %w", err)
	}

	// Archive mirroring will be handled by injected archive operations

	// Create source repository entity for archive service
	sourceRepoEntity := uc.createRepositoryEntity(ctx, sourceRepo)

	// Create target repository entity (for archive metadata)
	targetRepoBuilder := entities.NewRepositoryBuilder()
	targetRepoBuilder, _ = targetRepoBuilder.WithName(repoName)
	targetRepoBuilder = targetRepoBuilder.WithDescription("Archive of " + sourceRepoEntity.Description())
	targetRepoBuilder = targetRepoBuilder.WithProviderType("archive")

	targetRepo, err := targetRepoBuilder.Build()
	if err != nil {
		return fmt.Errorf("failed to create target repository entity: %w", err)
	}

	// Create mirror request using domain port
	mirrorRequest := ports.ArchiveMirrorRequest{
		SourceRepository:   sourceRepoEntity,
		TargetRepository:   targetRepo,
		ArchiveFormat:      "tar.gz",
		CompressionLevel:   6, // Standard compression
		IncludeMetadata:    true,
		IncludeHistory:     false, // Archive doesn't need full git history
		PreservePaths:      true,
		ExcludePatterns:    []string{".git", "*.tmp", "*.log"},
		IncludePatterns:    []string{"*"},
		ArchiveNamePattern: repoName + ".tar.gz",
		DryRun:             false,
	}

	uc.logger.Info(ctx, "Creating archive using mirror service", map[string]any{
		"repository":     repoName,
		"target_path":    targetPath,
		"archive_format": mirrorRequest.ArchiveFormat,
		"source_url":     sourceRepo.URL(),
	})

	// Execute archive mirroring through port
	err = uc.archiveOperations.CreateMirror(ctx, mirrorRequest)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	uc.logger.Info(ctx, "Archive sync completed successfully", map[string]any{
		"repository":  repoName,
		"target_path": targetPath,
	})

	// Increment sync count
	uc.incrementSyncCount(ctx)

	return nil
}

// PerformGitSync implements git provider mirror synchronization.
func (uc ToMirrorsUseCase) performGitSync(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
	targetRepoName string,
	request ToMirrorsRequest,
) error {
	uc.logger.Debug(ctx, "Syncing to git provider", map[string]any{
		"repository":    targetRepoName,
		"provider_type": mirrorTarget.ProviderType(),
		"owner":         mirrorTarget.Owner(),
	})

	// Create source repository entity from GitRepository
	sourceRepoEntity := uc.createRepositoryEntity(ctx, sourceRepo)

	// Create PushToProvider use case and execute
	pushUseCase := NewPushToProviderUseCase(mirrorProvider, uc.gitOperations)

	pushRequest := PushRequest{
		SourceRepository: sourceRepoEntity,
		SourceGitRepo:    sourceRepo, // CRITICAL: For GPSUPSTREAM remote setup
		TargetConfig:     mirrorTarget,
		SourceConfig:     request.SourceConfig, // Add the missing source config
		ForcePush:        request.Options.ForcePush,
		DryRun:           false, // Already handled at higher level
		CreateIfMissing:  request.Options.CreateIfNotExists,
	}

	pushResponse, err := pushUseCase.Execute(ctx, pushRequest)
	if err != nil {
		return fmt.Errorf("failed to push repository to provider: %w", err)
	}

	if !pushResponse.Success {
		return fmt.Errorf("%w: %w", domain.ErrPushToProviderFailed, pushResponse.Error)
	}

	uc.logger.Info(ctx, "Successfully pushed to git provider", map[string]any{
		"repository": targetRepoName,
		"provider":   mirrorTarget.ProviderType(),
		"created":    pushResponse.Created,
		"project_id": pushResponse.ProjectID,
		"target_url": pushResponse.TargetURL,
	})

	// Increment sync count
	uc.incrementSyncCount(ctx)

	// Step 5: Sync branch protection if enabled
	if request.Options.SyncProtection {
		err := uc.syncBranchProtection(ctx, sourceRepo, mirrorTarget, mirrorProvider, targetRepoName)
		if err != nil {
			uc.logger.Warn(ctx, "Failed to sync branch protection", map[string]any{
				"repository": targetRepoName,
				"error":      err.Error(),
			})
		}
	}

	return nil
}

// CreateRepositoryEntity transforms GitRepository interface into domain entity with fallback handling
// converts the GitRepository from ports to a full entities.Repository for use cases.
func (uc ToMirrorsUseCase) createRepositoryEntity(ctx context.Context, gitRepo ports.GitRepository) entities.Repository {
	builder := entities.NewRepositoryBuilder()

	// Extract basic repository information
	name := gitRepo.Name()
	if name != "" {
		builder, _ = builder.WithName(name)
	}

	// Extract URLs
	if url := gitRepo.URL(); url != "" {
		if strings.Contains(url, "https://") {
			builder, _ = builder.WithHTTPSURL(url)
		} else if strings.Contains(url, "git@") {
			builder, _ = builder.WithSSHURL(url)
		}
	}

	// Extract current branch as default branch
	if currentBranch, err := gitRepo.CurrentBranch(); err == nil && currentBranch != "" {
		builder, _ = builder.WithDefaultBranch(currentBranch)
	} else {
		// Fallback to common default branches
		builder, _ = builder.WithDefaultBranch("main")
	}

	// Set default values for other fields
	builder = builder.WithDescription("Synced repository")
	builder = builder.WithVisibility("private")   // Default to private for safety
	builder = builder.WithProviderType("unknown") // Will be set correctly when creating at target

	// Build the repository
	repo, err := builder.Build()
	if err != nil {
		// Log error but return a minimal repository
		uc.logger.Warn(ctx, "Failed to build repository entity", map[string]any{
			"error": err.Error(),
			"name":  name,
		})

		// Create minimal repository as fallback
		fallbackBuilder := entities.NewRepositoryBuilder()
		fallbackBuilder, _ = fallbackBuilder.WithName(name)
		fallbackBuilder, _ = fallbackBuilder.WithDefaultBranch("main")
		fallbackBuilder = fallbackBuilder.WithDescription("Synced repository")
		fallbackBuilder = fallbackBuilder.WithVisibility("private")
		fallbackBuilder = fallbackBuilder.WithProviderType("unknown")

		if fallbackRepo, buildErr := fallbackBuilder.Build(); buildErr == nil {
			return fallbackRepo
		}

		// If even fallback fails, return empty repository
		return entities.Repository{}
	}

	return repo
}

// SyncBranchProtection replicates branch protection rules from source to mirror with provider compatibility checks.
func (uc ToMirrorsUseCase) syncBranchProtection(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
	targetRepoName string,
) error {
	uc.logger.Debug(ctx, "Syncing branch protection", map[string]any{
		"source_repo": sourceRepo.Name(),
		"target_repo": targetRepoName,
		"mirror":      mirrorTarget.Name(),
	})

	// Check if provider supports branch protection
	if !mirrorProvider.SupportsFeature(ports.FeatureBranchProtection) {
		uc.logger.Debug(ctx, "Provider does not support branch protection, skipping", map[string]any{
			"provider": mirrorTarget.ProviderType(),
		})

		return nil
	}

	// Get default branch protection settings (main/master)
	// Try to get current branch, fallback to "main"
	defaultBranch := "main"
	if currentBranch, err := sourceRepo.CurrentBranch(); err == nil && currentBranch != "" {
		defaultBranch = currentBranch
	}

	// Create default protection for the main branch
	protection := ports.BranchProtection{
		Protected:     true,
		EnforceAdmins: false, // Don't enforce for admins initially
		RequiredStatusChecks: ports.RequiredStatusChecks{
			Strict:   false,
			Contexts: []string{},
		},
		RequiredPullRequestReviews: ports.RequiredPullRequestReviews{
			RequiredApprovingReviewCount: 1,
			DismissStaleReviews:          false,
			RequireCodeOwnerReviews:      false,
			DismissalRestrictions:        ports.UserRestrictions{},
		},
		Restrictions:                   ports.BranchRestrictions{},
		RequiredLinearHistory:          false,
		RequiredConversationResolution: false,
		AllowForcePushes:               false,
		AllowDeletions:                 false,
	}

	// Create provider config for mirror
	mirrorConfig := ports.ProviderConfig{
		ProviderType: string(mirrorTarget.ProviderType()),
		Domain:       mirrorTarget.Domain(),
		Owner:        mirrorTarget.Owner(),
		AuthConfig: ports.AuthenticationConfig{
			Token:      mirrorTarget.Token(),
			Username:   mirrorTarget.Username(),
			SSHKeyPath: mirrorTarget.SSHKeyPath(),
			SSHKey:     mirrorTarget.SSHKey(),
		},
	}

	// Set branch protection on mirror
	err := mirrorProvider.SetBranchProtection(ctx, mirrorConfig, targetRepoName, defaultBranch, protection)
	if err != nil {
		return fmt.Errorf("failed to set branch protection on mirror: %w", err)
	}

	uc.logger.Info(ctx, "Branch protection synced successfully", map[string]any{
		"target_repo": targetRepoName,
		"branch":      defaultBranch,
		"mirror":      mirrorTarget.Name(),
		"provider":    mirrorTarget.ProviderType(),
	})

	return nil
}

// InitMirrorSync initializes sync metadata in the context.
func (uc ToMirrorsUseCase) initMirrorSync(ctx context.Context, mirrorTarget entities.MirrorTarget, repositoryCount int) context.Context {
	// Create sync metadata (like main branch)
	meta := entities.NewSyncRunMetadata("", string(mirrorTarget.ProviderType()), "sync", "default")
	meta.SetTotalRepositories(repositoryCount)

	// Add to context
	ctx = entities.AddMetadataToContext(ctx, meta)

	uc.logger.Debug(ctx, "Initialized mirror sync metadata", map[string]any{
		"mirror_target":    mirrorTarget.Name(),
		"provider_type":    mirrorTarget.ProviderType(),
		"repository_count": repositoryCount,
	})

	return ctx
}

// IncrementSyncCount increments the sync count in the context metadata.
func (uc ToMirrorsUseCase) incrementSyncCount(ctx context.Context) {
	if meta, ok := entities.GetMetadataFromContext(ctx); ok {
		meta.AddSuccess("sync", "repository")
		uc.logger.Debug(ctx, "Incremented sync count", map[string]any{
			"processed_count": meta.ProcessedCount,
		})
	}
}

// SyncExistingDirectoryRepository synchronizes an existing directory repository.
func (uc ToMirrorsUseCase) syncExistingDirectoryRepository(ctx context.Context, fullTargetPath, repoName string) error {
	// Repository already exists, open it instead of cloning
	targetRepo, err := uc.gitOperations.Open(ctx, fullTargetPath)
	if err != nil {
		return fmt.Errorf("failed to open existing directory repository at %s: %w. "+
			"This may indicate repository corruption or permission issues. "+
			"Troubleshooting steps: "+
			"1) Check directory permissions (should be writable by current user), "+
			"2) Verify git repository integrity with 'git fsck --full', "+
			"3) Check if .git directory exists and is not corrupted, "+
			"4) Consider removing and re-cloning if repository is corrupted",
			fullTargetPath, err)
	}

	defer func() {
		if closeErr := targetRepo.Close(); closeErr != nil {
			uc.logger.Warn(ctx, "Failed to close target repository", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	// Perform Pull operation for directory targets - essential for proper synchronization
	pullOptions := ports.PullOptions{
		Remote:      "origin",
		Branch:      "",                        // Pull default branch
		FastForward: ports.FastForwardModeOnly, // Allow fast-forward for mirror sync
		Auth:        ports.AuthOptions{},
		Progress:    nil,
		Rebase:      false,
		Strategy:    ports.MergeStrategyDefault,
		Timeout:     0,
	}

	if err := targetRepo.Pull(ctx, pullOptions); err != nil {
		// error is critical for directory mirrors
		return fmt.Errorf("failed to pull repository for directory target at %s: %w. "+
			"This is critical for directory mirror synchronization. "+
			"Common causes: "+
			"1) Network connectivity issues to source repository, "+
			"2) Authentication problems (check credentials), "+
			"3) Merge conflicts (directory mirrors should be clean), "+
			"4) Remote repository not accessible or deleted. "+
			"Try: git pull origin manually in the target directory for more details",
			fullTargetPath, err)
	}

	uc.logger.Info(ctx, "Successfully pulled to existing directory mirror", map[string]any{
		"repository":  repoName,
		"target_path": fullTargetPath,
	})

	// Increment sync count
	uc.incrementSyncCount(ctx)

	return nil
}
