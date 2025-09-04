// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package mancmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestRunManGeneration_ToStdout_OutputsManPage(t *testing.T) { //nolint:paralleltest // Manipulates global os.Stdout
	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer

	// Create a dummy command
	cmd := &cli.Command{Name: "man"}

	// Run the function
	err = runManGeneration(context.Background(), cmd)

	// Close writer and restore stdout
	_ = writer.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer

	_, copyErr := io.Copy(&buf, reader)
	require.NoError(t, copyErr)

	// Verify no error occurred
	require.NoError(t, err)

	// Verify output contains expected content
	output := buf.String()
	assert.Contains(t, output, "# gitprovidersync(1)")
	assert.Contains(t, output, "## NAME")
	assert.Contains(t, output, "## SYNOPSIS")
	assert.Contains(t, output, "## DESCRIPTION")
	assert.Contains(t, output, "## COMMANDS")
	assert.Contains(t, output, "## OPTIONS")
	assert.Contains(t, output, "## AUTHOR")
	assert.Contains(t, output, "## COPYRIGHT")
	assert.Contains(t, output, "gitprovidersync - Utility for mirroring and storing Git repositories")
	assert.Contains(t, output, "EUPL-1.2 license")
}

func TestManCommand_Execution_GeneratesDocumentation(t *testing.T) { //nolint:paralleltest // Manipulates global os.Stdout
	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer

	// Create and execute the command
	cmd := NewManCommand()
	err = cmd.Action(context.Background(), cmd)

	// Close writer and restore stdout
	_ = writer.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer

	_, copyErr := io.Copy(&buf, reader)
	require.NoError(t, copyErr)

	// Verify execution
	require.NoError(t, err)

	output := buf.String()
	assert.NotEmpty(t, output)
	assert.True(t, strings.HasPrefix(output, "# gitprovidersync(1)"))
}

func TestManContentStructure(t *testing.T) { //nolint:paralleltest // Manipulates global os.Stdout
	// Capture stdout
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer

	// Run the generation
	err = runManGeneration(context.Background(), &cli.Command{})

	// Close writer and restore stdout
	_ = writer.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer

	_, copyErr := io.Copy(&buf, reader)
	require.NoError(t, copyErr)

	require.NoError(t, err)

	output := buf.String()

	// Test that sections appear in expected order
	nameIndex := strings.Index(output, "## NAME")
	synopsisIndex := strings.Index(output, "## SYNOPSIS")
	descriptionIndex := strings.Index(output, "## DESCRIPTION")
	commandsIndex := strings.Index(output, "## COMMANDS")

	assert.Less(t, nameIndex, synopsisIndex, "NAME should come before SYNOPSIS")
	assert.Less(t, synopsisIndex, descriptionIndex, "SYNOPSIS should come before DESCRIPTION")
	assert.Less(t, descriptionIndex, commandsIndex, "DESCRIPTION should come before COMMANDS")

	// Test specific command descriptions
	assert.Contains(t, output, "**sync**")
	assert.Contains(t, output, "**print**")
	assert.Contains(t, output, "Mirror repositories from a source Git provider to targets")
	assert.Contains(t, output, "Print the current configuration")
}
