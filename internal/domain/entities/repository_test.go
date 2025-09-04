// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package entities_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
)

func TestRepositoryBuilder_WithName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repoName    string
		expectError bool
	}{
		{
			name:        "valid_repository_name",
			repoName:    "my-repo",
			expectError: false,
		},
		{
			name:        "empty_repository_name",
			repoName:    "",
			expectError: true,
		},
		{
			name:        "repository_name_too_long",
			repoName:    strings.Repeat("a", 101), // 101 characters
			expectError: true,
		},
		{
			name:        "repository_name_with_invalid_chars",
			repoName:    "my-repo@invalid",
			expectError: true,
		},
		{
			name:        "repository_name_starting_with_period",
			repoName:    ".invalid-start",
			expectError: true,
		},
		{
			name:        "repository_name_ending_with_period",
			repoName:    "invalid-end.",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := entities.NewRepositoryBuilder()

			newBuilder, err := builder.WithName(test.repoName)

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Test building with valid name
				builderWithURL, urlErr := newBuilder.WithHTTPSURL("https://github.com/owner/repo.git")
				require.NoError(t, urlErr)

				repo, buildErr := builderWithURL.Build()

				require.NoError(t, buildErr)
				require.Equal(t, test.repoName, repo.Name())
			}
		})
	}
}

func TestRepositoryBuilder_Build(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupRepo   func() entities.RepositoryBuilder
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_repository_with_https_url",
			setupRepo: func() entities.RepositoryBuilder {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("test-repo")
				builder, _ = builder.WithHTTPSURL("https://github.com/owner/repo.git")

				return builder
			},
			expectError: false,
		},
		{
			name: "valid_repository_with_ssh_url",
			setupRepo: func() entities.RepositoryBuilder {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("test-repo")
				builder, _ = builder.WithSSHURL("git@github.com:owner/repo.git")

				return builder
			},
			expectError: false,
		},
		{
			name: "repository_without_name",
			setupRepo: func() entities.RepositoryBuilder {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithHTTPSURL("https://github.com/owner/repo.git")

				return builder
			},
			expectError: true,
			errorMsg:    "repository name cannot be empty",
		},
		{
			name: "repository_without_urls",
			setupRepo: func() entities.RepositoryBuilder {
				builder := entities.NewRepositoryBuilder()
				builder, _ = builder.WithName("test-repo")

				return builder
			},
			expectError: true,
			errorMsg:    "invalid repository URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := test.setupRepo()

			repo, err := builder.Build()

			if test.expectError {
				require.Error(t, err)

				if test.errorMsg != "" {
					require.Contains(t, err.Error(), test.errorMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, repo.Name())
			}
		})
	}
}

func TestRepository_IsActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		lastActivity   *time.Time
		expectedActive bool
	}{
		{
			name:           "no_activity_data_assumes_active",
			lastActivity:   nil,
			expectedActive: true,
		},
		{
			name:           "recent_activity_is_active",
			lastActivity:   timePtr(time.Now().AddDate(0, -1, 0)), // 1 month ago
			expectedActive: true,
		},
		{
			name:           "old_activity_is_inactive",
			lastActivity:   timePtr(time.Now().AddDate(0, -8, 0)), // 8 months ago
			expectedActive: false,
		},
		{
			name:           "exactly_six_months_is_inactive",
			lastActivity:   timePtr(time.Now().AddDate(0, -6, -1)), // 6 months and 1 day ago
			expectedActive: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := entities.NewRepositoryBuilder()
			builder, _ = builder.WithName("test-repo")
			builder, _ = builder.WithHTTPSURL("https://github.com/owner/repo.git")

			if test.lastActivity != nil {
				builder = builder.WithLastActivityAt(*test.lastActivity)
			}

			repo, err := builder.Build()
			require.NoError(t, err)

			isActive := repo.IsActive()
			require.Equal(t, test.expectedActive, isActive)
		})
	}
}

func TestRepository_ShouldIncludeInSync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		isFork          bool
		isArchived      bool
		includeForks    bool
		includeArchived bool
		expected        bool
	}{
		{
			name:            "regular_repo_always_included",
			isFork:          false,
			isArchived:      false,
			includeForks:    false,
			includeArchived: false,
			expected:        true,
		},
		{
			name:            "fork_excluded_when_forks_disabled",
			isFork:          true,
			isArchived:      false,
			includeForks:    false,
			includeArchived: false,
			expected:        false,
		},
		{
			name:            "fork_included_when_forks_enabled",
			isFork:          true,
			isArchived:      false,
			includeForks:    true,
			includeArchived: false,
			expected:        true,
		},
		{
			name:            "archived_excluded_when_archived_disabled",
			isFork:          false,
			isArchived:      true,
			includeForks:    false,
			includeArchived: false,
			expected:        false,
		},
		{
			name:            "archived_included_when_archived_enabled",
			isFork:          false,
			isArchived:      true,
			includeForks:    false,
			includeArchived: true,
			expected:        true,
		},
		{
			name:            "archived_fork_excluded_when_both_disabled",
			isFork:          true,
			isArchived:      true,
			includeForks:    false,
			includeArchived: false,
			expected:        false,
		},
		{
			name:            "archived_fork_included_when_both_enabled",
			isFork:          true,
			isArchived:      true,
			includeForks:    true,
			includeArchived: true,
			expected:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := entities.NewRepositoryBuilder()
			builder, _ = builder.WithName("test-repo")
			builder, _ = builder.WithHTTPSURL("https://github.com/owner/repo.git")
			builder = builder.WithFlags(false, test.isFork, test.isArchived)

			repo, err := builder.Build()
			require.NoError(t, err)

			shouldInclude := repo.ShouldIncludeInSync(test.includeForks, test.includeArchived)
			require.Equal(t, test.expected, shouldInclude)
		})
	}
}

func TestValidateRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repoName    string
		expectError bool
	}{
		{
			name:        "valid_name_with_hyphens",
			repoName:    "my-awesome-repo",
			expectError: false,
		},
		{
			name:        "valid_name_with_underscores",
			repoName:    "my_awesome_repo",
			expectError: false,
		},
		{
			name:        "valid_name_with_dots",
			repoName:    "my.awesome.repo",
			expectError: false,
		},
		{
			name:        "empty_name",
			repoName:    "",
			expectError: true,
		},
		{
			name:        "name_too_long",
			repoName:    strings.Repeat("a", 101),
			expectError: true,
		},
		{
			name:        "name_with_invalid_chars",
			repoName:    "my-repo@invalid",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := entities.ValidateRepositoryName(test.repoName)

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCleanRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already_clean_name",
			input:    "my-repo",
			expected: "my-repo",
		},
		{
			name:     "name_with_spaces",
			input:    "my repo name",
			expected: "my-repo-name",
		},
		{
			name:     "name_with_special_chars",
			input:    "my@repo#name!",
			expected: "my-repo-name",
		},
		{
			name:     "name_with_consecutive_invalid_chars",
			input:    "my@@@repo",
			expected: "my-repo",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "repository",
		},
		{
			name:     "only_invalid_chars",
			input:    "@#$%",
			expected: "repository",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cleaned := entities.CleanRepositoryName(test.input)
			require.Equal(t, test.expected, cleaned)
		})
	}
}

// Helper function.
func timePtr(t time.Time) *time.Time {
	return &t
}
