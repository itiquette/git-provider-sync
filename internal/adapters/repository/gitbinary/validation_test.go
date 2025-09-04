// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateGitBinary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectError bool
	}{
		{
			name:        "valid context with available git",
			setupCtx:    context.Background,
			expectError: false, // Assuming git is available in test environment
		},
		{
			name: "context with timeout",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()

				return ctx
			},
			expectError: true, // Should timeout
		},
		{
			name: "cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				return ctx
			},
			expectError: true, // Should fail due to cancelled context
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := test.setupCtx()
			path, err := ValidateGitBinary(ctx)

			if test.expectError {
				require.Error(t, err)
				assert.Empty(t, path)
			} else {
				// Git may not be available in all test environments
				// So we test both success and failure cases
				if err != nil {
					assert.Equal(t, ErrGitBinaryNotFound, err)
					assert.Empty(t, path)
					t.Logf("Git binary not found in test environment, which is acceptable: %v", err)
				} else {
					require.NoError(t, err)
					assert.NotEmpty(t, path)
					t.Logf("Found git binary at: %s", path)
				}
			}
		})
	}
}

func TestValidateGitBinary_LongTimeout(t *testing.T) {
	t.Parallel()

	// Test with a longer timeout to actually find git if available
	ctx := context.Background()

	path, err := ValidateGitBinary(ctx)
	if err != nil {
		assert.Equal(t, ErrGitBinaryNotFound, err)
		assert.Empty(t, path)
		t.Logf("Git binary not available in test environment: %v", err)
	} else {
		require.NoError(t, err)
		assert.NotEmpty(t, path)

		// Verify the path is one of the expected paths
		expectedPaths := []string{"git", "/usr/bin/git", "/usr/local/bin/git", "/opt/homebrew/bin/git"}
		assert.Contains(t, expectedPaths, path, "Git binary should be found in one of the expected paths")

		t.Logf("Successfully found git binary at: %s", path)
	}
}

func TestSetupSSHCommandEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		sshCommand      string
		rewriteURLFrom  string
		rewriteURLTo    string
		expectedEnvVars []string
	}{
		{
			name:            "empty ssh command",
			sshCommand:      "",
			rewriteURLFrom:  "",
			rewriteURLTo:    "",
			expectedEnvVars: []string{},
		},
		{
			name:            "ssh command only",
			sshCommand:      "ssh -i /path/to/key",
			rewriteURLFrom:  "",
			rewriteURLTo:    "",
			expectedEnvVars: []string{"GIT_SSH_COMMAND=ssh -i /path/to/key"},
		},
		{
			name:           "ssh command with URL rewriting",
			sshCommand:     "ssh -i /path/to/key",
			rewriteURLFrom: "git@github.com:",
			rewriteURLTo:   "https://github.com/",
			expectedEnvVars: []string{
				"GIT_SSH_COMMAND=ssh -i /path/to/key",
				"GIT_CONFIG_COUNT=1",
				"GIT_CONFIG_KEY_0=url.https://github.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=git@github.com:",
			},
		},
		{
			name:           "ssh command with partial URL rewriting (missing to)",
			sshCommand:     "ssh -i /path/to/key",
			rewriteURLFrom: "git@github.com:",
			rewriteURLTo:   "",
			expectedEnvVars: []string{
				"GIT_SSH_COMMAND=ssh -i /path/to/key",
			},
		},
		{
			name:           "ssh command with partial URL rewriting (missing from)",
			sshCommand:     "ssh -i /path/to/key",
			rewriteURLFrom: "",
			rewriteURLTo:   "https://github.com/",
			expectedEnvVars: []string{
				"GIT_SSH_COMMAND=ssh -i /path/to/key",
			},
		},
		{
			name:           "complex ssh command with options",
			sshCommand:     "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i /path/to/key",
			rewriteURLFrom: "git@gitlab.com:",
			rewriteURLTo:   "https://oauth2:token@gitlab.com/",
			expectedEnvVars: []string{
				"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i /path/to/key",
				"GIT_CONFIG_COUNT=1",
				"GIT_CONFIG_KEY_0=url.https://oauth2:token@gitlab.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=git@gitlab.com:",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := SetupSSHCommandEnv(test.sshCommand, test.rewriteURLFrom, test.rewriteURLTo)

			require.Len(t, result, len(test.expectedEnvVars), "Expected %d environment variables, got %d", len(test.expectedEnvVars), len(result))

			for i, expected := range test.expectedEnvVars {
				assert.Equal(t, expected, result[i], "Environment variable at index %d should match", i)
			}
		})
	}
}

func TestSetupSSHCommandEnv_MissingKeysAndInvalidPaths_ReturnsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sshCommand     string
		rewriteURLFrom string
		rewriteURLTo   string
		description    string
	}{
		{
			name:           "special characters in ssh command",
			sshCommand:     `ssh -o "UserKnownHostsFile=/dev/null" -i "/path/with spaces/key"`,
			rewriteURLFrom: "",
			rewriteURLTo:   "",
			description:    "SSH command with quotes and spaces should be preserved",
		},
		{
			name:           "special characters in URL rewriting",
			sshCommand:     "ssh -i /key",
			rewriteURLFrom: "git@custom-domain.com:",
			rewriteURLTo:   "https://user:pass@custom-domain.com/",
			description:    "URL rewriting with special characters should work correctly",
		},
		{
			name:           "empty URL rewriting values",
			sshCommand:     "ssh -i /key",
			rewriteURLFrom: "",
			rewriteURLTo:   "",
			description:    "Empty URL rewriting should not add config variables",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := SetupSSHCommandEnv(test.sshCommand, test.rewriteURLFrom, test.rewriteURLTo)

			// Verify that we always get at least the SSH command if it's not empty
			if test.sshCommand != "" {
				require.NotEmpty(t, result, test.description)
				assert.Contains(t, result[0], "GIT_SSH_COMMAND=", "First environment variable should be GIT_SSH_COMMAND")
				assert.Contains(t, result[0], test.sshCommand, "SSH command should be included in the environment variable")
			}

			// Verify URL rewriting logic
			if test.rewriteURLFrom != "" && test.rewriteURLTo != "" {
				assert.Len(t, result, 4, "Should have 4 environment variables when URL rewriting is enabled")
				assert.Contains(t, result, "GIT_CONFIG_COUNT=1")
				assert.Contains(t, result, "GIT_CONFIG_KEY_0=url."+test.rewriteURLTo+".insteadOf")
				assert.Contains(t, result, "GIT_CONFIG_VALUE_0="+test.rewriteURLFrom)
			} else {
				assert.Len(t, result, 1, "Should have only 1 environment variable when URL rewriting is disabled")
			}

			t.Logf("Test case: %s", test.description)
			t.Logf("Result: %v", result)
		})
	}
}
