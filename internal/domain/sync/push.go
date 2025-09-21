// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/shared"
)

const (
	defaultBranchMain = "main"
	remoteNameOrigin  = "origin"
)

// Error variables for common failure scenarios.
var (
	ErrTargetRepositoryName = errors.New("failed target repository name validation")
	ErrCreateRepository     = errors.New("failed to create repository")
	ErrPushChanges          = errors.New("failed to push changes")
	ErrDefaultBranch        = errors.New("failed to set default branch")
)

// PushToProviderUseCase handles pushing repositories to git providers (GitHub, GitLab, Gitea)
// Creates target repository if needed, temporarily disables branch protection during push,
// Then restores protection settings.
type PushToProviderUseCase struct {
	provider ports.RepositoryProvider
	gitOps   ports.GitOperations
}

// NewPushToProviderUseCase creates a new push to provider use case.
func NewPushToProviderUseCase(provider ports.RepositoryProvider, gitOps ports.GitOperations) *PushToProviderUseCase {
	return &PushToProviderUseCase{
		provider: provider,
		gitOps:   gitOps,
	}
}

// PushRequest contains parameters for push operation.
type PushRequest struct {
	SourceRepository entities.Repository
	SourceGitRepo    ports.GitRepository // CRITICAL: For GPSUPSTREAM remote setup
	TargetConfig     entities.MirrorTarget
	SourceConfig     ports.ProviderConfig
	ForcePush        bool
	DryRun           bool
	CreateIfMissing  bool
}

// PushResponse contains the result of push operation.
type PushResponse struct {
	Success   bool
	Created   bool
	ProjectID string
	TargetURL string
	Error     error
}

// Execute performs the push to provider operation.
func (uc *PushToProviderUseCase) Execute(ctx context.Context, request PushRequest) (PushResponse, error) {
	logger := ports.LoggerFromContext(ctx)

	var response PushResponse

	var err error

	// TRACE: Use case entry point (hexagonal boundary)
	logger.Trace(ctx, "PushToProviderUseCase.Execute entry", map[string]any{
		"source":     request.SourceRepository.Name(),
		"target":     request.TargetConfig.Name(),
		"force_push": request.ForcePush,
		"dry_run":    request.DryRun,
	})

	defer func() {
		// TRACE: Use case exit point with outcome
		logger.Trace(ctx, "PushToProviderUseCase.Execute exit", map[string]any{
			"success":    response.Success,
			"created":    response.Created,
			"project_id": response.ProjectID,
			"error":      err != nil,
		})
	}()

	if request.DryRun {
		// DEBUG: Branch decision
		logger.Debug(ctx, "Dry run mode enabled", map[string]any{
			"skipping_actual_push": true,
		})

		return uc.performDryRun(ctx, request), nil
	}

	return uc.executePush(ctx, request)
}

// ExecutePush performs the actual push operation with early returns.
func (uc *PushToProviderUseCase) executePush(ctx context.Context, request PushRequest) (PushResponse, error) {
	logger := ports.LoggerFromContext(ctx)

	// TRACE: Internal method entry
	logger.Trace(ctx, "executePush entry", map[string]any{
		"source": request.SourceRepository.Name(),
		"target": request.TargetConfig.Name(),
	})

	// TRACE: Step 1 - Setup remote
	logger.Trace(ctx, "setting up GPSUPSTREAM remote", map[string]any{
		"step": "1_setup_remote",
	})

	if err := uc.setupGPSUpstreamRemote(ctx, request.SourceGitRepo); err != nil {
		return uc.failResponse(err), fmt.Errorf("setup remote: %w", err)
	}

	// TRACE: Step 2 - Ensure repository exists
	logger.Trace(ctx, "ensuring repository exists", map[string]any{
		"step": "2_ensure_repo",
	})

	exists, projectID, err := uc.createRepositoryIfNeeded(ctx, request)
	if err != nil {
		return uc.failResponse(err), fmt.Errorf("ensure repo: %w", err)
	}
	// DEBUG: Repository state
	logger.Debug(ctx, "Repository existence check completed", map[string]any{
		"exists":     exists,
		"project_id": projectID,
	})

	// TRACE: Step 3 - Disable protection if needed
	if request.TargetConfig.DisableProtection() {
		logger.Trace(ctx, "disabling branch protection", map[string]any{
			"step": "3_disable_protection",
		})

		if err := uc.disableProtection(ctx, request, projectID); err != nil {
			return uc.failResponse(err), fmt.Errorf("disable protection: %w", err)
		}
	}

	// TRACE: Step 4 - Perform push (critical step)
	logger.Trace(ctx, "performing git push", map[string]any{
		"step": "4_git_push",
	})

	targetURL, err := uc.performPush(ctx, request)
	if err != nil {
		return uc.failResponse(err), fmt.Errorf("push: %w", err)
	}

	// TRACE: Step 5 - Set default branch
	logger.Trace(ctx, "setting default branch", map[string]any{
		"step": "5_default_branch",
	})

	if err := uc.setDefaultBranch(ctx, request, projectID); err != nil {
		return uc.failResponse(err), fmt.Errorf("set default branch: %w", err)
	}

	// TRACE: Step 6 - Re-enable protection if needed
	if request.TargetConfig.DisableProtection() {
		logger.Trace(ctx, "re-enabling branch protection", map[string]any{
			"step": "6_enable_protection",
		})

		if err := uc.enableProtection(ctx, request, projectID); err != nil {
			return uc.failResponse(err), fmt.Errorf("enable protection: %w", err)
		}
	}

	// DEBUG: Final operation state
	logger.Debug(ctx, "Push operation completed successfully", map[string]any{
		"target_url": targetURL,
		"created":    !exists,
		"project_id": projectID,
	})

	return PushResponse{
		Success:   true,
		Created:   !exists,
		ProjectID: projectID,
		TargetURL: targetURL,
	}, nil
}

// FailResponse creates a standardized failure response.
func (uc *PushToProviderUseCase) failResponse(err error) PushResponse {
	return PushResponse{Success: false, Error: err}
}

// SetupGPSUpstreamRemote sets up the GPSUPSTREAM remote from origin
//
//nolint:cyclop // Complex remote setup logic with multiple validation steps
func (uc *PushToProviderUseCase) setupGPSUpstreamRemote(ctx context.Context, gitRepo ports.GitRepository) error {
	logger := ports.LoggerFromContext(ctx)

	// DEBUG: Remote setup details (application state)
	logger.Debug(ctx, "Setting up GPSUPSTREAM remote", map[string]any{
		"repository": gitRepo.Name(),
		"path":       gitRepo.Path(),
	})

	// CRITICAL: Implement the main branch SetGPSUpstreamRemoteFromOrigin functionality
	// 1. Get origin remote URL
	remotes, err := gitRepo.ListRemotes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list remotes: %w", err)
	}

	var originURL string

	for _, remote := range remotes {
		if remote.Name == remoteNameOrigin {
			originURL = remote.URL

			break
		}
	}

	if originURL == "" {
		return domain.ErrOriginRemoteNotFound
	}

	// DEBUG: Remote discovery result
	logger.Debug(ctx, "Found origin remote", map[string]any{
		"origin_url": originURL,
	})

	// 2. Delete existing GPSUPSTREAM remote (ignore errors like main branch)
	err = gitRepo.RemoveRemote(ctx, "GPSUPSTREAM")
	if err != nil {
		// DEBUG: Expected removal failure
		logger.Debug(ctx, "Failed to remove existing GPSUPSTREAM remote (expected)", map[string]any{
			"error": err.Error(),
		})
	}

	// 3. Create new GPSUPSTREAM remote with origin URL
	if err := gitRepo.AddRemote(ctx, "GPSUPSTREAM", originURL); err != nil {
		return fmt.Errorf("failed to create GPSUPSTREAM remote: %w", err)
	}

	// 4. Verify URLs match (like main branch)
	newRemotes, err := gitRepo.ListRemotes(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify GPSUPSTREAM remote: %w", err)
	}

	for _, remote := range newRemotes {
		if remote.Name == "GPSUPSTREAM" {
			if remote.URL != originURL {
				return fmt.Errorf("%w: got %s, expected %s", domain.ErrRemoteMismatch, remote.URL, originURL)
			}

			// DEBUG: Remote setup completion (application state)
			logger.Info(ctx, "GPSUPSTREAM remote setup completed successfully", map[string]any{
				"origin_url":      originURL,
				"gpsupstream_url": remote.URL,
				"verified":        true,
			})

			return nil
		}
	}

	return domain.ErrUpstreamRemoteNotFound
}

// CreateRepositoryIfNeeded checks if repository exists and creates it if needed.
func (uc *PushToProviderUseCase) createRepositoryIfNeeded(ctx context.Context, request PushRequest) (bool, string, error) {
	logger := ports.LoggerFromContext(ctx)

	// DEBUG: Repository existence check (application state)
	logger.Debug(ctx, "Checking if repository exists", map[string]any{
		"owner": request.TargetConfig.Owner(),
		"name":  request.SourceRepository.Name(),
	})

	// For archive/directory providers, no check needed
	if uc.isArchiveOrDirectory(string(request.TargetConfig.ProviderType())) {
		return false, "", nil
	}

	// Check if repository exists
	exists, projectID, err := uc.provider.VerifyTarget(ctx,
		request.TargetConfig.Owner(),
		request.SourceRepository.Name())
	if err != nil {
		return false, "", fmt.Errorf("failed to check repository existence: %w", err)
	}

	if exists {
		// DEBUG: Repository existence result
		logger.Debug(ctx, "Repository exists at target", map[string]any{
			"project_id": projectID,
		})

		return true, projectID, nil
	}

	// Repository doesn't exist - create it if allowed
	if !request.CreateIfMissing {
		return false, "", domain.ErrRepositoryCreationDisabled
	}

	// DEBUG: Repository creation decision
	logger.Debug(ctx, "Repository doesn't exist, creating", map[string]any{
		"name": request.SourceRepository.Name(),
	})

	projectID, err = uc.createRepository(ctx, request)
	if err != nil {
		return false, "", fmt.Errorf("failed to create repository: %w", err)
	}

	return false, projectID, nil
}

// CreateRepository creates a new repository at the target provider.
func (uc *PushToProviderUseCase) createRepository(ctx context.Context, request PushRequest) (string, error) {
	logger := ports.LoggerFromContext(ctx)

	// DEBUG: Repository creation details (application state)
	logger.Debug(ctx, "Creating repository at target provider", map[string]any{
		"name": request.SourceRepository.Name(),
	})

	description := uc.buildDescription(request)
	visibility := uc.mapVisibility(request)

	createRequest := ports.CreateRepositoryRequest{
		Name:          request.SourceRepository.Name(),
		Description:   description,
		Visibility:    visibility,
		DefaultBranch: request.SourceRepository.DefaultBranch(),
		Private:       visibility == "private",
	}

	projectID, err := uc.provider.PrepareForPush(ctx, createRequest)
	if err != nil {
		return "", fmt.Errorf("%w: %s. err: %w", ErrCreateRepository, request.SourceRepository.Name(), err)
	}

	return projectID, nil
}

// BuildDescription creates repository description.
func (uc *PushToProviderUseCase) buildDescription(request PushRequest) string {
	// DescriptionPrefix not implemented in current mirror options
	descPrefix := "" // Placeholder for request.TargetConfig.Options().DescriptionPrefix

	var description string
	if descPrefix != "" {
		description = descPrefix
	} else {
		description = "Git Provider Sync cloned this from: " + request.SourceRepository.HTTPSURL() + ": "
	}

	if request.SourceRepository.Description() != "" {
		description += request.SourceRepository.Description()
	}

	// Remove line breaks
	return strings.ReplaceAll(strings.ReplaceAll(description, "\n", " "), "\r", "")
}

// MapVisibility maps source visibility to target visibility.
func (uc *PushToProviderUseCase) mapVisibility(request PushRequest) string {
	// Visibility option not implemented in current mirror options
	visibility := "" // Placeholder for request.TargetConfig.Options().Visibility
	if visibility != "" {
		return visibility
	}

	// Default mapping logic - could be enhanced
	return request.SourceRepository.Visibility()
}

// PerformPush performs the actual git push operation.
func (uc *PushToProviderUseCase) performPush(ctx context.Context, request PushRequest) (string, error) {
	logger := ports.LoggerFromContext(ctx)

	// TRACE: Critical method entry (git adapter boundary)
	logger.Trace(ctx, "performPush entry", map[string]any{
		"force": request.ForcePush,
	})

	targetURL := uc.buildTargetURL(ctx, request)

	// TRACE: Before crossing to git adapter (hexagonal boundary)
	logger.Trace(ctx, "crossing to git adapter: UpdateRemote", map[string]any{
		"operation":  "UpdateRemote",
		"remote":     remoteNameOrigin,
		"target_url": targetURL,
	})

	// CRITICAL: Update origin remote to point to the target (GitLab) instead of source (GitHub)
	if err := request.SourceGitRepo.UpdateRemote(ctx, remoteNameOrigin, targetURL); err != nil {
		return "", fmt.Errorf("failed to update origin remote to target: %w", err)
	}

	// DEBUG: State after remote update
	logger.Debug(ctx, "Origin remote updated successfully", map[string]any{
		"remote":     remoteNameOrigin,
		"target_url": targetURL,
	})

	// TRACE: Before git push operation (critical adapter boundary)
	logger.Trace(ctx, "crossing to git adapter: Push", map[string]any{
		"operation": "Push",
		"remote":    remoteNameOrigin,
		"force":     request.ForcePush,
	})

	pushOptions := ports.PushOptions{
		Remote:  remoteNameOrigin,
		Force:   request.ForcePush,
		Auth:    uc.createAuthOptions(ctx, request.TargetConfig.AuthConfig()),
		Timeout: time.Minute * 5, // 5 minute timeout
	}

	if err := request.SourceGitRepo.Push(ctx, pushOptions); err != nil {
		return "", fmt.Errorf("failed to push repository to target: %w", err)
	}

	// DEBUG: Final operation state
	logger.Debug(ctx, "Git push completed successfully", map[string]any{
		"target_url": targetURL,
		"force":      request.ForcePush,
	})

	// TRACE: Method exit
	logger.Trace(ctx, "performPush exit", map[string]any{
		"target_url": targetURL,
	})

	return targetURL, nil
}

// CreateAuthOptions creates authentication options for git operations.
func (uc *PushToProviderUseCase) createAuthOptions(_ /* ctx */ context.Context, authConfig entities.AuthConfig) ports.AuthOptions {
	// Create auth options based on configuration
	if authConfig.Token() != "" {
		return ports.AuthOptions{
			Type:     ports.AuthTypeToken,
			Token:    authConfig.Token(),
			Username: "git", // Standard username for token auth
		}
	}

	if authConfig.SSHKeyPath() != "" {
		return ports.AuthOptions{
			Type:       ports.AuthTypeSSHKey,
			SSHKeyPath: authConfig.SSHKeyPath(),
		}
	}

	if authConfig.SSHKey() != "" {
		return ports.AuthOptions{
			Type:   ports.AuthTypeSSHKey,
			SSHKey: []byte(authConfig.SSHKey()),
		}
	}

	// Default to no auth
	return ports.AuthOptions{Type: ports.AuthTypeNone}
}

// BuildTargetURL constructs the target URL for pushing.
func (uc *PushToProviderUseCase) buildTargetURL(_ context.Context, request PushRequest) string {
	repositoryName := request.SourceRepository.Name()

	// Check if we need alphanumeric name transformation ()
	// AlphaNumHyphName option not implemented
	alphaNumName := false // Placeholder for request.TargetConfig.Options().AlphaNumHyphName
	if alphaNumName {
		repositoryName = shared.RemoveNonAlphaNumericChars(repositoryName)
	}

	// Extract authentication details from target config
	authConfig := request.TargetConfig.AuthConfig()
	domain := strings.TrimRight(request.TargetConfig.Domain(), "/")
	owner := request.TargetConfig.Owner()

	// Get authentication protocol - default to https
	protocol := "https" // Protocol method not implemented - defaulting to HTTPS
	token := authConfig.Token()

	// Build the Git URL with authentication
	//  main branch toGitURL + AddBasicAuthToURL functionality
	baseURL := fmt.Sprintf("%s://%s/%s/%s.git", protocol, domain, owner, repositoryName)

	// Add authentication if token is provided
	if token != "" {
		baseURL = shared.AddBasicAuthToURL(baseURL, "git", token)
	}

	return baseURL
}

// SetDefaultBranch sets the default branch at the target.
func (uc *PushToProviderUseCase) setDefaultBranch(ctx context.Context, request PushRequest, projectID string) error { //nolint:unparam // Placeholder implementation - will return errors when implemented
	logger := ports.LoggerFromContext(ctx)

	// DEBUG: Branch configuration (application state)
	logger.Debug(ctx, "Setting default branch", map[string]any{
		"branch":     request.SourceRepository.DefaultBranch(),
		"project_id": projectID,
	})

	// Implementation would call provider to set default branch
	return nil
}

// DisableProtection disables branch protection.
func (uc *PushToProviderUseCase) disableProtection(ctx context.Context, request PushRequest, projectID string) error {
	logger := ports.LoggerFromContext(ctx)

	// DEBUG: Protection disabling (application state)
	logger.Debug(ctx, "Disabling branch protection", map[string]any{
		"project_id": projectID,
		"repository": request.SourceRepository.Name(),
		"branch":     request.SourceRepository.DefaultBranch(),
	})

	// Get the default branch to unprotect
	defaultBranch := request.SourceRepository.DefaultBranch()
	if defaultBranch == "" {
		defaultBranch = defaultBranchMain // Fallback to main
	}

	// Call provider Unprotect method (as in main branch)
	err := uc.provider.UnlockAfterSync(ctx, defaultBranch, projectID)
	if err != nil {
		return fmt.Errorf("failed to unprotect branch %s: %w", defaultBranch, err)
	}

	// DEBUG: Protection disabled successfully
	logger.Info(ctx, "Branch protection disabled successfully", map[string]any{
		"project_id": projectID,
		"branch":     defaultBranch,
	})

	return nil
}

// EnableProtection enables branch protection.
func (uc *PushToProviderUseCase) enableProtection(ctx context.Context, request PushRequest, projectID string) error {
	logger := ports.LoggerFromContext(ctx)

	logger.Debug(ctx, "Enabling branch protection", map[string]any{
		"project_id": projectID,
		"repository": request.SourceRepository.Name(),
		"branch":     request.SourceRepository.DefaultBranch(),
	})

	// Get the default branch to protect
	defaultBranch := request.SourceRepository.DefaultBranch()
	if defaultBranch == "" {
		defaultBranch = defaultBranchMain // Fallback to main
	}

	// Get owner from target config
	owner := request.TargetConfig.Owner()

	// Call provider Protect method (as in main branch)
	err := uc.provider.LockForSync(ctx, owner, defaultBranch, projectID)
	if err != nil {
		return fmt.Errorf("failed to protect branch %s: %w", defaultBranch, err)
	}

	logger.Info(ctx, "Branch protection enabled successfully", map[string]any{
		"project_id": projectID,
		"owner":      owner,
		"branch":     defaultBranch,
	})

	return nil
}

// IsArchiveOrDirectory checks if provider is archive or directory type.
func (uc *PushToProviderUseCase) isArchiveOrDirectory(providerType string) bool {
	return strings.EqualFold(providerType, "archive") || strings.EqualFold(providerType, "directory")
}

// PerformDryRun simulates the push operation.
func (uc *PushToProviderUseCase) performDryRun(ctx context.Context, request PushRequest) PushResponse {
	logger := ports.LoggerFromContext(ctx)

	logger.Info(ctx, "Performing dry run push", map[string]any{
		"source": request.SourceRepository.Name(),
		"target": request.TargetConfig.Name(),
	})

	targetURL := uc.buildTargetURL(ctx, request)

	return PushResponse{
		Success:   true,
		Created:   false,
		ProjectID: "dry-run-project-id",
		TargetURL: targetURL,
	}
}
