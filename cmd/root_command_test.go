// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteWithValidArgs tests the Execute function with various argument combinations.
func TestExecute_WithValidArgs_CompletesSuccessfully(t *testing.T) { //nolint:paralleltest // Manipulates global os.Args
	// Save original args
	originalArgs := os.Args

	defer func() {
		os.Args = originalArgs
	}()

	// Mock os.Exit to capture exit calls (not actually needed for these tests)
	var exitCode int

	exitCalled := false
	_ = exitCode   // avoid unused variable warning
	_ = exitCalled // avoid unused variable warning

	tests := []struct {
		name         string
		args         []string
		expectExit   bool
		expectedCode int
	}{
		{
			name:         "help argument",
			args:         []string{"gitprovidersync", "--help"},
			expectExit:   false,
			expectedCode: 0,
		},
		{
			name:         "version argument",
			args:         []string{"gitprovidersync", "--version"},
			expectExit:   false,
			expectedCode: 0,
		},
		{
			name:         "invalid flag returns error",
			args:         []string{"gitprovidersync", "--nonexistent-flag"},
			expectExit:   true,
			expectedCode: 1,
		},
	}

	for _, testCase := range tests { //nolint:paralleltest // Manipulates global os.Args
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			versionString := "test-version (Commit SHA: test-commit, Build date: test-date)"
			rootCmd := NewRootCommandForTesting(ctx, versionString)

			err := rootCmd.Run(ctx, testCase.args)

			if testCase.expectExit {
				// For invalid commands, urfave/cli returns an error
				// We verify that an error is returned, which indicates the command failed appropriately
				require.Error(t, err)
				t.Logf("Command correctly returned error for invalid input: %v", err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestExecuteVersionString tests that version information is properly formatted.
func TestExecuteVersionString(t *testing.T) { //nolint:paralleltest // Manipulates global os.Args
	originalArgs := os.Args

	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"gitprovidersync", "--version"}

	// This should not panic and should execute without error
	assert.NotPanics(t, func() {
		RunApplication("v1.0.0", "abc123", "2024-01-01T12:00:00Z")
	})
}

// TestNewRootCommand tests the root command creation.
func TestNewRootCommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	versionString := "test-version"

	cmd := NewRootCommand(ctx, versionString)

	require.NotNil(t, cmd)
	assert.Equal(t, "gitprovidersync", cmd.Name)
	assert.Contains(t, cmd.Description, "utility")
	assert.NotEmpty(t, cmd.Commands, "Should have subcommands")

	// Verify expected subcommands exist
	subcommandNames := make([]string, 0, len(cmd.Commands))
	for _, subCmd := range cmd.Commands {
		subcommandNames = append(subcommandNames, subCmd.Name)
	}

	expectedCommands := []string{"sync", "status", "print", "man"}
	for _, expectedCmd := range expectedCommands {
		assert.Contains(t, subcommandNames, expectedCmd)
	}
}
