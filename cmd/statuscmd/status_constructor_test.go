// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package statuscmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStatusCommand tests the status command constructor.
func TestNewStatusCommand_Constructor_CreatesCommandWithCorrectProperties(t *testing.T) {
	t.Parallel()

	cmd := NewStatusCommand()

	require.NotNil(t, cmd)
	assert.Equal(t, "status", cmd.Name)
	assert.Equal(t, "Show system status and suggest next actions", cmd.Usage)
	assert.Contains(t, cmd.Description, "status")
	assert.NotNil(t, cmd.Action)
	assert.False(t, cmd.Hidden)
}

// TestStatusCommandFlags tests that the status command has the expected flags.
func TestStatusCommand_Flags_ContainsExpectedOptions(t *testing.T) {
	t.Parallel()

	cmd := NewStatusCommand()

	// Collect flag names for easier testing
	flagNames := make([]string, 0, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		flagNames = append(flagNames, flag.Names()[0])
	}

	// Verify expected flags exist
	expectedFlags := []string{"connectivity-check", "skip-suggestions"}
	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag, "Should have %s flag", expectedFlag)
	}

	// Verify we have the correct number of flags
	assert.Len(t, cmd.Flags, 2, "Should have exactly 2 flags")
}

// TestStatusCommandDescription tests the command description content.
func TestStatusCommand_Description_ContainsRelevantText(t *testing.T) {
	t.Parallel()

	cmd := NewStatusCommand()

	description := cmd.Description
	assert.Contains(t, description, "status", "Description should mention status")
	assert.Contains(t, description, "configuration", "Description should mention configuration")
	assert.Contains(t, description, "connectivity", "Description should mention connectivity")
	assert.Contains(t, description, "suggestions", "Description should mention suggestions")
}
