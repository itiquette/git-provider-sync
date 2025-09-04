// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// TestHelpRenderer_RenderRootHelp tests root command help rendering.
func TestHelpRenderer_RenderRootHelp(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	rootCmd := &cli.Command{
		Name:        "gitprovidersync",
		Usage:       "Git repository synchronization utility",
		Description: "A tool for syncing Git repositories",
		Commands: []*cli.Command{
			{Name: "sync", Usage: "Synchronize repositories", Hidden: false},
			{Name: "hidden", Usage: "Hidden command", Hidden: true},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Usage: "Config file"},
			&cli.BoolFlag{Name: "hidden-flag", Usage: "Hidden", Hidden: true},
		},
	}

	result := renderer.RenderRootHelp(rootCmd)

	// Verify it produces non-empty output
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "gitprovidersync")

	// Verify hidden items are filtered
	content := renderer.extractRootHelpContent(rootCmd)
	assert.Len(t, content.Commands, 1) // Only visible command
	assert.Equal(t, "sync", content.Commands[0].Name)

	// Hidden flag should not be in flags
	flagNames := make([]string, len(content.Flags))

	for i, flag := range content.Flags {
		flagNames[i] = flag.Name
	}

	assert.NotContains(t, flagNames, "hidden-flag")
}

// TestHelpRenderer_RenderSubcommandHelp tests subcommand help rendering.
func TestHelpRenderer_RenderSubcommandHelp(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	cmd := &cli.Command{
		Name:  "sync",
		Usage: "Synchronize repositories",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run", Usage: "Preview changes"},
			&cli.BoolFlag{Name: "hidden", Usage: "Hidden", Hidden: true},
		},
	}

	result := renderer.RenderSubcommandHelp(cmd)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "sync")

	// Verify hidden flags are filtered
	content := renderer.extractSubcommandHelpContent(cmd)
	assert.Len(t, content.Flags, 1) // Only visible flag
	assert.Equal(t, "--dry-run", content.Flags[0].Name)
}

// TestIsHiddenFlag tests hidden flag detection.
func TestIsHiddenFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     cli.Flag
		expected bool
	}{
		{
			name:     "visible string flag",
			flag:     &cli.StringFlag{Name: "test"},
			expected: false,
		},
		{
			name:     "hidden string flag",
			flag:     &cli.StringFlag{Name: "test", Hidden: true},
			expected: true,
		},
		{
			name:     "visible bool flag",
			flag:     &cli.BoolFlag{Name: "test"},
			expected: false,
		},
		{
			name:     "hidden bool flag",
			flag:     &cli.BoolFlag{Name: "test", Hidden: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isHiddenFlag(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHelpRenderer_Integration tests with real command structure.
func TestHelpRenderer_Integration(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	// Create a realistic command structure
	app := &cli.Command{
		Name:  "gitprovidersync",
		Usage: "Git repository synchronization utility",
		Commands: []*cli.Command{
			{
				Name:  "sync",
				Usage: "Synchronize repositories",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dry-run", Usage: "Preview changes"},
				},
				Action: func(_ context.Context, _ *cli.Command) error {
					return nil
				},
			},
		},
	}

	// Test root help
	rootHelp := renderer.RenderRootHelp(app)
	require.NotEmpty(t, rootHelp)

	// Test subcommand help
	if len(app.Commands) > 0 {
		subHelp := renderer.RenderSubcommandHelp(app.Commands[0])
		require.NotEmpty(t, subHelp)
	}
}
