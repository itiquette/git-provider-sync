// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadTokenFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFile   func() (string, func())
		path        string
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name: "read_from_file_success",
			setupFile: func() (string, func()) {
				tmpFile, err := os.CreateTemp("", "token-test-*.txt")
				require.NoError(t, err)

				_, err = tmpFile.WriteString("test-token-12345\n")
				require.NoError(t, err)
				_ = tmpFile.Close()

				return tmpFile.Name(), func() {
					_ = os.Remove(tmpFile.Name())
				}
			},
			path:      "", // Will be set by setupFile
			wantToken: "test-token-12345",
			wantErr:   false,
		},
		{
			name: "read_from_file_with_whitespace",
			setupFile: func() (string, func()) {
				tmpFile, err := os.CreateTemp("", "token-test-*.txt")
				require.NoError(t, err)

				_, err = tmpFile.WriteString("  test-token-12345  \n\n")
				require.NoError(t, err)
				_ = tmpFile.Close()

				return tmpFile.Name(), func() {
					_ = os.Remove(tmpFile.Name())
				}
			},
			path:      "", // Will be set by setupFile
			wantToken: "test-token-12345",
			wantErr:   false,
		},
		{
			name: "file_not_found",
			setupFile: func() (string, func()) {
				return "/non/existent/file.txt", func() {}
			},
			path:        "/non/existent/file.txt",
			wantErr:     true,
			errContains: "failed to read token file",
		},
		{
			name: "empty_file",
			setupFile: func() (string, func()) {
				tmpFile, err := os.CreateTemp("", "token-test-*.txt")
				require.NoError(t, err)
				_ = tmpFile.Close()

				return tmpFile.Name(), func() {
					_ = os.Remove(tmpFile.Name())
				}
			},
			path:        "", // Will be set by setupFile
			wantErr:     true,
			errContains: "token file",
		},
	}

	for _, testCase := range tests {
		// Capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path, cleanup := testCase.setupFile()
			defer cleanup()

			if testCase.path == "" {
				testCase.path = path
			}

			token, err := ReadTokenFromFile(testCase.path)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errContains != "" {
					assert.Contains(t, err.Error(), testCase.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.wantToken, token)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid_token",
			token:   "ghp_xxxxxxxxxxxxxxxxxxxx",
			wantErr: false,
		},
		{
			name:        "empty_token",
			token:       "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "unexpanded_variable",
			token:       "${GITHUB_TOKEN}",
			wantErr:     true,
			errContains: "unexpanded variable",
		},
		{
			name:        "token_with_spaces",
			token:       "token with spaces",
			wantErr:     true,
			errContains: "spaces",
		},
	}

	for _, testCase := range tests {
		// Capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateToken(testCase.token)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errContains != "" {
					assert.Contains(t, err.Error(), testCase.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
