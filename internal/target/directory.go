// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package target

import (
	"context"
	"errors"
	"fmt"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
	"os"
	"path/filepath"
	"strconv"
	"time"

	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type Directory struct {
	gitClient Git
}

func NewDirectory(repository interfaces.GitRepository) Directory {
	return Directory{gitClient: NewGit(repository, repository.Metainfo().OriginalName)}
}

func (dir Directory) Push(ctx context.Context, repository interfaces.GitRepository, option model.PushOption, sourceGitOption gpsconfig.ProviderConfig, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory: Push")
	option.DebugLog(logger).Msg("Directory: Push")

	name := dir.getRepositoryName(ctx)
	targetDirPath := filepath.Join(option.Target, name)
	logger.Debug().Str("path", targetDirPath).Msg("Targeting directory")

	if err := os.MkdirAll(option.Target, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory at %s: %w", option.Target, err)
	}

	cliOption := model.CLIOptions(ctx)
	if cliOption.ForcePush || !directoryExists(targetDirPath) {
		return dir.handleClone(ctx, repository, targetDirPath)
	}

	return dir.handlePull(ctx, targetDirPath, sourceGitOption)
}

func (dir Directory) getRepositoryName(ctx context.Context) string {
	name := dir.gitClient.name
	cliOption := model.CLIOptions(ctx)

	if cliOption.CleanupName {
		name = stringconvert.RemoveNonAlphaNumericChars(ctx, name)
	}

	return name
}

func directoryExists(dir string) bool {
	_, err := os.Stat(dir)

	return !os.IsNotExist(err)
}

func (dir Directory) handleClone(ctx context.Context, repository interfaces.GitRepository, targetDirPath string) error {
	if err := moveDirToTmp(ctx, targetDirPath); err != nil {
		return fmt.Errorf("failed to move directory to tmp: %w", err)
	}

	err := writeInMemRepoToDisk(repository.GoGitRepository(), targetDirPath, false)
	if err != nil {
		return fmt.Errorf("failed to write in-memory repo to disk: %w", err)
	}

	return nil
}

func (dir Directory) handlePull(ctx context.Context, targetDirPath string, sourceGitinfo gpsconfig.ProviderConfig) error {
	pullOption := model.NewPullOption("", "", targetDirPath, sourceGitinfo.Git, sourceGitinfo.HTTPClient)

	if err := dir.gitClient.Pull(ctx, targetDirPath, pullOption); err != nil {
		return fmt.Errorf("failed to pull updates to %s: %w", targetDirPath, err)
	}

	return nil
}

func moveDirToTmp(ctx context.Context, dirPath string) error {
	logger := log.Logger(ctx)

	if !directoryExists(dirPath) {
		logger.Debug().Str("dirPath", dirPath).Msg("Directory does not exist")

		return nil
	}

	isDir, err := isDirectory(dirPath)
	if err != nil {
		return fmt.Errorf("failed to check if path is directory: %w", err)
	}

	if !isDir {
		logger.Warn().Str("path", dirPath).Msg("Path is not a directory")

		return nil
	}

	tmpDir := os.TempDir()
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	newPath := filepath.Join(tmpDir, filepath.Base(dirPath)+timestamp)

	logger.Debug().Str("from", dirPath).Str("to", newPath).Msg("Moving directory to tmp folder")

	if err := os.Rename(dirPath, newPath); err != nil {
		return fmt.Errorf("failed to move directory to tmp: %w", err)
	}

	return nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("failed to stat path: %w", err)
	}

	return fileInfo.IsDir(), nil
}

func writeInMemRepoToDisk(inMemRepo *git.Repository, targetPath string, makeBare bool) error {
	fileSystemRepo, err := initializeRepository(targetPath, makeBare)
	if err != nil {
		return err
	}

	if err := setupRemote(inMemRepo, fileSystemRepo); err != nil {
		return err
	}

	if err := copyObjects(inMemRepo, fileSystemRepo); err != nil {
		return err
	}

	if err := copyRefsAndBranches(inMemRepo, fileSystemRepo); err != nil {
		return err
	}

	if err := setDefaultBranch(inMemRepo, fileSystemRepo); err != nil {
		return err
	}

	if !makeBare {
		if err := checkoutDefaultBranch(fileSystemRepo); err != nil {
			return err
		}
	}

	if err := fetchRemoteBranches(fileSystemRepo); err != nil {
		return err
	}

	return nil
}

func initializeRepository(targetPath string, makeBare bool) (*git.Repository, error) {
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return nil, fmt.Errorf("error creating target directory: %w", err)
	}

	var fileSystemRepo *git.Repository

	var err error

	if makeBare {
		fs := osfs.New(targetPath)
		storage := filesystem.NewStorage(fs, nil)

		fileSystemRepo, err = git.Init(storage, nil)
		if err != nil {
			return nil, fmt.Errorf("error initializing bare repository: %w", err)
		}

		targetRepositoryConfig, _ := fileSystemRepo.Config()
		targetRepositoryConfig.Core.IsBare = true
		_ = fileSystemRepo.SetConfig(targetRepositoryConfig)
	} else {
		fileSystemRepo, err = git.PlainInit(targetPath, false)
		if err != nil {
			return nil, fmt.Errorf("error initializing non-bare repository: %w", err)
		}
	}

	return fileSystemRepo, nil
}

func setupRemote(src, dst *git.Repository) error {
	remotes, err := src.Remotes()
	if err != nil {
		return fmt.Errorf("error getting remotes: %w", err)
	}

	for _, remote := range remotes {
		_, err = dst.CreateRemote(&config.RemoteConfig{
			Name: remote.Config().Name,
			URLs: remote.Config().URLs,
		})
		if err != nil && !errors.Is(err, git.ErrRemoteExists) {
			return fmt.Errorf("error creating remote %s: %w", remote.Config().Name, err)
		}
	}

	return nil
}

func copyObjects(src, dst *git.Repository) error {
	srcObjects, err := src.Objects()
	if err != nil {
		return fmt.Errorf("failed to get objects from Git source Repository. err: %w", err)
	}

	dstStorer := dst.Storer

	err = srcObjects.ForEach(func(obj object.Object) error {
		encodedObj, err := src.Storer.EncodedObject(obj.Type(), obj.ID())
		if err != nil {
			return fmt.Errorf("error encoding object %s: %w", obj.ID(), err)
		}

		_, err = dstStorer.SetEncodedObject(encodedObj)
		if err != nil {
			return fmt.Errorf("failed to encode Git Object while copying. err: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed while iterating Git object for copy. Err: %w", err)
	}

	return nil
}

func copyRefsAndBranches(src, dst *git.Repository) error {
	refs, err := src.References()
	if err != nil {
		return fmt.Errorf("error getting references from source repository: %w", err)
	}

	if err := copyReferences(refs, dst); err != nil {
		return fmt.Errorf("failed while iterating Git reference for copy. Err: %w", err)
	}

	if err := setupBranchTracking(src, dst); err != nil {
		return fmt.Errorf("failed while setting up branch tracking. Err: %w", err)
	}

	return nil
}

func copyReferences(refs storer.ReferenceIter, dst *git.Repository) error {
	err := refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil // Skip symbolic references
		}

		return dst.Storer.SetReference(ref)
	})
	if err != nil {
		return fmt.Errorf("failed while iterating Git references for copy. Err: %w", err)
	}

	return nil
}

func setupBranchTracking(src, dst *git.Repository) error {
	branches, err := src.Branches()
	if err != nil {
		return fmt.Errorf("error getting branches: %w", err)
	}

	err = branches.ForEach(func(branch *plumbing.Reference) error {
		branchName := branch.Name().Short()

		if err := createRemoteTrackingBranch(dst, branchName, branch.Hash()); err != nil {
			return fmt.Errorf("failed while creating remote tracking branch. Err: %w", err)
		}

		if err := setupTrackingConfig(dst, branchName); err != nil {
			return fmt.Errorf("failed while setting up tracking. Err: %w", err)
		}

		if err := createLocalBranch(dst, branchName, branch.Hash()); err != nil {
			return fmt.Errorf("failed while creating local branch. Err: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed while setting up branch tracking. Err: %w", err)
	}

	return nil
}

func createRemoteTrackingBranch(repo *git.Repository, branchName string, hash plumbing.Hash) error {
	remoteRef := plumbing.NewHashReference(
		plumbing.ReferenceName("refs/remotes/origin/"+branchName),
		hash,
	)

	err := repo.Storer.SetReference(remoteRef)
	if err != nil {
		return fmt.Errorf("failed while setting reference. Err: %w", err)
	}

	return nil
}

func setupTrackingConfig(repo *git.Repository, branchName string) error {
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	cfg.Branches[branchName] = &config.Branch{
		Name:   branchName,
		Remote: "origin",
		Merge:  plumbing.ReferenceName("refs/heads/" + branchName),
	}

	err = repo.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed while setting  config. Err: %w", err)
	}

	return nil
}

func createLocalBranch(repo *git.Repository, branchName string, hash plumbing.Hash) error {
	localRef := plumbing.NewHashReference(
		plumbing.ReferenceName("refs/heads/"+branchName),
		hash,
	)

	err := repo.Storer.SetReference(localRef)
	if err != nil {
		return fmt.Errorf("error setting reference: %w", err)
	}

	return nil
}

func setDefaultBranch(src, dst *git.Repository) error {
	headRef, err := src.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD reference: %w", err)
	}

	err = dst.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, headRef.Name()))
	if err != nil {
		return fmt.Errorf("error setting HEAD reference: %w", err)
	}

	return nil
}

func checkoutDefaultBranch(repo *git.Repository) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %w", err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD reference: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: headRef.Name(),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("error checking out default branch: %w", err)
	}

	return nil
}

func fetchRemoteBranches(repo *git.Repository) error {
	err := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
		Force:      true,
	})
	if err != nil && errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("error fetching from origin: %w", err)
	}

	return nil
}
