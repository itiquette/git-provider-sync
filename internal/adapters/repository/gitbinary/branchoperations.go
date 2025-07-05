// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// BranchOperations provides sophisticated branch management for git binary operations.
// This restores the CreateTrackingBranches functionality from main branch gitbinary/operation.go.
type BranchOperations struct {
	executor GitCommandExecutor
	logger   ports.Logger
}

// NewBranchOperations creates a new branch operations service.
func NewBranchOperations(executor GitCommandExecutor, logger ports.Logger) *BranchOperations {
	return &BranchOperations{
		executor: executor,
		logger:   logger,
	}
}

// GitCommandExecutor defines the interface for executing git commands.
type GitCommandExecutor interface {
	RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error
	RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error)
}

// Fetch performs a comprehensive fetch operation with branch tracking.
// This restores the exact Fetch functionality from main branch.
func (bo *BranchOperations) Fetch(ctx context.Context, targetPath string) error {
	bo.logger.Debug(ctx, "Starting comprehensive fetch operation", map[string]interface{}{
		"target_path": targetPath,
	})

	// Execute the fetch commands sequence from main branch
	commands := [][]string{
		{"fetch", "--all", "--prune"},
		{"pull", "--all"},
	}

	for _, cmd := range commands {
		bo.logger.Debug(ctx, "Executing git command", map[string]interface{}{
			"command": strings.Join(cmd, " "),
			"path":    targetPath,
		})

		if err := bo.executor.RunGitCommand(ctx, nil, targetPath, cmd...); err != nil {
			return fmt.Errorf("failed to execute git %s: %w", strings.Join(cmd, " "), err)
		}
	}

	// Create tracking branches for all remote branches
	return bo.CreateTrackingBranches(ctx, targetPath)
}

// CreateTrackingBranches creates local tracking branches for all remote branches.
// This restores the exact CreateTrackingBranches functionality from main branch.
func (bo *BranchOperations) CreateTrackingBranches(ctx context.Context, targetPath string) error {
	bo.logger.Debug(ctx, "Creating tracking branches", map[string]interface{}{
		"target_path": targetPath,
	})

	// Get list of remote branches
	output, err := bo.executor.RunGitCommandWithOutput(ctx, targetPath, "branch", "-r")
	if err != nil {
		return fmt.Errorf("failed to get remote branches: %w", err)
	}

	// Process the remote branches output
	return bo.ProcessTrackingBranches(ctx, targetPath, output)
}

// ProcessTrackingBranches processes remote branch output and creates tracking branches.
// This restores the exact ProcessTrackingBranches functionality from main branch.
func (bo *BranchOperations) ProcessTrackingBranches(ctx context.Context, targetPath string, output []byte) error {
	bo.logger.Debug(ctx, "Processing tracking branches", map[string]interface{}{
		"target_path": targetPath,
		"output_size": len(output),
	})

	remoteBranches := strings.Split(strings.TrimSpace(string(output)), "\n")
	createdCount := 0
	skippedCount := 0

	for _, branch := range remoteBranches {
		branch = strings.TrimSpace(branch)

		// Skip symbolic references (HEAD -> main)
		if strings.Contains(branch, "->") {
			continue
		}

		// Skip empty lines
		if branch == "" {
			continue
		}

		// Extract local branch name (remove 'origin/' prefix)
		localBranch := strings.TrimPrefix(branch, "origin/")

		// Create tracking branch
		err := bo.executor.RunGitCommand(ctx, nil, targetPath, "branch", "--track", localBranch, branch)
		if err != nil {
			// Check if branch already exists (this is expected behavior)
			if strings.Contains(err.Error(), "already exists") {
				bo.logger.Debug(ctx, "Tracking branch already exists", map[string]interface{}{
					"branch":       branch,
					"local_branch": localBranch,
				})
				skippedCount++
			} else {
				bo.logger.Warn(ctx, "Could not create tracking branch", map[string]interface{}{
					"branch":       branch,
					"local_branch": localBranch,
					"error":        err.Error(),
				})
			}
		} else {
			bo.logger.Debug(ctx, "Created tracking branch", map[string]interface{}{
				"branch":       branch,
				"local_branch": localBranch,
			})
			createdCount++
		}
	}

	bo.logger.Info(ctx, "Tracking branch creation completed", map[string]interface{}{
		"total_branches":   len(remoteBranches),
		"created_branches": createdCount,
		"skipped_branches": skippedCount,
	})

	return nil
}

// ListRemoteBranches returns a list of all remote branches.
func (bo *BranchOperations) ListRemoteBranches(ctx context.Context, targetPath string) ([]string, error) {
	bo.logger.Debug(ctx, "Listing remote branches", map[string]interface{}{
		"target_path": targetPath,
	})

	output, err := bo.executor.RunGitCommandWithOutput(ctx, targetPath, "branch", "-r")
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}

	remoteBranches := strings.Split(strings.TrimSpace(string(output)), "\n")
	var cleanBranches []string

	for _, branch := range remoteBranches {
		branch = strings.TrimSpace(branch)

		// Skip symbolic references and empty lines
		if strings.Contains(branch, "->") || branch == "" {
			continue
		}

		cleanBranches = append(cleanBranches, branch)
	}

	return cleanBranches, nil
}

// ListLocalBranches returns a list of all local branches.
func (bo *BranchOperations) ListLocalBranches(ctx context.Context, targetPath string) ([]string, error) {
	bo.logger.Debug(ctx, "Listing local branches", map[string]interface{}{
		"target_path": targetPath,
	})

	output, err := bo.executor.RunGitCommandWithOutput(ctx, targetPath, "branch")
	if err != nil {
		return nil, fmt.Errorf("failed to list local branches: %w", err)
	}

	localBranches := strings.Split(strings.TrimSpace(string(output)), "\n")
	var cleanBranches []string

	for _, branch := range localBranches {
		// Remove the '* ' prefix from current branch and trim spaces
		branch = strings.TrimSpace(strings.TrimPrefix(branch, "* "))

		if branch != "" {
			cleanBranches = append(cleanBranches, branch)
		}
	}

	return cleanBranches, nil
}

// GetCurrentBranch returns the currently checked out branch.
func (bo *BranchOperations) GetCurrentBranch(ctx context.Context, targetPath string) (string, error) {
	bo.logger.Debug(ctx, "Getting current branch", map[string]interface{}{
		"target_path": targetPath,
	})

	output, err := bo.executor.RunGitCommandWithOutput(ctx, targetPath, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	currentBranch := strings.TrimSpace(string(output))

	bo.logger.Debug(ctx, "Current branch identified", map[string]interface{}{
		"current_branch": currentBranch,
	})

	return currentBranch, nil
}

// CreateBranch creates a new local branch.
func (bo *BranchOperations) CreateBranch(ctx context.Context, targetPath, branchName, startPoint string) error {
	bo.logger.Debug(ctx, "Creating new branch", map[string]interface{}{
		"target_path": targetPath,
		"branch_name": branchName,
		"start_point": startPoint,
	})

	args := []string{"branch", branchName}
	if startPoint != "" {
		args = append(args, startPoint)
	}

	err := bo.executor.RunGitCommand(ctx, nil, targetPath, args...)
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	bo.logger.Info(ctx, "Branch created successfully", map[string]interface{}{
		"branch_name": branchName,
		"start_point": startPoint,
	})

	return nil
}

// CheckoutBranch switches to the specified branch.
func (bo *BranchOperations) CheckoutBranch(ctx context.Context, targetPath, branchName string) error {
	bo.logger.Debug(ctx, "Checking out branch", map[string]interface{}{
		"target_path": targetPath,
		"branch_name": branchName,
	})

	err := bo.executor.RunGitCommand(ctx, nil, targetPath, "checkout", branchName)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}

	bo.logger.Info(ctx, "Branch checked out successfully", map[string]interface{}{
		"branch_name": branchName,
	})

	return nil
}
