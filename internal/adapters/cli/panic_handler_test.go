// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPanicHandler_HandlePanic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		panicValue     any
		expectInOutput []string
	}{
		{
			name:       "handles string panic",
			panicValue: "test panic message",
			expectInOutput: []string{
				"Unexpected Error",
				"test panic message",
				"What to do next:",
				"Report this issue:",
			},
		},
		{
			name:       "handles error panic",
			panicValue: assert.AnError,
			expectInOutput: []string{
				"Unexpected Error",
				"assert.AnError",
				"What to do next:",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create buffer to capture output
			var buf bytes.Buffer

			handler := NewPanicHandler(&buf, "test-version")

			// Simulate panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						handler.handlePanicRecovery(r)
					}
				}()

				panic(testCase.panicValue)
			}()

			// Check output contains expected strings
			output := buf.String()
			for _, expected := range testCase.expectInOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestPanicHandler_GenerateBugReportURL(t *testing.T) {
	t.Parallel()

	handler := NewPanicHandler(nil, "v1.0.0")

	urlStr := handler.generateBugReportURL("test error", "/tmp/crash.txt")

	// Parse URL to verify structure
	parsedURL, err := url.Parse(urlStr)
	require.NoError(t, err, "URL should be valid")

	// Check base URL
	assert.Equal(t, "https", parsedURL.Scheme)
	assert.Equal(t, "github.com", parsedURL.Host)
	assert.Equal(t, "/itiquette/git-provider-sync/issues/new", parsedURL.Path)

	// Parse query parameters
	params := parsedURL.Query()

	// Check required parameters
	assert.Equal(t, "bug", params.Get("labels"))
	assert.Equal(t, "bug_report.yml", params.Get("template"))

	// Check title format
	title := params.Get("title")
	assert.Contains(t, title, "[BUG]")
	assert.Contains(t, title, "test error")

	// Check body content
	body := params.Get("body")
	assert.Contains(t, body, "## Description")
	assert.Contains(t, body, "## Error Details")
	assert.Contains(t, body, "## System Information")
	assert.Contains(t, body, "v1.0.0")
	assert.Contains(t, body, "/tmp/crash.txt")
}

func TestPanicHandler_FormatIssueBody(t *testing.T) {
	t.Parallel()

	handler := NewPanicHandler(nil, "v1.0.0")

	body := handler.formatIssueBody("test panic", "/tmp/crash.txt")

	// Check markdown sections
	assert.Contains(t, body, "## Description")
	assert.Contains(t, body, "Unexpected error occurred: test panic")
	assert.Contains(t, body, "## Error Details")
	assert.Contains(t, body, "```\ntest panic\n```")
	assert.Contains(t, body, "## System Information")
	assert.Contains(t, body, "- Version: v1.0.0")
	assert.Contains(t, body, "## Crash Report")
	assert.Contains(t, body, "/tmp/crash.txt")
	assert.Contains(t, body, "## Additional Context")
}

func TestPanicHandler_SaveCrashReport(t *testing.T) {
	t.Parallel()

	handler := NewPanicHandler(nil, "v1.0.0")

	// Test saving crash report
	crashPath := handler.saveCrashReport("test panic", []byte("stack trace here"))

	// If path is returned, file was created
	if crashPath != "" {
		assert.True(t, strings.HasSuffix(crashPath, ".txt"))

		assert.Contains(t, crashPath, "crash-")
	}
}

func TestPanicHandler_FormatCrashReport(t *testing.T) {
	t.Parallel()

	handler := NewPanicHandler(nil, "v1.0.0")

	report := handler.formatCrashReport("test panic", []byte("stack trace"))

	// Check report contains expected sections
	assert.Contains(t, report, "Git Provider Sync Crash Report")
	assert.Contains(t, report, "Version: v1.0.0")
	assert.Contains(t, report, "Panic: test panic")
	assert.Contains(t, report, "Stack Trace:")
	assert.Contains(t, report, "stack trace")
}
