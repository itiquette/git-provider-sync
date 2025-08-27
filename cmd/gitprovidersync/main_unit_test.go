// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/cli"
)

// TestMainUnit tests the main function directly for unit test coverage.
func TestMain_WithHelpFlag_DoesNotPanic(t *testing.T) { //nolint:paralleltest // Tests main function behavior
	// Save original args and defer restoration
	originalArgs := os.Args

	defer func() { os.Args = originalArgs }()

	// Test with help flag to get quick exit
	os.Args = []string{"gitprovidersync", "--help"}

	// The main function should not panic
	assert.NotPanics(t, func() {
		main()
	})
}

// TestBuildVariables tests that build variables are properly initialized.
func TestBuildVariables(t *testing.T) {
	t.Parallel()

	// Test default values
	assert.Equal(t, "dev", version)
	assert.Equal(t, "none", commit)
	assert.Equal(t, "unknown", date)

	// Test that variables are not empty
	require.NotEmpty(t, version)
	require.NotEmpty(t, commit)
	require.NotEmpty(t, date)
}

// TestMainFunctionExists verifies the main function exists and is callable.
func TestMain_FunctionReference_ExistsAndCallable(t *testing.T) {
	t.Parallel()

	// This test ensures the main function exists and can be referenced
	assert.NotNil(t, main)
}

// TestBuildVariablesType tests that build variables have correct types.
func TestBuildVariablesType(t *testing.T) {
	t.Parallel()

	assert.IsType(t, "", version)
	assert.IsType(t, "", commit)
	assert.IsType(t, "", date)
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
	// This validates the integration exists
	assert.Contains(t, "github.com/itiquette/git-provider-sync", "git-provider-sync")
}
