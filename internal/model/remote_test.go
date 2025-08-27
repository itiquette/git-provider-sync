// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "remote with HTTPS URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote := Remote{
					URL: "https://github.com/user/repo.git",
				}

				assert.Equal(t, "https://github.com/user/repo.git", remote.URL)
			},
		},
		{
			name: "remote with SSH URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote := Remote{
					URL: "git@gitlab.com:user/repo.git",
				}

				assert.Equal(t, "git@gitlab.com:user/repo.git", remote.URL)
			},
		},
		{
			name: "remote with empty URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote := Remote{
					URL: "",
				}

				assert.Empty(t, remote.URL)
			},
		},
		{
			name: "zero value remote",
			testFunc: func(t *testing.T) {
				t.Helper()
				var remote Remote

				assert.Empty(t, remote.URL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestRemoteCreation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		url         string
		expectedURL string
	}{
		{
			name:        "create remote with GitHub HTTPS URL",
			url:         "https://github.com/owner/repository.git",
			expectedURL: "https://github.com/owner/repository.git",
		},
		{
			name:        "create remote with GitLab SSH URL",
			url:         "git@gitlab.com:owner/repository.git",
			expectedURL: "git@gitlab.com:owner/repository.git",
		},
		{
			name:        "create remote with Gitea URL",
			url:         "https://gitea.example.com/owner/repository.git",
			expectedURL: "https://gitea.example.com/owner/repository.git",
		},
		{
			name:        "create remote with Bitbucket URL",
			url:         "https://bitbucket.org/owner/repository.git",
			expectedURL: "https://bitbucket.org/owner/repository.git",
		},
		{
			name:        "create remote with custom domain",
			url:         "https://git.company.com/team/project.git",
			expectedURL: "https://git.company.com/team/project.git",
		},
		{
			name:        "create remote with URL containing authentication",
			url:         "https://user:token@github.com/owner/repo.git",
			expectedURL: "https://user:token@github.com/owner/repo.git",
		},
		{
			name:        "create remote with file protocol",
			url:         "file:///local/path/to/repo.git",
			expectedURL: "file:///local/path/to/repo.git",
		},
		{
			name:        "create remote with relative path",
			url:         "../relative/path/to/repo.git",
			expectedURL: "../relative/path/to/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			remote := Remote{URL: tt.url}
			assert.Equal(t, tt.expectedURL, remote.URL)
		})
	}
}

func TestRemoteComparison(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "remotes with same URL are equal",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote1 := Remote{URL: "https://github.com/user/repo.git"}
				remote2 := Remote{URL: "https://github.com/user/repo.git"}

				assert.Equal(t, remote1, remote2)
			},
		},
		{
			name: "remotes with different URLs are not equal",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote1 := Remote{URL: "https://github.com/user/repo1.git"}
				remote2 := Remote{URL: "https://github.com/user/repo2.git"}

				assert.NotEqual(t, remote1, remote2)
			},
		},
		{
			name: "remote with empty URL vs remote with URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote1 := Remote{URL: ""}
				remote2 := Remote{URL: "https://github.com/user/repo.git"}

				assert.NotEqual(t, remote1, remote2)
			},
		},
		{
			name: "two empty remotes are equal",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote1 := Remote{URL: ""}
				remote2 := Remote{URL: ""}

				assert.Equal(t, remote1, remote2)
			},
		},
		{
			name: "zero value remotes are equal",
			testFunc: func(t *testing.T) {
				t.Helper()
				var remote1 Remote
				var remote2 Remote

				assert.Equal(t, remote1, remote2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestRemote_InvalidURLsAndEmptyValues_HandlesCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "remote with very long URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				longURL := "https://very-long-domain-name-that-exceeds-normal-url-length-limits.example.com/organization-with-extremely-long-name/repository-with-very-long-name-that-might-cause-url-parsing-issues.git"

				remote := Remote{URL: longURL}
				assert.Equal(t, longURL, remote.URL)
			},
		},
		{
			name: "remote with special characters in URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				specialURL := "https://user%40domain:p%40ssw0rd@github.com/owner/repo-with-special@chars.git"

				remote := Remote{URL: specialURL}
				assert.Equal(t, specialURL, remote.URL)
			},
		},
		{
			name: "remote with unicode characters",
			testFunc: func(t *testing.T) {
				t.Helper()
				unicodeURL := "https://github.com/用户/项目-测试-🚀.git"

				remote := Remote{URL: unicodeURL}
				assert.Equal(t, unicodeURL, remote.URL)
			},
		},
		{
			name: "remote with whitespace in URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				whitespaceURL := "  https://github.com/user/repo.git  "

				remote := Remote{URL: whitespaceURL}
				// Remote struct just stores the URL as-is, doesn't trim
				assert.Equal(t, whitespaceURL, remote.URL)
			},
		},
		{
			name: "remote with newlines in URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				newlineURL := "https://github.com/user/repo.git\n"

				remote := Remote{URL: newlineURL}
				assert.Equal(t, newlineURL, remote.URL)
			},
		},
		{
			name: "remote with tab characters in URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				tabURL := "https://github.com/user/repo.git\t"

				remote := Remote{URL: tabURL}
				assert.Equal(t, tabURL, remote.URL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestRemoteUsageScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "remote as map value",
			testFunc: func(t *testing.T) {
				t.Helper()
				remotes := map[string]Remote{
					"origin":   {URL: "https://github.com/user/repo.git"},
					"upstream": {URL: "https://github.com/upstream/repo.git"},
					"fork":     {URL: "git@github.com:fork/repo.git"},
				}

				assert.Equal(t, "https://github.com/user/repo.git", remotes["origin"].URL)
				assert.Equal(t, "https://github.com/upstream/repo.git", remotes["upstream"].URL)
				assert.Equal(t, "git@github.com:fork/repo.git", remotes["fork"].URL)
			},
		},
		{
			name: "remote in slice",
			testFunc: func(t *testing.T) {
				t.Helper()
				remotes := []Remote{
					{URL: "https://github.com/user/repo1.git"},
					{URL: "https://github.com/user/repo2.git"},
					{URL: "git@gitlab.com:user/repo3.git"},
				}

				assert.Len(t, remotes, 3)
				assert.Equal(t, "https://github.com/user/repo1.git", remotes[0].URL)
				assert.Equal(t, "https://github.com/user/repo2.git", remotes[1].URL)
				assert.Equal(t, "git@gitlab.com:user/repo3.git", remotes[2].URL)
			},
		},
		{
			name: "remote modification",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote := Remote{URL: "https://github.com/user/old-repo.git"}

				// Verify initial URL
				assert.Equal(t, "https://github.com/user/old-repo.git", remote.URL)

				// Modify URL
				remote.URL = "https://github.com/user/new-repo.git"
				assert.Equal(t, "https://github.com/user/new-repo.git", remote.URL)
			},
		},
		{
			name: "remote pointer operations",
			testFunc: func(t *testing.T) {
				t.Helper()
				remote := &Remote{URL: "https://github.com/user/repo.git"}

				assert.Equal(t, "https://github.com/user/repo.git", remote.URL)

				// Modify through pointer
				remote.URL = "https://github.com/user/modified-repo.git"
				assert.Equal(t, "https://github.com/user/modified-repo.git", remote.URL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}
