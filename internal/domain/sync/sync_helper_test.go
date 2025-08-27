// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"itiquette/git-provider-sync/internal/domain/entities"
)

func TestRepositoriesUseCase_convertProviderType(t *testing.T) {
	t.Parallel()

	// Create a use case instance
	useCase := RepositoriesUseCase{}

	tests := []struct {
		name         string
		providerType string
		want         entities.ProviderType
	}{
		{
			name:         "github lowercase",
			providerType: "github",
			want:         entities.ProviderTypeGitHub,
		},
		{
			name:         "GitHub mixed case",
			providerType: "GitHub",
			want:         entities.ProviderTypeGitHub,
		},
		{
			name:         "GITHUB uppercase",
			providerType: "GITHUB",
			want:         entities.ProviderTypeGitHub,
		},
		{
			name:         "gitlab lowercase",
			providerType: "gitlab",
			want:         entities.ProviderTypeGitLab,
		},
		{
			name:         "GitLab mixed case",
			providerType: "GitLab",
			want:         entities.ProviderTypeGitLab,
		},
		{
			name:         "GITLAB uppercase",
			providerType: "GITLAB",
			want:         entities.ProviderTypeGitLab,
		},
		{
			name:         "gitea lowercase",
			providerType: "gitea",
			want:         entities.ProviderTypeGitea,
		},
		{
			name:         "Gitea mixed case",
			providerType: "Gitea",
			want:         entities.ProviderTypeGitea,
		},
		{
			name:         "GITEA uppercase",
			providerType: "GITEA",
			want:         entities.ProviderTypeGitea,
		},
		{
			name:         "directory lowercase",
			providerType: "directory",
			want:         entities.ProviderTypeDirectory,
		},
		{
			name:         "Directory mixed case",
			providerType: "Directory",
			want:         entities.ProviderTypeDirectory,
		},
		{
			name:         "DIRECTORY uppercase",
			providerType: "DIRECTORY",
			want:         entities.ProviderTypeDirectory,
		},
		{
			name:         "archive lowercase",
			providerType: "archive",
			want:         entities.ProviderTypeArchive,
		},
		{
			name:         "Archive mixed case",
			providerType: "Archive",
			want:         entities.ProviderTypeArchive,
		},
		{
			name:         "ARCHIVE uppercase",
			providerType: "ARCHIVE",
			want:         entities.ProviderTypeArchive,
		},
		{
			name:         "unknown provider defaults to github",
			providerType: "unknown",
			want:         entities.ProviderTypeGitHub,
		},
		{
			name:         "empty string defaults to github",
			providerType: "",
			want:         entities.ProviderTypeGitHub,
		},
		{
			name:         "whitespace defaults to github",
			providerType: "   ",
			want:         entities.ProviderTypeGitHub,
		},
		{
			name:         "invalid provider defaults to github",
			providerType: "bitbucket",
			want:         entities.ProviderTypeGitHub,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.convertProviderType(test.providerType)
			assert.Equal(t, test.want, result)
		})
	}
}
