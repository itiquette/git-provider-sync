// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package targetfilter

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsInInterval(t *testing.T) {
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
		t.Run(tabletest.name, func(t *testing.T) {
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
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestFilterIncludedExcluded(t *testing.T) {
	tests := []struct {
		name     string
		projects []model.ProjectInfo
		opt      model.ProviderOption
		expected []model.ProjectInfo
	}{
		{
			name: "empty include/exclude returns all",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
			},
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
			},
			opt: model.ProviderOption{},
		},
		{
			name: "included list filters",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
				{OriginalName: "repo3"},
			},
			opt: model.ProviderOption{
				IncludedRepositories: []string{"repo1", "repo3"},
			},
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo3"},
			},
		},
		{
			name: "excluded list filters",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
				{OriginalName: "repo3"},
			},
			opt: model.ProviderOption{
				ExcludedRepositories: []string{"repo2"},
			},
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo3"},
			},
		},
		{
			name: "include takes precedence",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
				{OriginalName: "repo3"},
			},
			opt: model.ProviderOption{
				IncludedRepositories: []string{"repo1"},
				ExcludedRepositories: []string{"repo1", "repo2"},
			},
			expected: []model.ProjectInfo{
				{OriginalName: "repo1"},
			},
		},
		{
			name: "include/exclude with no matching repos",
			projects: []model.ProjectInfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
			},
			opt: model.ProviderOption{
				IncludedRepositories: []string{"repo3", "repo4"},
			},
			expected: []model.ProjectInfo{},
		},
		{
			name: "case sensitive matching",
			projects: []model.ProjectInfo{
				{OriginalName: "Repo1"},
				{OriginalName: "repo1"},
			},
			opt: model.ProviderOption{
				IncludedRepositories: []string{"Repo1"},
			},
			expected: []model.ProjectInfo{
				{OriginalName: "Repo1"},
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			filterFunc := FilterIncludedExcludedGen()
			result, err := filterFunc(context.Background(), tabletest.opt, tabletest.projects)
			require.NoError(t, err)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestShouldIncludeRepo(t *testing.T) {
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
			name:     "included list - repo in list",
			repoName: "repo1",
			included: []string{"repo1", "repo2"},
			expected: true,
		},
		{
			name:     "included list - repo not in list",
			repoName: "repo3",
			included: []string{"repo1", "repo2"},
			expected: false,
		},
		{
			name:     "excluded list - repo in list",
			repoName: "repo1",
			excluded: []string{"repo1", "repo2"},
			expected: false,
		},
		{
			name:     "excluded list - repo not in list",
			repoName: "repo3",
			excluded: []string{"repo1", "repo2"},
			expected: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := shouldIncludeRepo(tabletest.repoName, tabletest.included, tabletest.excluded)
			require.Equal(t, tabletest.expected, result)
		})
	}
}
