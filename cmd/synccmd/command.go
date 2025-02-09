// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// command.go - Command setup, entry points and error definitions
package synccmd

import (
	"context"

	baseOpt "itiquette/git-provider-sync/cmd/baseoption"
	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"

	"github.com/spf13/cobra"
)

func NewSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Mirror repositories from a source Git provider to targets",
		Long: `The 'sync' command mirrors your repositories from a source Git provider to one or more targets.
It allows for various options to control the synchronization process.`,
		Run: runSync,
	}

	addSyncInputOptions(cmd)

	return cmd
}

func runSync(cmd *cobra.Command, _ []string) {
	ctx := cmd.Root().Context()
	ctx = baseOpt.AddRootInputOptionsToContext(ctx, cmd)

	flags, err := getSyncInputOptions(ctx, cmd)
	model.HandleError(ctx, err)

	ctx = initLogger(ctx, cmd)
	ctx = addInputOptionsToContext(ctx, flags)

	config, err := configuration.DefaultConfigLoader{}.LoadConfiguration(ctx)
	model.HandleError(ctx, err)

	err = sync(ctx, config)
	model.HandleError(ctx, err)
}

func addInputOptionsToContext(ctx context.Context, flags *syncInputOption) context.Context {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering addInputOptionsToContext")
	flags.DebugLog(logger).Msg("addInputOptionsToContext")

	cliOpts := model.CLIOptions(ctx)

	cliOpts.AlphaNumHyphName = flags.alphaNumHyphName
	cliOpts.ActiveFromLimit = flags.activeFromLimit
	cliOpts.DryRun = flags.dryRun
	cliOpts.ForcePush = flags.forcePush
	cliOpts.IgnoreInvalidName = flags.ignoreInvalidName

	return model.WithCLIOpt(ctx, cliOpts)
}

func initLogger(ctx context.Context, cmd *cobra.Command) context.Context {
	withCaller := model.CLIOptions(ctx).VerbosityWithCaller
	outputFormat := model.CLIOptions(ctx).OutputFormat

	ctx = log.InitLogger(ctx, cmd, withCaller, outputFormat)
	log.Logger(ctx).Trace().Msg("Logger initialized")

	return ctx
}
