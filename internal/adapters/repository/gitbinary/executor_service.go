// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ExecutorService defines the interface for executing git commands.
// This restores the critical git binary execution functionality from main branch.
type ExecutorService interface {
	RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error
	RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error)
}

// executorService implements ExecutorService for git binary operations.
// This restores the sophisticated git command execution from main branch.
type executorService struct {
	gitBinaryPath string
	logger        ports.Logger
}

// NewExecutorService creates a new git executor service.
func NewExecutorService(binaryPath string) ExecutorService {
	return &executorService{
		gitBinaryPath: binaryPath,
	}
}

// NewExecutorServiceWithLogger creates a new git executor service with logger.
func NewExecutorServiceWithLogger(binaryPath string, logger ports.Logger) ExecutorService {
	return &executorService{
		gitBinaryPath: binaryPath,
		logger:        logger,
	}
}

// RunGitCommand executes a git command with environment setup and working directory.
// This restores the main branch git command execution functionality.
func (e *executorService) RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error {
	if e.logger != nil {
		e.logger.Trace(ctx, "Entering RunGitCommand", map[string]interface{}{
			"args":       strings.Join(args, " "),
			"workingDir": workingDir,
		})
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	// #nosec G204 - Git binary path is validated at startup
	cmd := exec.CommandContext(ctx, e.gitBinaryPath, args...)

	cmd.Env = append(os.Environ(), env...)

	if len(workingDir) == 0 {
		return errors.New("failed to run git command, workingDir was empty")
	}

	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing '%s %s': %w. output: %s", e.gitBinaryPath, strings.Join(args, " "), err, output)
	}

	if e.logger != nil {
		e.logger.Debug(ctx, "Git command output", map[string]interface{}{
			"output": string(output),
		})
	}

	return nil
}

// RunGitCommandWithOutput executes a git command and returns its output.
// This restores the main branch git command with output functionality.
func (e *executorService) RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	// #nosec G204 - Git binary path is validated at startup
	cmd := exec.CommandContext(ctx, e.gitBinaryPath, args...)

	if len(workingDir) != 0 {
		cmd.Dir = workingDir
	}

	return cmd.Output()
}
