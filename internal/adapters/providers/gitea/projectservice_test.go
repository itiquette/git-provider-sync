// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// testGiteaProjectLogger is a simple no-op logger for testing ProjectService.
type testGiteaProjectLogger struct{}

func (l testGiteaProjectLogger) Trace(_ context.Context, _ string, _ map[string]any) {
}
func (l testGiteaProjectLogger) Debug(_ context.Context, _ string, _ map[string]any) {
}
func (l testGiteaProjectLogger) Info(_ context.Context, _ string, _ map[string]any) {
}
func (l testGiteaProjectLogger) Warn(_ context.Context, _ string, _ map[string]any) {
}
func (l testGiteaProjectLogger) Error(_ context.Context, _ string, _ map[string]any) {
}
func (l testGiteaProjectLogger) Fatal(_ context.Context, _ string, _ map[string]any) {
}
func (l testGiteaProjectLogger) IsLevelEnabled(_ ports.LogLevel) bool { return true }

// Removed TestNewProjectService - constructor test with nil dependency adds no value
// The actual service behavior is tested in TestProjectService_ValidateProjectName

func TestProjectService_ValidateProjectName(t *testing.T) {
	t.Parallel()

	logger := testGiteaProjectLogger{}
	service := NewProjectService(nil, logger)

	tests := []struct {
		name        string
		projectName string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid repository name",
			projectName: "my-valid-repo",
			expectError: false,
		},
		{
			name:        "valid repository name with underscores",
			projectName: "my_valid_repo",
			expectError: false,
		},
		{
			name:        "valid repository name with numbers",
			projectName: "repo123",
			expectError: false,
		},
		{
			name:        "valid repository name with dots",
			projectName: "my.valid.repo",
			expectError: false,
		},
		{
			name:        "empty repository name",
			projectName: "",
			expectError: true,
			expectedErr: ErrProjectNameEmpty,
		},
		{
			name:        "repository name too long",
			projectName: "this-is-a-very-long-repository-name-that-exceeds-the-maximum-allowed-length-of-one-hundred-characters-for-gitea-repositories",
			expectError: true,
			expectedErr: ErrProjectNameTooLong,
		},
		{
			name:        "repository name starts with period",
			projectName: ".invalid-repo",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithPeriod,
		},
		{
			name:        "repository name ends with period",
			projectName: "invalid-repo.",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithPeriod,
		},
		{
			name:        "repository name starts with hyphen",
			projectName: "-invalid-repo",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithHyphen,
		},
		{
			name:        "repository name ends with hyphen",
			projectName: "invalid-repo-",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithHyphen,
		},
		{
			name:        "repository name with invalid character - space",
			projectName: "invalid repo",
			expectError: true,
			expectedErr: ErrProjectNameInvalidCharacter,
		},
		{
			name:        "repository name with invalid character - special symbol",
			projectName: "invalid@repo",
			expectError: true,
			expectedErr: ErrProjectNameInvalidCharacter,
		},
		{
			name:        "repository name with reserved name - admin",
			projectName: "admin",
			expectError: true,
			expectedErr: ErrProjectNameReserved,
		},
		{
			name:        "repository name with reserved name - api",
			projectName: "api",
			expectError: true,
			expectedErr: ErrProjectNameReserved,
		},
		{
			name:        "repository name with reserved name - case insensitive",
			projectName: "ADMIN",
			expectError: true,
			expectedErr: ErrProjectNameReserved,
		},
		{
			name:        "repository name with reserved name - .git",
			projectName: ".git",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithPeriod, // This will be caught first
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateProjectName(test.projectName)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectService_validateBasicRequirements(t *testing.T) {
	t.Parallel()

	logger := testGiteaProjectLogger{}
	service := NewProjectService(nil, logger)

	tests := []struct {
		name        string
		projectName string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid name",
			projectName: "valid-repo",
			expectError: false,
		},
		{
			name:        "empty name",
			projectName: "",
			expectError: true,
			expectedErr: ErrProjectNameEmpty,
		},
		{
			name:        "name at maximum length",
			projectName: strings.Repeat("a", 100), // 100 characters total
			expectError: false,
		},
		{
			name:        "name too long",
			projectName: strings.Repeat("a", 101), // 101 characters total
			expectError: true,
			expectedErr: ErrProjectNameTooLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := service.validateBasicRequirements(test.projectName)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectService_validateNamingRules(t *testing.T) {
	t.Parallel()

	logger := testGiteaProjectLogger{}
	service := NewProjectService(nil, logger)

	tests := []struct {
		name        string
		projectName string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid name",
			projectName: "valid-repo",
			expectError: false,
		},
		{
			name:        "name with period in middle",
			projectName: "my.repo",
			expectError: false,
		},
		{
			name:        "name with hyphen in middle",
			projectName: "my-repo",
			expectError: false,
		},
		{
			name:        "name starts with period",
			projectName: ".invalid",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithPeriod,
		},
		{
			name:        "name ends with period",
			projectName: "invalid.",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithPeriod,
		},
		{
			name:        "name starts with hyphen",
			projectName: "-invalid",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithHyphen,
		},
		{
			name:        "name ends with hyphen",
			projectName: "invalid-",
			expectError: true,
			expectedErr: ErrProjectNameStartsOrEndsWithHyphen,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := service.validateNamingRules(test.projectName)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectService_validateCharacters(t *testing.T) {
	t.Parallel()

	logger := testGiteaProjectLogger{}
	service := NewProjectService(nil, logger)

	tests := []struct {
		name        string
		projectName string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid characters - lowercase",
			projectName: "myrepo",
			expectError: false,
		},
		{
			name:        "valid characters - uppercase",
			projectName: "MYREPO",
			expectError: false,
		},
		{
			name:        "valid characters - numbers",
			projectName: "repo123",
			expectError: false,
		},
		{
			name:        "valid characters - hyphen",
			projectName: "my-repo",
			expectError: false,
		},
		{
			name:        "valid characters - underscore",
			projectName: "my_repo",
			expectError: false,
		},
		{
			name:        "valid characters - period",
			projectName: "my.repo",
			expectError: false,
		},
		{
			name:        "valid characters - mixed",
			projectName: "My-Repo_123.test",
			expectError: false,
		},
		{
			name:        "invalid character - space",
			projectName: "my repo",
			expectError: true,
			expectedErr: ErrProjectNameInvalidCharacter,
		},
		{
			name:        "invalid character - at symbol",
			projectName: "my@repo",
			expectError: true,
			expectedErr: ErrProjectNameInvalidCharacter,
		},
		{
			name:        "invalid character - hash",
			projectName: "my#repo",
			expectError: true,
			expectedErr: ErrProjectNameInvalidCharacter,
		},
		{
			name:        "invalid character - forward slash",
			projectName: "my/repo",
			expectError: true,
			expectedErr: ErrProjectNameInvalidCharacter,
		},
		{
			name:        "invalid character - backslash",
			projectName: "my\\repo",
			expectError: true,
			expectedErr: ErrProjectNameInvalidCharacter,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := service.validateCharacters(test.projectName)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectService_validateReservedNames(t *testing.T) {
	t.Parallel()

	logger := testGiteaProjectLogger{}
	service := NewProjectService(nil, logger)

	// Test all reserved names
	reservedNames := []string{
		".", "..", ".git", ".well-known", "admin", "api", "root", "help",
		"install", "assets", "css", "js", "img", "debug", "raw", "user",
		"org", "repo", "issues", "pulls", "commits", "releases", "wiki",
	}

	tests := []struct {
		name        string
		projectName string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid name - not reserved",
			projectName: "my-awesome-project",
			expectError: false,
		},
		{
			name:        "valid name - similar to reserved but not exact",
			projectName: "admin-panel",
			expectError: false,
		},
		{
			name:        "valid name - reserved as substring",
			projectName: "user-management",
			expectError: false,
		},
	}

	// Add tests for all reserved names
	for _, reserved := range reservedNames {
		tests = append(tests, struct {
			name        string
			projectName string
			expectError bool
			expectedErr error
		}{
			name:        "reserved name - " + reserved,
			projectName: reserved,
			expectError: true,
			expectedErr: ErrProjectNameReserved,
		})

		// Test case insensitive
		tests = append(tests, struct {
			name        string
			projectName string
			expectError bool
			expectedErr error
		}{
			name:        "reserved name case insensitive - " + reserved,
			projectName: strings.ToUpper(reserved),
			expectError: true,
			expectedErr: ErrProjectNameReserved,
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := service.validateReservedNames(test.projectName)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsValidGiteaRepoChar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		// Valid characters
		{name: "lowercase letter", char: 'a', expected: true},
		{name: "uppercase letter", char: 'Z', expected: true},
		{name: "digit", char: '5', expected: true},
		{name: "hyphen", char: '-', expected: true},
		{name: "underscore", char: '_', expected: true},
		{name: "period", char: '.', expected: true},

		// Invalid characters
		{name: "space", char: ' ', expected: false},
		{name: "at symbol", char: '@', expected: false},
		{name: "hash", char: '#', expected: false},
		{name: "forward slash", char: '/', expected: false},
		{name: "backslash", char: '\\', expected: false},
		{name: "question mark", char: '?', expected: false},
		{name: "exclamation mark", char: '!', expected: false},
		{name: "percent", char: '%', expected: false},
		{name: "ampersand", char: '&', expected: false},
		{name: "asterisk", char: '*', expected: false},
		{name: "plus", char: '+', expected: false},
		{name: "equals", char: '=', expected: false},
		{name: "pipe", char: '|', expected: false},
		{name: "colon", char: ':', expected: false},
		{name: "semicolon", char: ';', expected: false},
		{name: "comma", char: ',', expected: false},
		{name: "less than", char: '<', expected: false},
		{name: "greater than", char: '>', expected: false},
		{name: "parenthesis open", char: '(', expected: false},
		{name: "parenthesis close", char: ')', expected: false},
		{name: "bracket open", char: '[', expected: false},
		{name: "bracket close", char: ']', expected: false},
		{name: "brace open", char: '{', expected: false},
		{name: "brace close", char: '}', expected: false},
		{name: "single quote", char: '\'', expected: false},
		{name: "double quote", char: '"', expected: false},
		{name: "backtick", char: '`', expected: false},
		{name: "tilde", char: '~', expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isValidGiteaRepoChar(test.char)
			assert.Equal(t, test.expected, result, "Character '%c' validation should return %v", test.char, test.expected)
		})
	}
}

func TestIsValidGiteaRepoChar_EdgeCases(t *testing.T) {
	t.Parallel()

	// Test boundary conditions
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		// Boundary characters around valid ranges
		{name: "char before 'a'", char: 'a' - 1, expected: false},
		{name: "char after 'z'", char: 'z' + 1, expected: false},
		{name: "char before 'A'", char: 'A' - 1, expected: false},
		{name: "char after 'Z'", char: 'Z' + 1, expected: false},
		{name: "char before '0'", char: '0' - 1, expected: false},
		{name: "char after '9'", char: '9' + 1, expected: false},

		// Unicode characters
		{name: "unicode letter", char: 'é', expected: false},
		{name: "unicode symbol", char: '€', expected: false},
		{name: "emoji", char: '😀', expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isValidGiteaRepoChar(test.char)
			assert.Equal(t, test.expected, result, "Character '%c' (U+%04X) validation should return %v", test.char, test.char, test.expected)
		})
	}
}
