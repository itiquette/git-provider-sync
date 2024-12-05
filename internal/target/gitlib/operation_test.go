// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:wrapcheck
package gitlib

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTmpGitBareRepo(tmpDir, tmpRepo string) (*git.Repository, error) {
	tempPath := filepath.Join(tmpDir, "temp-"+tmpRepo)

	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return nil, err
	}

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
	if err := setupRemotes(tempRep, tmpRepo, "https://origin.dot/"+tmpRepo+".git"); err != nil {
		return nil, err
	}

	// Create and commit content
	if err := createInitialCommit(tempRep, tempPath); err != nil {
		return nil, err
	}

	return tempRep, nil
}
func createTmpGitRepo(tmpDir, tmpRepo, url string) (*git.Repository, error) { //nolint:unparam
	tempPath := filepath.Join(tmpDir, "temp-"+tmpRepo)

	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return nil, err
	}

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
	if err := setupRemotes(tempRep, tmpRepo, url); err != nil {
		return nil, err
	}

	// Create and commit content
	if err := createInitialCommit(tempRep, tempPath); err != nil {
		return nil, err
	}

	return tempRep, nil
}

func setupRemotes(repo *git.Repository, _, url string) error {
	remotes := []struct {
		name string
		url  string
	}{
		{gpsconfig.ORIGIN, url},
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

func TestNewOperation(t *testing.T) {
	oper := NewOperation()
	assert.NotNil(t, oper)
	assert.IsType(t, &operation{}, oper)
}

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()
	oper := NewOperation()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Invalid path",
			path:    "/nonexistent/path",
			wantErr: true,
		},
		{
			name:    "Valid repository",
			path:    filepath.Join(tmpDir, "temp-test-open"),
			wantErr: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			if tabletest.wantErr {
				repo, err := oper.Open(context.Background(), tabletest.path)
				require.Error(t, err)
				require.Nil(t, repo)
			} else {
				repo, err := createTmpGitBareRepo(tmpDir, "test-open")
				require.NoError(t, err)
				require.NotNil(t, repo)
				repo, err = oper.Open(context.Background(), tabletest.path)
				require.NoError(t, err)
				require.NotNil(t, repo)
			}
		})
	}
}

func TestGetWorktree(t *testing.T) {
	tmpDir := t.TempDir()
	oper := NewOperation()

	repo, err := createTmpGitBareRepo(tmpDir, "test-worktree")
	require.NoError(t, err)

	worktree, err := oper.GetWorktree(context.Background(), repo)
	require.NoError(t, err)
	require.NotNil(t, worktree)
}

func TestWorktreeStatus(t *testing.T) {
	tmpDir := t.TempDir()
	oper := NewOperation()

	repo, err := createTmpGitBareRepo(tmpDir, "test-status")
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	tests := []struct {
		name    string
		setup   func(*git.Worktree) error
		wantErr bool
	}{
		{
			name:    "Clean worktree",
			setup:   func(*git.Worktree) error { return nil },
			wantErr: false,
		},
		{
			name: "Unclean worktree",
			setup: func(w *git.Worktree) error {
				return os.WriteFile(
					filepath.Join(w.Filesystem.Root(), "newfile"),
					[]byte("test"),
					0600,
				)
			},
			wantErr: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			err := tabletest.setup(worktree)
			require.NoError(t, err)

			err = oper.WorktreeStatus(context.Background(), worktree)
			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrUncleanWorkspace)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// func TestFetchBranches(t *testing.T) {
// 	tmpDir := t.TempDir()
// 	op := NewOperation()

// 	repo, err := createTmpGitBareRepo(tmpDir, "test-fetch")
// 	require.NoError(t, err)

// 	auth := &http.BasicAuth{
// 		Username: "test",
// 		Password: "test",
// 	}

// 	err = op.FetchBranches(context.Background(), repo, auth, "test-repo")
// 	// Should get NoErrAlreadyUpToDate since we're using a local repo
// 	assert.NoError(t, err)
// }

func TestSetRemoteAndBranch(t *testing.T) {
	tmpDir := t.TempDir()
	oper := NewOperation()

	// Create source repository
	sourceRepo, err := createTmpGitBareRepo(tmpDir, "source-repo")
	require.NoError(t, err)

	// Create target repository
	targetPath := filepath.Join(tmpDir, "target-repo")
	targetRepo, err := git.PlainInit(targetPath, false)
	require.NoError(t, err)

	repo, _ := model.NewRepository(sourceRepo)
	// Test setting remote and branch
	err = oper.SetRemoteAndBranch(context.Background(), repo, targetPath)
	require.NoError(t, err)

	// Verify remote was set correctly
	remote, err := targetRepo.Remote(gpsconfig.ORIGIN)
	require.NoError(t, err)
	require.NotNil(t, remote)

	// Verify remote URLs match source repository
	sourceRemote, err := sourceRepo.Remote(gpsconfig.ORIGIN)
	require.NoError(t, err)
	require.Equal(t, sourceRemote.Config().URLs, remote.Config().URLs)
}

func TestSetDefaultBranch(t *testing.T) {
	tmpDir := t.TempDir()
	oper := NewOperation()

	repo, err := createTmpGitBareRepo(tmpDir, "test-branch")
	require.NoError(t, err)

	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{
			name:    "Set main branch",
			branch:  "main",
			wantErr: false,
		},
		{
			name:    "Set nonexistent branch",
			branch:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			err := oper.SetDefaultBranch(context.Background(), repo, tabletest.branch)
			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, "failed to checkout branch")
			} else {
				require.NoError(t, err)

				// Verify HEAD points to correct branch
				head, err := repo.Head()
				require.NoError(t, err)
				require.Equal(t, plumbing.NewBranchReferenceName(tabletest.branch), head.Name())
			}
		})
	}
}
