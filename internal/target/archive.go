// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mholt/archiver/v4"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Archive represents operations for a compressed archive repository.
// It encapsulates a Git client for performing git-related operations.
type Archive struct {
	gitClient Git
}

func serializeConfig(cfg *gitconfig.Config) string {
	var buf strings.Builder

	if cfg.Raw == nil {
		return ""
	}

	for _, section := range cfg.Raw.Sections {
		if section.Name != "core" {
			buf.WriteString(fmt.Sprintf("[%s]\n", section.Name))
		}

		for _, option := range section.Options {
			buf.WriteString(fmt.Sprintf("\t%s = %s\n", option.Key, option.Value))
		}

		for _, subsection := range section.Subsections {
			buf.WriteString(fmt.Sprintf("[%s \"%s\"]\n", section.Name, subsection.Name))
			for _, option := range subsection.Options {
				buf.WriteString(fmt.Sprintf("\t%s = %s\n", option.Key, option.Value))
			}
		}

		buf.WriteString("\n")
	}

	return buf.String()
}
func createArchiveFile(name string, content []byte, modTime time.Time) archiver.File {
	return archiver.File{
		NameInArchive: name,
		FileInfo: &fileInfo{
			name:    path.Base(name),
			size:    int64(len(content)),
			mode:    0644,
			modTime: modTime,
		},
		Open: func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(content)), nil
		},
	}
}

// fileInfo implements fs.FileInfo
type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return false }
func (fi *fileInfo) Sys() interface{}   { return nil }
func BareRepoToTar(repo *git.Repository, outputPath string) error {
	var files []archiver.File

	// Add HEAD file
	head, err := repo.Head()
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}
	if err == nil {
		headContent := []byte(fmt.Sprintf("ref: %s", head.Name()))
		files = append(files, createArchiveFile("HEAD", headContent, time.Now()))
	} else {
		// If HEAD is not found, create an empty one
		files = append(files, createArchiveFile("HEAD", []byte("ref: refs/heads/master"), time.Now()))
	}

	// Add config file
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	cfgContent := []byte(serializeConfig(cfg))
	files = append(files, createArchiveFile("config", cfgContent, time.Now()))

	// Get all references
	refs, err := repo.References()
	if err != nil {
		return fmt.Errorf("failed to get references: %w", err)
	}

	// Process each reference
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		refContent := []byte(ref.Hash().String())
		refPath := path.Join("refs", ref.Name().String())
		files = append(files, createArchiveFile(refPath, refContent, time.Now()))
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process repository references: %w", err)
	}

	// Process all objects
	objectIter, err := repo.Objects()
	if err != nil {
		return fmt.Errorf("failed to get object iterator: %w", err)
	}

	err = objectIter.ForEach(func(obj object.Object) error {
		objBytes, err := processObject(obj, repo)
		if err != nil {
			return fmt.Errorf("failed to process object %s: %w", obj.ID(), err)
		}

		objPath := path.Join("objects", obj.ID().String()[:2], obj.ID().String()[2:])
		files = append(files, createArchiveFile(objPath, objBytes, time.Now()))
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process repository objects: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in the repository")
	}

	// Create the tar file
	tarFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarFile.Close()

	format := archiver.Tar{}
	err = format.Archive(context.Background(), tarFile, files)
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}

	return nil
}

func processObject(obj object.Object, repo *git.Repository) ([]byte, error) {
	encodedObj := repo.Storer.NewEncodedObject()
	err := obj.Encode(encodedObj)
	if err != nil {
		return nil, fmt.Errorf("failed to encode object: %w", err)
	}

	reader, err := encodedObj.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to get reader for encoded object: %w", err)
	}
	defer reader.Close()

	return io.ReadAll(reader)
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
func (a Archive) Push(ctx context.Context, repository interfaces.GitRepository, option model.PushOption, _ gpsconfig.ProviderConfig, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Archive:Push")
	option.DebugLog(logger).Msg("Archive:Push")

	if err := os.MkdirAll(filepath.Dir(option.Target), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", option.Target, err)
	}

	// files, err := mapFilesToArchiveV(sourceRepositoryDir)
	// if err != nil {
	// 	return err
	// }

	// return createArchive(ctx, option.Target, files)
	err := BareRepoToTar(a.gitClient.goGitRepository, option.Target)
	if err != nil {
		return fmt.Errorf("failed to create tar arch. Err: %w", err)
	}
	return nil
}

// validateSourceDir checks if the source directory exists.
// It returns an error if the directory does not exist.
// func validateSourceDir(dir string) error {
// 	if _, err := os.Stat(dir); os.IsNotExist(err) {
// 		return fmt.Errorf("source directory %s does not exist", dir)
// 	}
//
// 	return nil
// }

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
func NewArchive(repository interfaces.GitRepository) Archive {
	gitClient := NewGit(repository, repository.Metainfo().OriginalName)

	return Archive{gitClient: gitClient}
}
