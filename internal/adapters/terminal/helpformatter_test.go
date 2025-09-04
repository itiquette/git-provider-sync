// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package terminal

import (
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
)

func TestHelpFormatterWithColorSupport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		colorEnabled   bool
		input          string
		expectedOutput string
		description    string
	}{
		{
			name:           "bold_with_color_enabled",
			colorEnabled:   true,
			input:          "USAGE",
			expectedOutput: "\x1b[1mUSAGE\x1b[22m", // ANSI bold escape sequence
			description:    "Should return bold ANSI when color is enabled",
		},
		{
			name:           "bold_with_color_disabled",
			colorEnabled:   false,
			input:          "USAGE",
			expectedOutput: "USAGE", // Already uppercase, no change needed
			description:    "Should return uppercase when color is disabled",
		},
		{
			name:           "section_with_color_enabled",
			colorEnabled:   true,
			input:          "Quick Start",
			expectedOutput: "\x1b[1mQuick Start\x1b[22m", // ANSI bold escape sequence
			description:    "Section headers should be bold when color is enabled",
		},
		{
			name:           "section_with_color_disabled",
			colorEnabled:   false,
			input:          "Quick Start",
			expectedOutput: "QUICK START", // Uppercase fallback
			description:    "Section headers should be uppercase when color is disabled",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Create formatter with explicit color setting
			formatter := NewHelpFormatterWithColorSupport(test.colorEnabled)

			// Test Bold method
			result := formatter.Bold(test.input)

			if test.colorEnabled {
				// When color is enabled, should contain ANSI escape sequences
				require.Contains(t, result, "\x1b[1m", "Bold text should contain ANSI bold escape sequence")
				require.Contains(t, result, "\x1b[22m", "Bold text should contain ANSI reset escape sequence")
			} else {
				// When color is disabled, should be uppercase
				require.Equal(t, strings.ToUpper(test.input), result, "Text should be uppercase when color is disabled")
			}

			// Test Section method (which uses Bold internally)
			sectionResult := formatter.Section(test.input)
			require.Equal(t, result, sectionResult, "Section should use same formatting as Bold")

			// Test IsColorSupported method
			require.Equal(t, test.colorEnabled, formatter.IsColorSupported(), "IsColorSupported should return correct value")
		})
	}
}

func TestHelpFormatterColorDetection(t *testing.T) {
	t.Parallel()

	// Test automatic color detection
	formatter := NewHelpFormatter()

	// Formatter should detect color support based on environment
	require.NotNil(t, formatter, "Formatter should be created successfully")

	// Test that Bold and Section methods work without panicking
	boldResult := formatter.Bold("TEST")
	require.NotEmpty(t, boldResult, "Bold should return non-empty result")

	sectionResult := formatter.Section("TEST SECTION")
	require.NotEmpty(t, sectionResult, "Section should return non-empty result")
}

func TestHelpFormatterWithNOCOLOR(t *testing.T) { //nolint:paralleltest // Cannot run in parallel due to color.NoColor global variable modification
	// Cannot run in parallel due to color.NoColor global variable modification
	// Test that NO_COLOR environment variable is respected
	originalNoColor := color.NoColor

	defer func() {
		color.NoColor = originalNoColor
	}()

	// Force color disabled
	color.NoColor = true
	formatter := NewHelpFormatter()

	// Should fallback to uppercase formatting
	result := formatter.Bold("test")
	require.Equal(t, "TEST", result, "Should use uppercase fallback when NO_COLOR is set")

	require.False(t, formatter.IsColorSupported(), "Should report color as not supported")
}

func TestHelpFormatterTerminalIndependence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "simple_text",
			input:       "COMMANDS",
			description: "Simple section header",
		},
		{
			name:        "text_with_spaces",
			input:       "Quick Start",
			description: "Section header with spaces",
		},
		{
			name:        "empty_text",
			input:       "",
			description: "Empty string handling",
		},
		{
			name:        "special_characters",
			input:       "User's Guide",
			description: "Text with special characters",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Test both color enabled and disabled scenarios
			for _, colorEnabled := range []bool{true, false} {
				formatter := NewHelpFormatterWithColorSupport(colorEnabled)

				// Methods should not panic and should return some result
				boldResult := formatter.Bold(test.input)
				sectionResult := formatter.Section(test.input)

				require.NotNil(t, boldResult, "Bold should not return nil")
				require.NotNil(t, sectionResult, "Section should not return nil")
				require.Equal(t, boldResult, sectionResult, "Bold and Section should return same result")

				if test.input == "" {
					require.Empty(t, boldResult, "Empty input should return empty output")
				} else {
					require.NotEmpty(t, boldResult, "Non-empty input should return non-empty output")
				}
			}
		})
	}
}

func TestHelpFormatterInterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Verify all ports.HelpFormatter interface methods work correctly
	formatter := NewHelpFormatter()

	// Test all interface methods
	boldResult := formatter.Bold("TEST")
	require.NotEmpty(t, boldResult, "Bold method should work")

	sectionResult := formatter.Section("TEST")
	require.NotEmpty(t, sectionResult, "Section method should work")

	colorSupport := formatter.IsColorSupported()
	require.IsType(t, true, colorSupport, "IsColorSupported should return boolean")
}
