// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// command.go - Command setup, entry points and error definitions
package synccmd

import (
	"context"
	"errors"

	"itiquette/git-provider-sync/cmd/baseoption"
	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"

	"github.com/spf13/cobra"
)

// Package-level sentinel errors.
var (
	ErrInvalidRepoName    = errors.New("invalid repository name")
	ErrEmptyMetainfo      = errors.New("empty repository metainfo")
	ErrMissingSyncRunMeta = errors.New("missing sync run metadata")
)

func NewSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Mirror repositories from a source Git provider to targets",
		Long: `The 'sync' command mirrors your repositories from a source Git provider to one or more targets.
It allows for various options to control the synchronization process.`,
		Run: runSync,
	}

	addSyncFlags(cmd)

	return cmd
}

func runSync(cmd *cobra.Command, _ []string) {
	ctx := cmd.Root().Context()
	ctx = baseoption.AddRootInputOptionsToContext(ctx, cmd)

	flags, err := getSyncFlags(ctx, cmd)
	model.HandleError(ctx, err)

	ctx = initLogger(ctx, cmd)
	ctx = addFlagsToContext(ctx, flags)

	config, err := configuration.DefaultConfigLoader{}.LoadConfiguration(ctx)
	model.HandleError(ctx, err)

	err = sync(ctx, config)
	model.HandleError(ctx, err)
}

func addFlagsToContext(ctx context.Context, flags *syncFlags) context.Context {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering addInputOptionsToContext")
	flags.DebugLog(logger).Msg("addInputOptionsToContext")

	opts := model.CLIOptions(ctx)

	opts.ASCIIName = flags.asciiName
	opts.ActiveFromLimit = flags.activeFromLimit
	opts.DryRun = flags.dryRun
	opts.ForcePush = flags.forcePush
	opts.IgnoreInvalidName = flags.ignoreInvalidName

	return model.WithCLIOption(ctx, opts)
}

func initLogger(ctx context.Context, cmd *cobra.Command) context.Context {
	withCaller := model.CLIOptions(ctx).VerbosityWithCaller
	outputFormat := model.CLIOptions(ctx).OutputFormat

	ctx = log.InitLogger(ctx, cmd, withCaller, outputFormat)
	log.Logger(ctx).Trace().Msg("Logger initialized")

	return ctx
}
