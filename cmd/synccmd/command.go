// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package synccmd implements the sync command
package synccmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	baseOpt "itiquette/git-provider-sync/cmd/baseoption"
	cliAdapters "itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/adapters/configuration"
	"itiquette/git-provider-sync/internal/adapters/log"
	"itiquette/git-provider-sync/internal/adapters/terminal"
	"itiquette/git-provider-sync/internal/domain/entities"
)

// contextKey is a custom type for context keys to avoid collisions..
type contextKey string

const (
	// logLevelKey is the context key for log level..
	logLevelKey contextKey = "logLevel"

	// Exit codes.
	exitMisuse      = 2  // Bad command usage
	exitConfigError = 78 // Configuration error
	// logLevelExplicitKey is the context key for whether log level was explicitly set.
	logLevelExplicitKey contextKey = "logLevelExplicit"
)

// NewSyncCommand creates the sync command for repository mirroring.
func NewSyncCommand() *cli.Command {
	cmd := &cli.Command{
		Name:        "sync",
		Usage:       "Mirror repositories from a source Git provider to targets",
		Description: `Mirror repositories from source to target providers.`,
		Action:      runSync,
		CustomHelpTemplate: `{{.Usage}}

Usage:
  {{.FullName}} [options]
{{if .Description}}
Description:
  {{.Description}}
{{end}}
Options:
{{range .VisibleFlags}}  {{.}}
{{end}}`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "dry-run",
				Aliases:  []string{"n"},
				Usage:    "Show what would be synced without making changes",
				Category: "Operations",
			},
			&cli.BoolFlag{
				Name:     "force-push",
				Aliases:  []string{"f"},
				Usage:    "Force push to target repositories",
				Category: "Operations",
			},
			&cli.StringFlag{
				Name:     "since",
				Aliases:  []string{"s"},
				Usage:    "Only sync repositories active since this date/time",
				Category: "Filtering",
			},
			&cli.BoolFlag{
				Name:     "sanitize-names",
				Usage:    "Clean repository names to alphanumeric + hyphens only",
				Category: "Processing",
			},
			&cli.BoolFlag{
				Name:     "skip-invalid",
				Usage:    "Ignore repositories with invalid names",
				Category: "Processing",
			},
		},
	}

	return cmd
}

// exitFunc allows overriding os.Exit for testing.
//
//nolint:gochecknoglobals // Required for testing
var exitFunc = os.Exit

// RunSync executes sync using simple functional approach.
func runSync(ctx context.Context, cmd *cli.Command) error {
	// Extract CLI options
	cliConfig, err := baseOpt.ExtractRootInputOptions(cmd)
	if err != nil {
		return handleError(ctx, err, "Failed to extract CLI options", withExitCode(exitMisuse))
	}

	// Setup context with config and logging
	updatedConfig := mergeSyncOptionsWithCLIConfig(cliConfig, cmd)
	ctx = cliAdapters.WithCLIConfig(ctx, updatedConfig)
	ctx = initLogger(ctx, cmd)

	// Load configuration
	config, err := configuration.DefaultConfigLoader{}.LoadConfiguration(ctx)
	if err != nil {
		return handleError(ctx, err, "Failed to load configuration", withExitCode(exitConfigError))
	}

	// Check dangerous operations
	if !confirmDangerousOperations(ctx, cmd) {
		return handleError(ctx, errors.New("operation cancelled by user"), "Operation cancelled by user", withExitCode(1))
	}

	// Execute sync - happy path
	if err := performSync(ctx, config); err != nil {
		return handleError(ctx, err, "Sync operation failed", withExitCode(1))
	}

	return nil
}

// errorOption is a functional option for error handling.
type errorOption func(*errorConfig)

// errorConfig holds error handling configuration.
type errorConfig struct {
	exitCode int
	logLevel string
}

// withExitCode sets the exit code for the error.
func withExitCode(code int) errorOption {
	return func(c *errorConfig) {
		c.exitCode = code
	}
}

// handleError handles errors with functional options.
func handleError(ctx context.Context, err error, msg string, opts ...errorOption) error {
	// Default config
	cfg := &errorConfig{
		exitCode: 1,
		logLevel: "error",
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Log the error
	log.Logger(ctx).Error().Err(err).Msg(msg)

	// Exit if not in test mode
	if !testing.Testing() {
		exitFunc(cfg.exitCode)
	}

	return fmt.Errorf("%s: %w", strings.ToLower(msg), err)
}

// MergeSyncOptionsWithCLIConfig merges sync-specific flags with existing CLI dto.
func mergeSyncOptionsWithCLIConfig(cliConfig entities.CLIConfig, cmd *cli.Command) entities.CLIConfig {
	// Extract sync flags directly from command and merge with existing CLI config
	return entities.NewCLIConfigBuilder().
		WithAlphaNumHyphName(cmd.Bool("sanitize-names")).
		WithActiveFromLimit(cmd.String("since")).
		WithDryRun(cmd.Bool("dry-run")).
		WithForcePush(cmd.Bool("force-push")).
		WithIgnoreInvalidName(cmd.Bool("skip-invalid")).
		WithOutputFormat(cliConfig.OutputFormat()).
		WithVerbosityWithCaller(cliConfig.VerbosityWithCaller()).
		WithQuiet(cliConfig.Quiet()).
		WithConfigFilePath(cliConfig.ConfigFilePath()).
		WithConfigFileOnly(cliConfig.ConfigFileOnly()).
		Build()
}

// InitLogger initializes logging with CLI options.
func initLogger(ctx context.Context, cmd *cli.Command) context.Context {
	cliConfig, ok := cliAdapters.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	withCaller := cliConfig.VerbosityWithCaller()
	outputFormat := cliConfig.OutputFormat()

	// Extract log level using the new helper
	logLevel, explicitlySet := extractLogLevel(cmd)
	quiet := logLevel == VerbosityError

	ctx = log.InitLogger(ctx, logLevel, quiet, withCaller, outputFormat)
	// Store log level in context for container configuration
	ctx = context.WithValue(ctx, logLevelKey, logLevel)
	// Store whether log level was explicitly set
	ctx = context.WithValue(ctx, logLevelExplicitKey, explicitlySet)
	log.Logger(ctx).Trace().Msg("Logger initialized")

	return ctx
}

// ExtractLogLevel determines the effective log level from various flags
// Duplicated here to avoid circular dependency with baseoption.
// Returns the log level and whether it was explicitly set by the user.
func extractLogLevel(cmd *cli.Command) (string, bool) {
	// Handle nil command
	if cmd == nil {
		return VerbosityInfo, false
	}

	// Explicit log-level takes highest precedence
	if level := cmd.String("log-level"); level != "" {
		return level, true
	}

	// Then check shortcuts (most specific first)
	if cmd.Bool("quiet") {
		return VerbosityError, true
	}

	if cmd.Bool("debug") {
		return VerbosityDebug, true
	}

	return VerbosityInfo, false // default, not explicitly set
}

// ConfirmDangerousOperations checks if dangerous operations need confirmation
// Returns true if operation should proceed, false if cancelled
// Follows idiomatic patterns: explicit, simple, no magic.
func confirmDangerousOperations(ctx context.Context, cmd *cli.Command) bool {
	// Skip confirmation if in dry-run mode (always safe)
	if cmd.Bool("dry-run") {
		return true
	}

	// Skip confirmation if --yes flag is set
	if cmd.Bool("yes") {
		return true
	}

	// Build list of dangerous operations
	var operations []string
	if cmd.Bool("force-push") {
		operations = append(operations, "force push (may overwrite remote history)")
	}

	// No dangerous operations, proceed
	if len(operations) == 0 {
		return true
	}

	// Check if we're in non-interactive mode
	if !terminal.IsInput() || !terminal.IsError() {
		// In non-interactive mode, require explicit --yes for dangerous operations
		log.Logger(ctx).Error().
			Str("operations", strings.Join(operations, ", ")).
			Msg("Dangerous operations require --yes flag in non-interactive mode")

		return false
	}

	// Interactive mode - prompt user
	operation := "This will perform: " + strings.Join(operations, ", ")

	return terminal.ConfirmOperation(operation)
}
