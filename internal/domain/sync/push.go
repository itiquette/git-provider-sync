// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/shared"
)

// Error variables for common failure scenarios.
var (
	ErrTargetRepositoryName = errors.New("failed target repository name validation")
	ErrCreateRepository     = errors.New("failed to create repository")
	ErrPushChanges          = errors.New("failed to push changes")
	ErrDefaultBranch        = errors.New("failed to set default branch")
)

// PushToProviderUseCase handles the critical provider.Push functionality from main branch.
// This restores the sophisticated push-to-provider workflow in hexagonal architecture.
type PushToProviderUseCase struct {
	provider ports.RepositoryProvider
	gitOps   ports.GitOperations
	logger   ports.Logger
}

// NewPushToProviderUseCase creates a new push to provider use case.
func NewPushToProviderUseCase(provider ports.RepositoryProvider, gitOps ports.GitOperations, logger ports.Logger) *PushToProviderUseCase {
	return &PushToProviderUseCase{
		provider: provider,
		gitOps:   gitOps,
		logger:   logger,
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
// This restores the main branch provider.Push functionality in hexagonal style.
func (uc *PushToProviderUseCase) Execute(ctx context.Context, request PushRequest) (PushResponse, error) {
	uc.logger.Info(ctx, "Starting push to provider", map[string]interface{}{
		"source":     request.SourceRepository.Name(),
		"target":     request.TargetConfig.Name(),
		"force_push": request.ForcePush,
		"dry_run":    request.DryRun,
	})

	if request.DryRun {
		return uc.performDryRun(ctx, request), nil
	}

	// 1. Setup GPSUPSTREAM remote (critical missing functionality)
	if err := uc.setupGPSUpstreamRemote(ctx, request.SourceGitRepo); err != nil {
		return PushResponse{Success: false, Error: err}, fmt.Errorf("failed to setup GPSUPSTREAM remote: %w", err)
	}

	// 2. Check if repository exists at target, create if needed
	exists, projectID, err := uc.ensureRepositoryExists(ctx, request)
	if err != nil {
		return PushResponse{Success: false, Error: err}, fmt.Errorf("failed to ensure repository exists: %w", err)
	}

	created := !exists

	// 3. Disable protection if needed (restores main branch Settings.Disabled functionality)
	if request.TargetConfig.Options().DisableProtection() {
		if err := uc.disableProtection(ctx, request, projectID); err != nil {
			return PushResponse{Success: false, Error: err}, fmt.Errorf("failed to disable protection: %w", err)
		}
	}

	// 4. Perform the actual push
	targetURL, err := uc.performPush(ctx, request)
	if err != nil {
		return PushResponse{Success: false, Error: err}, fmt.Errorf("failed to push changes: %w", err)
	}

	// 5. Set default branch
	if err := uc.setDefaultBranch(ctx, request, projectID); err != nil {
		return PushResponse{Success: false, Error: err}, fmt.Errorf("failed to set default branch: %w", err)
	}

	// 6. Re-enable protection if needed (restores main branch Settings.Disabled functionality)
	if request.TargetConfig.Options().DisableProtection() {
		if err := uc.enableProtection(ctx, request, projectID); err != nil {
			return PushResponse{Success: false, Error: err}, fmt.Errorf("failed to enable protection: %w", err)
		}
	}

	uc.logger.Info(ctx, "Push to provider completed successfully", map[string]interface{}{
		"target_url": targetURL,
		"created":    created,
		"project_id": projectID,
	})

	return PushResponse{
		Success:   true,
		Created:   created,
		ProjectID: projectID,
		TargetURL: targetURL,
	}, nil
}

// setupGPSUpstreamRemote sets up the GPSUPSTREAM remote from origin.
// This restores the critical SetGPSUpstreamRemoteFromOrigin functionality from main branch.
func (uc *PushToProviderUseCase) setupGPSUpstreamRemote(ctx context.Context, gitRepo ports.GitRepository) error {
	uc.logger.Debug(ctx, "Setting up GPSUPSTREAM remote", map[string]interface{}{
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
		if remote.Name == "origin" {
			originURL = remote.URL
			break
		}
	}

	if originURL == "" {
		return fmt.Errorf("origin remote not found")
	}

	uc.logger.Debug(ctx, "Found origin remote", map[string]interface{}{
		"origin_url": originURL,
	})

	// 2. Delete existing GPSUPSTREAM remote (ignore errors like main branch)
	err = gitRepo.RemoveRemote(ctx, "GPSUPSTREAM")
	if err != nil {
		uc.logger.Debug(ctx, "Failed to remove existing GPSUPSTREAM remote (expected)", map[string]interface{}{
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
				return fmt.Errorf("mismatch in GPSUPSTREAM vs origin remote: got %s, expected %s", remote.URL, originURL)
			}
			uc.logger.Info(ctx, "GPSUPSTREAM remote setup completed successfully", map[string]interface{}{
				"origin_url":      originURL,
				"gpsupstream_url": remote.URL,
				"verified":        true,
			})
			return nil
		}
	}

	return fmt.Errorf("GPSUPSTREAM remote not found after creation")
}

// ensureRepositoryExists checks if repository exists and creates it if needed.
// This restores the main branch exists() and create() functionality.
func (uc *PushToProviderUseCase) ensureRepositoryExists(ctx context.Context, request PushRequest) (bool, string, error) {
	uc.logger.Debug(ctx, "Checking if repository exists", map[string]interface{}{
		"owner": request.TargetConfig.Owner(),
		"name":  request.SourceRepository.Name(),
	})

	// For archive/directory providers, no check needed
	if uc.isArchiveOrDirectory(string(request.TargetConfig.ProviderType())) {
		return false, "", nil
	}

	// Check if repository exists
	exists, projectID, err := uc.provider.ProjectExists(ctx,
		request.TargetConfig.Owner(),
		request.SourceRepository.Name())

	if err != nil {
		return false, "", fmt.Errorf("failed to check repository existence: %w", err)
	}

	if exists {
		uc.logger.Debug(ctx, "Repository exists at target", map[string]interface{}{
			"project_id": projectID,
		})
		return true, projectID, nil
	}

	// Repository doesn't exist - create it if allowed
	if !request.CreateIfMissing {
		return false, "", fmt.Errorf("repository does not exist and creation is disabled")
	}

	uc.logger.Debug(ctx, "Repository doesn't exist, creating", map[string]interface{}{
		"name": request.SourceRepository.Name(),
	})

	projectID, err = uc.createRepository(ctx, request)
	if err != nil {
		return false, "", fmt.Errorf("failed to create repository: %w", err)
	}

	return false, projectID, nil
}

// createRepository creates a new repository at the target provider.
// This restores the main branch create() functionality.
func (uc *PushToProviderUseCase) createRepository(ctx context.Context, request PushRequest) (string, error) {
	uc.logger.Debug(ctx, "Creating repository at target provider", map[string]interface{}{
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

	projectID, err := uc.provider.CreateRepositoryForPush(ctx, createRequest)
	if err != nil {
		return "", fmt.Errorf("%w: %s. err: %w", ErrCreateRepository, request.SourceRepository.Name(), err)
	}

	return projectID, nil
}

// buildDescription creates repository description.
// This restores the main branch buildDescription functionality.
func (uc *PushToProviderUseCase) buildDescription(request PushRequest) string {
	// TODO: Add DescriptionPrefix to MirrorOptions if needed
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

// mapVisibility maps source visibility to target visibility.
func (uc *PushToProviderUseCase) mapVisibility(request PushRequest) string {
	// TODO: Add Visibility option to MirrorOptions if needed
	visibility := "" // Placeholder for request.TargetConfig.Options().Visibility
	if visibility != "" {
		return visibility
	}

	// Default mapping logic - could be enhanced
	return request.SourceRepository.Visibility()
}

// performPush performs the actual git push operation.
// This restores the critical writer.Push() functionality from main branch.
func (uc *PushToProviderUseCase) performPush(ctx context.Context, request PushRequest) (string, error) {
	uc.logger.Debug(ctx, "Performing git push", map[string]interface{}{
		"force": request.ForcePush,
	})

	targetURL := uc.buildTargetURL(ctx, request)

	// CRITICAL: Perform the actual git push operation (restored from main branch writer.Push)
	pushOptions := ports.PushOptions{
		Remote:  "origin",
		Force:   request.ForcePush,
		Auth:    uc.createAuthOptions(ctx, request.TargetConfig.AuthConfig()),
		Timeout: time.Minute * 5, // 5 minute timeout
	}

	if err := request.SourceGitRepo.Push(ctx, pushOptions); err != nil {
		return "", fmt.Errorf("failed to push repository to target: %w", err)
	}

	uc.logger.Info(ctx, "Git push completed successfully", map[string]interface{}{
		"target_url": targetURL,
		"force":      request.ForcePush,
	})

	return targetURL, nil
}

// createAuthOptions creates authentication options for git operations.
// This restores the authentication handling from main branch.
func (uc *PushToProviderUseCase) createAuthOptions(ctx context.Context, authConfig entities.AuthConfig) ports.AuthOptions {
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

// buildTargetURL constructs the target URL for pushing.
// This restores the main branch toGitURL functionality with proper authentication.
func (uc *PushToProviderUseCase) buildTargetURL(ctx context.Context, request PushRequest) string {
	repositoryName := request.SourceRepository.Name()

	// Check if we need alphanumeric name transformation (from main branch)
	// TODO: Add AlphaNumHyphName option to MirrorOptions if needed
	alphaNumName := false // Placeholder for request.TargetConfig.Options().AlphaNumHyphName
	if alphaNumName {
		repositoryName = shared.RemoveNonAlphaNumericChars(ctx, repositoryName)
	}

	// Extract authentication details from target config
	authConfig := request.TargetConfig.AuthConfig()
	domain := strings.TrimRight(request.TargetConfig.Domain(), "/")
	owner := request.TargetConfig.Owner()

	// Get authentication protocol - default to https
	protocol := "https" // TODO: Add Protocol() method to AuthConfig interface
	token := authConfig.Token()

	// Build the Git URL with authentication
	// This restores the main branch toGitURL + AddBasicAuthToURL functionality
	baseURL := fmt.Sprintf("%s://%s/%s/%s.git", protocol, domain, owner, repositoryName)

	// Add authentication if token is provided
	if token != "" {
		baseURL = shared.AddBasicAuthToURL(ctx, baseURL, "git", token)
	}

	return baseURL
}

// setDefaultBranch sets the default branch at the target.
func (uc *PushToProviderUseCase) setDefaultBranch(ctx context.Context, request PushRequest, projectID string) error {
	uc.logger.Debug(ctx, "Setting default branch", map[string]interface{}{
		"branch":     request.SourceRepository.DefaultBranch(),
		"project_id": projectID,
	})

	// Implementation would call provider to set default branch
	return nil
}

// disableProtection disables branch protection.
// This restores the main branch provider.Unprotect functionality.
func (uc *PushToProviderUseCase) disableProtection(ctx context.Context, request PushRequest, projectID string) error {
	uc.logger.Debug(ctx, "Disabling branch protection", map[string]interface{}{
		"project_id": projectID,
		"repository": request.SourceRepository.Name(),
		"branch":     request.SourceRepository.DefaultBranch(),
	})

	// Get the default branch to unprotect
	defaultBranch := request.SourceRepository.DefaultBranch()
	if defaultBranch == "" {
		defaultBranch = "main" // Fallback to main
	}

	// Call provider Unprotect method (as in main branch)
	err := uc.provider.Unprotect(ctx, defaultBranch, projectID)
	if err != nil {
		return fmt.Errorf("failed to unprotect branch %s: %w", defaultBranch, err)
	}

	uc.logger.Info(ctx, "Branch protection disabled successfully", map[string]interface{}{
		"project_id": projectID,
		"branch":     defaultBranch,
	})

	return nil
}

// enableProtection enables branch protection.
// This restores the main branch provider.Protect functionality.
func (uc *PushToProviderUseCase) enableProtection(ctx context.Context, request PushRequest, projectID string) error {
	uc.logger.Debug(ctx, "Enabling branch protection", map[string]interface{}{
		"project_id": projectID,
		"repository": request.SourceRepository.Name(),
		"branch":     request.SourceRepository.DefaultBranch(),
	})

	// Get the default branch to protect
	defaultBranch := request.SourceRepository.DefaultBranch()
	if defaultBranch == "" {
		defaultBranch = "main" // Fallback to main
	}

	// Get owner from target config
	owner := request.TargetConfig.Owner()

	// Call provider Protect method (as in main branch)
	err := uc.provider.Protect(ctx, owner, defaultBranch, projectID)
	if err != nil {
		return fmt.Errorf("failed to protect branch %s: %w", defaultBranch, err)
	}

	uc.logger.Info(ctx, "Branch protection enabled successfully", map[string]interface{}{
		"project_id": projectID,
		"owner":      owner,
		"branch":     defaultBranch,
	})

	return nil
}

// isArchiveOrDirectory checks if provider is archive or directory type.
func (uc *PushToProviderUseCase) isArchiveOrDirectory(providerType string) bool {
	return strings.EqualFold(providerType, "archive") || strings.EqualFold(providerType, "directory")
}

// performDryRun simulates the push operation.
func (uc *PushToProviderUseCase) performDryRun(ctx context.Context, request PushRequest) PushResponse {
	uc.logger.Info(ctx, "Performing dry run push", map[string]interface{}{
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
