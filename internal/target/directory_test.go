// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/model"
	configuration "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/target"
)

func TestDirectoryPush(t *testing.T) {
	require := require.New(t)
	// Setup
	ctx, tmpDirPath := setupTestEnvironment(t)
	//defer model.DeleteTmpDir(ctx, tmpDirPath)

	repoName := "directoryTestRepo"
	// regularTargetDirectory := "atargetdirectory"
	// emptyDir := "dirtargettestempty"

	tmpRepository := createTmpGitRepod(t, tmpDirPath, repoName)
	require.NoError(tmpRepository.CreateBranch(&config.Branch{Name: "atestbranch"}))

	rep, err := model.NewRepository(tmpRepository)
	require.NoError(err)

	tests := map[string]struct {
		targetDirectory       string
		modifyContext         func(context.Context) context.Context
		modifySourceDir       func(*testing.T, string) error
		expectedErrorContains string
	}{
		// "create dir repository - success": {
		// 	targetDirectory: regularTargetDirectory,
		// 	modifyContext:   func(ctx context.Context) context.Context { return ctx },
		// },
		// "create dir repository - fail if source dir does not exist": {
		// 	targetDirectory: regularTargetDirectory,
		// 	modifyContext: func(ctx context.Context) context.Context {
		// 		return context.WithValue(ctx, model.TmpDirKey{}, "notavalidtmp")
		// 	},
		// 	expectedErrorContains: "does not exist",
		// },
		// "create dir repository - fail if source dir is empty": {
		// 	targetDirectory: emptyDir,
		// 	modifyContext:   func(ctx context.Context) context.Context { return ctx },
		// 	modifySourceDir: func(t *testing.T, tmpDirPath string) error {
		// 		return os.Mkdir(filepath.Join(tmpDirPath, emptyDir), os.ModePerm)
		// 	},
		// 	expectedErrorContains: "no files found",
		// },
		// "successful push with ForcePush disabled": {
		// 	targetDirectory: regularTargetDirectory,
		// 	modifyContext: func(ctx context.Context) context.Context {
		// 		return model.WithCLIOption(ctx, model.CLIOption{ForcePush: false})
		// 	},
		// },
	}

	for name, tabletest := range tests {
		t.Run(name, func(t *testing.T) {
			ctxTest := tabletest.modifyContext(ctx)

			if tabletest.modifySourceDir != nil {
				require.NoError(tabletest.modifySourceDir(t, tmpDirPath))
			}

			dirTarget := target.NewDirectory(rep)
			option := model.NewPushOption(filepath.Join(tmpDirPath, tabletest.targetDirectory), false, false, configuration.HTTPClientOption{})

			err := dirTarget.Push(ctxTest, option, configuration.GitOption{}, configuration.GitOption{})

			if tabletest.expectedErrorContains != "" {
				require.ErrorContains(err, tabletest.expectedErrorContains)
			} else {
				require.NoError(err)
				verifyPushResult(t, tmpDirPath, tabletest.targetDirectory, repoName, rep)
			}
		})
	}
}

func setupTestEnvironment(t *testing.T) (context.Context, string) {
	t.Helper()

	ctx := context.Background()
	ctx, err := model.CreateTmpDir(ctx, "", "directorytargettest")
	require.NoError(t, err)

	tmpDirPath, err := model.GetTmpDirPath(ctx)
	if err != nil {
		panic(4)
	}

	cliOption := model.CLIOption{ForcePush: true}
	ctx = context.WithValue(ctx, model.CLIOptionKey{}, cliOption)

	return ctx, tmpDirPath
}

func createTmpGitRepod(t *testing.T, path, name string) *git.Repository {
	t.Helper()

	repo, err := git.PlainInit(filepath.Join(path, name), false)
	require.NoError(t, err)

	_, _ = repo.CreateRemote(&config.RemoteConfig{
		Name: configuration.GPSUPSTREAM,
		URLs: []string{"https://bla.bla"},
	})

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(worktree.Filesystem.Root(), "example-git-file")
	//nolint:gosec
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = worktree.Add("example-git-file")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	require.NoError(t, err)

	return repo
}

func verifyPushResult(t *testing.T, tmpDirPath, targetDirectory, repoName string, rep model.Repository) {
	t.Helper()

	repoPath := filepath.Join(tmpDirPath, targetDirectory, repoName)
	_, err := os.Stat(repoPath)
	require.NoError(t, err)

	ref, err := rep.GoGitRepository().Head()
	require.NoError(t, err)

	commit, err := rep.GoGitRepository().CommitObject(ref.Hash())
	require.NoError(t, err)

	tree, err := commit.Tree()
	require.NoError(t, err)

	err = tree.Files().ForEach(func(f *object.File) error {
		require.Equal(t, "example-git-file", f.Name)
		require.Positive(t, f.Size)

		return nil
	})
	require.NoError(t, err)
}
