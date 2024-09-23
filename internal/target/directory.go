// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
)

// Directory represents a directory target for git repositories.
// It manages the storage and synchronization of git repositories in a local directory structure.
// By default, it clones repositories to a working copy and fetches all branches locally.
type Directory struct {
	gitClient Git // gitClient is the interface for interacting with git operations
}

// Push writes an existing repository to a target directory according to the given push options.
// It handles the process of cloning a new repository or updating an existing one.
func (dir Directory) Push(ctx context.Context, option model.PushOption, sourceGitOption model.GitOption, _ model.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory: Push")
	option.DebugLog(logger).Msg("Directory: Push")

	name := dir.gitClient.name
	tmpDir, _ := model.GetTmpDirPath(ctx)
	sourceRepoDir := filepath.Join(tmpDir, name)

	// // Verify the source repository exists and is not empty
	// if err := checkSourceRepoExists(sourceRepoDir); err != nil {
	// 	return fmt.Errorf("source repository check failed: %w", err)
	// }

	// Clean up the repository name if required by CLI options
	cliOption := model.CLIOptions(ctx)
	if cliOption.CleanupName {
		name = stringconvert.RemoveNonAlphaNumericChars(ctx, name)
	}

	targetDirPath := filepath.Join(option.Target, name)
	logger.Debug().Str("path", targetDirPath).Msg("Targeting directory")

	// Ensure the target directory exists
	if err := os.MkdirAll(option.Target, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory at %s: %w", option.Target, err)
	}

	// Determine whether to clone or pull based on CLI options and directory existence
	if cliOption.ForcePush || !directoryExists(targetDirPath) {
		return dir.handleClone(ctx, sourceRepoDir, targetDirPath, model.GitOption{})
	}

	return dir.handlePull(ctx, targetDirPath, sourceGitOption)
}

// checkSourceRepoExists verifies the existence of the source repository directory and checks for files.
// It ensures that the directory exists and is not empty.
//
// Parameters:
// - dir: The path to the source repository directory.
//
// Returns an error if the directory doesn't exist or is empty.
// func checkSourceRepoExists(dir string) error {
// 	if _, err := os.Stat(dir); os.IsNotExist(err) {
// 		return fmt.Errorf("source directory/repository %s does not exist", dir)
// 	}
//
// 	files, err := os.ReadDir(dir)
// 	if err != nil {
// 		return fmt.Errorf("failed to read directory %s: %w", dir, err)
// 	}
//
// 	if len(files) == 0 {
// 		return errors.New("no files found in source directory")
// 	}
//
// 	return nil
// }

// directoryExists checks if a directory exists at the given path.
//
// Parameters:
// - dir: The path to check for existence.
//
// Returns true if the directory exists, false otherwise.
func directoryExists(dir string) bool {
	_, err := os.Stat(dir)

	return !os.IsNotExist(err)
}

// handleClone manages the process of cloning a repository to the target directory.
// It moves any existing directory to a temporary location, clones the repository,
// fetches branches, and sets up the remote.
//
// Parameters:
// - ctx: The context for the operation.
// - sourceRepoDir: The path to the source repository.
// - targetDirPath: The path where the repository should be cloned.
//
// Returns an error if any step of the cloning process fails.
func (dir Directory) handleClone(ctx context.Context, sourceRepoDir, targetDirPath string, protocol model.GitOption) error {
	logger := log.Logger(ctx)

	if err := moveDirToTmp(ctx, targetDirPath); err != nil {
		return fmt.Errorf("failed to move directory to tmp: %w", err)
	}

	metainfo := model.RepositoryMetainfo{
		HTTPSURL: sourceRepoDir,
		SSHURL:   sourceRepoDir,
	}
	cloneOption := model.NewCloneOption(ctx, metainfo, false, targetDirPath, protocol, model.HTTPClientOption{}, false)

	repo, err := dir.gitClient.Clone(ctx, cloneOption)
	if err != nil {
		if strings.Contains(err.Error(), "repository is empty") {
			logger.Info().Str("Repository", dir.gitClient.name).Msg("No content in Repository, ignoring Clone")

			return nil
		}

		return fmt.Errorf("failed to clone repository to %s: %w", targetDirPath, err)
	}

	if err := dir.gitClient.Fetch(ctx, repo); err != nil && !strings.Contains(err.Error(), "already up-to-date") {
		return fmt.Errorf("failed to fetch branches: %w", err)
	}

	return dir.setupRemote(ctx, repo)
}

// handlePull manages the process of pulling updates for an existing repository.
//
// Parameters:
// - ctx: The context for the operation.
// - targetDirPath: The path to the existing repository.
//
// Returns an error if the pull operation fails.
func (dir Directory) handlePull(ctx context.Context, targetDirPath string, sourceGitinfo model.GitOption) error {
	pullOption := model.NewPullOption("", "", targetDirPath, sourceGitinfo)
	if err := dir.gitClient.Pull(ctx, dir.gitClient.goGitRepository, pullOption); err != nil {
		return fmt.Errorf("failed to pull updates to %s: %w", targetDirPath, err)
	}

	return nil
}

// setupRemote configures the remote for a repository.
// It creates a new 'origin' remote based on the 'gpsupstream' remote URL.
//
// Parameters:
// - ctx: The context for the operation.
// - repo: The git repository interface.
//
// Returns an error if setting up the remote fails.
func (dir Directory) setupRemote(_ context.Context, repo interfaces.GitRepository) error {
	repository, err := model.NewRepository(dir.gitClient.goGitRepository)
	if err != nil {
		return fmt.Errorf("failed to create repository abstraction: %w", err)
	}

	remote, err := repository.Remote(configuration.GPSUPSTREAM)
	if err != nil {
		return fmt.Errorf("failed to get remote %s: %w", configuration.GPSUPSTREAM, err)
	}

	if err := repo.DeleteRemote(configuration.ORIGIN); err != nil {
		return fmt.Errorf("failed to delete remote %s: %w", configuration.ORIGIN, err)
	}

	if err := repo.CreateRemote(configuration.ORIGIN, remote.URL, true); err != nil {
		return fmt.Errorf("failed to create remote %s: %w", configuration.ORIGIN, err)
	}

	return nil
}

// moveDirToTmp moves an existing directory to a temporary location.
// This is typically used to backup an existing directory before performing operations that might modify it.
//
// Parameters:
// - ctx: The context for the operation.
// - dirPath: The path of the directory to move.
//
// Returns an error if moving the directory fails.
func moveDirToTmp(ctx context.Context, dirPath string) error {
	logger := log.Logger(ctx)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
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

// isDirectory checks if the given path is a directory.
//
// Parameters:
// - path: The file system path to check.
//
// Returns true if the path is a directory, false otherwise. An error is returned if the path cannot be accessed.
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("failed to stat path: %w", err)
	}

	return fileInfo.IsDir(), nil
}

// NewDirectory creates a new Directory instance.
// This is the constructor for the Directory type.
//
// Parameters:
// - repository: The git repository interface.
// - repositoryName: The name of the repository.
//
// Returns a new Directory instance.
func NewDirectory(repository interfaces.GitRepository, repositoryName string) Directory {
	return Directory{gitClient: NewGit(repository, repositoryName)}
}
