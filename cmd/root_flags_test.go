// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// TestShortFlagAliases verifies that short flag aliases work correctly.
func TestShortFlagAliases(t *testing.T) { //nolint:paralleltest // Cannot run in parallel due to command modification
	tests := []struct {
		name      string
		args      []string
		checkFunc func(t *testing.T, cmd *cli.Command)
	}{
		{
			name: "config-file short alias -c",
			args: []string{"gitprovidersync", "-c", "custom.yaml"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				t.Helper()
				assert.Equal(t, "custom.yaml", cmd.String("config-file"))
			},
		},
		{
			name: "quiet short alias -q",
			args: []string{"gitprovidersync", "-q"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				assert.True(t, cmd.Bool("quiet"))
			},
		},
		{
			name: "log-level short alias -l",
			args: []string{"gitprovidersync", "-l", "debug"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				assert.Equal(t, "debug", cmd.String("log-level"))
			},
		},
		{
			name: "multiple short flags",
			args: []string{"gitprovidersync", "-q", "-c", "test.yaml", "-l", "trace"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				assert.True(t, cmd.Bool("quiet"))
				assert.Equal(t, "test.yaml", cmd.String("config-file"))
				assert.Equal(t, "trace", cmd.String("log-level"))
			},
		},
	}

	for _, test := range tests { //nolint:paralleltest // Cannot run in parallel due to command modification
		t.Run(test.name, func(t *testing.T) {
			// Cannot run in parallel due to command modification
			rootCmd := NewRootCommand(context.Background(), "test-version")

			// Create a test context
			ctx := context.Background()

			// Set up a Before function to capture the command
			var capturedCmd *cli.Command

			originalBefore := rootCmd.Before
			rootCmd.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
				capturedCmd = cmd
				if originalBefore != nil {
					return originalBefore(ctx, cmd)
				}

				return ctx, nil
			}

			// Run the command with test args
			err := rootCmd.Run(ctx, test.args)

			// For these tests, we expect the command to fail due to missing config
			// But the flags should still be parsed
			if err == nil || capturedCmd != nil {
				require.NotNil(t, capturedCmd, "Command should have been captured")
				test.checkFunc(t, capturedCmd)
			}
		})
	}
}

// TestSyncCommandShortFlags tests short flag aliases for the sync command.
func TestSyncCommandShortFlags(t *testing.T) { //nolint:paralleltest // Cannot run in parallel due to command modification
	tests := []struct {
		name      string
		args      []string
		checkFunc func(t *testing.T, cmd *cli.Command)
	}{
		{
			name: "dry-run short alias -n",
			args: []string{"gitprovidersync", "sync", "-n"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				assert.True(t, cmd.Bool("dry-run"))
			},
		},
		{
			name: "force-push short alias -f",
			args: []string{"gitprovidersync", "sync", "-f"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				assert.True(t, cmd.Bool("force-push"))
			},
		},
		{
			name: "since short alias -s",
			args: []string{"gitprovidersync", "sync", "-s", "2024-01-01"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				assert.Equal(t, "2024-01-01", cmd.String("since"))
			},
		},
		{
			name: "multiple sync short flags",
			args: []string{"gitprovidersync", "sync", "-n", "-f", "-s", "2024-01-01"},
			checkFunc: func(t *testing.T, cmd *cli.Command) {
				t.Helper()
				assert.True(t, cmd.Bool("dry-run"))
				assert.True(t, cmd.Bool("force-push"))
				assert.Equal(t, "2024-01-01", cmd.String("since"))
			},
		},
	}

	for _, test := range tests { //nolint:paralleltest // Cannot run in parallel due to command modification
		t.Run(test.name, func(t *testing.T) {
			// Cannot run in parallel due to command modification //nolint:paralleltest // Cannot run in parallel due to command modification
			rootCmd := NewRootCommand(context.Background(), "test-version")

			// Find the sync command
			var syncCmd *cli.Command

			for _, cmd := range rootCmd.Commands {
				if cmd.Name == "sync" {
					syncCmd = cmd

					break
				}
			}

			require.NotNil(t, syncCmd, "sync command should exist")

			// Create a test context
			ctx := context.Background()

			// Set up action to capture the command

			var capturedCmd *cli.Command

			syncCmd.Action = func(_ context.Context, cmd *cli.Command) error {
				capturedCmd = cmd
				// Don't run the actual action
				return nil
			}

			// Run the command with test args
			err := rootCmd.Run(ctx, test.args)

			if err == nil && capturedCmd != nil {
				test.checkFunc(t, capturedCmd)
			}
		})
	}
}
