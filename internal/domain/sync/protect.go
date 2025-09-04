// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// BranchProtectionUseCase handles branch protection operations.
type BranchProtectionUseCase struct {
	repositoryProvider ports.RepositoryProvider
	logger             ports.Logger
}

// NewBranchProtectionUseCase creates a new branch protection use case.
func NewBranchProtectionUseCase(
	repositoryProvider ports.RepositoryProvider,
	logger ports.Logger,
) BranchProtectionUseCase {
	return BranchProtectionUseCase{
		repositoryProvider: repositoryProvider,
		logger:             logger,
	}
}

// ProtectionRequest represents the input for protection operations.
type ProtectionRequest struct {
	ProviderConfig ports.ProviderConfig
	Repository     entities.Repository
	Branch         string
	Protection     ports.BranchProtection
	Operation      ProtectionOperation
}

// ProtectionOperation represents the type of protection operation.
type ProtectionOperation string

const (
	// ProtectionOperationEnable represents enable protection operation.
	ProtectionOperationEnable ProtectionOperation = "enable"
	// ProtectionOperationDisable represents disable protection operation.
	ProtectionOperationDisable ProtectionOperation = "disable"
	// ProtectionOperationUpdate represents update protection operation.
	ProtectionOperationUpdate ProtectionOperation = "update"
)

// ProtectionResponse represents the result of protection operations.
type ProtectionResponse struct {
	Success    bool
	Repository string
	Branch     string
	Operation  ProtectionOperation
	Protected  bool
	Error      error
}

// ExecuteProtection executes branch protection operations.
func (uc BranchProtectionUseCase) ExecuteProtection(
	ctx context.Context,
	request ProtectionRequest,
) (ProtectionResponse, error) {
	uc.logger.Info(ctx, "Executing branch protection operation", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
		"operation":  string(request.Operation),
		"provider":   request.ProviderConfig.ProviderType,
	})

	response := ProtectionResponse{
		Repository: request.Repository.Name(),
		Branch:     request.Branch,
		Operation:  request.Operation,
	}

	switch request.Operation {
	case ProtectionOperationEnable:
		err := uc.enableProtection(ctx, request)
		if err != nil {
			response.Error = fmt.Errorf("failed to enable protection: %w", err)
		} else {
			response.Success = true
			response.Protected = true
		}

	case ProtectionOperationDisable:
		err := uc.disableProtection(ctx, request)
		if err != nil {
			response.Error = fmt.Errorf("failed to disable protection: %w", err)
		} else {
			response.Success = true
			response.Protected = false
		}

	case ProtectionOperationUpdate:
		err := uc.updateProtection(ctx, request)
		if err != nil {
			response.Error = fmt.Errorf("failed to update protection: %w", err)
		} else {
			response.Success = true
			response.Protected = true
		}

	default:
		response.Error = fmt.Errorf("%w: %s", domain.ErrUnknownProtectionOperation, request.Operation)
	}

	uc.logger.Info(ctx, "Branch protection operation completed", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
		"operation":  string(request.Operation),
		"success":    response.Success,
		"protected":  response.Protected,
	})

	return response, response.Error
}

// GetProtectionStatus retrieves the current protection status for a branch.
func (uc BranchProtectionUseCase) GetProtectionStatus(
	ctx context.Context,
	providerConfig ports.ProviderConfig,
	repoName, branch string,
) (ports.BranchProtection, error) {
	uc.logger.Debug(ctx, "Getting branch protection status", map[string]any{
		"repository": repoName,
		"branch":     branch,
	})

	protection, err := uc.repositoryProvider.GetBranchProtection(ctx, providerConfig, repoName, branch)
	if err != nil {
		return ports.BranchProtection{}, fmt.Errorf("failed to get branch protection: %w", err)
	}

	return protection, nil
}

// ListProtectedBranches lists all protected branches for a repository.
func (uc BranchProtectionUseCase) ListProtectedBranches(
	ctx context.Context,
	providerConfig ports.ProviderConfig,
	repoName string,
) ([]string, error) {
	uc.logger.Debug(ctx, "Listing protected branches", map[string]any{
		"repository": repoName,
	})

	branches, err := uc.repositoryProvider.ListProtectedBranches(ctx, providerConfig, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to list protected branches: %w", err)
	}

	return branches, nil
}

// EnableProtection enables branch protection for a repository branch.
func (uc BranchProtectionUseCase) enableProtection(
	ctx context.Context,
	request ProtectionRequest,
) error {
	uc.logger.Debug(ctx, "Enabling branch protection", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
	})

	// Check if provider supports branch protection
	if !uc.repositoryProvider.SupportsFeature(ports.FeatureBranchProtection) {
		return fmt.Errorf("%w: %s", domain.ErrProviderNoProtectionSupport, request.ProviderConfig.ProviderType)
	}

	// Set branch protection using the repository provider
	err := uc.repositoryProvider.SetBranchProtection(
		ctx,
		request.ProviderConfig,
		request.Repository.Name(),
		request.Branch,
		request.Protection,
	)
	if err != nil {
		return fmt.Errorf("failed to set branch protection: %w", err)
	}

	uc.logger.Info(ctx, "Branch protection enabled successfully", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
	})

	return nil
}

// DisableProtection disables branch protection for a repository branch
// ports the protection service Unprotect functionality.
func (uc BranchProtectionUseCase) disableProtection(
	ctx context.Context,
	request ProtectionRequest,
) error {
	uc.logger.Debug(ctx, "Disabling branch protection", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
	})

	// Check if provider supports branch protection
	if !uc.repositoryProvider.SupportsFeature(ports.FeatureBranchProtection) {
		return fmt.Errorf("%w: %s", domain.ErrProviderNoProtectionSupport, request.ProviderConfig.ProviderType)
	}

	// Remove branch protection using the repository provider
	err := uc.repositoryProvider.RemoveBranchProtection(
		ctx,
		request.ProviderConfig,
		request.Repository.Name(),
		request.Branch,
	)
	if err != nil {
		return fmt.Errorf("failed to remove branch protection: %w", err)
	}

	uc.logger.Info(ctx, "Branch protection disabled successfully", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
	})

	return nil
}

// UpdateProtection updates branch protection settings for a repository branch.
func (uc BranchProtectionUseCase) updateProtection(
	ctx context.Context,
	request ProtectionRequest,
) error {
	uc.logger.Debug(ctx, "Updating branch protection", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
	})

	// Check if provider supports branch protection
	if !uc.repositoryProvider.SupportsFeature(ports.FeatureBranchProtection) {
		return fmt.Errorf("%w: %s", domain.ErrProviderNoProtectionSupport, request.ProviderConfig.ProviderType)
	}

	// Update branch protection using the repository provider
	err := uc.repositoryProvider.SetBranchProtection(
		ctx,
		request.ProviderConfig,
		request.Repository.Name(),
		request.Branch,
		request.Protection,
	)
	if err != nil {
		return fmt.Errorf("failed to update branch protection: %w", err)
	}

	uc.logger.Info(ctx, "Branch protection updated successfully", map[string]any{
		"repository": request.Repository.Name(),
		"branch":     request.Branch,
	})

	return nil
}
