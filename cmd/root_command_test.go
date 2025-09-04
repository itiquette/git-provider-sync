// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRootCommand_HelpOutput tests help flag behavior without global state.
func TestRootCommand_HelpOutput(t *testing.T) { //nolint:paralleltest // urfave/cli has race conditions
	// Cannot run parallel - urfave/cli has race conditions
	ctx := context.Background()
	versionString := "test-version (Commit SHA: test-commit, Build date: test-date)"
	rootCmd := NewRootCommandForTesting(ctx, versionString)

	// Suppress output
	rootCmd.Writer = io.Discard
	rootCmd.ErrWriter = io.Discard

	// Test help flag directly without modifying os.Args
	err := rootCmd.Run(ctx, []string{"gitprovidersync", "--help"})
	require.NoError(t, err)
}

// TestRootCommand_VersionOutput tests version flag behavior without global state.
func TestRootCommand_VersionOutput(t *testing.T) { //nolint:paralleltest // urfave/cli has race conditions
	// Cannot run parallel - urfave/cli has race conditions
	ctx := context.Background()
	versionString := "test-version (Commit SHA: test-commit, Build date: test-date)"
	rootCmd := NewRootCommandForTesting(ctx, versionString)

	// Suppress output
	rootCmd.Writer = io.Discard
	rootCmd.ErrWriter = io.Discard

	// Test version flag directly without modifying os.Args
	err := rootCmd.Run(ctx, []string{"gitprovidersync", "--version"})
	require.NoError(t, err)
}
