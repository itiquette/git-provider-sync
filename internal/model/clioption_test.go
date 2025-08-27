// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
)

// TestErrorHandler is a test error handler that captures output for verification.
type TestErrorHandler struct {
	Writer io.Writer
}

// HandleError implements ErrorHandler interface for testing.
func (h *TestErrorHandler) HandleError(_ context.Context, err error) {
	_, _ = h.Writer.Write([]byte("CLI Option Error: " + err.Error() + "\n"))
}

func TestCLIOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		contextValue    interface{}
		expectError     bool
		expectedOptions CLIOption
	}{
		{
			name: "valid CLI options in context",
			contextValue: CLIOption{
				AlphaNumHyphName:    true,
				ActiveFromLimit:     "30d",
				ConfigFileOnly:      false,
				ConfigFilePath:      "/path/to/config.yaml",
				DryRun:              true,
				ForcePush:           false,
				IgnoreInvalidName:   true,
				IncludeForks:        false,
				IncludeArchived:     true,
				UseGitBinary:        false,
				OutputFormat:        "json",
				Quiet:               true,
				VerbosityWithCaller: false,
			},
			expectError: false,
			expectedOptions: CLIOption{
				AlphaNumHyphName:    true,
				ActiveFromLimit:     "30d",
				ConfigFileOnly:      false,
				ConfigFilePath:      "/path/to/config.yaml",
				DryRun:              true,
				ForcePush:           false,
				IgnoreInvalidName:   true,
				IncludeForks:        false,
				IncludeArchived:     true,
				UseGitBinary:        false,
				OutputFormat:        "json",
				Quiet:               true,
				VerbosityWithCaller: false,
			},
		},
		{
			name:            "no CLI options in context",
			contextValue:    nil,
			expectError:     true,
			expectedOptions: CLIOption{}, // zero value
		},
		{
			name:            "wrong type in context",
			contextValue:    "not a CLIOption",
			expectError:     true,
			expectedOptions: CLIOption{}, // zero value
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			if testCase.contextValue != nil {
				ctx = context.WithValue(context.Background(), CLIOptionKey{}, testCase.contextValue)
			} else {
				ctx = context.Background()
			}

			// Create a buffer to capture error output
			var buf bytes.Buffer

			errorHandler := &TestErrorHandler{Writer: &buf}
			ctx = WithErrorHandler(ctx, errorHandler)

			options := CLIOptions(ctx)

			// Read captured output
			stderrOutput := buf.String()

			if testCase.expectError {
				assert.Contains(t, stderrOutput, "CLI Option Error")
				assert.Contains(t, stderrOutput, domain.ErrCLIOptionRetrievalFailed.Error())
			} else {
				assert.Empty(t, stderrOutput)
			}

			assert.Equal(t, testCase.expectedOptions, options)
		})
	}
}

func TestWithCLIOpt_AddToContext_StoresAndRetrievesCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cliOption   CLIOption
		verifyValue bool
	}{
		{
			name: "add CLI option to context",
			cliOption: CLIOption{
				AlphaNumHyphName: true,
				ActiveFromLimit:  "7d",
				ConfigFilePath:   "/test/config.yaml",
				DryRun:           true,
				OutputFormat:     "plain",
			},
			verifyValue: true,
		},
		{
			name:        "add empty CLI option to context",
			cliOption:   CLIOption{},
			verifyValue: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			newCtx := WithCLIOpt(ctx, testCase.cliOption)

			assert.NotEqual(t, ctx, newCtx)

			if testCase.verifyValue {
				retrievedOption, ok := newCtx.Value(CLIOptionKey{}).(CLIOption)
				require.True(t, ok)
				assert.Equal(t, testCase.cliOption, retrievedOption)
			}
		})
	}
}

func TestCLIOptionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cliOption      CLIOption
		expectedFields []string
	}{
		{
			name: "all fields populated",
			cliOption: CLIOption{
				AlphaNumHyphName:    true,
				ActiveFromLimit:     "30d",
				ConfigFileOnly:      false,
				ConfigFilePath:      "/path/to/config.yaml",
				DryRun:              true,
				ForcePush:           false,
				IgnoreInvalidName:   true,
				IncludeForks:        false,
				IncludeArchived:     true,
				UseGitBinary:        false,
				OutputFormat:        "json",
				Quiet:               true,
				VerbosityWithCaller: false,
			},
			expectedFields: []string{
				"ForcePush: false",
				"IgnoreInvalidName: true",
				"ASCIIName: true",
				"ActiveFromLimit: 30d",
				"DryRun: true",
				"ConfigFilePath: /path/to/config.yaml",
				"ConfigFileOnly: false",
				"Quiet: true",
				"OutputFormat: json",
			},
		},
		{
			name:      "zero value CLI option",
			cliOption: CLIOption{},
			expectedFields: []string{
				"ForcePush: false",
				"IgnoreInvalidName: false",
				"ASCIIName: false",
				"ActiveFromLimit: ",
				"DryRun: false",
				"ConfigFilePath: ",
				"ConfigFileOnly: false",
				"Quiet: false",
				"OutputFormat: ",
			},
		},
		{
			name: "CLI option with force push enabled",
			cliOption: CLIOption{
				ForcePush:         true,
				IgnoreInvalidName: false,
				AlphaNumHyphName:  false,
				OutputFormat:      "plain",
			},
			expectedFields: []string{
				"ForcePush: true",
				"IgnoreInvalidName: false",
				"ASCIIName: false",
				"OutputFormat: plain",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.cliOption.String()

			assert.Contains(t, result, "CLIOption{")

			for _, expectedField := range testCase.expectedFields {
				assert.Contains(t, result, expectedField)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		inputError    error
		expectedInErr string
	}{
		{
			name:          "domain error",
			inputError:    domain.ErrCLIOptionRetrievalFailed,
			expectedInErr: "failed to retrieve or type-assert CLIOption from context",
		},
		{
			name:          "custom error",
			inputError:    assert.AnError,
			expectedInErr: "assert.AnError",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a buffer to capture error output
			var buf bytes.Buffer

			errorHandler := &TestErrorHandler{Writer: &buf}
			ctx := context.Background()
			ctx = WithErrorHandler(ctx, errorHandler)

			HandleError(ctx, testCase.inputError)

			// Read captured output
			stderrOutput := buf.String()

			assert.Contains(t, stderrOutput, "CLI Option Error:")
			assert.Contains(t, stderrOutput, testCase.expectedInErr)
		})
	}
}

func TestCLIOption_ContextWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupOption CLIOption
		testAction  func(t *testing.T, ctx context.Context)
	}{
		{
			name: "complete workflow with valid options",
			setupOption: CLIOption{
				AlphaNumHyphName: true,
				DryRun:           true,
				ConfigFilePath:   "/test/config.yaml",
				OutputFormat:     "json",
				Quiet:            false,
			},
			testAction: func(t *testing.T, ctx context.Context) {
				t.Helper()
				// Test that we can retrieve and use the options
				options := CLIOptions(ctx)
				assert.True(t, options.AlphaNumHyphName)
				assert.True(t, options.DryRun)
				assert.Equal(t, "/test/config.yaml", options.ConfigFilePath)
				assert.Equal(t, "json", options.OutputFormat)
				assert.False(t, options.Quiet)

				// Test string representation
				str := options.String()
				assert.Contains(t, str, "DryRun: true")
				assert.Contains(t, str, "ASCIIName: true")
			},
		},
		{
			name:        "empty context workflow",
			setupOption: CLIOption{},
			testAction: func(t *testing.T, ctx context.Context) {
				t.Helper()
				// Create a buffer to capture error output
				var buf bytes.Buffer
				errorHandler := &TestErrorHandler{Writer: &buf}
				ctx = WithErrorHandler(ctx, errorHandler)

				// This should trigger error handling
				options := CLIOptions(ctx)

				// Read captured output
				stderrOutput := buf.String()

				// Should have gotten error output
				assert.Contains(t, stderrOutput, "CLI Option Error")

				// Should have gotten zero value
				assert.Equal(t, CLIOption{}, options)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			if testCase.name == "empty context workflow" {
				ctx = context.Background()
			} else {
				ctx = WithCLIOpt(context.Background(), testCase.setupOption)
			}

			testCase.testAction(t, ctx)
		})
	}
}

func TestCLIOption_SpecialValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "context with nil value",
			testFunc: func(t *testing.T) {
				t.Helper()
				ctx := context.WithValue(context.Background(), CLIOptionKey{}, nil)

				// Create a buffer to capture error output
				var buf bytes.Buffer
				errorHandler := &TestErrorHandler{Writer: &buf}
				ctx = WithErrorHandler(ctx, errorHandler)

				options := CLIOptions(ctx)

				stderrOutput := buf.String()

				assert.Contains(t, stderrOutput, "CLI Option Error")
				assert.Equal(t, CLIOption{}, options)
			},
		},
		{
			name: "very long string values",
			testFunc: func(t *testing.T) {
				t.Helper()
				longPath := strings.Repeat("/very/long/path", 100)
				longFormat := strings.Repeat("json", 50)

				option := CLIOption{
					ConfigFilePath: longPath,
					OutputFormat:   longFormat,
				}

				str := option.String()
				assert.Contains(t, str, longPath)
				assert.Contains(t, str, longFormat)
			},
		},
		{
			name: "special characters in string values",
			testFunc: func(t *testing.T) {
				t.Helper()
				option := CLIOption{
					ConfigFilePath:  "/path/with spaces/and\"quotes\"",
					ActiveFromLimit: "30d\nwith\nnewlines",
					OutputFormat:    "json\twith\ttabs",
				}

				str := option.String()
				assert.Contains(t, str, option.ConfigFilePath)
				assert.Contains(t, str, option.ActiveFromLimit)
				assert.Contains(t, str, option.OutputFormat)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}
