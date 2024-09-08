// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidGitLabRepositoryNameCharacters(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		want     bool
	}{
		{"valid name", "valid-repo_name", true},
		{"valid name with numbers", "repo123", true},
		{"valid name with dots", "repo.name", true},
		{"valid name with plus", "repo+name", true},
		{"valid name with space", "repo name", true},
		{"invalid name with exclamation", "invalid!", false},
		{"invalid name with at symbol", "invalid@repo", false},
		{"invalid name starting with dot", ".invalidrepo", false},
		{"invalid name starting with hyphen", "-invalidrepo", false},
		{"empty name", "", false},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := isValidGitLabRepositoryNameCharacters(tabletest.repoName)
			assert.Equal(t, tabletest.want, got)
		})
	}
}

func TestIsInvalidGitLabRepositoryName(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		want     bool
	}{
		{"valid name", "validrepo", false},
		{"invalid name - commits", "commits", true},
		{"invalid name - case insensitive", "CoMmItS", true},
		{"invalid name - create", "create", true},
		{"invalid name - gitlab-lfs/objects", "gitlab-lfs/objects", true},
		{"valid name similar to invalid", "commitsrepo", false},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := isInvalidGitLabRepositoryName(tabletest.repoName)
			assert.Equal(t, tabletest.want, got)
		})
	}
}

func TestIsValidGitLabRepositoryName(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		want     bool
	}{
		{"valid name", "valid-repo_name", true},
		{"valid name with numbers", "repo123", true},
		{"valid name with dots", "repo.name", true},
		{"invalid character", "invalid!", false},
		{"invalid reserved name", "commits", false},
		{"invalid reserved name - case insensitive", "CoMmItS", false},
		{"invalid name starting with hyphen", "-invalidrepo", false},
		{"empty name", "", false},
		{"name with space", "repo name", true},
		{"name with plus", "repo+name", true},
		{"name similar to reserved but valid", "commitsrepo", true},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := IsValidGitLabRepositoryName(tabletest.repoName)
			assert.Equal(t, tabletest.want, got)
		})
	}
}
