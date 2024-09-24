// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/target"

	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func TestArchivePush(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	ctx, _ = model.CreateTmpDir(ctx, "", "tarclienttest")

	type args struct {
		archive string
	}

	tmpDirPath, _ := model.GetTmpDirPath(ctx)

	const tmpRepo = "aTmpRep"

	const tmpArch = "aTmpArch"

	tmpRepository := createTmpGitRepo(tmpDirPath, tmpRepo)
	_ = tmpRepository.CreateBranch(&gogitconfig.Branch{Name: "atestbranch"})

	tests := map[string]struct {
		args                    args
		wantErrNotExist         bool
		wantErrEmptySourceDir   bool
		wantErrFailCreateTarArc bool
	}{
		"create tar archive - success": {
			args: args{archive: tmpArch},
		},
		// "create tar archive - fail source dir does not exist": {
		// 	args:            args{archive: tmpArch},
		// 	wantErrNotExist: true,
		// },
		// "create tar archive - fail source dir is empty": {
		// 	args:                  args{archive: tmpArch},
		// 	wantErrEmptySourceDir: true,
		// },
	}

	for name, tableTest := range tests {
		t.Run(name, func(_ *testing.T) {
			tarClient := target.Archive{}
			sourceGitOption := config.GitOption{}
			targetGitOption := config.GitOption{}

			if tableTest.wantErrNotExist {
				ctxErr := context.WithValue(ctx, model.TmpDirKey{}, "notavalidtmp")
				option := model.NewPushOption(filepath.Join(tmpDirPath, tableTest.args.archive), false, false, config.HTTPClientOption{})
				err := tarClient.Push(ctxErr, option, sourceGitOption, targetGitOption)
				require.ErrorContains(err, "does not exist")
			} else if tableTest.wantErrEmptySourceDir {
				ctxEmptySourceDir, _ := model.CreateTmpDir(ctx, "", "tarclienttestempty")
				tmpDirPath, _ := model.GetTmpDirPath(ctxEmptySourceDir)

				option := model.NewPushOption(filepath.Join(tmpDirPath, tableTest.args.archive), false, false, config.HTTPClientOption{})
				err := tarClient.Push(ctxEmptySourceDir, option, sourceGitOption, targetGitOption)
				require.ErrorContains(err, "no files found to archive")

				_ = model.DeleteTmpDir(ctx)
			} else {
				tmpDirPath, _ := model.GetTmpDirPath(ctx)
				option := model.NewPushOption(filepath.Join(tmpDirPath, tableTest.args.archive), false, false, config.HTTPClientOption{})
				_ = tarClient.Push(ctx, option, sourceGitOption, targetGitOption)

				if _, err := os.Stat(filepath.Join(tmpDirPath, tableTest.args.archive)); os.IsNotExist(err) {
					panic(1)
				}
			}
		})
	}

	_ = model.DeleteTmpDir(ctx)
}

func createTmpGitRepo(tmpDir string, tmpRepo string) *git.Repository {
	path := filepath.Join(tmpDir, tmpRepo)
	_ = os.Mkdir(path, 0777)

	rep, _ := git.PlainInit(path, false)

	_, _ = rep.CreateRemote(&gogitconfig.RemoteConfig{
		Name: config.ORIGIN,
		URLs: []string{"https://origin.dot/" + tmpRepo + ".git"},
	})
	_, _ = rep.CreateRemote(&gogitconfig.RemoteConfig{
		Name: config.GPSUPSTREAM,
		URLs: []string{"http://gpsupstream.dot/anotherrepo.git"},
	})

	worktree, _ := rep.Worktree()

	// create an file in the worktree
	filename := filepath.Join(path, "example-git-file")
	_ = os.WriteFile(filename, []byte("hello world!"), 0600)

	// git add example-git-file
	// Adds the example file to the staging/index area.
	_, _ = worktree.Add("example-git-file")

	// git commit -m 'chore: add example file'
	_, _ = worktree.Commit("chore: add example file", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Laval",
			Email: "laval@cavora.chi",
			When:  time.Now(),
		},
	})

	return rep
}
