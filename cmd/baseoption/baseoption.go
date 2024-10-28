// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package baseoption

import (
	"context"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"

	"github.com/spf13/cobra"
)

// AddRootInputOptionsToContext adds CLI options to the context.
func AddRootInputOptionsToContext(ctx context.Context, cmd *cobra.Command) context.Context {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering addRootInputOptionsToContext:")

	configFilePath, err := cmd.Flags().GetString("config-file")
	model.HandleError(ctx, err)
	configFileOnly, err := cmd.Flags().GetBool("config-file-only")
	model.HandleError(ctx, err)
	verbosityWithCaller, err := cmd.Flags().GetBool("verbosity-with-caller")
	model.HandleError(ctx, err)

	outputFormat, err := cmd.Flags().GetString("output-format")
	model.HandleError(ctx, err)

	cliOption := model.CLIOption{
		ConfigFilePath:      configFilePath,
		ConfigFileOnly:      configFileOnly,
		VerbosityWithCaller: verbosityWithCaller,
		OutputFormat:        outputFormat,
	}

	return model.WithCLIOption(ctx, cliOption)
}
