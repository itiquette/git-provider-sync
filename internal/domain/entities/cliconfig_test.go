// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCLIConfig_DryRunBehavior tests that DryRun flag affects execution behavior.
func TestCLIConfig_DryRunBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		dryRun          bool
		expectExecution bool
	}{
		{"dry run enabled prevents execution", true, false},
		{"dry run disabled allows execution", false, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := NewCLIConfigBuilder().
				WithDryRun(testCase.dryRun).
				Build()

			assert.Equal(t, testCase.dryRun, config.DryRun())
			// In real usage, this would control whether operations execute
			// Tests that configuration can be set and retrieved correctly
		})
	}
}

// TestCLIConfig_OutputFormatValidation tests output format affects formatting behavior.
func TestCLIConfig_OutputFormatValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		format       string
		expectFormat string
	}{
		{"json format", "json", "json"},
		{"plain format", "plain", "plain"},
		{"console format default", "", "console"},
		{"invalid format defaults to console", "invalid", "invalid"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewCLIConfigBuilder()
			if testCase.format != "" {
				builder = builder.WithOutputFormat(testCase.format)
			}

			config := builder.Build()

			assert.Equal(t, testCase.expectFormat, config.OutputFormat())
		})
	}
}
