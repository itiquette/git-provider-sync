// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"itiquette/git-provider-sync/cmd/mancmd"
	"itiquette/git-provider-sync/cmd/printcmd"
	"itiquette/git-provider-sync/cmd/statuscmd"
	"itiquette/git-provider-sync/cmd/synccmd"
	"itiquette/git-provider-sync/internal/log"
)

// NewRootCommand returns the root CLI command.
// Exported for testing purposes.
func NewRootCommand(_ context.Context, versionString string) *cli.Command {
	return newRootCommandWithOptions(versionString, true)
}

// NewRootCommandForTesting creates a root command with disabled suggestions for testing.
func NewRootCommandForTesting(_ context.Context, versionString string) *cli.Command {
	return newRootCommandWithOptions(versionString, false)
}

// newRootCommandWithOptions creates the root command with configurable suggestion behavior.
func newRootCommandWithOptions(versionString string, enableSuggestions bool) *cli.Command {
	rootCmd := &cli.Command{
		Name:    "gitprovidersync",
		Version: versionString,
		Usage:   "Utility for mirroring and storing Git repositories",
		Description: `A utility for mirroring Git repositories to various Git providers or storage.
Supports GitHub, Gitea, GitLab, uncompressed directories, and a compressed archive format (tar.gz).
Allows syncing to multiple mirror destinations.`,

		// Enable shell completion
		EnableShellCompletion: true,

		// Enable command suggestions for typos (configurable for testing)
		Suggest: enableSuggestions,

		// Show version flag in help and enable --version
		HideVersion: false,

		// Global flags
		Flags: []cli.Flag{
			// Common flags (most frequently used)
			&cli.StringFlag{
				Name:     "config-file",
				Aliases:  []string{"c"},
				Value:    "gitprovidersync.yaml",
				Usage:    "Path to the configuration file",
				Category: "Configuration",
			},
			&cli.BoolFlag{
				Name:     "quiet",
				Aliases:  []string{"q"},
				Usage:    "Equivalent to --log-level=quiet. Only output errors",
				Category: "Output Control",
			},
			&cli.BoolFlag{
				Name:     "yes",
				Aliases:  []string{"y"},
				Usage:    "Automatic yes to prompts (assume yes)",
				Category: "Operations",
			},
			&cli.BoolFlag{
				Name:     "verbose",
				Aliases:  []string{"v"},
				Usage:    "Verbose output (same as --log-level=verbose)",
				Category: "Output Control",
			},
			&cli.BoolFlag{
				Name:     "debug",
				Aliases:  []string{"d"},
				Usage:    "Debug output (same as --log-level=debug)",
				Category: "Output Control",
			},

			// Additional flags
			&cli.StringFlag{
				Name:     "log-level",
				Aliases:  []string{"l"},
				Value:    "brief",
				Usage:    "Set logging level: quiet | brief | verbose | debug | trace",
				Category: "Output Control",
			},
			&cli.StringFlag{
				Name:     "format",
				Value:    "", // Auto-detect based on TTY
				Usage:    "Output format (console,json,plain) - auto-detects if not specified",
				Category: "Output Control",
			},
			&cli.StringFlag{
				Name:     "color",
				Value:    "auto",
				Usage:    "When to use colors: auto, always, never (respects NO_COLOR env)",
				Category: "Output Control",
			},
			&cli.BoolFlag{
				Name:     "config-file-only",
				Usage:    "Read configuration from file only (ignore ENV, dotenv, XDG_CONFIG_HOME)",
				Category: "Configuration",
			},
			&cli.BoolFlag{
				Name:     "plain",
				Usage:    "Equivalent to --format=plain. Tabular text output for pipelines",
				Category: "Output Control",
			},
			&cli.BoolFlag{
				Name:     "json",
				Usage:    "Equivalent to --format=json. Structured JSON output",
				Category: "Output Control",
			},

			// Developer flags (rarely used, for debugging)
			&cli.BoolFlag{
				Name:     "log-caller",
				Usage:    "Add caller information to log output (for development)",
				Category: "Debug and Logging",
			},
		},

		// Before hook for global setup
		Before: func(ctx context.Context, _ *cli.Command) (context.Context, error) {
			return ctx, nil
		},

		// Add subcommands by importance: sync (primary), status (visibility), print (secondary), man (hidden utility)
		Commands: []*cli.Command{
			synccmd.NewSyncCommand(),
			statuscmd.NewStatusCommand(),
			printcmd.NewPrintCommand(),
			mancmd.NewManCommand(),
		},
	}

	return rootCmd
}

// RunApplication runs the root command with the provided version information.
// This is the application boundary where errors can cause program exit.
func RunApplication(version, commitSHA, buildDate string) {
	ctx := context.Background()

	// Set up signal handling for graceful shutdown
	ctx = SetupSignalContext(ctx)

	zerologLogger := log.Logger(ctx)

	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"
	err := NewRootCommand(ctx, versionString).Run(ctx, os.Args)

	// Simple error handling at application boundary
	if err != nil {
		zerologLogger.Error().Err(err).Msg("Command execution failed")

		// If debug logging was enabled, show the debug log path
		showDebugLogPath(ctx)

		os.Exit(1)
	}

	// Show debug log path on successful completion too (if debug enabled)
	showDebugLogPath(ctx)
}

// showDebugLogPath displays the debug log file path if debug logging was enabled.
func showDebugLogPath(ctx context.Context) {
	// Use the GetDebugLogPath helper from the log package
	if debugPath := log.GetDebugLogPath(ctx); debugPath != "" {
		fmt.Fprintf(os.Stderr, "\nDebug log saved to: %s\n", debugPath)
	}
}

// SetupSignalContext creates a context that will be cancelled on interrupt signals.
func SetupSignalContext(ctx context.Context) context.Context {
	// Create a context that will be cancelled on SIGINT (Ctrl-C) or SIGTERM
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// Provide immediate feedback when interrupted
	go func() {
		<-ctx.Done()
		// Immediate feedback - user knows their Ctrl-C was received
		fmt.Fprintf(os.Stderr, "\nInterrupted\n")
		stop()
	}()

	return ctx
}
