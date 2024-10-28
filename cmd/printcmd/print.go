// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

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
// It can be modified for testing purposes.
var configPrintWriter io.Writer = os.Stdout

// NewPrintCommand creates and returns a new cobra.Command for the 'print' subcommand.
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
// It loads the configuration, initializes the logger, and prints the configuration.
func runPrint(cmd *cobra.Command, _ []string) {
	ctx := cmd.Root().Context()
	ctx = baseoption.AddRootInputOptionsToContext(ctx, cmd)

	withCaller := model.CLIOptions(ctx).VerbosityWithCaller
	ctx = log.InitLogger(ctx, cmd, withCaller, string(log.CONSOLE))

	var configLoaderInstance configuration.ConfigLoader = configuration.DefaultConfigLoader{}

	conf, err := configLoaderInstance.LoadConfiguration(ctx)
	if err != nil {
		model.HandleError(ctx, err)
	}

	configuration.PrintConfiguration(*conf, configPrintWriter)
}
