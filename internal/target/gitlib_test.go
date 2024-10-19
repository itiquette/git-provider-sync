// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package target

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

func TestClone(t *testing.T) {
	t.Run("Successful clone", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := context.Background()
		repoPath := filepath.Join(tmpDir, "test-clone-repo")
		sourceRepo := setupTestRepository(t, repoPath, "clone")

		cloneOption := model.CloneOption{
			URL:       repoPath,
			Mirror:    false,
			PlainRepo: true,
			Git: gpsconfig.GitOption{
				Type: gpsconfig.HTTPS,
			},
			HTTPClient: gpsconfig.HTTPClientOption{
				Token: "dummy-token",
			},
		}

		clonedRepo, err := GitLib{}.Clone(ctx, cloneOption)
		require.NoError(t, err)
		require.NotNil(t, clonedRepo)

		verifyClonedRepository(t, sourceRepo, clonedRepo.GoGitRepository())
	})

	t.Run("Clone with invalid URL", func(t *testing.T) {
		ctx := context.Background()
		cloneOption := model.CloneOption{
			URL:       "invalid-url",
			PlainRepo: true,
			Git: gpsconfig.GitOption{
				Type: gpsconfig.HTTPS,
			},
		}

		_, err := GitLib{}.Clone(ctx, cloneOption)
		assert.Error(t, err)
	})
}

func TestPull(t *testing.T) {
	t.Run("Successful pull", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := context.Background()
		repoPath := filepath.Join(tmpDir, "test-pull-repo")
		repo := setupTestRepository(t, repoPath, "pull")

		pullOption := model.PullOption{
			Name: "origin",
			URL:  repoPath,
			GitOption: gpsconfig.GitOption{
				Type: gpsconfig.HTTPS,
			},
			HTTPClientOption: gpsconfig.HTTPClientOption{Token: ""},
			SSHClient:        gpsconfig.SSHClientOption{},
		}

		err := GitLib{}.Pull(ctx, repoPath, pullOption)
		require.NoError(t, err)

		verifyFile(t, repoPath, "test.txt", "test contentpull")
		verifyCommitMessages(t, repo, []string{"chore: add commitpull"}, 1)
	})

	t.Run("Pull with no changes", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := context.Background()
		repoPath := filepath.Join(tmpDir, "test-pull-no-changes-repo")
		_ = setupTestRepository(t, repoPath, "pull-no-changes")

		pullOption := model.PullOption{
			Name: "origin",
			URL:  repoPath,
			GitOption: gpsconfig.GitOption{
				Type: gpsconfig.HTTPS,
			},
		}

		err := GitLib{}.Pull(ctx, repoPath, pullOption)
		assert.NoError(t, err) // Should not return an error when already up-to-date
	})
}

//	t.Run("Pull with conflicts", func(t *testing.T) {
//		tmpDir, cleanup := setupTestEnvironment(t)
//		defer cleanup()
//		ctx := context.Background()
//		repoPath := filepath.Join(tmpDir, "test-pull-conflicts-repo")
//		repo := setupTestRepository(t, repoPath, "pull-conflicts")
//		// Create a conflicting change
//		createConflictingCommit(t, repo)
//		pullOption := model.PullOption{
//			Name: "origin",
//			URL:  repoPath,
//			GitOption: gpsconfig.GitOption{
//				Type: gpsconfig.HTTPS,
//			},
//		}
//		err := Git{}.Pull(ctx, repoPath, pullOption)
//		assert.Error(t, err) // Should return an error due to conflicts
//	})
//}

func testContext() context.Context {
	ctx := context.Background()
	input := model.CLIOption{CleanupName: true}
	//ctx, _ = model.CreateTmpDir(ctx, "", "testadir")

	return model.WithCLIOption(ctx, input)
}
func TestPush(t *testing.T) {
	t.Run("Successful push", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := testContext()
		repoPath := filepath.Join(tmpDir, "test-push-repo")
		repo := setupTestRepository(t, repoPath, "push")

		pushOption := model.NewPushOption(repoPath, false, false, gpsconfig.HTTPClientOption{})

		modelRepo, err := model.NewRepository(repo)
		modelRepo.Meta = model.RepositoryMetainfo{OriginalName: "test-push-repo"}

		require.NoError(t, err)

		err = NewGitLib().Push(ctx, modelRepo, pushOption, gpsconfig.ProviderConfig{}, gpsconfig.GitOption{})
		require.NoError(t, err)

		verifyFile(t, repoPath, "test.txt", "test contentpush")
		verifyCommitMessages(t, repo, []string{"chore: add commitpush"}, 1)
	})

	t.Run("Push with no changes", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := testContext()
		repoPath := filepath.Join(tmpDir, "test-push-no-changes-repo")
		repo := setupTestRepository(t, repoPath, "push-no-changes")

		po := model.NewPushOption(repoPath, false, false, gpsconfig.HTTPClientOption{})

		modelRepo, err := model.NewRepository(repo)
		require.NoError(t, err)

		err = NewGitLib().Push(ctx, modelRepo, po, gpsconfig.ProviderConfig{}, gpsconfig.GitOption{})
		assert.NoError(t, err) // Should not return an error when already up-to-date
	})

	t.Run("Push to non-existent remote", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := testContext()
		repoPath := filepath.Join(tmpDir, "test-push-non-existent-remote")
		r := setupTestRepository(t, repoPath, "push-non-existent")

		po := model.NewPushOption("non-existent-remote", false, false, gpsconfig.HTTPClientOption{})

		R, err := model.NewRepository(r)
		require.NoError(t, err)

		err = NewGitLib().Push(ctx, R, po, gpsconfig.ProviderConfig{}, gpsconfig.GitOption{})
		assert.Error(t, err)
	})
}

func TestFetch(t *testing.T) {
	t.Run("Successful fetch", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := context.Background()
		repoPath := filepath.Join(tmpDir, "test-repo-fetch")
		repo := setupTestRepository(t, repoPath, "fetch")

		modelRepo, err := model.NewRepository(repo)
		require.NoError(t, err)

		err = NewGitLib().fetch(ctx, "", modelRepo.GoGitRepository())
		require.NoError(t, err)

		verifyFile(t, repoPath, "test.txt", "test contentfetch")
		verifyCommitMessages(t, repo, []string{"chore: add commitfetch"}, 1)
	})

	t.Run("Fetch with no changes", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := context.Background()
		repoPath := filepath.Join(tmpDir, "test-repo-fetch-no-changes")
		repo := setupTestRepository(t, repoPath, "fetch-no-changes")

		modelRepo, err := model.NewRepository(repo)
		require.NoError(t, err)

		err = NewGitLib().fetch(ctx, "", modelRepo.GoGitRepository())
		assert.NoError(t, err) // Should not return an error when already up-to-date
	})

	t.Run("Fetch from non-existent remote", func(t *testing.T) {
		tmpDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		ctx := context.Background()
		repoPath := filepath.Join(tmpDir, "test-repo-fetch-non-existent")
		repo := setupTestRepository(t, repoPath, "fetch-non-existent")

		// Remove the "origin" remote
		err := repo.DeleteRemote("origin")
		require.NoError(t, err)

		modelRepo, err := model.NewRepository(repo)
		require.NoError(t, err)

		err = NewGitLib().fetch(ctx, "", modelRepo.GoGitRepository())
		assert.Error(t, err)
	})
}

// func createConflictingCommit(t *testing.T, repo *git.Repository) {
// 	w, err := repo.Worktree()
// 	require.NoError(t, err)

// 	// Create a new branch
// 	err = w.Checkout(&git.CheckoutOptions{
// 		Branch: plumbing.NewBranchReferenceName("conflict-branch"),
// 		Create: true,
// 	})
// 	require.NoError(t, err)

// 	// Make a conflicting change
// 	filename := "test.txt"
// 	filePath := filepath.Join(w.Filesystem.Root(), filename)
// 	err = os.WriteFile(filePath, []byte("conflicting content"), 0644)
// 	require.NoError(t, err)

// 	_, err = w.Add(filename)
// 	require.NoError(t, err)

// 	_, err = w.Commit("Conflicting commit", &git.CommitOptions{
// 		Author: &object.Signature{
// 			Name:  "test",
// 			Email: "test@example.com",
// 		},
// 	})
// 	require.NoError(t, err)

// 	// Switch back to the main branch
// 	err = w.Checkout(&git.CheckoutOptions{
// 		Branch: plumbing.NewBranchReferenceName("master"),
// 	})
// 	require.NoError(t, err)
// }

func verifyClonedRepository(t *testing.T, sourceRepo, clonedRepo *git.Repository) {
	t.Helper()

	sourceHead, err := sourceRepo.Head()
	require.NoError(t, err)

	clonedHead, err := clonedRepo.Head()
	require.NoError(t, err)

	assert.Equal(t, sourceHead.Hash(), clonedHead.Hash(), "HEAD references should match")

	verifyCommitHistory(t, sourceRepo, clonedRepo)
	verifyFileContents(t, sourceRepo, clonedRepo)
}

func verifyFile(t *testing.T, repoPath, filename, expectedContent string) {
	t.Helper()

	filePath := filepath.Join(repoPath, filename)
	_, err := os.Stat(filePath)
	require.NoError(t, err, "File should exist after operation")

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(content), "File content should match")
}

func verifyCommitMessages(t *testing.T, repo *git.Repository, expectedMessages []string, expectedCount int) {
	t.Helper()

	ref, err := repo.Head()
	require.NoError(t, err)

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	require.NoError(t, err)

	commitMessages := make([]string, 0)
	err = commitIter.ForEach(func(c *object.Commit) error {
		commitMessages = append(commitMessages, c.Message)

		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, expectedCount, len(commitMessages), fmt.Sprintf("Expected %d commits, but found %d", expectedCount, len(commitMessages))) //nolint

	for _, expectedMsg := range expectedMessages {
		assert.Contains(t, commitMessages, expectedMsg, "Expected commit message not found: "+expectedMsg)
	}
}

func verifyCommitHistory(t *testing.T, sourceRepo, clonedRepo *git.Repository) {
	t.Helper()

	sourceIter, err := sourceRepo.Log(&git.LogOptions{})
	require.NoError(t, err)

	clonedIter, err := clonedRepo.Log(&git.LogOptions{})
	require.NoError(t, err)

	for {
		sourceCommit, sourceErr := sourceIter.Next()
		clonedCommit, clonedErr := clonedIter.Next()

		if sourceErr != nil || clonedErr != nil {
			assert.Equal(t, sourceErr, clonedErr, "Commit history should end at the same point")

			break
		}

		assert.Equal(t, sourceCommit.Hash, clonedCommit.Hash, "Commit hashes should match")
		assert.Equal(t, sourceCommit.Message, clonedCommit.Message, "Commit messages should match")
	}
}

func verifyFileContents(t *testing.T, sourceRepo, clonedRepo *git.Repository) {
	t.Helper()

	sourceWorktree, err := sourceRepo.Worktree()
	require.NoError(t, err)

	clonedWorktree, err := clonedRepo.Worktree()
	require.NoError(t, err)

	sourceFiles, err := sourceWorktree.Filesystem.ReadDir("/")
	require.NoError(t, err)

	for _, file := range sourceFiles {
		if !file.IsDir() {
			sourceContent, err := sourceWorktree.Filesystem.Open(file.Name())
			require.NoError(t, err)
			defer sourceContent.Close()

			clonedContent, err := clonedWorktree.Filesystem.Open(file.Name())
			require.NoError(t, err)
			defer clonedContent.Close()

			sourceBytes, err := io.ReadAll(sourceContent)
			require.NoError(t, err)

			clonedBytes, err := io.ReadAll(clonedContent)
			require.NoError(t, err)

			assert.Equal(t, sourceBytes, clonedBytes, "File contents should match")
		}
	}
}

func createCommit(t *testing.T, repo *git.Repository, message, content string) {
	t.Helper()

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	filename := "test.txt"
	filePath := filepath.Join(worktree.Filesystem.Root(), filename)
	err = os.WriteFile(filePath, []byte("test content"+content), 0600)
	require.NoError(t, err)

	_, err = worktree.Add(filename)
	require.NoError(t, err)

	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)
}

func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "git-test")
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func setupTestRepository(t *testing.T, repoPath string, message string) *git.Repository {
	t.Helper()

	repo, err := git.PlainInit(repoPath, false)
	require.NoError(t, err)

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoPath},
	})
	require.NoError(t, err)

	createCommit(t, repo, "chore: add commit"+message, message)

	return repo
}
