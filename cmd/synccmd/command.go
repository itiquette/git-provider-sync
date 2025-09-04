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
	"itiquette/git-provider-sync/internal/adapters/terminal"
	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/log"
)

// NewSyncCommand creates the sync command for repository mirroring.
func NewSyncCommand() *cli.Command {
	cmd := &cli.Command{
		Name:  "sync",
		Usage: "Mirror repositories from a source Git provider to targets",
		Description: `The 'sync' command mirrors your repositories from a source Git provider to one or more targets.
It allows for various options to control the synchronization process.`,
		Action: runSync,
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

// RunSync executes sync using simple functional approach.
func runSync(ctx context.Context, cmd *cli.Command) error {
	cliConfig, err := baseOpt.ExtractRootInputOptions(cmd)
	if err != nil {
		log.Logger(ctx).Error().Err(err).Msg("Failed to extract CLI options")

		if !testing.Testing() {
			os.Exit(2) // Configuration error
		}

		return fmt.Errorf("failed to extract CLI options: %w", err)
	}

	// Extract sync-specific flags and merge with CLI config
	updatedConfig := mergeSyncOptionsWithCLIConfig(cliConfig, cmd)

	ctx = cliAdapters.WithCLIConfig(ctx, updatedConfig)
	ctx = initLogger(ctx, cmd)

	// Use original proven configuration loader
	config, err := configuration.DefaultConfigLoader{}.LoadConfiguration(ctx)
	if err != nil {
		log.Logger(ctx).Error().Err(err).Msg("Failed to load configuration")

		if !testing.Testing() {
			os.Exit(2) // Configuration error
		}

		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check for dangerous operations requiring confirmation
	if !confirmDangerousOperations(ctx, cmd) {
		log.Logger(ctx).Info().Msg("Operation cancelled by user")

		if !testing.Testing() {
			os.Exit(1) // User cancelled
		}

		return errors.New("operation cancelled by user")
	}

	// Execute sync using proper hexagonal architecture
	err = performSync(ctx, config)
	if err != nil {
		log.Logger(ctx).Error().Err(err).Msg("Sync operation failed")

		if !testing.Testing() {
			os.Exit(1) // Operation failure
		}

		return err
	}

	return nil
}

// MergeSyncOptionsWithCLIConfig merges sync-specific flags with existing CLI config.
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
	logLevel := extractLogLevel(cmd)
	quiet := logLevel == "quiet"

	ctx = log.InitLogger(ctx, logLevel, quiet, withCaller, outputFormat)
	log.Logger(ctx).Trace().Msg("Logger initialized")

	return ctx
}

// ExtractLogLevel determines the effective log level from various flags
// Duplicated here to avoid circular dependency with baseoption.
func extractLogLevel(cmd *cli.Command) string {
	// Handle nil command
	if cmd == nil {
		return "brief"
	}

	// Explicit log-level takes highest precedence
	if level := cmd.String("log-level"); level != "" {
		return level
	}

	// Then check shortcuts (most specific first)
	if cmd.Bool("quiet") {
		return "quiet"
	}

	if cmd.Bool("debug") {
		return "debug"
	}

	if cmd.Bool("verbose") {
		return "verbose"
	}

	return "brief" // default
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
