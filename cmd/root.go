// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"

	"itiquette/git-provider-sync/cmd/mancmd"
	"itiquette/git-provider-sync/cmd/printcmd"
	"itiquette/git-provider-sync/cmd/synccmd"
	"itiquette/git-provider-sync/internal/model"

	"github.com/spf13/cobra"
)

// newRootCommand creates and returns the root command for the Git Provider Sync CLI.
func newRootCommand(ctx context.Context, versionString string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gitprovidersync",
		Version: versionString,
		Short:   "Utility for mirroring and storing Git repositories",
		Long: `A utility for mirroring Git repositories to various Git providers or storage.
Supports GitHub, Gitea, GitLab, uncompressed directories, and a compressed archive format (tar.gz).
Allows syncing to multiple mirror destinations.`,
	}

	// Add persistent flags
	rootCmd.PersistentFlags().String("verbosity", "brief", "Set output verbosity: quiet | brief | verbose | debug | trace")
	rootCmd.PersistentFlags().Bool("verbosity-with-caller", false, "Add caller information to verbosity output (for development)")
	rootCmd.PersistentFlags().Bool("quiet", false, "Equivalent to --verbosity=quiet. Only output errors")
	rootCmd.PersistentFlags().String("config-file", "gitprovidersync.yaml", "Path to the configuration file")
	rootCmd.PersistentFlags().Bool("config-file-only", false, "Read configuration from file only (ignore ENV, dotenv, XDG_CONFIG_HOME)")
	rootCmd.PersistentFlags().String("output-format", "console", "Output format (console,json)")

	rootCmd.SetContext(ctx)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Add subcommands,
	rootCmd.AddCommand(mancmd.NewManCommand(), printcmd.NewPrintCommand(), synccmd.NewSyncCommand())

	return rootCmd
}

// Execute runs the root command with the provided version information.
func Execute(version, commitSHA, buildDate string) {
	ctx := context.Background()

	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"
	err := newRootCommand(ctx, versionString).Execute()
	model.HandleError(ctx, err)
}
