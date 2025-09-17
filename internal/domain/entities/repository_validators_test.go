// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package entities_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// TestRepositoryNameValidation consolidates all repository name validation tests
// Tests business rules for valid repository names across different providers.
func TestRepositoryNameValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		valid        bool
		errorContent string // What the error should contain
		description  string // Business rule being tested
	}{
		// Valid cases - these should work across all providers
		{
			name:        "standard_name_with_hyphens",
			input:       "my-awesome-repo",
			valid:       true,
			description: "Standard repository names with hyphens are universally accepted",
		},
		{
			name:        "name_with_underscores",
			input:       "my_awesome_repo",
			valid:       true,
			description: "Underscores are valid in repository names",
		},
		{
			name:        "name_with_dots",
			input:       "my.awesome.repo",
			valid:       true,
			description: "Dots are allowed within repository names",
		},
		{
			name:        "alphanumeric_only",
			input:       "repo123",
			valid:       true,
			description: "Pure alphanumeric names are valid",
		},
		{
			name:        "single_character",
			input:       "x",
			valid:       true,
			description: "Single character names are valid",
		},

		// Invalid cases - business rule violations
		{
			name:         "empty_name",
			input:        "",
			valid:        false,
			errorContent: "empty",
			description:  "Empty repository names are not allowed",
		},
		{
			name:         "name_too_long",
			input:        strings.Repeat("a", 101),
			valid:        false,
			errorContent: "too long",
			description:  "Repository names must be 100 characters or less",
		},
		{
			name:         "special_characters",
			input:        "my-repo@invalid",
			valid:        false,
			errorContent: "invalid character",
			description:  "Special characters like @ are not allowed",
		},
		{
			name:         "starts_with_period",
			input:        ".hidden-repo",
			valid:        false,
			errorContent: "cannot start or end with period",
			description:  "Repository names cannot start with a period",
		},
		{
			name:         "ends_with_period",
			input:        "repo.",
			valid:        false,
			errorContent: "cannot start or end with period",
			description:  "Repository names cannot end with a period",
		},
		{
			name:         "spaces_not_allowed",
			input:        "my repo",
			valid:        false,
			errorContent: "invalid character",
			description:  "Spaces are not allowed in repository names",
		},
		{
			name:        "consecutive_dots",
			input:       "repo..name",
			valid:       true, // Allowed by current validation
			description: "Consecutive dots are allowed (may want to reconsider)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := entities.ValidateRepositoryName(testCase.input)

			if testCase.valid {
				require.NoError(t, err, "Expected valid name but got error: %v", err)
			} else {
				require.Error(t, err, "Expected validation error for: %s", testCase.description)

				if testCase.errorContent != "" {
					assert.Contains(t, err.Error(), testCase.errorContent,
						"Error should mention: %s", testCase.errorContent)
				}
			}
		})
	}
}

// TestRepositoryNameCleaning tests the business logic for cleaning repository names
// Names can be automatically fixed for different provider requirements.
func TestRepositoryNameCleaning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		reason   string // Business reason for the transformation
	}{
		{
			name:     "already_clean",
			input:    "my-repo",
			expected: "my-repo",
			reason:   "Clean names should not be modified",
		},
		{
			name:     "spaces_to_hyphens",
			input:    "my awesome repo",
			expected: "my-awesome-repo",
			reason:   "Spaces are converted to hyphens for URL compatibility",
		},
		{
			name:     "special_chars_removed",
			input:    "my@repo#name!",
			expected: "my-repo-name",
			reason:   "Special characters are replaced with hyphens",
		},
		{
			name:     "consecutive_special_chars",
			input:    "my@@@repo",
			expected: "my-repo",
			reason:   "Multiple consecutive special chars become single hyphen",
		},
		{
			name:     "empty_string_fallback",
			input:    "",
			expected: "repository",
			reason:   "Empty input gets default name",
		},
		{
			name:     "only_special_chars_fallback",
			input:    "@#$%^&*()",
			expected: "repository",
			reason:   "Input with only special chars gets default name",
		},
		{
			name:     "leading_trailing_hyphens_removed",
			input:    "-repo-name-",
			expected: "repo-name",
			reason:   "Leading and trailing hyphens should be removed",
		},
		{
			name:     "unicode_handling",
			input:    "my-répo-name",
			expected: "my-r-po-name",
			reason:   "Non-ASCII characters are replaced for compatibility",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cleaned := entities.CleanRepositoryName(testCase.input)
			assert.Equal(t, testCase.expected, cleaned,
				"Cleaning failed. Reason: %s", testCase.reason)

			// Verify cleaned name is always valid
			err := entities.ValidateRepositoryName(cleaned)
			require.NoError(t, err,
				"Cleaned name should always be valid, but got: %v", err)
		})
	}
}

// TestRepositoryBuilderValidation tests the builder's validation behavior
// Builder enforces business rules during construction.
func TestRepositoryBuilderValidation(t *testing.T) {
	t.Parallel()

	t.Run("builder_enforces_required_fields", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name      string
			build     func() (entities.Repository, error)
			shouldErr bool
			errMsg    string
		}{
			{
				name: "missing_name",
				build: func() (entities.Repository, error) {
					builder := entities.NewRepositoryBuilder()
					builder, _ = builder.WithHTTPSURL("https://github.com/user/repo.git")

					return builder.Build()
				},
				shouldErr: true,
				errMsg:    "name cannot be empty",
			},
			{
				name: "missing_url",
				build: func() (entities.Repository, error) {
					builder := entities.NewRepositoryBuilder()
					builder, _ = builder.WithName("test-repo")

					return builder.Build()
				},
				shouldErr: true,
				errMsg:    "URL",
			},
			{
				name: "valid_with_https",
				build: func() (entities.Repository, error) {
					builder := entities.NewRepositoryBuilder()
					builder, _ = builder.WithName("test-repo")
					builder, _ = builder.WithHTTPSURL("https://github.com/user/repo.git")

					return builder.Build()
				},
				shouldErr: false,
			},
			{
				name: "valid_with_ssh",
				build: func() (entities.Repository, error) {
					builder := entities.NewRepositoryBuilder()
					builder, _ = builder.WithName("test-repo")
					builder, _ = builder.WithSSHURL("git@github.com:user/repo.git")

					return builder.Build()
				},
				shouldErr: false,
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				repo, err := testCase.build()

				if testCase.shouldErr {
					require.Error(t, err)

					if testCase.errMsg != "" {
						assert.Contains(t, err.Error(), testCase.errMsg)
					}
				} else {
					require.NoError(t, err)
					assert.NotEmpty(t, repo.Name())
				}
			})
		}
	})

	t.Run("builder_validates_name_format", func(t *testing.T) {
		t.Parallel()

		invalidNames := []string{
			"",                       // empty
			strings.Repeat("a", 101), // too long
			".hidden",                // starts with period
			"repo.",                  // ends with period
			"repo name",              // contains space
			"repo@host",              // contains @
		}

		for _, invalidName := range invalidNames {
			builder := entities.NewRepositoryBuilder()
			_, err := builder.WithName(invalidName)
			require.Error(t, err, "Should reject invalid name: %q", invalidName)
		}
	})
}
