// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package printcmd provides functionality to print Git Provider Sync configuration.
// It allows users to view their current configuration settings in a readable format.
package printcmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"itiquette/git-provider-sync/cmd/baseoption"
	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/log"
)

// configPrintWriter is the default writer for configuration output.
// It defaults to os.Stdout but can be modified for testing purposes.
// This variable should only be modified in tests.
var configPrintWriter io.Writer = os.Stdout

// NewPrintCommand creates and returns a new cobra.Command for the 'print' subcommand.
// It displays the current Git Provider Sync configuration using the default formatter.
//
// Example usage:
//
//	git-provider-sync print
//
// The command will output the full configuration including all sources
// and their respective settings.
func NewPrintCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: "Print the current configuration",
		Long: `The 'print' command outputs the current, aggregated Git Provider Sync configuration to stdout.
It loads the configuration from available sources and displays it in a formatted manner.`,
		Run: runPrint,
	}
}

// runPrint executes the logic for the 'print' command.
// Uses pure functional error handling without side effects.
func runPrint(cmd *cobra.Command, _ []string) {
	ctx := cmd.Root().Context()
	logger := log.Logger(ctx)

	// Create CLI configuration functionally
	cliConfig, err := baseoption.ExtractRootInputOptions(cmd)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create CLI configuration")
		return
	}

	ctx = cli.WithCLIConfig(ctx, cliConfig)

	// CLI config already added to context by baseOpt.AddRootInputOptionsToContext
	// Initialize logger using proven approach
	ctx = log.InitLogger(ctx, cmd, false, "console")

	configLoaderInstance := configuration.DefaultConfigLoader{}

	conf, err := configLoaderInstance.LoadConfiguration(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load configuration")

		return
	}

	configuration.PrintConfiguration(*conf, configPrintWriter)
}
