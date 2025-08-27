// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// Mock HelpFormatter for testing

type mockHelpFormatter struct {
	returnValue string
}

func (m *mockHelpFormatter) Bold(text string) string {
	return "**" + text + "**"
}

func (m *mockHelpFormatter) Section(title string) string {
	return "== " + title + " =="
}

func (m *mockHelpFormatter) IsColorSupported() bool {
	return false
}

// Test HelpRenderer constructors

func TestNewHelpRenderer(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	require.NotNil(t, renderer)
	assert.NotNil(t, renderer.helpService)
}

func TestNewHelpRendererWithFormatter(t *testing.T) {
	t.Parallel()

	formatter := &mockHelpFormatter{returnValue: "test output"}
	renderer := NewHelpRendererWithFormatter(formatter)

	require.NotNil(t, renderer)
	assert.NotNil(t, renderer.helpService)
}

// Test RenderRootHelp

func TestHelpRenderer_RenderRootHelp(t *testing.T) {
	t.Parallel()

	formatter := &mockHelpFormatter{returnValue: "formatted root help"}
	renderer := NewHelpRendererWithFormatter(formatter)

	// Create a test root command
	rootCmd := &cli.Command{
		Name:        "gitprovidersync",
		Usage:       "Git repository synchronization utility",
		Description: "A comprehensive tool for syncing Git repositories across providers.\nSupports GitHub, GitLab, and Gitea.",
		Commands: []*cli.Command{
			{
				Name:   "sync",
				Usage:  "Synchronize repositories",
				Hidden: false,
			},
			{
				Name:   "print",
				Usage:  "Print configuration",
				Hidden: false,
			},
			{
				Name:   "hidden-cmd",
				Usage:  "Hidden command",
				Hidden: true,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config-file",
				Usage: "Path to configuration file",
			},
			&cli.BoolFlag{
				Name:  "quiet",
				Usage: "Suppress output",
			},
			&cli.BoolFlag{
				Name:   "hidden-flag",
				Usage:  "Hidden flag",
				Hidden: true,
			},
		},
	}

	result := renderer.RenderRootHelp(rootCmd)

	// Just check that the result is a non-empty formatted string
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Git repository synchronization utility")

	// Test content extraction separately
	content := renderer.extractRootHelpContent(rootCmd)
	assert.Equal(t, "Git repository synchronization utility", content.Title)
	assert.Equal(t, "A comprehensive tool for syncing Git repositories across providers.", content.Description)
	assert.Equal(t, "gitprovidersync [command]", content.Usage)

	// Should have examples
	require.Len(t, content.Examples, 3)
	assert.Equal(t, "verify configuration", content.Examples[0].Description)
	assert.Equal(t, "gitprovidersync print", content.Examples[0].Command)
	assert.Equal(t, "preview changes", content.Examples[1].Description)
	assert.Equal(t, "gitprovidersync sync --dry-run", content.Examples[1].Command)
	assert.Equal(t, "execute sync", content.Examples[2].Description)
	assert.Equal(t, "gitprovidersync sync", content.Examples[2].Command)

	// Should have commands (excluding hidden ones)
	require.Len(t, content.Commands, 2)
	commandNames := []string{content.Commands[0].Name, content.Commands[1].Name}
	assert.Contains(t, commandNames, "sync")
	assert.Contains(t, commandNames, "print")
	assert.NotContains(t, commandNames, "hidden-cmd")

	// Should have flags (excluding hidden ones)
	require.GreaterOrEqual(t, len(content.Flags), 2)

	flagNames := make([]string, len(content.Flags))
	for i, flag := range content.Flags {
		flagNames[i] = flag.Name
	}

	assert.Contains(t, flagNames, "--config-file")
	assert.Contains(t, flagNames, "--quiet")

	// Should have support info
	assert.NotEmpty(t, content.Support.Documentation)
	assert.NotEmpty(t, content.Support.Issues)
	assert.NotEmpty(t, content.Support.Discussions)
}

func TestHelpRenderer_RenderSubcommandHelp(t *testing.T) {
	t.Parallel()

	formatter := &mockHelpFormatter{returnValue: "formatted subcommand help"}
	renderer := NewHelpRendererWithFormatter(formatter)

	// Create a test subcommand
	syncCmd := &cli.Command{
		Name:        "sync",
		Usage:       "Synchronize repositories",
		Description: "Synchronizes Git repositories from source to mirror destinations.\nSupports dry-run mode for preview.",
		Action:      func(_ context.Context, _ *cli.Command) error { return nil }, // Add action to make it show [options]
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview changes without making them",
			},
			&cli.BoolFlag{
				Name:  "force-push",
				Usage: "Force push to mirrors",
			},
		},
	}

	result := renderer.RenderSubcommandHelp(syncCmd)

	// Just check that the result is a non-empty formatted string
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Synchronize repositories")

	// Test content extraction separately
	content := renderer.extractSubcommandHelpContent(syncCmd)
	assert.Equal(t, "Synchronize repositories", content.Title)
	assert.Equal(t, "Synchronizes Git repositories from source to mirror destinations.", content.Description)
	assert.Equal(t, "sync [options]", content.Usage)

	// Should have sync-specific examples
	require.Len(t, content.Examples, 2)
	assert.Equal(t, "run synchronization", content.Examples[0].Description)
	assert.Equal(t, "gitprovidersync sync", content.Examples[0].Command)
	assert.Equal(t, "preview changes only", content.Examples[1].Description)
	assert.Equal(t, "gitprovidersync sync --dry-run", content.Examples[1].Command)

	// Should not have nested commands
	assert.Empty(t, content.Commands)

	// Should have subcommand flags
	require.GreaterOrEqual(t, len(content.Flags), 2)

	flagNames := make([]string, len(content.Flags))
	for i, flag := range content.Flags {
		flagNames[i] = flag.Name
	}

	assert.Contains(t, flagNames, "--dry-run")
	assert.Contains(t, flagNames, "--force-push")

	// Should have documentation link
	assert.Contains(t, content.Support.Documentation, "#sync-command")
}

// Test extractRootHelpContent

func TestHelpRenderer_extractRootHelpContent(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	rootCmd := &cli.Command{
		Name:        "gitprovidersync",
		Usage:       "Test utility",
		Description: "First line of description.\nSecond line should be ignored.",
		Commands: []*cli.Command{
			{Name: "cmd1", Usage: "Command 1", Hidden: false},
			{Name: "cmd2", Usage: "Command 2", Hidden: true},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Usage: "Config file path"},
			&cli.BoolFlag{Name: "quiet", Usage: "Suppress output"},
		},
	}

	content := renderer.extractRootHelpContent(rootCmd)

	assert.Equal(t, "Test utility", content.Title)
	assert.Equal(t, "First line of description.", content.Description)
	assert.Equal(t, "gitprovidersync [command]", content.Usage)
	assert.Len(t, content.Examples, 3) // Standard root examples
	assert.Len(t, content.Commands, 1) // Only non-hidden commands
	assert.Equal(t, "cmd1", content.Commands[0].Name)
	assert.GreaterOrEqual(t, len(content.Flags), 2)
}

func TestHelpRenderer_extractSubcommandHelpContent(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	tests := []struct {
		name             string
		cmdName          string
		cmdUsage         string
		expectedExamples int
	}{
		{
			name:             "sync command",
			cmdName:          "sync",
			cmdUsage:         "Synchronize repositories",
			expectedExamples: 2,
		},
		{
			name:             "print command",
			cmdName:          "print",
			cmdUsage:         "Print configuration",
			expectedExamples: 2,
		},
		{
			name:             "unknown command",
			cmdName:          "unknown",
			cmdUsage:         "Unknown command",
			expectedExamples: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Name:   test.cmdName,
				Usage:  test.cmdUsage,
				Action: func(_ context.Context, _ *cli.Command) error { return nil }, // Has action, so usage should be [options]
			}

			content := renderer.extractSubcommandHelpContent(cmd)

			assert.Equal(t, test.cmdUsage, content.Title)
			assert.Equal(t, test.cmdUsage, content.Description)
			assert.Equal(t, test.cmdName+" [options]", content.Usage)
			assert.Len(t, content.Examples, test.expectedExamples)
			assert.Empty(t, content.Commands)
		})
	}
}

// Test helper functions

func TestHelpRenderer_extractDescription(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	tests := []struct {
		name        string
		description string
		usage       string
		expected    string
	}{
		{
			name:        "multi-line description",
			description: "First line of description.\nSecond line.\nThird line.",
			usage:       "Fallback usage",
			expected:    "First line of description.",
		},
		{
			name:        "single line description",
			description: "Single line description.",
			usage:       "Fallback usage",
			expected:    "Single line description.",
		},
		{
			name:        "empty description",
			description: "",
			usage:       "Fallback usage",
			expected:    "Fallback usage",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Usage:       test.usage,
				Description: test.description,
			}

			result := renderer.extractDescription(cmd)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_extractUsage(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	tests := []struct {
		name      string
		cmdName   string
		hasAction bool
		expected  string
	}{
		{
			name:      "command with action",
			cmdName:   "sync",
			hasAction: true,
			expected:  "sync [options]",
		},
		{
			name:      "command without action",
			cmdName:   "gitprovidersync",
			hasAction: false,
			expected:  "gitprovidersync [command]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Name: test.cmdName,
			}

			if test.hasAction {
				cmd.Action = func(_ context.Context, _ *cli.Command) error { return nil } // Dummy action
			}

			result := renderer.extractUsage(cmd)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_extractCommands(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	cmd := &cli.Command{
		Commands: []*cli.Command{
			{Name: "visible-cmd", Usage: "Visible command", Hidden: false},
			{Name: "hidden-cmd", Usage: "Hidden command", Hidden: true},
			{Name: "default-cmd", Usage: "Default command"}, // Hidden defaults to false
		},
	}

	commands := renderer.extractCommands(cmd)

	require.Len(t, commands, 2)
	commandNames := []string{commands[0].Name, commands[1].Name}
	assert.Contains(t, commandNames, "visible-cmd")
	assert.Contains(t, commandNames, "default-cmd")
	assert.NotContains(t, commandNames, "hidden-cmd")
}

func TestHelpRenderer_extractRootFlags(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config-file",
				Usage: "Configuration file path",
			},
			&cli.BoolFlag{
				Name:  "quiet",
				Usage: "Suppress output",
			},
			&cli.StringFlag{
				Name:  "verbosity",
				Usage: "Set verbosity level",
			},
			&cli.BoolFlag{
				Name:   "hidden-flag",
				Usage:  "Hidden flag",
				Hidden: true,
			},
		},
	}

	flags := renderer.extractRootFlags(cmd)

	// Should exclude hidden flags
	flagNames := make([]string, len(flags))
	for i, flag := range flags {
		flagNames[i] = flag.Name
	}

	assert.Contains(t, flagNames, "--config-file")
	assert.Contains(t, flagNames, "--quiet")
	assert.Contains(t, flagNames, "--verbosity")
	assert.NotContains(t, flagNames, "--hidden-flag")

	// Check common flag categorization
	for _, flag := range flags {
		switch flag.Name {
		case "--config-file", "--quiet":
			assert.True(t, flag.IsCommon, "Flag %s should be marked as common", flag.Name)
		case "--verbosity":
			assert.False(t, flag.IsCommon, "Flag %s should not be marked as common", flag.Name)
		}
	}
}

func TestHelpRenderer_extractSubcommandFlags(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	tests := []struct {
		name                string
		cmdName             string
		flags               []cli.Flag
		expectedCommonFlags []string
	}{
		{
			name:    "sync command flags",
			cmdName: "sync",
			flags: []cli.Flag{
				&cli.BoolFlag{Name: "dry-run", Usage: "Preview changes"},
				&cli.BoolFlag{Name: "force-push", Usage: "Force push"},
				&cli.StringFlag{Name: "config", Usage: "Config file"},
			},
			expectedCommonFlags: []string{"--dry-run", "--force-push"},
		},
		{
			name:    "other command flags",
			cmdName: "print",
			flags: []cli.Flag{
				&cli.BoolFlag{Name: "verbose", Usage: "Verbose output"},
				&cli.StringFlag{Name: "format", Usage: "Output format"},
			},
			expectedCommonFlags: []string{}, // Only help is common for non-sync commands
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Name:  test.cmdName,
				Flags: test.flags,
			}

			flags := renderer.extractSubcommandFlags(cmd)

			// Check common flag categorization
			commonFlags := []string{}

			for _, flag := range flags {
				if flag.IsCommon {
					commonFlags = append(commonFlags, flag.Name)
				}
			}

			for _, expectedCommon := range test.expectedCommonFlags {
				assert.Contains(t, commonFlags, expectedCommon, "Flag %s should be marked as common", expectedCommon)
			}
		})
	}
}

func TestHelpRenderer_extractSupportInfo(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	support := renderer.extractSupportInfo()

	assert.NotEmpty(t, support.Documentation)
	assert.NotEmpty(t, support.Issues)
	assert.NotEmpty(t, support.Discussions)
	assert.Contains(t, support.Documentation, "github.com/itiquette/git-provider-sync")
	assert.Contains(t, support.Issues, "github.com/itiquette/git-provider-sync/issues")
	assert.Contains(t, support.Discussions, "github.com/itiquette/git-provider-sync/discussions")
}

func TestHelpRenderer_extractDocumentationLink(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	tests := []struct {
		name             string
		cmdName          string
		expectedFragment string
	}{
		{
			name:             "sync command",
			cmdName:          "sync",
			expectedFragment: "#sync-command",
		},
		{
			name:             "print command",
			cmdName:          "print",
			expectedFragment: "#print-command",
		},
		{
			name:             "unknown command",
			cmdName:          "unknown",
			expectedFragment: "README.adoc", // Falls back to README
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Name: test.cmdName,
			}

			support := renderer.extractDocumentationLink(cmd)

			assert.NotEmpty(t, support.Documentation)
			assert.Contains(t, support.Documentation, test.expectedFragment)
		})
	}
}

// Test flag utility functions

func TestHelpRenderer_isHiddenFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     cli.Flag
		expected bool
	}{
		{
			name:     "visible string flag",
			flag:     &cli.StringFlag{Name: "config", Hidden: false},
			expected: false,
		},
		{
			name:     "hidden string flag",
			flag:     &cli.StringFlag{Name: "secret", Hidden: true},
			expected: true,
		},
		{
			name:     "visible bool flag",
			flag:     &cli.BoolFlag{Name: "verbose", Hidden: false},
			expected: false,
		},
		{
			name:     "hidden bool flag",
			flag:     &cli.BoolFlag{Name: "debug", Hidden: true},
			expected: true,
		},
		{
			name:     "visible int flag",
			flag:     &cli.IntFlag{Name: "count", Hidden: false},
			expected: false,
		},
		{
			name:     "hidden int flag",
			flag:     &cli.IntFlag{Name: "internal", Hidden: true},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isHiddenFlag(test.flag)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_getFlagUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     cli.Flag
		expected string
	}{
		{
			name:     "string flag",
			flag:     &cli.StringFlag{Name: "config", Usage: "Configuration file"},
			expected: "Configuration file",
		},
		{
			name:     "bool flag",
			flag:     &cli.BoolFlag{Name: "verbose", Usage: "Enable verbose output"},
			expected: "Enable verbose output",
		},
		{
			name:     "int flag",
			flag:     &cli.IntFlag{Name: "count", Usage: "Number of items"},
			expected: "Number of items",
		},
		{
			name:     "duration flag",
			flag:     &cli.DurationFlag{Name: "timeout", Usage: "Operation timeout"},
			expected: "Operation timeout",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := getFlagUsage(test.flag)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_splitIntoLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single line",
			text:     "Single line text",
			expected: []string{"Single line text"},
		},
		{
			name:     "empty text",
			text:     "",
			expected: []string{},
		},
		{
			name:     "multi-line text",
			text:     "First line\nSecond line\nThird line",
			expected: []string{"First line", "Second line", "Third line"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := splitIntoLines(test.text)
			assert.Equal(t, test.expected, result)
		})
	}
}

// Integration tests

func TestHelpRenderer_Integration_RealCommand(t *testing.T) {
	t.Parallel()

	formatter := &mockHelpFormatter{returnValue: "integration test output"}
	renderer := NewHelpRendererWithFormatter(formatter)

	// Create a realistic command structure
	rootCmd := &cli.Command{
		Name:        "gitprovidersync",
		Usage:       "Git repository synchronization utility",
		Description: "Synchronize Git repositories across multiple providers.\nSupports GitHub, GitLab, Gitea, and more.",
		Commands: []*cli.Command{
			{
				Name:   "sync",
				Usage:  "Synchronize repositories",
				Hidden: false,
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dry-run", Usage: "Preview changes"},
					&cli.BoolFlag{Name: "force-push", Usage: "Force push"},
				},
			},
			{
				Name:   "print",
				Usage:  "Print configuration",
				Hidden: false,
			},
			{
				Name:   "status",
				Usage:  "Show status",
				Hidden: false,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Usage: "Path to config file"},
			&cli.BoolFlag{Name: "quiet", Usage: "Suppress output"},
			&cli.StringFlag{Name: "verbosity", Usage: "Set verbosity level"},
		},
	}

	result := renderer.RenderRootHelp(rootCmd)

	// Just check that the result is a non-empty formatted string
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Git repository synchronization utility")

	// Verify comprehensive content extraction
	content := renderer.extractRootHelpContent(rootCmd)
	assert.Equal(t, "Git repository synchronization utility", content.Title)
	assert.Contains(t, content.Description, "Synchronize Git repositories across multiple providers.")
	assert.Equal(t, "gitprovidersync [command]", content.Usage)
	assert.Len(t, content.Examples, 3)
	assert.Len(t, content.Commands, 3)
	assert.GreaterOrEqual(t, len(content.Flags), 3)
	assert.NotEmpty(t, content.Support.Documentation)
}

// Benchmark help rendering operations

func BenchmarkHelpRenderer_RenderRootHelp(b *testing.B) {
	formatter := &mockHelpFormatter{returnValue: "benchmark output"}
	renderer := NewHelpRendererWithFormatter(formatter)

	rootCmd := &cli.Command{
		Name:        "gitprovidersync",
		Usage:       "Git repository synchronization utility",
		Description: "Synchronize Git repositories across multiple providers.",
		Commands: []*cli.Command{
			{Name: "sync", Usage: "Synchronize repositories"},
			{Name: "print", Usage: "Print configuration"},
			{Name: "status", Usage: "Show status"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Usage: "Path to config file"},
			&cli.BoolFlag{Name: "quiet", Usage: "Suppress output"},
		},
	}

	b.ResetTimer()

	for range b.N {
		renderer.RenderRootHelp(rootCmd)
	}
}

func BenchmarkHelpRenderer_extractRootHelpContent(b *testing.B) {
	renderer := NewHelpRenderer()

	rootCmd := &cli.Command{
		Name:        "gitprovidersync",
		Usage:       "Git repository synchronization utility",
		Description: "Synchronize Git repositories across multiple providers.",
		Commands: []*cli.Command{
			{Name: "sync", Usage: "Synchronize repositories"},
			{Name: "print", Usage: "Print configuration"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Usage: "Path to config file"},
			&cli.BoolFlag{Name: "quiet", Usage: "Suppress output"},
		},
	}

	b.ResetTimer()

	for range b.N {
		renderer.extractRootHelpContent(rootCmd)
	}
}

// Additional tests for edge cases and better coverage

func TestHelpRenderer_extractRootFlags_WithAliases(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Verbose output",
			},
			&cli.BoolFlag{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "Show help",
			},
			&cli.IntFlag{
				Name:    "count",
				Aliases: []string{"c", "num"},
				Usage:   "Set count",
			},
		},
	}

	flags := renderer.extractRootFlags(cmd)

	// Find flags with aliases
	for _, flag := range flags {
		if flag.Name == "-v, --verbose" {
			assert.Equal(t, "Verbose output", flag.Description)
		}

		if flag.Name == "-h, --help" {
			assert.Equal(t, "Show help", flag.Description)
			assert.True(t, flag.IsCommon) // help should be common
		}

		if flag.Name == "-c, --count" {
			assert.Equal(t, "Set count", flag.Description)
		}
	}
}

func TestHelpRenderer_extractSubcommandFlags_WithAliases(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	cmd := &cli.Command{
		Name: "sync",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"d"},
				Usage:   "Preview changes",
			},
			&cli.BoolFlag{
				Name:    "force-push",
				Aliases: []string{"f"},
				Usage:   "Force push",
			},
		},
	}

	flags := renderer.extractSubcommandFlags(cmd)

	// Find flags with aliases
	for _, flag := range flags {
		if flag.Name == "-d, --dry-run" {
			assert.Equal(t, "Preview changes", flag.Description)
			assert.True(t, flag.IsCommon) // dry-run should be common for sync
		}

		if flag.Name == "-f, --force-push" {
			assert.Equal(t, "Force push", flag.Description)
			assert.True(t, flag.IsCommon) // force-push should be common for sync
		}
	}
}

func TestHelpRenderer_isHiddenFlag_AdditionalTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     cli.Flag
		expected bool
	}{
		{
			name:     "visible float64 flag",
			flag:     &cli.Float64Flag{Name: "rate", Hidden: false},
			expected: false,
		},
		{
			name:     "hidden float64 flag",
			flag:     &cli.Float64Flag{Name: "internal-rate", Hidden: true},
			expected: true,
		},
		{
			name:     "visible duration flag",
			flag:     &cli.DurationFlag{Name: "timeout", Hidden: false},
			expected: false,
		},
		{
			name:     "hidden duration flag",
			flag:     &cli.DurationFlag{Name: "internal-timeout", Hidden: true},
			expected: true,
		},
		{
			name:     "visible string slice flag",
			flag:     &cli.StringSliceFlag{Name: "items", Hidden: false},
			expected: false,
		},
		{
			name:     "hidden string slice flag",
			flag:     &cli.StringSliceFlag{Name: "internal-items", Hidden: true},
			expected: true,
		},
		{
			name:     "visible int slice flag",
			flag:     &cli.IntSliceFlag{Name: "numbers", Hidden: false},
			expected: false,
		},
		{
			name:     "hidden int slice flag",
			flag:     &cli.IntSliceFlag{Name: "internal-numbers", Hidden: true},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isHiddenFlag(test.flag)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_getFlagUsage_AdditionalTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     cli.Flag
		expected string
	}{
		{
			name:     "float64 flag",
			flag:     &cli.Float64Flag{Name: "rate", Usage: "Processing rate"},
			expected: "Processing rate",
		},
		{
			name:     "string slice flag",
			flag:     &cli.StringSliceFlag{Name: "items", Usage: "List of items"},
			expected: "List of items",
		},
		{
			name:     "int slice flag",
			flag:     &cli.IntSliceFlag{Name: "numbers", Usage: "List of numbers"},
			expected: "List of numbers",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := getFlagUsage(test.flag)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_splitIntoLines_WhitespaceAndEmptyLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "text with empty lines",
			text:     "First line\n\nThird line\n\nFifth line",
			expected: []string{"First line", "Third line", "Fifth line"},
		},
		{
			name:     "text with whitespace-only lines",
			text:     "First line\n   \nThird line\n\t\nFifth line",
			expected: []string{"First line", "Third line", "Fifth line"},
		},
		{
			name:     "text with leading/trailing whitespace",
			text:     "  First line  \n  Second line  \n  Third line  ",
			expected: []string{"First line", "Second line", "Third line"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := splitIntoLines(test.text)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_extractDescription_EmptyAndMultilineText(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	tests := []struct {
		name        string
		description string
		usage       string
		expected    string
	}{
		{
			name:        "description with only empty lines",
			description: "\n\n\n",
			usage:       "Fallback usage",
			expected:    "Fallback usage",
		},
		{
			name:        "description with whitespace only",
			description: "   \t   \n   \t   ",
			usage:       "Fallback usage",
			expected:    "Fallback usage",
		},
		{
			name:        "description with empty first line but content later",
			description: "\nSecond line content\nThird line",
			usage:       "Fallback usage",
			expected:    "Second line content",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Usage:       test.usage,
				Description: test.description,
			}

			result := renderer.extractDescription(cmd)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHelpRenderer_extractRootExamples(t *testing.T) {
	t.Parallel()

	renderer := NewHelpRenderer()

	examples := renderer.extractRootExamples()

	require.Len(t, examples, 3)
	assert.Equal(t, "verify configuration", examples[0].Description)
	assert.Equal(t, "gitprovidersync print", examples[0].Command)
	assert.Equal(t, "preview changes", examples[1].Description)
	assert.Equal(t, "gitprovidersync sync --dry-run", examples[1].Command)
	assert.Equal(t, "execute sync", examples[2].Description)
	assert.Equal(t, "gitprovidersync sync", examples[2].Command)
}
