// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugWriter_NewDebugWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		enabled    bool
		expectFile bool
	}{
		{
			name:       "creates file when enabled",
			enabled:    true,
			expectFile: true,
		},
		{
			name:       "returns console when disabled",
			enabled:    false,
			expectFile: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var console bytes.Buffer

			writer, path, err := NewDebugWriter(&console, testCase.enabled)

			require.NoError(t, err)
			require.NotNil(t, writer)

			if testCase.expectFile {
				assert.NotEmpty(t, path)
				assert.Contains(t, path, "gps-debug-")

				// Clean up - close the writer but leave debug file in temp
				// Debug files in temp are fine - OS will clean them up
				if debugWriter, ok := writer.(*DebugWriter); ok {
					defer func() { _ = debugWriter.Close() }()
				}
			} else {
				assert.Empty(t, path)
				assert.Equal(t, &console, writer)
			}
		})
	}
}

func TestDebugWriter_Write(t *testing.T) {
	t.Parallel()

	var console bytes.Buffer

	writer, path, err := NewDebugWriter(&console, true)
	require.NoError(t, err)

	debugWriter, ok := writer.(*DebugWriter)
	require.True(t, ok)

	defer func() { _ = debugWriter.Close() }()

	// Debug files in temp are fine - OS will clean them up

	// Write test data
	testData := []byte("test log message\n")
	n, err := debugWriter.Write(testData)

	// Check write succeeded
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)

	// Check console got the data
	assert.Equal(t, "test log message\n", console.String())

	// Check file got the data
	if path != "" {
		content, err := os.ReadFile(path) //nolint:gosec // Test file path is controlled
		require.NoError(t, err)

		assert.Contains(t, string(content), "test log message")
	}
}

func TestDebugWriter_TeeWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		logLevel   string
		expectFile bool
	}{
		{
			name:       "debug level creates file",
			logLevel:   "debug",
			expectFile: true,
		},
		{
			name:       "trace level creates file",
			logLevel:   "trace",
			expectFile: true,
		},
		{
			name:       "info level no file",
			logLevel:   "info",
			expectFile: false,
		},
		{
			name:     "error level no file",
			logLevel: "error",

			expectFile: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var console bytes.Buffer

			writer, debugPath := TeeWriter(&console, testCase.logLevel)

			require.NotNil(t, writer)

			if testCase.expectFile {
				assert.NotEmpty(t, debugPath)
				assert.True(t, strings.HasSuffix(debugPath, ".log"))

				// Clean up - close the writer but leave debug file in temp
				if debugWriter, ok := writer.(*DebugWriter); ok {
					defer func() { _ = debugWriter.Close() }()
				}
			} else {
				assert.Empty(t, debugPath)
			}
		})
	}
}
