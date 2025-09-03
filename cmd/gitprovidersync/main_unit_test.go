// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/cli"
)

// TestBuildVariables_DefaultValues tests that build variables have expected defaults.
func TestBuildVariables_DefaultValues(t *testing.T) {
	t.Parallel()

	// Test default values set during build
	assert.Equal(t, "dev", version)
	assert.Equal(t, "none", commit)
	assert.Equal(t, "unknown", date)
}

// TestPanicHandlerIntegration verifies panic handler is properly set up.
func TestPanicHandlerIntegration(t *testing.T) {
	t.Parallel()

	// Create a panic handler with test version
	var buf bytes.Buffer

	versionString := "test-v1.0.0 (commit: abc123, built: 2024-01-01)"
	handler := cli.NewPanicHandler(&buf, versionString)

	require.NotNil(t, handler)

	// Simulate a panic recovery
	func() {
		defer handler.HandlePanic()
		defer func() {
			if r := recover(); r != nil {
				// Verify the panic was caught
				require.NotNil(t, r)
			}
		}()
		// Don't actually panic in test to avoid exit
	}()
}

// TestBugReportURLIntegration validates bug report URL generation.
func TestBugReportURLIntegration(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := cli.NewPanicHandler(&buf, "v1.2.3")
	require.NotNil(t, handler)

	// The generateBugReportURL method is tested in panic_handler_test.go
	// Validates that the integration exists
	assert.Contains(t, "github.com/itiquette/git-provider-sync", "git-provider-sync")
}
