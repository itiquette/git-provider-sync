// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strings"
)

type operation struct {
	executor *executorService
}

func NewOperation(executor *executorService) *operation { //nolint
	return &operation{
		executor: executor,
	}
}

func (b *operation) Fetch(ctx context.Context, targetPath string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Fetch")
	logger.Debug().Str("targetPath", targetPath).Msg("Fetch")

	commands := [][]string{
		{"fetch", "--all", "--prune"},
		{"pull", "--all"},
	}

	for _, cmd := range commands {
		if err := b.executor.RunGitCommand(ctx, nil, targetPath, cmd...); err != nil {
			return err
		}
	}

	return b.CreateTrackingBranches(ctx, targetPath)
}

func (b *operation) CreateTrackingBranches(ctx context.Context, targetPath string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering CreateTrackingBranches")
	logger.Debug().Str("targetPath", targetPath).Msg("CreateTrackingBranches")

	output, err := b.executor.RunGitCommandWithOutput(ctx, targetPath, "branch", "-r")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetRemoteBranches, err)
	}

	return b.ProcessTrackingBranches(ctx, targetPath, output)
}

func (b *operation) ProcessTrackingBranches(ctx context.Context, targetPath string, output []byte) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering ProcessTrackingBranches")
	logger.Debug().Str("targetPath", targetPath).Msg("ProcessTrackingBranches")

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
