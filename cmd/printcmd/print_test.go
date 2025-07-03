// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2
package printcmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/domain/entities"
)

// setupTestCommand creates a new command with standard test flags.
func setupTestCommand() *cobra.Command {
	cmd := NewPrintCommand()
	cmd.PersistentFlags().String("config-file", "", "path to a git provider sync configuration file")
	cmd.PersistentFlags().Bool("config-file-only", false, "read configuration from file only")
	cmd.PersistentFlags().Bool("verbosity-with-caller", false, "")
	cmd.PersistentFlags().String("output-format", "co", "")

	return cmd
}

// setupTestContext creates a context with zerolog logger writing to provided buffer.
func setupTestContext(output *bytes.Buffer) context.Context {
	logger := zerolog.New(output).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	// Create CLI config using domain entities
	cliConfig := entities.NewCLIConfigBuilder().Build()

	return cli.WithCLIConfig(ctx, cliConfig)
}

func TestExecutePrintCommandNoArgNoConfPanics(t *testing.T) {
	require := require.New(t)

	// Setup error capture
	errorOutput := bytes.NewBufferString("")
	ctx := setupTestContext(errorOutput)

	// Setup and configure command with invalid config path
	cmd := setupTestCommand()
	_ = cmd.PersistentFlags().Set("config-file", "testdasadfasdfta/testconfig.yaml")
	cmd.Root().SetContext(ctx)
	cmd.SetErr(errorOutput)

	// Execute command - should handle errors gracefully now
	_ = cmd.Execute()

	// Verify error was logged (functional error handling doesn't panic)
	require.Contains(errorOutput.String(), "Failed to load configuration",
		"Expected error message in stderr")
}

func TestExecutePrintCommandFileConfArgSuccess(t *testing.T) {
	require := require.New(t)

	// Backup and restore configPrintWriter
	originalWriter := configPrintWriter
	testBuffer := new(bytes.Buffer)
	configPrintWriter = testBuffer

	defer func() { configPrintWriter = originalWriter }()

	// Setup and configure command
	cmd := setupTestCommand()
	_ = cmd.PersistentFlags().Set("config-file", "testdata/testconfig.yaml")
	require.Empty(cmd.Commands())

	// Setup context and execute
	ctx := context.Background()
	cmd.Root().SetContext(ctx)
	_ = cmd.Execute()

	// Verify output
	require.Contains(testBuffer.String(), "Sync Configuration",
		"Expected configuration output")
}
