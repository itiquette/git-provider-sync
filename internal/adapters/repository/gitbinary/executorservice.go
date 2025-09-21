// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/shared"
)

// ExecutorService defines the interface for executing git commands.
type ExecutorService interface {
	RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error
	RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error)
}

// ExecutorService implements ExecutorService for git binary operations
//
//	git command execution .
type executorService struct {
	gitBinaryPath string
	logger        ports.Logger
	gitTimeout    time.Duration
}

// NewExecutorService creates a new git executor service.
func NewExecutorService(binaryPath string) ExecutorService { //nolint:ireturn
	return &executorService{
		gitBinaryPath: binaryPath,
		gitTimeout:    DefaultTimeout,
	}
}

// NewExecutorServiceWithLogger creates a new git executor service with logger.
func NewExecutorServiceWithLogger(binaryPath string, logger ports.Logger) ExecutorService { //nolint:ireturn
	return &executorService{
		gitBinaryPath: binaryPath,
		logger:        logger,
		gitTimeout:    DefaultTimeout,
	}
}

// NewExecutorServiceWithTimeout creates a new git executor service with configurable timeout.
func NewExecutorServiceWithTimeout(binaryPath string, logger ports.Logger, timeout time.Duration) ExecutorService { //nolint:ireturn
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	return &executorService{
		gitBinaryPath: binaryPath,
		logger:        logger,
		gitTimeout:    timeout,
	}
}

// RunGitCommand executes a git command with environment setup and working directory

func (e *executorService) RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error {
	if e.logger != nil {
		e.logger.Trace(ctx, "Entering RunGitCommand", map[string]any{
			"args":       strings.Join(args, " "),
			"workingDir": workingDir,
		})
	}

	ctx, cancel := context.WithTimeout(ctx, e.gitTimeout)
	defer cancel()

	// #nosec G204 - Git binary path is validated at startup
	cmd := exec.CommandContext(ctx, e.gitBinaryPath, args...)

	cmd.Env = append(os.Environ(), env...)

	if len(workingDir) == 0 {
		return domain.ErrWorkingDirEmpty
	}

	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Sanitize command arguments and output to prevent credential leakage
		sanitizedArgs := shared.SanitizeURL(strings.Join(args, " "))
		sanitizedOutput := shared.SanitizeURL(string(output))

		return fmt.Errorf("error executing git %s: %w. output: %s", sanitizedArgs, err, sanitizedOutput)
	}

	if e.logger != nil {
		// Sanitize output before logging to prevent credential leakage
		e.logger.Debug(ctx, "Git command output", map[string]any{
			"output": shared.SanitizeURL(string(output)),
		})
	}

	return nil
}

// RunGitCommandWithOutput executes a git command and returns its output

func (e *executorService) RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, e.gitTimeout)
	defer cancel()

	// #nosec G204 - Git binary path is validated at startup
	cmd := exec.CommandContext(ctx, e.gitBinaryPath, args...)

	if len(workingDir) != 0 {
		cmd.Dir = workingDir
	}

	output, err := cmd.Output()
	if err != nil {
		// Sanitize command arguments to prevent credential leakage
		sanitizedArgs := shared.SanitizeURL(strings.Join(args, " "))

		return nil, fmt.Errorf("failed to execute git %s: %w", sanitizedArgs, err)
	}

	return output, nil
}
