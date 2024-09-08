// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archiver/v4"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
)

// Archive represents operations for a compressed archive repository.
// It encapsulates a Git client for performing git-related operations.
type Archive struct {
	gitClient Git
}

// Push writes an existing repository to a tar archive directory according to given push options.
// It creates a compressed tar archive (.tar.gz) of the specified repository.
//
// Parameters:
// - ctx: The context for the operation, which can be used for cancellation and passing values.
// - option: The PushOption containing details about the push operation, including the target directory.
//
// Returns an error if any step of the process fails, including source directory validation,
// target directory creation, file mapping, or archive creation.
func (a Archive) Push(ctx context.Context, option model.PushOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Archive:Push")
	option.DebugLog(logger).Msg("Archive:Push")

	tmpDir, _ := model.GetTmpDirPath(ctx)

	sourceRepositoryDir := filepath.Join(tmpDir, a.gitClient.name)
	if err := validateSourceDir(sourceRepositoryDir); err != nil {
		return fmt.Errorf("source directory validation failed: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(option.Target), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", option.Target, err)
	}

	files, err := mapFilesToArchive(sourceRepositoryDir)
	if err != nil {
		return err
	}

	return createArchive(ctx, option.Target, files)
}

// validateSourceDir checks if the source directory exists.
// It returns an error if the directory does not exist.
func validateSourceDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist", dir)
	}

	return nil
}

// mapFilesToArchive creates a mapping of files from the source directory to be included in the archive.
// It returns an error if no files are found or if there's an issue mapping the files.
func mapFilesToArchive(sourceDir string) ([]archiver.File, error) {
	files, err := archiver.FilesFromDisk(nil, map[string]string{
		sourceDir: "", // contents added recursively
	})
	if err != nil {
		return nil, fmt.Errorf("failed to map files at %s to tar archive: %w", sourceDir, err)
	}

	if len(files) <= 1 {
		return nil, fmt.Errorf("no files found to archive at %s", sourceDir)
	}

	return files, nil
}

// createArchive creates a compressed tar archive (.tar.gz) at the specified target path,
// including all the provided files.
// It returns an error if there's an issue creating the file, setting permissions, or compressing the archive.
func createArchive(ctx context.Context, targetPath string, files []archiver.File) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create archive file %s: %w", targetPath, err)
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
		return fmt.Errorf("failed to compress archive: %w", err)
	}

	return nil
}

// ArchiveTargetPath generates the target path for the archive file.
// It combines the provided name, target directory, and a timestamp to create a unique file name.
//
// Parameters:
// - name: The base name for the archive file.
// - targetDir: The directory where the archive will be created.
//
// Returns the full path to the target archive file.
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

// NewArchive creates a new Archive instance.
// It initializes the Archive with a new Git client using the provided repository and name.
//
// Parameters:
// - repository: The GitRepository interface for interacting with the git repository.
// - repositoryName: The name of the repository.
//
// Returns a new Archive instance.
func NewArchive(repository interfaces.GitRepository, repositoryName string) Archive {
	gitClient := NewGit(repository, repositoryName)

	return Archive{gitClient: gitClient}
}
