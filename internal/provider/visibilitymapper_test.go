// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapVisibility(t *testing.T) {
	tests := []struct {
		name          string
		fromProvider  string
		toProvider    string
		visibility    string
		expected      string
		expectedError string
	}{
		// Same source as target
		{"GitHub to GitHub", "github", "github", "public", "public", ""},
		{"GitLab to GitLab", "gitlab", "gitlab", "private", "private", ""},
		{"Gitea to Gitea", "gitea", "gitea", "private", "private", ""},
		// Successful mappings
		{"GitLab Public to GitHub", "gitlab", "github", "public", "public", ""},
		{"GitLab Internal to GitHub", "gitlab", "github", "internal", "private", ""},
		{"GitLab Private to GitHub", "gitlab", "github", "private", "private", ""},

		{"GitLab Public to Gitea", "gitlab", "gitea", "public", "public", ""},
		{"GitLab Internal to Gitea", "gitlab", "gitea", "internal", "private", ""},
		{"GitLab Private to Gitea", "gitlab", "gitea", "private", "private", ""},

		{"GitHub Public to GitLab", "github", "gitlab", "public", "public", ""},
		{"GitHub Private to GitLab", "github", "gitlab", "private", "private", ""},

		{"GitHub Public to Gitea", "github", "gitea", "public", "public", ""},
		{"GitHub Private to Gitea", "github", "gitea", "private", "private", ""},

		{"Gitea Public to GitLab", "gitea", "gitlab", "public", "public", ""},
		{"Gitea Private to GitLab", "gitea", "gitlab", "private", "private", ""},
		{"Gitea Limited to GitLab", "gitea", "gitlab", "limited", "private", ""},
		{"Gitea Public to GitHub", "gitea", "github", "public", "public", ""},
		{"Gitea Private to GitHub", "gitea", "github", "private", "private", ""},
		{"Gitea Limited to GitHub", "gitea", "github", "limited", "private", ""},

		// Case insensitivity tests
		{"Case Insensitive Provider", "GitLab", "GitHub", "Public", "public", ""},
		{"Case Insensitive Visibility", "gitlab", "github", "INTERNAL", "private", ""},

		// Error cases
		{"Invalid Source Provider", "invalid", "github", "public", "", "invalid source provider: invalid"},
		{"Invalid Target Provider", "gitlab", "invalid", "public", "", "invalid target provider: invalid"},
		{"Invalid GitLab Visibility", "gitlab", "github", "invalid", "", "invalid visibility for gitlab: invalid"},
		{"Invalid GitHub Visibility", "github", "gitlab", "internal", "", "invalid visibility for github: internal"},
		{"Invalid Gitea Visibility", "gitea", "github", "internal", "", "invalid visibility for gitea: internal"},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result, err := mapVisibility(tabletest.fromProvider, tabletest.toProvider, tabletest.visibility)

			if tabletest.expectedError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tabletest.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tabletest.expected, result)
			}
		})
	}
}
