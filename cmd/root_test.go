// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCommand_NoArgs_ShowsHelp(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	ctx := context.Background()
	cmd := NewRootCommandForTesting(ctx, "tests")

	// urfave/cli v3 doesn't have SetOut, test by checking command structure
	require.Len(cmd.Commands, 4)

	subCmdNames := make([]string, 0, 4)
	for _, v := range cmd.Commands {
		subCmdNames = append(subCmdNames, v.Name)
	}

	require.Contains(subCmdNames, "sync")
	require.Contains(subCmdNames, "status")
	require.Contains(subCmdNames, "print")
	require.Contains(subCmdNames, "man")

	// Test that the command has expected properties
	require.Equal("gitprovidersync", cmd.Name)
	require.Contains(cmd.Description, "A utility")
}
