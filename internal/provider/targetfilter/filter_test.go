// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package targetfilter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

func TestIsInInterval(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		name        string
		updatedAt   time.Time
		limitOption string
		expected    bool
		expectError bool
	}{
		{
			name:      "zero time returns true",
			updatedAt: time.Time{},
			expected:  true,
		},
		{
			name:        "empty limit returns true",
			updatedAt:   time.Now(),
			limitOption: "",
			expected:    true,
		},
		{
			name:        "invalid duration returns error",
			updatedAt:   time.Now(),
			limitOption: "invalid",
			expectError: true,
		},
		{
			name:        "time within interval",
			updatedAt:   time.Now(),
			limitOption: "-24h",
			expected:    true,
		},
		{
			name:        "time outside interval",
			updatedAt:   time.Now().Add(-48 * time.Hour),
			limitOption: "-24h",
			expected:    false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			ctx := context.Background()
			if tabletest.limitOption != "" {
				ctx = context.WithValue(ctx, model.CLIOptionKey{}, model.CLIOption{
					ActiveFromLimit: tabletest.limitOption,
				})
			} else {
				ctx = context.WithValue(ctx, model.CLIOptionKey{}, model.CLIOption{})
			}

			result, err := IsInInterval(ctx, tabletest.updatedAt)

			if tabletest.expectError {
				assert.Error(err)

				return
			}

			assert.NoError(err)
			assert.Equal(tabletest.expected, result)
		})
	}
}

func TestFilterIncludedExcluded(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		name     string
		projects []model.ProjectInfo
		included string
		excluded string
		expected []model.ProjectInfo
	}{
		{
			name: "empty include/exclude lists returns all",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
			},
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
			},
			included: "",
			excluded: "",
		},
		{
			name: "include list filters correctly",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
				{OriginalName: "repo3"},
			},
			included: "repo1, repo3",
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo3"},
			},
		},
		{
			name: "exclude list filters correctly",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
				{OriginalName: "repo3"},
			},
			excluded: "repo2",
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo3"},
			},
		},
		{
			name: "include list takes precedence",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
				{OriginalName: "repo3"},
			},
			included: "repo1",
			excluded: "repo1,repo2",
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			filterFunc := FilterIncludedExcludedGen()

			cfg := config.ProviderConfig{
				Repositories: config.RepositoriesOption{
					Include: tabletest.included,
					Exclude: tabletest.excluded,
				},
			}

			result, err := filterFunc(context.Background(), cfg, tabletest.projects)

			assert.NoError(err)
			assert.Equal(tabletest.expected, result)
		})
	}
}

func TestShouldIncludeRepo(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		name     string
		repoName string
		included []string
		excluded []string
		expected bool
	}{
		{
			name:     "empty lists includes all",
			repoName: "repo1",
			expected: true,
		},
		{
			name:     "included list only - repo in list",
			repoName: "repo1",
			included: []string{"repo1", "repo2"},
			expected: true,
		},
		{
			name:     "included list only - repo not in list",
			repoName: "repo3",
			included: []string{"repo1", "repo2"},
			expected: false,
		},
		{
			name:     "excluded list only - repo in list",
			repoName: "repo1",
			excluded: []string{"repo1", "repo2"},
			expected: false,
		},
		{
			name:     "excluded list only - repo not in list",
			repoName: "repo3",
			excluded: []string{"repo1", "repo2"},
			expected: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			result := shouldIncludeRepo(tabletest.repoName, tabletest.included, tabletest.excluded)
			assert.Equal(tabletest.expected, result)
		})
	}
}
