// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package shared

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic auth cases
		{
			name:     "basic auth in https URL",
			input:    "https://username:password@github.com/user/repo.git",
			expected: "https://***:***@github.com/user/repo.git",
		},
		{
			name:     "basic auth in http URL",
			input:    "http://admin:secret123@internal.server.com/path",
			expected: "http://***:***@internal.server.com/path",
		},
		{
			name:     "oauth2 token as username",
			input:    "https://oauth2:glpat-xxxxxxxxxxxxxxxxxxxx@gitlab.com/user/repo.git",
			expected: "https://***:***@gitlab.com/user/repo.git",
		},
		{
			name:     "GitHub token in URL",
			input:    "https://x-access-token:ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx@github.com/user/repo.git",
			expected: "https://***:***@github.com/user/repo.git",
		},

		// Query parameter cases
		{
			name:     "token in query parameter",
			input:    "https://api.example.com/endpoint?token=secret123&user=john",
			expected: "https://api.example.com/endpoint?token=***&user=john",
		},
		{
			name:     "api_key in query parameter",
			input:    "https://api.service.com/v1/data?api_key=abc123xyz&format=json",
			expected: "https://api.service.com/v1/data?api_key=***&format=json",
		},
		{
			name:     "multiple sensitive params",
			input:    "https://api.example.com?token=secret&api_key=key123&access_token=token456",
			expected: "https://api.example.com?access_token=***&api_key=***&token=***",
		},

		// SSH URLs
		{
			name:     "SSH URL with git user",
			input:    "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git", // SSH URLs typically don't have passwords
		},
		{
			name:     "SSH URL with custom user",
			input:    "customuser@gitlab.com:group/project.git",
			expected: "customuser@gitlab.com:group/project.git",
		},

		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "URL without credentials",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "URL with only username",
			input:    "https://username@github.com/user/repo.git",
			expected: "https://***:***@github.com/user/repo.git",
		},
		{
			name:     "malformed URL falls back to pattern matching",
			input:    "not-a-url-but-has-token:supersecret123",
			expected: "not-a-url-but-has-token:***",
		},

		// Complex cases
		{
			name:     "URL with fragment and credentials",
			input:    "https://user:pass@example.com/path#section",
			expected: "https://***:***@example.com/path#section",
		},
		{
			name:     "URL with port and credentials",
			input:    "https://admin:password@server.com:8080/api",
			expected: "https://***:***@server.com:8080/api",
		},
		{
			name:     "URL with special characters in password",
			input:    "https://user:p@ss!w0rd%40@github.com/repo.git",
			expected: "https://***:***@github.com/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := SanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeStringMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "map with URL containing credentials",
			input: map[string]interface{}{
				"clone_url": "https://user:token@github.com/repo.git",
				"name":      "my-repo",
				"public":    true,
			},
			expected: map[string]interface{}{
				"clone_url": "https://***:***@github.com/repo.git",
				"name":      "my-repo",
				"public":    true,
			},
		},
		{
			name: "map with sensitive keys",
			input: map[string]interface{}{
				"password":     "supersecret",
				"api_key":      "xyz123abc",
				"access_token": "ghp_xxxxxxxxxxxx",
				"username":     "john",
			},
			expected: map[string]interface{}{
				"password":     "***",
				"api_key":      "***",
				"access_token": "***",
				"username":     "john",
			},
		},
		{
			name: "nested map with credentials",
			input: map[string]interface{}{
				"config": map[string]interface{}{
					"auth_token": "secret123",
					"server_url": "https://admin:pass@server.com",
				},
				"status": "active",
			},
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"auth_token": "***",
					"server_url": "https://***:***@server.com",
				},
				"status": "active",
			},
		},
		{
			name: "map with token-like strings",
			input: map[string]interface{}{
				"id":       "abc123",                                     // Too short to be a token
				"long_key": "abcdefghijklmnopqrstuvwxyz0123456789ABCDEF", // Looks like a token
				"name":     "regular-name",
			},
			expected: map[string]interface{}{
				"id":       "abc123",
				"long_key": "***",
				"name":     "regular-name",
			},
		},
		{
			name:     "nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := SanitizeStringMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "error with URL containing credentials",
			err:      errors.New("failed to clone https://user:password@github.com/repo.git: authentication failed"),
			expected: "failed to clone https://***:***@github.com/repo.git: authentication failed",
		},
		{
			name:     "error with token",
			err:      errors.New("API call failed with token:ghp_1234567890abcdef"),
			expected: "API call failed with token:***",
		},
		{
			name:     "error with authorization header",
			err:      errors.New("request failed: Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"),
			expected: "request failed: Authorization: ***",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "error without sensitive data",
			err:      errors.New("connection timeout"),
			expected: "connection timeout",
		},
		{
			name:     "git error with multiple URLs",
			err:      errors.New("failed to push https://oauth2:token@gitlab.com/repo.git to https://user:pass@github.com/backup.git"),
			expected: "failed to push https://***:***@gitlab.com/repo.git to https://***:***@github.com/backup.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := SanitizeError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsSensitiveKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key       string
		sensitive bool
	}{
		{"password", true},
		{"api_key", true},
		{"access_token", true},
		{"auth_token", true},
		{"secret", true},
		{"private_key", true},
		{"username", false},
		{"email", false},
		{"name", false},
		{"url", false},
		{"PASSWORD", true}, // Case insensitive
		{"ApiKey", true},   // Case variations
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			result := containsSensitiveKey(strings.ToLower(tt.key))
			assert.Equal(t, tt.sensitive, result)
		})
	}
}

func TestLooksLikeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"https://github.com/repo.git", true},
		{"http://example.com", true},
		{"ssh://git@server.com/repo", true},
		{"git://server.com/repo.git", true},
		{"git@github.com:user/repo.git", true},
		{"ftp://server.com/file", true},
		{"not a url", false},
		{"example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			result := looksLikeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLooksLikeToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		// Real-looking tokens
		{"ghp_1234567890abcdefghijklmnopqrstuvwxyz", true},
		{"glpat-xxxxxxxxxxxxxxxxxxxx1234567890", true},
		{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0", true},

		// Not tokens
		{"short", false},
		{"just-some-text", false},
		{"12345678901234567890", false},       // Only numbers
		{"abcdefghijklmnopqrstuvwxyz", false}, // Only letters
		{"", false},
		{"normal-string-value", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			result := looksLikeToken(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"user:pass@domain.com", true},
		{"oauth2:token@gitlab.com", true},
		{"contains api_key value", true},
		{"has Bearer token", true},
		{"Authorization: something", true},
		{"normal text", false},
		{"email@domain.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			result := containsCredentials(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
