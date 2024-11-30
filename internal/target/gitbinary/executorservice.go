// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type executorService struct {
	gitBinaryPath string
}

func NewExecutorService(binaryPath string) *executorService { //nolint
	return &executorService{
		gitBinaryPath: binaryPath,
	}
}

func (e *executorService) RunGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering RunGitCommand")

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

func (e *executorService) RunGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, e.gitBinaryPath, args...) //nolint:gosec
	if len(workingDir) != 0 {
		cmd.Dir = workingDir
	}

	return cmd.Output() //nolint
}
