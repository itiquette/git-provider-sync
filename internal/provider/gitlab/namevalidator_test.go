// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
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
		{"valid name", "valid-repo-name", true},
		{"valid name with numbers", "repo123", true},
		{"valid name with dots", "repo.name", true},
		{"valid name with plus", "repo+name", true},
		{"valid name with space", "repo name", true},
		{"invalid name with underscore", "repo_name", false}, //todo, for now dont allow underscores as gitlab have a possible bug with repos nameed for example a-_b-c
		{"invalid name with exclamation", "invalid!", false},
		{"invalid name with at symbol", "invalid@repo", false},
		{"invalid name starting with dot", ".invalidrepo", false},
		{"invalid name starting with hyphen", "-invalidrepo", false},
		{"invalid name starting with underscore", "_invalidrepo", false},
		{"invalid name ending with underscore", "invalidrepo_", false},
		{"empty name", "", false},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := isValidGitLabNameCharacters(tabletest.repoName)
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
		{"valid name length under 256", "commitsrepo", false},
		{"invalid name length over 256", "commitsriiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiepoaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
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
		{"valid name", "valid-repo-name", true},
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
			got := IsValidGitLabName(tabletest.repoName)
			assert.Equal(t, tabletest.want, got)
		})
	}
}
