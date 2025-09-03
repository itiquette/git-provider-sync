// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package baseoption provides base command option types.
package baseoption

import (
	"errors"

	"github.com/urfave/cli/v3"

	"itiquette/git-provider-sync/internal/adapters/terminal"
	"itiquette/git-provider-sync/internal/domain/entities"
)

// ExtractRootInputOptions creates CLI configuration from command flags.
func ExtractRootInputOptions(cmd *cli.Command) (entities.CLIConfig, error) {
	// Check for conflicting log level flags
	if hasConflictingLogFlags(cmd) {
		return entities.CLIConfig{}, errors.New("cannot use multiple log level flags together (--quiet, --verbose, --debug, --log-level)")
	}

	configFilePath := cmd.String("config-file")
	configFileOnly := cmd.Bool("config-file-only")
	logCaller := cmd.Bool("log-caller")
	outputFormat := cmd.String("format")
	plain := cmd.Bool("plain")
	json := cmd.Bool("json")
	colorMode := cmd.String("color")

	// Priority: explicit format > plain flag > json flag > auto-detect
	// Be idiomatic: simple precedence rules
	if outputFormat == "" {
		switch {
		case plain:
			outputFormat = "plain"
		case json:
			outputFormat = "json"
		case terminal.IsOutput():
			outputFormat = "console" // Human-readable for TTY
		default:
			outputFormat = "plain" // Machine-readable for pipes
		}
	}

	// Build CLI configuration using functional builder
	cliConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath(configFilePath).
		WithConfigFileOnly(configFileOnly).
		WithVerbosityWithCaller(logCaller).
		WithOutputFormat(outputFormat).
		WithColorMode(colorMode).
		Build()

	return cliConfig, nil
}

// hasConflictingLogFlags checks if multiple log level flags are set.
// Returns true if there's a conflict that would be confusing.
func hasConflictingLogFlags(cmd *cli.Command) bool {
	count := 0

	// Count how many log-related flags are set
	if cmd.String("log-level") != "" {
		count++
	}

	if cmd.Bool("quiet") {
		count++
	}

	if cmd.Bool("verbose") {
		count++
	}

	if cmd.Bool("debug") {
		count++
	}

	// More than one is a conflict
	return count > 1
}
