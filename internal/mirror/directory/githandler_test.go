// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:wrapcheck
package directory

import (
	"context"
	"io"
	"itiquette/git-provider-sync/internal/mirror/gitlib"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func createTmpGitBareRepo(tmpDir, tmpRepo string) (*git.Repository, error) {
	path := filepath.Join(tmpDir, tmpRepo)
	tempPath := filepath.Join(tmpDir, "temp-"+tmpRepo)

	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempPath) // Clean up temp repo at the end

	// Initialize temporary repo
	tempRep, err := git.PlainInitWithOptions(tempPath, &git.PlainInitOptions{
		Bare: false,
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName("main"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create remotes
	if err := setupRemotes(tempRep, tmpRepo); err != nil {
		return nil, err
	}

	// Create and commit content
	if err := createInitialCommit(tempRep, tempPath); err != nil {
		return nil, err
	}

	// Create and setup bare repo
	bareRep, err := setupBareRepo(path, tempRep, tmpRepo)
	if err != nil {
		return nil, err
	}

	return bareRep, nil
}

func setupRemotes(repo *git.Repository, repoName string) error {
	remotes := []struct {
		name string
		url  string
	}{
		{gpsconfig.ORIGIN, "https://origin.dot/" + repoName + ".git"},
		{gpsconfig.GPSUPSTREAM, "http://gpsupstream.dot/anotherrepo.git"},
	}

	for _, remote := range remotes {
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: remote.name,
			URLs: []string{remote.url},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func createInitialCommit(repo *git.Repository, repoPath string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	filename := filepath.Join(repoPath, "example-git-file")
	if err := os.WriteFile(filename, []byte("hello world!"), 0600); err != nil {
		return err
	}

	if _, err := worktree.Add("example-git-file"); err != nil {
		return err
	}

	_, err = worktree.Commit("chore: add example file", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Laval",
			Email: "laval@cavora.chi",
			When:  time.Now(),
		},
	})

	return err
}

func setupBareRepo(path string, sourceRepo *git.Repository, repoName string) (*git.Repository, error) {
	bareRepo, err := git.PlainInit(path, true)
	if err != nil {
		return nil, err
	}

	// Setup remote for push
	_, err = sourceRepo.CreateRemote(&config.RemoteConfig{
		Name: "bare",
		URLs: []string{path},
	})
	if err != nil {
		return nil, err
	}

	// Push to bare repo
	err = sourceRepo.Push(&git.PushOptions{
		RemoteName: "bare",
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/heads/*:refs/heads/*"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Setup remotes in bare repo
	if err := setupRemotes(bareRepo, repoName); err != nil {
		return nil, err
	}

	return bareRepo, nil
}

func validateRepoRestored(t *testing.T, restoredRepo *git.Repository, tmpRepo string) {
	t.Helper()

	// Validate branch
	ref, err := restoredRepo.Reference(plumbing.NewBranchReferenceName("main"), true)
	require.NoError(t, err)
	require.NotNil(t, ref)

	// Validate remotes
	origin, err := restoredRepo.Remote(gpsconfig.ORIGIN)
	require.NoError(t, err)
	require.NotNil(t, origin)
	require.Equal(t, []string{"https://origin.dot/" + tmpRepo + ".git"}, origin.Config().URLs)

	upstream, err := restoredRepo.Remote(gpsconfig.GPSUPSTREAM)
	require.Error(t, err)
	require.Nil(t, upstream)

	validateContent(t, restoredRepo)
	validateCommit(t, restoredRepo)
	validateConfig(t, restoredRepo)
}

func validateContent(t *testing.T, repo *git.Repository) {
	t.Helper()

	w, err := repo.Worktree()
	require.NoError(t, err)

	fileContent, err := w.Filesystem.Open("example-git-file")
	require.NoError(t, err)
	defer fileContent.Close()

	content, err := io.ReadAll(fileContent)
	require.NoError(t, err)
	require.Equal(t, "hello world!", string(content))
}

func validateCommit(t *testing.T, repo *git.Repository) {
	t.Helper()

	head, err := repo.Head()
	require.NoError(t, err)

	commit, err := repo.CommitObject(head.Hash())
	require.NoError(t, err)
	require.Equal(t, "chore: add example file", commit.Message)
	require.Equal(t, "Laval", commit.Author.Name)
	require.Equal(t, "laval@cavora.chi", commit.Author.Email)
}

func validateConfig(t *testing.T, repo *git.Repository) {
	t.Helper()

	config, err := repo.Config()
	require.NoError(t, err)
	require.False(t, config.Core.IsBare)
}

func testContext() context.Context {
	return model.WithCLIOpt(context.Background(), model.CLIOption{
		AlphaNumHyphName: true,
		DryRun:           false,
	})
}

func TestInitializeRepository(t *testing.T) {
	tmpDir := t.TempDir()
	initTestRepoName := "inittestrepo"

	tests := []struct {
		name          string
		setupRepo     func(t *testing.T) *git.Repository
		wantErr       bool
		expectedError error
	}{
		{
			name: "successful initialization in empty directory",
			setupRepo: func(t *testing.T) *git.Repository {
				t.Helper()
				repo, err := createTmpGitBareRepo(tmpDir, initTestRepoName)
				require.NoError(t, err)

				return repo
			},
			wantErr: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			bareRepo := tabletest.setupRepo(t)
			handler := NewGitHandler(gitlib.NewService())

			bareRepository, err := model.NewRepository(bareRepo)
			require.NoError(t, err)

			bareRepository.ProjectMetaInfo = &model.ProjectInfo{}

			bareRepository.ProjectMetaInfo.DefaultBranch = "main"

			targetTmpDir := t.TempDir()
			err = handler.InitializeRepository(testContext(), targetTmpDir, bareRepository)

			if tabletest.wantErr {
				require.Error(t, err)

				if tabletest.expectedError != nil {
					require.ErrorIs(t, err, tabletest.expectedError)
				}

				return
			}

			require.NoError(t, err)
			restoredRepo, err := git.PlainOpen(targetTmpDir)
			require.NoError(t, err)

			validateRepoRestored(t, restoredRepo, initTestRepoName)
		})
	}
}
