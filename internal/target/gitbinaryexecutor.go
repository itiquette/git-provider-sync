// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CommandExecutor handles git command execution.
type CommandExecutor interface {
	RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error
	RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error)
}

type gitBinaryExec struct {
	gitBinaryPath string
}

func newExecService(binaryPath string) CommandExecutor {
	return &gitBinaryExec{
		gitBinaryPath: binaryPath,
	}
}

func (e *gitBinaryExec) RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering runGitCommand")

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.gitBinaryPath, args...) //nolint:gosec

	cmd.Env = append(os.Environ(), env...)
	if len(workingDir) != 0 {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing '%s %s': %w. err: %s", e.gitBinaryPath, strings.Join(args, " "), err, output)
	}

	logger.Debug().Msgf("Git command output: %s", output)

	return nil
}

func (e *gitBinaryExec) RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, e.gitBinaryPath, args...) //nolint:gosec
	if len(workingDir) != 0 {
		cmd.Dir = workingDir
	}

	return cmd.Output() //nolint
}
