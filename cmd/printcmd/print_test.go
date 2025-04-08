// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2
package printcmd

import (
	"bytes"
	"context"
	"testing"

	"itiquette/git-provider-sync/internal/model"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
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

	return model.WithCLIOpt(ctx, model.CLIOption{})
}

func TestExecutePrintCommandNoArgNoConfPanics(t *testing.T) {
	require := require.New(t)

	// Save original error handler
	originalHandleError := model.HandleError
	defer func() { model.HandleError = originalHandleError }()

	// Setup error capture
	errorOutput := bytes.NewBufferString("")
	ctx := setupTestContext(errorOutput)

	// Configure test error handler
	model.HandleError = func(ctx context.Context, err error) {
		logger := zerolog.Ctx(ctx)
		logger.Error().Err(err).Msg("A fatal error occurred")
		panic(4)
	}

	// Setup and configure command
	cmd := setupTestCommand()
	_ = cmd.PersistentFlags().Set("config-file", "testdasadfasdfta/testconfig.yaml")
	cmd.Root().SetContext(ctx)
	cmd.SetErr(errorOutput)

	// Execute and verify
	require.Panics(func() {
		_ = cmd.Execute()
	}, "Expected command to exit")
	require.Contains(errorOutput.String(), "A fatal error occurred",
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
