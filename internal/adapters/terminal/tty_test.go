// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package terminal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsError_BehaviorInTestEnvironment(t *testing.T) {
	t.Parallel()

	// Test that IsError correctly detects stderr is not a terminal in test environment
	result := IsError()
	assert.False(t, result, "stderr should not be detected as terminal in test environment")
}

func TestIsInput_BehaviorInTestEnvironment(t *testing.T) {
	t.Parallel()

	// Test that IsInput correctly detects stdin is not a terminal in test environment
	result := IsInput()
	assert.False(t, result, "stdin should not be detected as terminal in test environment")
}

func TestNewColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mode   ColorMode
		isTTY  bool
		expect func(Color)
	}{
		{
			name:  "auto mode with TTY enables colors",
			mode:  ColorAuto,
			isTTY: true,
			expect: func(c Color) {
				assert.NotEmpty(t, c.Red)
				assert.NotEmpty(t, c.Green)
				assert.NotEmpty(t, c.Bold)
				assert.NotEmpty(t, c.Reset)
			},
		},
		{
			name:  "auto mode without TTY disables colors",
			mode:  ColorAuto,
			isTTY: false,
			expect: func(c Color) {
				assert.Empty(t, c.Red)
				assert.Empty(t, c.Green)
				assert.Empty(t, c.Bold)
				assert.Empty(t, c.Reset)
			},
		},
		{
			name:  "always mode enables colors even without TTY",
			mode:  ColorAlways,
			isTTY: false,
			expect: func(c Color) {
				assert.NotEmpty(t, c.Red)
				assert.NotEmpty(t, c.Green)
				assert.NotEmpty(t, c.Bold)
				assert.NotEmpty(t, c.Reset)
			},
		},
		{
			name:  "never mode disables colors even with TTY",
			mode:  ColorNever,
			isTTY: true,
			expect: func(c Color) {
				assert.Empty(t, c.Red)
				assert.Empty(t, c.Green)
				assert.Empty(t, c.Bold)
				assert.Empty(t, c.Reset)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			color := NewColor(tt.mode, tt.isTTY)
			tt.expect(color)
		})
	}
}

func TestColorMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		color Color
		input string
		want  struct {
			success string
			error   string
			header  string
		}
	}{
		{
			name: "with colors",
			color: Color{
				Red:   "\033[31m",
				Green: "\033[32m",
				Bold:  "\033[1m",
				Reset: "\033[0m",
			},
			input: "test",
			want: struct {
				success string
				error   string
				header  string
			}{
				success: "\033[32mtest\033[0m",
				error:   "\033[31mtest\033[0m",
				header:  "\033[1mtest\033[0m",
			},
		},
		{
			name:  "without colors",
			color: Color{},
			input: "test",
			want: struct {
				success string
				error   string
				header  string
			}{
				success: "test",
				error:   "test",
				header:  "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want.success, tt.color.Success(tt.input))
			assert.Equal(t, tt.want.error, tt.color.Error(tt.input))
			assert.Equal(t, tt.want.header, tt.color.Header(tt.input))
		})
	}
}

func TestColorModeHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      ColorMode
		isTTY     bool
		wantColor bool
	}{
		{
			name:      "auto mode with TTY enables colors",
			mode:      ColorAuto,
			isTTY:     true,
			wantColor: true,
		},
		{
			name:      "auto mode without TTY disables colors",
			mode:      ColorAuto,
			isTTY:     false,
			wantColor: false,
		},
		{
			name:      "always mode with TTY enables colors",
			mode:      ColorAlways,
			isTTY:     true,
			wantColor: true,
		},
		{
			name:      "always mode without TTY still enables colors",
			mode:      ColorAlways,
			isTTY:     false,
			wantColor: true,
		},
		{
			name:      "never mode with TTY disables colors",
			mode:      ColorNever,
			isTTY:     true,
			wantColor: false,
		},
		{
			name:      "never mode without TTY disables colors",
			mode:      ColorNever,
			isTTY:     false,
			wantColor: false,
		},
		{
			name:      "empty string defaults to auto with TTY",
			mode:      "",
			isTTY:     true,
			wantColor: true,
		},
		{
			name:      "empty string defaults to auto without TTY",
			mode:      "",
			isTTY:     false,
			wantColor: false,
		},
		{
			name:      "unknown mode defaults to auto with TTY",
			mode:      "invalid",
			isTTY:     true,
			wantColor: true,
		},
		{
			name:      "unknown mode defaults to auto without TTY",
			mode:      "invalid",
			isTTY:     false,
			wantColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			color := NewColor(tt.mode, tt.isTTY)
			if tt.wantColor {
				assert.NotEmpty(t, color.Red, "Red color code should be set")
				assert.NotEmpty(t, color.Green, "Green color code should be set")
				assert.NotEmpty(t, color.Bold, "Bold color code should be set")
				assert.NotEmpty(t, color.Reset, "Reset color code should be set")
			} else {
				assert.Empty(t, color.Red, "Red color code should be empty")
				assert.Empty(t, color.Green, "Green color code should be empty")
				assert.Empty(t, color.Bold, "Bold color code should be empty")
				assert.Empty(t, color.Reset, "Reset color code should be empty")
			}
		})
	}
}

func TestColorModeWithEnvironmentVariables(t *testing.T) { //nolint:paralleltest // Cannot run in parallel due to environment variable modification
	tests := []struct {
		name      string
		mode      ColorMode
		isTTY     bool
		envVar    string
		envValue  string
		wantColor bool
	}{
		{
			name:      "NO_COLOR env disables colors even with always mode",
			mode:      ColorAlways,
			isTTY:     true,
			envVar:    "NO_COLOR",
			envValue:  "1",
			wantColor: false,
		},
		{
			name:      "NO_COLOR empty string enables colors with always mode",
			mode:      ColorAlways,
			isTTY:     true,
			envVar:    "NO_COLOR",
			envValue:  "",
			wantColor: true,
		},
		{
			name:      "TERM=dumb disables colors with auto mode",
			mode:      ColorAuto,
			isTTY:     true,
			envVar:    "TERM",
			envValue:  "dumb",
			wantColor: false,
		},
		{
			name:      "TERM=xterm enables colors with auto mode and TTY",
			mode:      ColorAuto,
			isTTY:     true,
			envVar:    "TERM",
			envValue:  "xterm",
			wantColor: true,
		},
	}

	for _, testCase := range tests { //nolint:paralleltest // Cannot run in parallel due to environment variable modification
		testCase := testCase //nolint:copyloopvar // Keep for Go compatibility
		t.Run(testCase.name, func(t *testing.T) {
			// Save original env value
			originalValue := os.Getenv(testCase.envVar)

			defer func() {
				if originalValue == "" {
					_ = os.Unsetenv(testCase.envVar)
				} else {
					_ = os.Setenv(testCase.envVar, originalValue)
				}
			}()

			// Set test env value
			if testCase.envValue == "" {
				_ = os.Unsetenv(testCase.envVar)
			} else {
				_ = os.Setenv(testCase.envVar, testCase.envValue)
			}

			// Test color creation
			color := NewColor(testCase.mode, testCase.isTTY)
			if testCase.wantColor {
				assert.NotEmpty(t, color.Red, "Red color code should be set")
				assert.NotEmpty(t, color.Green, "Green color code should be set")
				assert.NotEmpty(t, color.Bold, "Bold color code should be set")
				assert.NotEmpty(t, color.Reset, "Reset color code should be set")
			} else {
				assert.Empty(t, color.Red, "Red color code should be empty")
				assert.Empty(t, color.Green, "Green color code should be empty")
				assert.Empty(t, color.Bold, "Bold color code should be empty")
				assert.Empty(t, color.Reset, "Reset color code should be empty")
			}
		})
	}
}
