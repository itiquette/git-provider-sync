// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_RemoteOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "create_and_retrieve_remote",
			testFunc: func(t *testing.T) {
				t.Helper()

				// Create in-memory repository for testing
				gitRepo, err := git.Init(memory.NewStorage(), nil)
				require.NoError(t, err)

				repo := Repository{
					goGitRepository: gitRepo,
					ProjectMetaInfo: &ProjectInfo{
						OriginalName: "test-repo",
						CleanName:    "test-repo",
					},
				}

				// Test creating a remote
				err = repo.CreateRemote("origin", "https://github.com/test/repo.git", false)
				require.NoError(t, err)

				// Test retrieving the remote
				remote, err := repo.Remote("origin")
				require.NoError(t, err)
				assert.Equal(t, "https://github.com/test/repo.git", remote.URL)
			},
		},
		{
			name: "delete_existing_remote",
			testFunc: func(t *testing.T) {
				t.Helper()

				gitRepo, err := git.Init(memory.NewStorage(), nil)
				require.NoError(t, err)

				// Add a remote first
				_, err = gitRepo.CreateRemote(&config.RemoteConfig{
					Name: "upstream",
					URLs: []string{"https://github.com/upstream/repo.git"},
				})
				require.NoError(t, err)

				repo := Repository{
					goGitRepository: gitRepo,
				}

				// Test deleting the remote
				err = repo.DeleteRemote("upstream")
				require.NoError(t, err)

				// Verify remote is deleted
				_, err = gitRepo.Remote("upstream")
				require.ErrorIs(t, err, git.ErrRemoteNotFound)
			},
		},
		{
			name: "delete_nonexistent_remote_succeeds",
			testFunc: func(t *testing.T) {
				t.Helper()

				gitRepo, err := git.Init(memory.NewStorage(), nil)
				require.NoError(t, err)

				repo := Repository{
					goGitRepository: gitRepo,
				}

				// Deleting non-existent remote should succeed (idempotent)
				err = repo.DeleteRemote("nonexistent")
				require.NoError(t, err)
			},
		},
		{
			name: "get_nonexistent_remote_returns_error",
			testFunc: func(t *testing.T) {
				t.Helper()

				gitRepo, err := git.Init(memory.NewStorage(), nil)
				require.NoError(t, err)

				repo := Repository{
					goGitRepository: gitRepo,
				}

				// Should return error for non-existent remote
				_, err = repo.Remote("nonexistent")
				require.Error(t, err)
				assert.Contains(t, err.Error(), "nonexistent")
			},
		},
		{
			name: "create_mirror_remote",
			testFunc: func(t *testing.T) {
				t.Helper()

				gitRepo, err := git.Init(memory.NewStorage(), nil)
				require.NoError(t, err)

				repo := Repository{
					goGitRepository: gitRepo,
				}

				// Create mirror remote
				err = repo.CreateRemote("mirror", "https://gitlab.com/mirror/repo.git", true)
				require.NoError(t, err)

				// Verify mirror flag is set
				remote, err := gitRepo.Remote("mirror")
				require.NoError(t, err)
				assert.True(t, remote.Config().Mirror)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestNewGitGoRemoteOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		remoteName string
		urls       []string
		isMirror   bool
	}{
		{
			name:       "standard_remote",
			remoteName: "origin",
			urls:       []string{"https://github.com/user/repo.git"},
			isMirror:   false,
		},
		{
			name:       "mirror_remote_with_multiple_urls",
			remoteName: "backup",
			urls:       []string{"https://gitlab.com/user/repo.git", "https://gitea.com/user/repo.git"},
			isMirror:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			remoteConfig := NewGitGoRemoteOption(test.remoteName, test.urls, test.isMirror)

			assert.Equal(t, test.remoteName, remoteConfig.Name)
			assert.Equal(t, test.urls, remoteConfig.URLs)
			assert.Equal(t, test.isMirror, remoteConfig.Mirror)
		})
	}
}
