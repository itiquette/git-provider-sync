// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

var (
	// ErrGetRemoteBranches indicates failure to retrieve remote branches.
	ErrGetRemoteBranches = errors.New("failed to get remote branches")
)

// Remote represents a git remote.
type Remote struct {
	Name string
	URL  string
}

// OperationServiceInterface defines git operations like fetch and branch management.
//
//	sophisticated git operations .
//
// Note: interfacebloat is suppressed - this interface needs to maintain backward compatibility.
type OperationServiceInterface interface { //nolint:interfacebloat // Required for backward compatibility
	Fetch(ctx context.Context, targetPath string) error
	CreateTrackingBranches(ctx context.Context, targetPath string) error
	ProcessTrackingBranches(ctx context.Context, targetPath string, output []byte) error

	// Branch operations
	GetBranches(ctx context.Context, repoPath string) ([]string, error)
	GetCurrentBranch(ctx context.Context, repoPath string) (string, error)
	CreateBranch(ctx context.Context, repoPath, branchName string) error
	DeleteBranch(ctx context.Context, repoPath, branchName string, force bool) error

	// Remote operations
	GetRemotes(ctx context.Context, repoPath string) ([]Remote, error)
	AddRemote(ctx context.Context, repoPath, name, url string) error
	RemoveRemote(ctx context.Context, repoPath, name string) error

	// Tag operations
	GetTags(ctx context.Context, repoPath string) ([]string, error)

	// Status operations
	GetStatus(ctx context.Context, repoPath string) (ports.StatusResult, error)
}

// operationService implements sophisticated git operations.
//
//	critical git operation functionality .
type operationService struct {
	executor ExecutorService
	logger   ports.Logger
}

// NewOperationServiceImpl creates a new git operations service.
func NewOperationServiceImpl(executor ExecutorService) OperationServiceInterface { //nolint:ireturn // Factory method must return interface
	return &operationService{
		executor: executor,
	}
}

// NewOperationServiceImplWithLogger creates a new git operations service with logger.
func NewOperationServiceImplWithLogger(executor ExecutorService, logger ports.Logger) OperationServiceInterface { //nolint:ireturn // Factory method must return interface
	return &operationService{
		executor: executor,
		logger:   logger,
	}
}

// Fetch performs comprehensive fetch operations: fetch --all --prune and pull --all.
//
//	critical fetch functionality .
func (b *operationService) Fetch(ctx context.Context, targetPath string) error {
	if b.logger != nil {
		b.logger.Trace(ctx, "Entering Fetch", map[string]interface{}{
			"targetPath": targetPath,
		})
		b.logger.Debug(ctx, "Fetch", map[string]interface{}{
			"targetPath": targetPath,
		})
	}

	// Execute fetch commands in sequence - critical for proper sync
	commands := [][]string{
		{"fetch", "--all", "--prune"},
		{"pull", "--all"},
	}

	for _, cmd := range commands {
		if err := b.executor.RunGitCommand(ctx, nil, targetPath, cmd...); err != nil {
			return fmt.Errorf("failed to execute git %s: %w", strings.Join(cmd, " "), err)
		}
	}

	// Create tracking branches after fetch - critical for proper branch sync
	return b.CreateTrackingBranches(ctx, targetPath)
}

// CreateTrackingBranches creates local tracking branches for all remote branches.
//
//	critical branch tracking functionality .
func (b *operationService) CreateTrackingBranches(ctx context.Context, targetPath string) error {
	if b.logger != nil {
		b.logger.Trace(ctx, "Entering CreateTrackingBranches", map[string]interface{}{
			"targetPath": targetPath,
		})
		b.logger.Debug(ctx, "CreateTrackingBranches", map[string]interface{}{
			"targetPath": targetPath,
		})
	}

	output, err := b.executor.RunGitCommandWithOutput(ctx, targetPath, "branch", "-r")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetRemoteBranches, err)
	}

	return b.ProcessTrackingBranches(ctx, targetPath, output)
}

// ProcessTrackingBranches processes git branch -r output to create tracking branches.
//
//	sophisticated branch processing .
func (b *operationService) ProcessTrackingBranches(ctx context.Context, targetPath string, output []byte) error {
	if b.logger != nil {
		b.logger.Trace(ctx, "Entering ProcessTrackingBranches", map[string]interface{}{
			"targetPath": targetPath,
		})
		b.logger.Debug(ctx, "ProcessTrackingBranches", map[string]interface{}{
			"targetPath": targetPath,
		})
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)

		// Skip symbolic refs like "origin/HEAD -> origin/main"
		if strings.Contains(branch, "->") {
			continue
		}

		// Extract local branch name from "origin/branch-name"
		localBranch := strings.TrimPrefix(branch, "origin/")

		// Skip if it's already the local branch name (no origin prefix)
		if localBranch == branch {
			continue
		}

		// Create tracking branch
		err := b.executor.RunGitCommand(ctx, nil, targetPath, "branch", "--track", localBranch, branch)
		b.handleTrackingBranchResult(ctx, err, branch)
	}

	return nil
}

// GetBranches returns all branches in the repository.
func (b *operationService) GetBranches(ctx context.Context, repoPath string) ([]string, error) {
	output, err := b.executor.RunGitCommandWithOutput(ctx, repoPath, "branch", "-a")
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")

	var result []string

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		// Remove the * marker for current branch
		branch = strings.TrimPrefix(branch, "* ")
		if branch != "" {
			result = append(result, branch)
		}
	}

	return result, nil
}

// GetCurrentBranch returns the current branch name.
func (b *operationService) GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	output, err := b.executor.RunGitCommandWithOutput(ctx, repoPath, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CreateBranch creates a new branch.
func (b *operationService) CreateBranch(ctx context.Context, repoPath, branchName string) error {
	if err := b.executor.RunGitCommand(ctx, nil, repoPath, "branch", branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// DeleteBranch deletes a branch.
func (b *operationService) DeleteBranch(ctx context.Context, repoPath, branchName string, force bool) error {
	args := []string{"branch"}
	if force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}

	args = append(args, branchName)

	if err := b.executor.RunGitCommand(ctx, nil, repoPath, args...); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// GetRemotes returns all remotes in the repository.
func (b *operationService) GetRemotes(ctx context.Context, repoPath string) ([]Remote, error) {
	output, err := b.executor.RunGitCommandWithOutput(ctx, repoPath, "remote", "-v")
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	remoteMap := make(map[string]string)

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]
			// Only keep fetch URLs for simplicity
			if len(parts) < 3 || parts[2] == "(fetch)" {
				remoteMap[name] = url
			}
		}
	}

	result := make([]Remote, 0, len(remoteMap))
	for name, url := range remoteMap {
		result = append(result, Remote{Name: name, URL: url})
	}

	return result, nil
}

// AddRemote adds a new remote.
func (b *operationService) AddRemote(ctx context.Context, repoPath, name, url string) error {
	if err := b.executor.RunGitCommand(ctx, nil, repoPath, "remote", "add", name, url); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

// RemoveRemote removes a remote.
func (b *operationService) RemoveRemote(ctx context.Context, repoPath, name string) error {
	if err := b.executor.RunGitCommand(ctx, nil, repoPath, "remote", "remove", name); err != nil {
		return fmt.Errorf("failed to remove remote: %w", err)
	}

	return nil
}

// GetTags returns all tags in the repository.
func (b *operationService) GetTags(ctx context.Context, repoPath string) ([]string, error) {
	output, err := b.executor.RunGitCommandWithOutput(ctx, repoPath, "tag", "-l")
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")

	var result []string

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result, nil
}

// GetStatus returns the status of the repository.
func (b *operationService) GetStatus(ctx context.Context, repoPath string) (ports.StatusResult, error) {
	output, err := b.executor.RunGitCommandWithOutput(ctx, repoPath, "status", "--porcelain")
	if err != nil {
		return ports.StatusResult{}, fmt.Errorf("failed to get status: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	result := ports.StatusResult{
		IsClean: len(lines) == 1 && lines[0] == "",
	}

	b.parseStatusLines(lines, &result)

	return result, nil
}

// parseStatusLines processes git status output lines and populates the result.
func (b *operationService) parseStatusLines(lines []string, result *ports.StatusResult) {
	for _, line := range lines {
		if !b.isValidStatusLine(line) {
			continue
		}

		status := line[:2]
		filename := strings.TrimSpace(line[2:])

		b.categorizeFileStatus(status, filename, result)
	}
}

// isValidStatusLine checks if a status line is valid for processing.
func (b *operationService) isValidStatusLine(line string) bool {
	return line != "" && len(line) >= 3
}

// categorizeFileStatus categorizes a file based on its git status.
func (b *operationService) categorizeFileStatus(status, filename string, result *ports.StatusResult) {
	switch {
	case b.isModified(status):
		result.Modified = append(result.Modified, filename)
	case b.isAdded(status):
		result.Added = append(result.Added, filename)
	case b.isDeleted(status):
		result.Deleted = append(result.Deleted, filename)
	case b.isRenamed(status):
		result.Renamed = append(result.Renamed, filename)
	case b.isUntracked(status):
		result.Untracked = append(result.Untracked, filename)
	case b.isConflicted(status):
		result.Conflicted = append(result.Conflicted, filename)
	}
}

// isModified checks if the status indicates a modified file.
func (b *operationService) isModified(status string) bool {
	return status[0] == 'M' || status[1] == 'M'
}

// isAdded checks if the status indicates an added file.
func (b *operationService) isAdded(status string) bool {
	return status[0] == 'A'
}

// isDeleted checks if the status indicates a deleted file.
func (b *operationService) isDeleted(status string) bool {
	return status[0] == 'D' || status[1] == 'D'
}

// isRenamed checks if the status indicates a renamed file.
func (b *operationService) isRenamed(status string) bool {
	return status[0] == 'R'
}

// isUntracked checks if the status indicates an untracked file.
func (b *operationService) isUntracked(status string) bool {
	return status[0] == '?' && status[1] == '?'
}

// isConflicted checks if the status indicates a conflicted file.
func (b *operationService) isConflicted(status string) bool {
	return status[0] == 'U' || status[1] == 'U' || status == "AA" || status == "DD"
}

// handleTrackingBranchResult handles the result of creating a tracking branch.
func (b *operationService) handleTrackingBranchResult(ctx context.Context, err error, branch string) {
	if err != nil {
		// Don't fail if branch already exists - this is expected
		if !strings.Contains(err.Error(), "already exists") && b.logger != nil {
			b.logger.Debug(ctx, "Could not create tracking branch", map[string]interface{}{
				"branch": branch,
				"error":  err.Error(),
			})
		}
	} else if b.logger != nil {
		b.logger.Debug(ctx, "Created tracking branch", map[string]interface{}{
			"branch": branch,
		})
	}
}
