// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestToMirrorsUseCase_transformRepositoryName(t *testing.T) {
	t.Parallel()

	// Create a use case instance - we only need to test the method
	useCase := ToMirrorsUseCase{}

	tests := []struct {
		name         string
		originalName string
		options      ports.NameTransformOptions
		want         string
	}{
		{
			name:         "no transformations",
			originalName: "my-repo",
			options:      ports.NameTransformOptions{},
			want:         "my-repo",
		},
		{
			name:         "prefix only",
			originalName: "repo",
			options: ports.NameTransformOptions{
				Prefix: "backup-",
			},
			want: "backup-repo",
		},
		{
			name:         "suffix only",
			originalName: "repo",
			options: ports.NameTransformOptions{
				Suffix: "-mirror",
			},
			want: "repo-mirror",
		},
		{
			name:         "prefix and suffix",
			originalName: "repo",
			options: ports.NameTransformOptions{
				Prefix: "org-",
				Suffix: "-backup",
			},
			want: "org-repo-backup",
		},
		{
			name:         "single replacement",
			originalName: "my-old-repo",
			options: ports.NameTransformOptions{
				Replacements: map[string]string{
					"old": "new",
				},
			},
			want: "my-new-repo",
		},
		{
			name:         "multiple replacements",
			originalName: "old-repo-old",
			options: ports.NameTransformOptions{
				Replacements: map[string]string{
					"old":  "new",
					"repo": "project",
				},
			},
			want: "new-project-new",
		},
		{
			name:         "to lowercase",
			originalName: "MyRepo",
			options: ports.NameTransformOptions{
				ToLowercase: true,
			},
			want: "myrepo",
		},
		{
			name:         "to uppercase",
			originalName: "myrepo",
			options: ports.NameTransformOptions{
				ToUppercase: true,
			},
			want: "MYREPO",
		},
		{
			name:         "lowercase takes precedence over uppercase",
			originalName: "MyRepo",
			options: ports.NameTransformOptions{
				ToLowercase: true,
				ToUppercase: true, // This should be ignored
			},
			want: "myrepo",
		},
		{
			name:         "complex transformation",
			originalName: "OldRepo",
			options: ports.NameTransformOptions{
				Prefix: "backup-",
				Suffix: "-v2",
				Replacements: map[string]string{
					"Old": "New",
				},
				ToLowercase: true,
			},
			want: "backup-newrepo-v2",
		},
		{
			name:         "empty original name",
			originalName: "",
			options: ports.NameTransformOptions{
				Prefix: "prefix-",
				Suffix: "-suffix",
			},
			want: "prefix--suffix",
		},
		{
			name:         "replacement with empty string",
			originalName: "remove-this-part",
			options: ports.NameTransformOptions{
				Replacements: map[string]string{
					"-this": "",
				},
			},
			want: "remove-part",
		},
		{
			name:         "replacement that doesn't match",
			originalName: "my-repo",
			options: ports.NameTransformOptions{
				Replacements: map[string]string{
					"notfound": "replacement",
				},
			},
			want: "my-repo",
		},
		{
			name:         "case transformation with special characters",
			originalName: "My-Repo_123",
			options: ports.NameTransformOptions{
				ToUppercase: true,
			},
			want: "MY-REPO_123",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.transformRepositoryName(test.originalName, test.options)
			assert.Equal(t, test.want, result)
		})
	}
}
