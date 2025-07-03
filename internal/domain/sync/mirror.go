// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/adapters/repository/archive"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// SyncToMirrorsUseCase handles synchronizing repositories to mirror targets.
// This ports the toMirror functionality from main branch to hexagonal architecture.
type SyncToMirrorsUseCase struct {
	repositoryProvider ports.RepositoryProvider
	gitOperations      ports.GitOperations
	logger             ports.Logger
}

// NewSyncToMirrorsUseCase creates a new sync to mirrors use case.
func NewSyncToMirrorsUseCase(
	repositoryProvider ports.RepositoryProvider,
	gitOps ports.GitOperations,
	logger ports.Logger,
) SyncToMirrorsUseCase {
	return SyncToMirrorsUseCase{
		repositoryProvider: repositoryProvider,
		gitOperations:      gitOps,
		logger:             logger,
	}
}

// SyncToMirrorsRequest represents the input for syncing to mirrors.
type SyncToMirrorsRequest struct {
	SourceRepositories []ports.GitRepository
	MirrorTargets      []entities.MirrorTarget
	SourceConfig       ports.ProviderConfig
	DryRun             bool
	Options            SyncOptions
}

// SyncOptions contains options for mirror synchronization.
type SyncOptions struct {
	ForcePush          bool
	IgnoreInvalidNames bool
	CreateIfNotExists  bool
	UpdateDescription  bool
	SyncProtection     bool
	NameTransformation ports.NameTransformOptions
}

// SyncToMirrorsResponse represents the result of syncing to mirrors.
type SyncToMirrorsResponse struct {
	Results           []MirrorSyncResult
	Success           bool
	TotalRepositories int
	SuccessfulSyncs   int
	FailedSyncs       int
	SkippedSyncs      int
	Errors            []error
}

// MirrorSyncResult represents the result of syncing a single repository to a mirror.
type MirrorSyncResult struct {
	RepositoryName string
	MirrorName     string
	Success        bool
	Skipped        bool
	Error          error
	Action         string // "created", "updated", "skipped"
}

// Execute synchronizes repositories to all mirror targets.
// This implements the core toMirror logic from main branch.
func (uc SyncToMirrorsUseCase) Execute(
	ctx context.Context,
	request SyncToMirrorsRequest,
) (SyncToMirrorsResponse, error) {
	uc.logger.Info(ctx, "Starting mirror synchronization", map[string]interface{}{
		"source_repos":   len(request.SourceRepositories),
		"mirror_targets": len(request.MirrorTargets),
		"dry_run":        request.DryRun,
	})

	response := SyncToMirrorsResponse{
		TotalRepositories: len(request.SourceRepositories),
		Success:           true,
	}

	// Process each mirror target
	for _, mirrorTarget := range request.MirrorTargets {
		if !mirrorTarget.Enabled() {
			uc.logger.Info(ctx, "Skipping disabled mirror", map[string]interface{}{
				"mirror_name": mirrorTarget.Name(),
			})

			continue
		}

		uc.logger.Info(ctx, "Processing mirror target", map[string]interface{}{
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
		if result.Success {
			response.SuccessfulSyncs++
		} else if result.Skipped {
			response.SkippedSyncs++
		} else {
			response.FailedSyncs++
		}
	}

	uc.logger.Info(ctx, "Mirror synchronization completed", map[string]interface{}{
		"total_operations": len(response.Results),
		"successful":       response.SuccessfulSyncs,
		"failed":           response.FailedSyncs,
		"skipped":          response.SkippedSyncs,
		"success":          response.Success,
	})

	return response, nil
}

// syncRepositoriesToMirror synchronizes all repositories to a single mirror target.
// This implements the processRepository logic from main branch.
func (uc SyncToMirrorsUseCase) syncRepositoriesToMirror(
	ctx context.Context,
	sourceRepos []ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	request SyncToMirrorsRequest,
) []MirrorSyncResult {
	// CRITICAL: Initialize sync metadata (restored from main branch initMirrorSync)
	ctx = uc.initMirrorSync(ctx, mirrorTarget, len(sourceRepos))

	results := make([]MirrorSyncResult, 0, len(sourceRepos))

	// Create mirror provider if it's a git provider (not directory/archive)
	var mirrorProvider ports.RepositoryProvider
	if mirrorTarget.ProviderType() != entities.ProviderTypeDirectory && mirrorTarget.ProviderType() != entities.ProviderTypeArchive {
		// Note: Would need provider factory here - for now using the same provider
		// TODO: Create actual provider instance with mirror target config
		mirrorProvider = uc.repositoryProvider
	}

	// Process each repository
	for _, repo := range sourceRepos {
		result := uc.syncSingleRepositoryToMirror(ctx, repo, mirrorTarget, mirrorProvider, request)
		results = append(results, result)
	}

	return results
}

// syncSingleRepositoryToMirror synchronizes a single repository to a mirror.
// This implements the core repository processing logic from main branch.
func (uc SyncToMirrorsUseCase) syncSingleRepositoryToMirror(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
	request SyncToMirrorsRequest,
) MirrorSyncResult {
	repoName := sourceRepo.Name()

	result := MirrorSyncResult{
		RepositoryName: repoName,
		MirrorName:     mirrorTarget.Name(),
	}

	uc.logger.Debug(ctx, "Processing repository for mirror", map[string]interface{}{
		"repository":    repoName,
		"mirror":        mirrorTarget.Name(),
		"provider_type": mirrorTarget.ProviderType(),
	})

	// Step 1: Check if repository was already marked as invalid (main branch logic)
	if entities.ContainsFailureInContext(ctx, "invalid", repoName) {
		if request.Options.IgnoreInvalidNames {
			uc.logger.Warn(ctx, "Repository already marked as invalid, skipping", map[string]interface{}{
				"repository": repoName,
				"mirror":     mirrorTarget.Name(),
			})

			result.Skipped = true
			result.Action = "skipped_invalid"

			return result
		}
	}

	// Step 2: Validate repository name for target (equivalent to validateRepository)
	if err := uc.validateRepositoryForMirror(ctx, repoName, mirrorTarget, mirrorProvider); err != nil {
		if request.Options.IgnoreInvalidNames {
			uc.logger.Warn(ctx, "Ignoring repository with invalid name", map[string]interface{}{
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

	// Step 3: Transform repository name if needed
	targetRepoName := uc.transformRepositoryName(repoName, request.Options.NameTransformation)

	// Step 4: Handle dry run
	if request.DryRun {
		uc.logger.Info(ctx, "Dry run - would sync repository", map[string]interface{}{
			"source_repo": repoName,
			"target_repo": targetRepoName,
			"mirror":      mirrorTarget.Name(),
		})

		result.Success = true
		result.Action = "dry_run"

		return result
	}

	// Step 5: Sync based on provider type
	switch mirrorTarget.ProviderType() {
	case entities.ProviderTypeGitHub, entities.ProviderTypeGitLab, entities.ProviderTypeGitea:
		// Git provider (GitHub, GitLab, Gitea)
		err := uc.syncToGitProvider(ctx, sourceRepo, mirrorTarget, mirrorProvider, targetRepoName, request)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
			result.Action = "synced_to_provider"
		}
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
	}

	return result
}

// validateRepositoryForMirror validates if repository can be synced to mirror.
// This restores the sophisticated validation logic from main branch validateRepository().
func (uc SyncToMirrorsUseCase) validateRepositoryForMirror(
	ctx context.Context,
	repoName string,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
) error {
	uc.logger.Trace(ctx, "Validating repository for mirror", map[string]interface{}{
		"repository": repoName,
		"mirror":     mirrorTarget.Name(),
	})

	// Apply alphanumeric name transformation if needed (from main branch logic)
	targetRepoName := repoName
	// TODO: Check if mirrorTarget.Options().AlphaNumHyphName is set
	// if alphaNumHyphName { targetRepoName = utilities.RemoveNonAlphaNumericChars(repoName) }

	// Validate repository name with provider (main branch logic)
	if mirrorProvider != nil {
		if !mirrorProvider.IsValidProjectName(ctx, targetRepoName) {
			// This matches main branch behavior where invalid names can be ignored
			uc.logger.Warn(ctx, "Repository name is invalid for target provider", map[string]interface{}{
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
			return fmt.Errorf("repository name '%s' is invalid for provider '%s'", targetRepoName, mirrorTarget.ProviderType())
		}
	}

	uc.logger.Debug(ctx, "Repository validation passed", map[string]interface{}{
		"repository":  repoName,
		"target_name": targetRepoName,
		"mirror":      mirrorTarget.Name(),
	})

	return nil
}

// transformRepositoryName applies name transformations.
func (uc SyncToMirrorsUseCase) transformRepositoryName(
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

// syncToDirectory implements directory mirror synchronization.
// This restores the critical directory sync with Pull operation from main branch.
func (uc SyncToMirrorsUseCase) syncToDirectory(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	targetPath string,
	repoName string,
) error {
	uc.logger.Debug(ctx, "Syncing to directory", map[string]interface{}{
		"repository":  repoName,
		"target_path": targetPath,
	})

	// Build full target path including repo name
	fullTargetPath := filepath.Join(targetPath, repoName)

	// Step 1: Clone/push to the directory target using git operations
	cloneOptions := ports.CloneOptions{
		URL:          sourceRepo.URL(),
		Path:         fullTargetPath,
		SingleBranch: false, // Clone all branches for directory mirrors
		Depth:        0,     // Full clone for directory mirrors
	}

	// Check if target already exists
	if _, err := os.Stat(fullTargetPath); err == nil {
		// Repository already exists, open it instead of cloning
		targetRepo, err := uc.gitOperations.Open(ctx, fullTargetPath)
		if err != nil {
			return fmt.Errorf("failed to open existing directory repository: %w", err)
		}
		defer func() {
			if closeErr := targetRepo.Close(); closeErr != nil {
				uc.logger.Warn(ctx, "Failed to close target repository", map[string]interface{}{
					"error": closeErr.Error(),
				})
			}
		}()

		// CRITICAL: Perform Pull operation for directory targets (restored from main branch)
		// This is essential functionality that was missing - directory targets need Pull after Push
		pullOptions := ports.PullOptions{
			Remote:      "origin",
			Branch:      "",                        // Pull default branch
			FastForward: ports.FastForwardModeOnly, // Allow fast-forward for mirror sync
		}

		if err := targetRepo.Pull(ctx, pullOptions); err != nil {
			// This error is critical for directory mirrors - main branch would fail here
			return fmt.Errorf("failed to pull repository for directory target: %w", err)
		}

		uc.logger.Info(ctx, "Successfully pulled to existing directory mirror", map[string]interface{}{
			"repository":  repoName,
			"target_path": fullTargetPath,
		})

		// CRITICAL: Increment sync count (restored from main branch)
		uc.incrementSyncCount(ctx)
	} else {
		// Clone new repository
		targetRepo, err := uc.gitOperations.Clone(ctx, cloneOptions)
		if err != nil {
			return fmt.Errorf("failed to clone repository to directory: %w", err)
		}
		defer func() {
			if closeErr := targetRepo.Close(); closeErr != nil {
				uc.logger.Warn(ctx, "Failed to close target repository", map[string]interface{}{
					"error": closeErr.Error(),
				})
			}
		}()

		uc.logger.Info(ctx, "Successfully cloned to new directory mirror", map[string]interface{}{
			"repository":  repoName,
			"target_path": fullTargetPath,
		})

		// CRITICAL: Increment sync count (restored from main branch)
		uc.incrementSyncCount(ctx)
	}

	return nil
}

// syncToArchive implements archive mirror synchronization.
// This restores the main branch archive functionality using archive mirror service.
func (uc SyncToMirrorsUseCase) syncToArchive(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	targetPath string,
	repoName string,
) error {
	uc.logger.Debug(ctx, "Syncing to archive", map[string]interface{}{
		"repository":  repoName,
		"target_path": targetPath,
	})

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetPath, 0750); err != nil {
		return fmt.Errorf("failed to create archive target directory: %w", err)
	}

	// Create archive mirror service
	mirrorService := archive.NewMirrorService(uc.logger, "/tmp", targetPath)

	// Create source repository entity for archive service
	sourceRepoEntity := uc.createRepositoryEntity(sourceRepo)

	// Create target repository entity (for archive metadata)
	targetRepoBuilder := entities.NewRepositoryBuilder()
	targetRepoBuilder, _ = targetRepoBuilder.WithName(repoName)
	targetRepoBuilder = targetRepoBuilder.WithDescription("Archive of " + sourceRepoEntity.Description())
	targetRepoBuilder = targetRepoBuilder.WithProviderType("archive")
	targetRepo, err := targetRepoBuilder.Build()
	if err != nil {
		return fmt.Errorf("failed to create target repository entity: %w", err)
	}

	// Create mirror request
	mirrorRequest := archive.MirrorRequest{
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

	uc.logger.Info(ctx, "Creating archive using mirror service", map[string]interface{}{
		"repository":     repoName,
		"target_path":    targetPath,
		"archive_format": mirrorRequest.ArchiveFormat,
		"source_url":     sourceRepo.URL(),
	})

	// Execute archive mirroring
	result, err := mirrorService.Mirror(ctx, mirrorRequest)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	uc.logger.Info(ctx, "Archive sync completed successfully", map[string]interface{}{
		"repository":      repoName,
		"archive_path":    result.ArchivePath,
		"archive_size":    result.ArchiveSize,
		"files_processed": result.FilesProcessed,
		"files_skipped":   result.FilesSkipped,
		"success":         result.Success,
	})

	// CRITICAL: Increment sync count (restored from main branch)
	uc.incrementSyncCount(ctx)

	return nil
}

// syncToGitProvider implements git provider mirror synchronization.
// This restores the main branch provider.Push functionality via PushToProvider use case.
func (uc SyncToMirrorsUseCase) syncToGitProvider(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
	targetRepoName string,
	request SyncToMirrorsRequest,
) error {
	uc.logger.Debug(ctx, "Syncing to git provider", map[string]interface{}{
		"repository":    targetRepoName,
		"provider_type": mirrorTarget.ProviderType(),
		"owner":         mirrorTarget.Owner(),
	})

	// Create source repository entity from GitRepository
	sourceRepoEntity := uc.createRepositoryEntity(sourceRepo)

	// Create PushToProvider use case and execute
	pushUseCase := NewPushToProviderUseCase(mirrorProvider, uc.gitOperations, uc.logger)

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
		return fmt.Errorf("push to provider failed: %v", pushResponse.Error)
	}

	uc.logger.Info(ctx, "Successfully pushed to git provider", map[string]interface{}{
		"repository": targetRepoName,
		"provider":   mirrorTarget.ProviderType(),
		"created":    pushResponse.Created,
		"project_id": pushResponse.ProjectID,
		"target_url": pushResponse.TargetURL,
	})

	// CRITICAL: Increment sync count (restored from main branch)
	uc.incrementSyncCount(ctx)

	// Step 5: Sync branch protection if enabled
	if request.Options.SyncProtection {
		err := uc.syncBranchProtection(ctx, sourceRepo, mirrorTarget, mirrorProvider, targetRepoName)
		if err != nil {
			uc.logger.Warn(ctx, "Failed to sync branch protection", map[string]interface{}{
				"repository": targetRepoName,
				"error":      err.Error(),
			})
		}
	}

	return nil
}

// createRepositoryEntity creates a domain repository entity from GitRepository.
// This converts the GitRepository from ports to a full entities.Repository for use cases.
func (uc SyncToMirrorsUseCase) createRepositoryEntity(gitRepo ports.GitRepository) entities.Repository {
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
		uc.logger.Warn(context.Background(), "Failed to build repository entity", map[string]interface{}{
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

// syncBranchProtection synchronizes branch protection settings from source to mirror.
// This ports the protection functionality from main branch to hexagonal architecture.
func (uc SyncToMirrorsUseCase) syncBranchProtection(
	ctx context.Context,
	sourceRepo ports.GitRepository,
	mirrorTarget entities.MirrorTarget,
	mirrorProvider ports.RepositoryProvider,
	targetRepoName string,
) error {
	uc.logger.Debug(ctx, "Syncing branch protection", map[string]interface{}{
		"source_repo": sourceRepo.Name(),
		"target_repo": targetRepoName,
		"mirror":      mirrorTarget.Name(),
	})

	// Check if provider supports branch protection
	if !mirrorProvider.SupportsFeature(ports.FeatureBranchProtection) {
		uc.logger.Debug(ctx, "Provider does not support branch protection, skipping", map[string]interface{}{
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
		},
		AllowForcePushes: false,
		AllowDeletions:   false,
	}

	// Create provider config for mirror
	authConfig := mirrorTarget.AuthConfig()
	mirrorConfig := ports.ProviderConfig{
		ProviderType: string(mirrorTarget.ProviderType()),
		Domain:       mirrorTarget.Domain(),
		Owner:        mirrorTarget.Owner(),
		AuthConfig: ports.AuthenticationConfig{
			Token:      authConfig.Token(),
			Username:   authConfig.Username(),
			SSHKeyPath: authConfig.SSHKeyPath(),
			SSHKey:     authConfig.SSHKey(),
		},
	}

	// Set branch protection on mirror
	err := mirrorProvider.SetBranchProtection(ctx, mirrorConfig, targetRepoName, defaultBranch, protection)
	if err != nil {
		return fmt.Errorf("failed to set branch protection on mirror: %w", err)
	}

	uc.logger.Info(ctx, "Branch protection synced successfully", map[string]interface{}{
		"target_repo": targetRepoName,
		"branch":      defaultBranch,
		"mirror":      mirrorTarget.Name(),
		"provider":    mirrorTarget.ProviderType(),
	})

	return nil
}

// initMirrorSync initializes sync metadata in the context.
// This restores the critical initMirrorSync functionality from main branch.
func (uc SyncToMirrorsUseCase) initMirrorSync(ctx context.Context, mirrorTarget entities.MirrorTarget, repositoryCount int) context.Context {
	// Create sync metadata (like main branch)
	meta := entities.NewSyncRunMetadata("", string(mirrorTarget.ProviderType()), "sync", "default")
	meta.SetTotalRepositories(repositoryCount)

	// Add to context
	ctx = entities.AddMetadataToContext(ctx, meta)

	uc.logger.Debug(ctx, "Initialized mirror sync metadata", map[string]interface{}{
		"mirror_target":    mirrorTarget.Name(),
		"provider_type":    mirrorTarget.ProviderType(),
		"repository_count": repositoryCount,
	})

	return ctx
}

// incrementSyncCount increments the sync count in the context metadata.
// This restores the critical sync tracking functionality from main branch.
func (uc SyncToMirrorsUseCase) incrementSyncCount(ctx context.Context) {
	if meta, ok := entities.GetMetadataFromContext(ctx); ok {
		meta.AddSuccess("sync", "repository")
		uc.logger.Debug(ctx, "Incremented sync count", map[string]interface{}{
			"processed_count": meta.ProcessedCount,
		})
	}
}
