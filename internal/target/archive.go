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
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mholt/archiver/v4"
)

// Custom error types.
var (
	ErrDirectoryCreation  = errors.New("failed to create target directory")
	ErrRepoInitialization = errors.New("failed to initialize target repository")
	ErrRepoPush           = errors.New("failed to push to target repository")
	ErrRepoOpen           = errors.New("failed to open repository")
	ErrRemoteCreation     = errors.New("failed to set remote in target repository")
	ErrBranchCheckout     = errors.New("failed to checkout branch")
	ErrHeadSet            = errors.New("failed to set HEAD reference")
	ErrNoFilesToArchive   = errors.New("no files found to archive")
	ErrArchiveCreation    = errors.New("failed to create archive file")
	ErrArchiveCompression = errors.New("failed to compress archive")
)

// Archive represents a structure capable of pushing Git repositories to archive files.
type Archive struct {
	gitClient Git
}

func (a Archive) Push(ctx context.Context, repository interfaces.GitRepository, option model.PushOption, _ gpsconfig.ProviderConfig, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Archive:Push")
	option.DebugLog(logger).Msg("Archive:Push")

	sourceDirPath := strings.TrimSuffix(option.Target, ".tar.gz")
	if err := os.MkdirAll(filepath.Dir(sourceDirPath), os.ModePerm); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrDirectoryCreation, filepath.Dir(sourceDirPath), err)
	}

	if err := a.initializeAndPushRepo(repository, sourceDirPath); err != nil {
		return err
	}

	if err := a.setRemoteAndBranch(repository, sourceDirPath); err != nil {
		return err
	}

	files, err := mapFilesToArchive(sourceDirPath, repository.Metainfo().Name(ctx))
	if err != nil {
		return err
	}

	return createArchive(ctx, option.Target, files)
}

func (a Archive) initializeAndPushRepo(repository interfaces.GitRepository, sourceDirPath string) error {
	if _, err := git.PlainInit(sourceDirPath, false); err != nil {
		return fmt.Errorf("%w: %w", ErrRepoInitialization, err)
	}

	pushOptions := git.PushOptions{
		RemoteURL: sourceDirPath,
		RefSpecs:  []config.RefSpec{"+refs/*:refs/*"},
	}

	if err := repository.GoGitRepository().Push(&pushOptions); err != nil {
		return fmt.Errorf("%w: %w", ErrRepoPush, err)
	}

	return nil
}

func (a Archive) setRemoteAndBranch(repository interfaces.GitRepository, sourceDirPath string) error {
	targetRepo, err := git.PlainOpen(sourceDirPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrRepoOpen, sourceDirPath, err)
	}

	remote, err := repository.GoGitRepository().Remote("origin")
	if err == nil {
		if _, err := targetRepo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: remote.Config().URLs,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrRemoteCreation, err)
		}
	}

	return setArcDefaultBranch(sourceDirPath, repository.Metainfo().DefaultBranch)
}

func setArcDefaultBranch(repoPath, branchName string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrRepoOpen, repoPath, err)
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Force:  true,
	}); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrBranchCheckout, branchName, err)
	}

	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, branchRef)
	if err := repo.Storer.SetReference(headRef); err != nil {
		return fmt.Errorf("%w: %w", ErrHeadSet, err)
	}

	return nil
}

func mapFilesToArchive(sourceDir string, targetName string) ([]archiver.File, error) {
	files, err := archiver.FilesFromDisk(nil, map[string]string{
		sourceDir: targetName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to map files at %s to tar archive: %w", sourceDir, err)
	}

	if len(files) <= 1 {
		return nil, fmt.Errorf("%w: %s", ErrNoFilesToArchive, sourceDir)
	}

	return files, nil
}

func createArchive(ctx context.Context, targetPath string, files []archiver.File) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrArchiveCreation, targetPath, err)
	}
	defer file.Close()

	if err := os.Chmod(targetPath, 0o644); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", targetPath, err)
	}

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	if err := format.Archive(ctx, file, files); err != nil {
		return fmt.Errorf("%w: %w", ErrArchiveCompression, err)
	}

	return nil
}

func NewArchive(repository interfaces.GitRepository) Archive {
	gitClient := NewGit(repository, repository.Metainfo().OriginalName)

	return Archive{gitClient: gitClient}
}

func ArchiveTargetPath(name, targetDir string) string {
	tarArchive := fmt.Sprintf("%s%s.tar.gz", name, nowString())

	return filepath.Join(targetDir, tarArchive)
}

// nowString returns a string representation of the current time.
// The format is *yearmonthday*hourminutesecondunixmilli.
// This is used to create unique timestamps for archive file names.
func nowString() string {
	currentTime := time.Now()

	return fmt.Sprintf("_%d%02d%02d_%02d%02d%02d_%d",
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.UnixMilli())
}
