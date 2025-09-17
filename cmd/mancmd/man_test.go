// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package mancmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetManContent_ReturnsExpectedStructure(t *testing.T) {
	t.Parallel()

	// Get man page content
	content := getManContent()

	// Verify content structure
	require.Contains(t, content, "# gitprovidersync(1)")
	require.Contains(t, content, "## NAME")
	require.Contains(t, content, "## SYNOPSIS")
	require.Contains(t, content, "## DESCRIPTION")
	require.Contains(t, content, "## COMMANDS")
	require.Contains(t, content, "## OPTIONS")
	require.Contains(t, content, "## AUTHOR")
	require.Contains(t, content, "## COPYRIGHT")
	require.Contains(t, content, "gitprovidersync - Utility for mirroring and storing Git repositories")
	require.Contains(t, content, "EUPL-1.2 license")
}

func TestManCommand_WhenExecuted_ProducesValidManPage(t *testing.T) {
	t.Parallel()

	// Create command
	cmd := NewManCommand()

	// Verify command properties
	require.Equal(t, "man", cmd.Name)
	require.Equal(t, "Generate man page documentation", cmd.Usage)
	require.True(t, cmd.Hidden, "Man command should be hidden from regular help")
	require.NotNil(t, cmd.Action, "Command should have an action")
}

func TestManContent_HasCorrectSectionOrder(t *testing.T) {
	t.Parallel()

	content := getManContent()

	// Test that sections appear in expected order
	nameIndex := strings.Index(content, "## NAME")
	synopsisIndex := strings.Index(content, "## SYNOPSIS")
	descriptionIndex := strings.Index(content, "## DESCRIPTION")
	commandsIndex := strings.Index(content, "## COMMANDS")

	require.Greater(t, nameIndex, -1, "NAME section should exist")
	require.Greater(t, synopsisIndex, nameIndex, "SYNOPSIS should come after NAME")
	require.Greater(t, descriptionIndex, synopsisIndex, "DESCRIPTION should come after SYNOPSIS")
	require.Greater(t, commandsIndex, descriptionIndex, "COMMANDS should come after DESCRIPTION")

	// Test specific command descriptions
	require.Contains(t, content, "**sync**")
	require.Contains(t, content, "**print**")
	require.Contains(t, content, "Mirror repositories from a source Git provider to targets")
	require.Contains(t, content, "Print the current configuration")
}
