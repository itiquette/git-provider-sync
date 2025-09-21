// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli/v3"

	"itiquette/git-provider-sync/cmd/mancmd"
	"itiquette/git-provider-sync/cmd/printcmd"
	"itiquette/git-provider-sync/cmd/statuscmd"
	"itiquette/git-provider-sync/cmd/synccmd"
	"itiquette/git-provider-sync/internal/adapters/log"
)

// NewRootCommand returns the root CLI command
// Exported for testing purposes.
func NewRootCommand(_ context.Context, versionString string) *cli.Command {
	return newRootCommandWithOptions(versionString, true)
}

// NewRootCommandForTesting creates a root command with disabled suggestions for testing.
func NewRootCommandForTesting(_ context.Context, versionString string) *cli.Command {
	return newRootCommandWithOptions(versionString, false)
}

// NewRootCommandWithOptions creates the root command with configurable suggestion behavior.
func newRootCommandWithOptions(versionString string, enableSuggestions bool) *cli.Command {
	rootCmd := &cli.Command{
		Name:    "gitprovidersync",
		Version: versionString,
		Usage:   "mirror git repositories across providers",
		// Description removed - details are in man page and subcommand help

		// Enable shell completion
		EnableShellCompletion: true,

		// Enable command suggestions for typos (configurable for testing)
		Suggest: enableSuggestions,

		// Show version flag in help and enable --version
		HideVersion: false,

		// Clean, minimal help output like kubectl/docker
		CustomRootCommandHelpTemplate: `{{.Usage}}

Usage:
  {{.Name}} [global options] command [command options]

Commands:{{range .Commands}}{{if not .Hidden}}
  {{.Name}}{{if .Aliases}}, {{join .Aliases ", "}}{{end}}{{"\t"}}{{.Usage}}{{end}}{{end}}

Global Options:
{{range $category := .FlagCategories}}{{if $category.Name}}  {{$category.Name}}
{{end}}{{range $flag := $category.Flags}}  {{$flag}}
{{end}}{{end}}
Use '{{.Name}} <command> --help' for more information about a command.
`,

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
				Usage:    "Set logging level: error | warn | info | debug | trace",
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

// RunApplication runs the root command with the provided version information
// is the application boundary where errors can cause program exit.
func RunApplication(version, commitSHA, buildDate string) {
	ctx := context.Background()

	// Set up signal handling for graceful shutdown
	ctx, getSignalExitCode := SetupSignalContext(ctx)

	zerologLogger := log.Logger(ctx)

	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"
	err := NewRootCommand(ctx, versionString).Run(ctx, os.Args)

	// Simple error handling at application boundary
	if err != nil {
		zerologLogger.Error().Err(err).Msg("Command execution failed")

		// If debug logging was enabled, show the debug log path
		showDebugLogPath(ctx)

		// Check if the error is due to context cancellation from a signal
		if errors.Is(err, context.Canceled) {
			// Exit with the signal exit code if a signal was received
			if exitCode := getSignalExitCode(); exitCode != 0 {
				os.Exit(exitCode)
			}
		}

		// Use DetermineExitCode to get the appropriate exit code
		os.Exit(DetermineExitCode(err))
	}

	// Check if we received a signal even without error (shouldn't happen but be safe)
	if exitCode := getSignalExitCode(); exitCode != 0 {
		os.Exit(exitCode)
	}

	// Show debug log path on successful completion too (if debug enabled)
	showDebugLogPath(ctx)
}

// ShowDebugLogPath displays the debug log file path if debug logging was enabled.
func showDebugLogPath(ctx context.Context) {
	// Use the GetDebugLogPath helper from the log package
	if debugPath := log.GetDebugLogPath(ctx); debugPath != "" {
		fmt.Fprintf(os.Stderr, "\nDebug log saved to: %s\n", debugPath)
	}
}

// SetupSignalContext creates a context that will be cancelled on interrupt signals
// and returns a function to get the exit code for the received signal.
func SetupSignalContext(ctx context.Context) (context.Context, func() int) {
	// Track which signal was received
	var (
		mutex          sync.Mutex
		receivedSignal os.Signal
	)

	// Create a new context that we can cancel
	ctx, cancel := context.WithCancel(ctx)

	// Set up signal notification channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,  // Terminal hangup
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl+\ (quit with core dump)
		syscall.SIGTERM, // Termination request
	)

	// Handle signals in goroutine
	go func() {
		sig := <-sigChan

		mutex.Lock()

		receivedSignal = sig

		mutex.Unlock()

		// Provide appropriate feedback based on signal
		switch sig {
		case syscall.SIGQUIT:
			fmt.Fprintf(os.Stderr, "\nQuit\n")
		case syscall.SIGHUP:
			fmt.Fprintf(os.Stderr, "\nHangup\n")
		case syscall.SIGTERM:
			fmt.Fprintf(os.Stderr, "\nTerminated\n")
		default:
			fmt.Fprintf(os.Stderr, "\nInterrupted\n")
		}

		// Cancel the context to trigger shutdown
		cancel()
	}()

	// Return function to get exit code based on signal
	getExitCode := func() int {
		mutex.Lock()
		defer mutex.Unlock()

		if receivedSignal == nil {
			return ExitSuccess
		}

		// Use the helper function to convert signal to exit code
		if signum, ok := receivedSignal.(syscall.Signal); ok {
			return SignalToExitCode(signum)
		}

		return ExitSuccess
	}

	return ctx, getExitCode
}
