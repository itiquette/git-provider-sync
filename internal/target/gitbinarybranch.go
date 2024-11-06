// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strings"
)

// BranchManager handles branch-related operations.
type BranchManager interface {
	CreateTrackingBranches(ctx context.Context, repoPath string) error
	Fetch(ctx context.Context, workingDirPath string) error
	ProcessTrackingBranches(ctx context.Context, targetPath string, input []byte) error
}

type gitBinaryBranch struct {
	executor CommandExecutor
}

func newGitBranch(executor CommandExecutor) BranchManager {
	return &gitBinaryBranch{
		executor: executor,
	}
}

func (b *gitBinaryBranch) Fetch(ctx context.Context, targetPath string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:fetch")
	logger.Debug().Str("targetPath", targetPath).Msg("GitBinary:fetch")

	commands := [][]string{
		{"fetch", "--all", "--prune"},
		{"pull", "--all"},
	}

	for _, cmd := range commands {
		if err := b.executor.RunGitCommand(ctx, nil, targetPath, cmd...); err != nil {
			return err //nolint
		}
	}

	return b.CreateTrackingBranches(ctx, targetPath)
}

func (b *gitBinaryBranch) CreateTrackingBranches(ctx context.Context, targetPath string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:CreateTrackingBranches")
	logger.Debug().Str("targetPath", targetPath).Msg("GitBinary:CreateTrackingBranches")

	output, err := b.executor.RunGitCommandWithOutput(ctx, targetPath, "branch", "-r")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetRemoteBranches, err)
	}

	return b.ProcessTrackingBranches(ctx, targetPath, output)
}

func (b *gitBinaryBranch) ProcessTrackingBranches(ctx context.Context, targetPath string, output []byte) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:processTrackingBranches")
	logger.Debug().Str("targetPath", targetPath).Msg("GitBinary:processTrackingBranches")

	for _, branch := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		branch = strings.TrimSpace(branch)
		if strings.Contains(branch, "->") {
			continue
		}

		localBranch := strings.TrimPrefix(branch, "origin/")
		if err := b.executor.RunGitCommand(ctx, nil, targetPath, "branch", "--track", localBranch, branch); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				logger.Debug().Msgf("Could not create tracking branch for %s: %s", branch, err.Error())
			}
		} else {
			logger.Debug().Msgf("Created tracking branch for %s", branch)
		}
	}

	return nil
}
