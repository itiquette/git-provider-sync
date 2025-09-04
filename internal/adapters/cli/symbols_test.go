// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"itiquette/git-provider-sync/internal/adapters/terminal"
)

func TestGetSymbols(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tests := []struct {
		name          string
		colorMode     terminal.ColorMode
		envVar        string
		envValue      string
		expectUnicode bool
	}{
		{
			name:          "auto mode returns ASCII when NO_COLOR is set",
			colorMode:     terminal.ColorAuto,
			envVar:        "NO_COLOR",
			envValue:      "1",
			expectUnicode: false,
		},
		{
			name:          "never mode returns ASCII",
			colorMode:     terminal.ColorNever,
			envVar:        "",
			envValue:      "",
			expectUnicode: false,
		},
		{
			name:          "always mode with UTF-8 returns Unicode",
			colorMode:     terminal.ColorAlways,
			envVar:        "LANG",
			envValue:      "en_US.UTF-8",
			expectUnicode: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup - use t.Setenv for automatic cleanup
			if testCase.envVar != "" {
				t.Setenv(testCase.envVar, testCase.envValue)
			}

			// Test
			symbols := GetSymbols(testCase.colorMode)

			// Assert
			if testCase.expectUnicode {
				assert.Equal(t, "✓", symbols.Check)
				assert.Equal(t, "✗", symbols.Cross)
				assert.Equal(t, "→", symbols.Arrow)
			} else {
				assert.Equal(t, "[OK]", symbols.Check)
				assert.Equal(t, "[!!]", symbols.Cross)
				assert.Equal(t, "->", symbols.Arrow)
			}
		})
	}
}

func TestGetErrorSuggestion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		errMsg         string
		wantSuggestion bool
	}{
		{
			name:           "configuration error gets suggestion",
			errMsg:         "failed to read configuration file",
			wantSuggestion: true,
		},
		{
			name:           "authentication error gets suggestion",
			errMsg:         "401 unauthorized",
			wantSuggestion: true,
		},
		{
			name:           "network error gets suggestion",
			errMsg:         "connection timeout",
			wantSuggestion: true,
		},
		{
			name:           "unknown error gets no suggestion",
			errMsg:         "unknown error occurred",
			wantSuggestion: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := &testError{msg: testCase.errMsg}

			suggestion := GetErrorSuggestion(err)

			if testCase.wantSuggestion {
				assert.NotEmpty(t, suggestion, "Expected suggestion for error: %s", testCase.errMsg)
			} else {
				assert.Empty(t, suggestion, "Expected no suggestion for error: %s", testCase.errMsg)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
