// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// command.go - Command setup, entry points (restored from main, hexagonal adapters)
package synccmd

import (
	"context"

	baseOpt "itiquette/git-provider-sync/cmd/baseoption"
	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/log"

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

// runSync executes sync using simple functional approach from main branch.
func runSync(cmd *cobra.Command, _ []string) {
	ctx := cmd.Root().Context()
	cliConfig, err := baseOpt.ExtractRootInputOptions(cmd)

	if err != nil {
		// TODO: Implement proper error handling via CLI adapter
		return
	}

	ctx = cli.WithCLIConfig(ctx, cliConfig)

	flags, err := getSyncInputOptions(ctx, cmd)
	if err != nil {
		// TODO: Implement proper error handling via CLI adapter
		return
	}

	ctx = initLogger(ctx, cmd)
	ctx = addInputOptionsToContext(ctx, flags)

	// Use original proven configuration loader
	config, err := configuration.DefaultConfigLoader{}.LoadConfiguration(ctx)
	if err != nil {
		// TODO: Implement proper error handling via CLI adapter
		return
	}

	// Execute sync using proper hexagonal architecture
	err = syncHexagonal(ctx, config)
	if err != nil {
		// TODO: Implement proper error handling via CLI adapter
		return
	}
}

// addInputOptionsToContext adds sync input options to context (restored from main).
func addInputOptionsToContext(ctx context.Context, flags *syncInputOption) context.Context {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering addInputOptionsToContext")
	flags.DebugLog(logger).Msg("addInputOptionsToContext")

	// Get existing CLI config or create default
	cliConfig, ok := cli.CLIConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	// Build new config with updated values
	updatedConfig := entities.NewCLIConfigBuilder().
		WithAlphaNumHyphName(flags.alphaNumHyphName).
		WithActiveFromLimit(flags.activeFromLimit).
		WithDryRun(flags.dryRun).
		WithForcePush(flags.forcePush).
		WithIgnoreInvalidName(flags.ignoreInvalidName).
		WithOutputFormat(cliConfig.OutputFormat()).
		WithVerbosityWithCaller(cliConfig.VerbosityWithCaller()).
		WithQuiet(cliConfig.Quiet()).
		WithConfigFilePath(cliConfig.ConfigFilePath()).
		WithConfigFileOnly(cliConfig.ConfigFileOnly()).
		Build()

	return cli.WithCLIConfig(ctx, updatedConfig)
}

// initLogger initializes logging with CLI options (restored from main).
func initLogger(ctx context.Context, cmd *cobra.Command) context.Context {
	cliConfig, ok := cli.CLIConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	withCaller := cliConfig.VerbosityWithCaller()
	outputFormat := cliConfig.OutputFormat()

	ctx = log.InitLogger(ctx, cmd, withCaller, outputFormat)
	log.Logger(ctx).Trace().Msg("Logger initialized")

	return ctx
}
