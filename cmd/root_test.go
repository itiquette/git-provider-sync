// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand_RegistersAllRequiredSubcommands(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cmd := NewRootCommandForTesting(ctx, "tests")

	// Verify expected subcommands exist
	require.Len(t, cmd.Commands, 4, "Should have 4 subcommands")

	subCmdNames := make([]string, 0, len(cmd.Commands))
	for _, v := range cmd.Commands {
		subCmdNames = append(subCmdNames, v.Name)
	}

	// Test behavior: command has the required subcommands
	assert.Contains(t, subCmdNames, "sync", "Missing sync subcommand")
	assert.Contains(t, subCmdNames, "status", "Missing status subcommand")
	assert.Contains(t, subCmdNames, "print", "Missing print subcommand")
	assert.Contains(t, subCmdNames, "man", "Missing man subcommand")
}

func TestRootCommand_SetsCorrectMetadataAndDescription(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cmd := NewRootCommandForTesting(ctx, "v1.0.0")

	// Test metadata is properly set
	assert.Equal(t, "gitprovidersync", cmd.Name)
	assert.Contains(t, cmd.Description, "utility", "Description should mention utility")
}
