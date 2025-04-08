// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package printcmd provides functionality to print Git Provider Sync configuration.
// It allows users to view their current configuration settings in a readable format.
package printcmd

import (
	"io"
	"os"

	"itiquette/git-provider-sync/cmd/baseoption"
	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"

	"github.com/spf13/cobra"
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
// Any errors encountered during execution are handled using model.HandleError.
func runPrint(cmd *cobra.Command, _ []string) {
	ctx := cmd.Root().Context()
	ctx = baseoption.AddRootInputOptionsToContext(ctx, cmd)
	opts := model.CLIOptions(ctx)

	ctx = log.InitLogger(ctx, cmd, opts.VerbosityWithCaller, string(log.CONSOLE))

	configLoaderInstance := configuration.DefaultConfigLoader{}

	conf, err := configLoaderInstance.LoadConfiguration(ctx)
	if err != nil {
		model.HandleError(ctx, err)
	}

	configuration.PrintConfiguration(*conf, configPrintWriter)
}
