// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package printcmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPrintCommand tests the standard print command constructor.
func TestNewPrintCommand_Constructor_CreatesCommandWithDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewPrintCommand()

	require.NotNil(t, cmd)
	assert.Equal(t, "print", cmd.Name)
	assert.Equal(t, "Print the current configuration", cmd.Usage)
	assert.Contains(t, cmd.Description, "print")
	assert.NotNil(t, cmd.Action)
	assert.False(t, cmd.Hidden)
}

// TestNewPrintCommandWithWriter tests the print command constructor with custom writer.
func TestNewPrintCommand_WithCustomWriter_CreatesCommandCorrectly(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := NewPrintCommandWithWriter(&buf)

	require.NotNil(t, cmd)
	assert.Equal(t, "print", cmd.Name)
	assert.Equal(t, "Print the current configuration", cmd.Usage)
	assert.Contains(t, cmd.Description, "print")
	assert.NotNil(t, cmd.Action)
	assert.False(t, cmd.Hidden)

	// Verify the command has expected flags
	flags := cmd.Flags
	assert.NotEmpty(t, flags, "Print command should have flags")

	// Check for connectivity-check flag
	var hasConnectivityFlag bool

	for _, flag := range flags {
		if flag.Names()[0] == "connectivity-check" {
			hasConnectivityFlag = true

			break
		}
	}

	assert.True(t, hasConnectivityFlag, "Should have connectivity-check flag")
}

// TestNewPrintCommandConstructorConsistency tests that both constructors create consistent commands.
func TestNewPrintCommandConstructorConsistency(t *testing.T) {
	t.Parallel()

	// Create commands using both constructors
	standardCmd := NewPrintCommand()

	var buf bytes.Buffer

	writerCmd := NewPrintCommandWithWriter(&buf)

	// They should have the same basic properties
	assert.Equal(t, standardCmd.Name, writerCmd.Name)
	assert.Equal(t, standardCmd.Usage, writerCmd.Usage)
	assert.Equal(t, standardCmd.Description, writerCmd.Description)
	assert.Equal(t, standardCmd.Hidden, writerCmd.Hidden)
	assert.Len(t, writerCmd.Flags, len(standardCmd.Flags))
}

// TestNewPrintCommandWithStdout tests that NewPrintCommand uses stdout by default.
func TestNewPrintCommand_UsesStdout_ByDefault(t *testing.T) { //nolint:paralleltest // Tests stdout behavior
	// This test verifies that NewPrintCommand() creates a command that would use stdout
	// We can't directly test the writer without executing the command, but we can verify
	// the command structure is correct
	cmd := NewPrintCommand()

	// Verify command structure
	require.NotNil(t, cmd)
	assert.Equal(t, "print", cmd.Name)

	// The actual writer (os.Stdout) is used internally by NewPrintCommandWithWriter
	// This test ensures the constructor chain works correctly
	assert.NotNil(t, cmd.Action, "Command should have an action function")
}

// TestPrintCommandFlags tests that the print command has the expected flags.
func TestPrintCommandFlags(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := NewPrintCommandWithWriter(&buf)

	// Collect flag names for easier testing
	flagNames := make([]string, 0, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		flagNames = append(flagNames, flag.Names()[0])
	}

	// Verify expected flags exist
	expectedFlags := []string{"connectivity-check"}
	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag, "Should have %s flag", expectedFlag)
	}
}

// TestPrintCommandDescription tests the command description content.
func TestPrintCommandDescription(t *testing.T) {
	t.Parallel()

	cmd := NewPrintCommand()

	description := cmd.Description
	assert.Contains(t, description, "configuration", "Description should mention configuration")
	assert.Contains(t, description, "connectivity-check", "Description should mention connectivity-check option")
	assert.Contains(t, description, "output", "Description should mention output")
}
