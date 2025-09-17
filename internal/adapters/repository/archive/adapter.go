// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Static errors for err113 compliance.
var (
	ErrUnsupportedArchiveFormat = errors.New("unsupported archive format, expected .tar.gz or .tgz")
	ErrArchiveOnlySupportsFile  = errors.New("archive adapter only supports file:// URLs")
	ErrFetchNotSupported        = errors.New("fetch operation not supported for archive adapter")
	ErrInitNotSupported         = errors.New("init is not supported for archive repositories")
	ErrRepositoryPathNotExist   = errors.New("repository path does not exist")
	ErrUnsafePathInArchive      = errors.New("unsafe path in archive")
	ErrInvalidPathInArchive     = errors.New("invalid path in archive")
	ErrFileTooLargeInArchive    = errors.New("file too large in archive")
)

// Adapter implements the GitOperations interface for archive-based operations
// adapter provides functionality for creating and extracting tar.gz archives
// Of repositories, useful for backup and distribution purposes.
type Adapter struct {
	config ports.GitConfig
	fs     ports.FileSystem
}

// New creates a new archive adapter with OS filesystem.
func New(config ports.GitConfig, optionalFS ...ports.FileSystem) *Adapter {
	// Use provided filesystem if given, otherwise use OS filesystem
	fs := filesystem.NewOSFileSystem()
	if len(optionalFS) > 0 {
		fs = optionalFS[0]
	}

	return NewWithFileSystem(config, fs)
}

// NewWithFileSystem creates a new archive adapter with custom filesystem.
func NewWithFileSystem(config ports.GitConfig, fs ports.FileSystem) *Adapter {
	return &Adapter{
		config: config,
		fs:     fs,
	}
}

// Clone extracts an archive to a destination directory
// For archive adapter, this means extracting a tar.gz archive.
func (a *Adapter) Clone(_ context.Context, options ports.CloneOptions) (ports.GitRepository, error) { //nolint:ireturn
	// Ensure destination directory exists
	err := a.fs.MkdirAll(options.Path, 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := a.extractFromURL(options.URL, options.Path); err != nil {
		return nil, err
	}

	// Return a simple repository representation
	return &Repository{
		path:   options.Path,
		config: a.config,
		fs:     a.fs,
	}, nil
}

// ExtractFromURL handles extraction from different URL types.
func (a *Adapter) extractFromURL(url, destinationPath string) error {
	if !strings.HasPrefix(url, "file://") {
		return ErrArchiveOnlySupportsFile
	}

	archivePath := strings.TrimPrefix(url, "file://")
	if !isSupportedArchiveFormat(archivePath) {
		return ErrUnsupportedArchiveFormat
	}

	if err := a.extractArchive(archivePath, destinationPath); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// IsSupportedArchiveFormat checks if the file extension is supported.
func isSupportedArchiveFormat(path string) bool {
	return strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz")
}

// Push creates an archive from the repository
// For archive adapter, this means creating a tar.gz archive.
func (a *Adapter) Push(_ context.Context, repo ports.GitRepository, _ ports.PushOptions) error {
	// Get the repository path
	repoPath := repo.Path()

	// For archive operations, we'll create an archive in the repository directory
	// Since archive adapter doesn't use remotes, we'll create a default archive name
	archivePath := a.fs.Join(repoPath, "repository-archive.tar.gz")

	err := a.createArchive(repoPath, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	return nil
}

// Fetch is not applicable for archive operations.
func (a *Adapter) Fetch(_ context.Context, _ ports.GitRepository, _ ports.FetchOptions) error {
	return ErrFetchNotSupported
}

// Open creates a repository instance for the given path.
func (a *Adapter) Open(_ context.Context, path string) (ports.GitRepository, error) { //nolint:ireturn
	// Verify the path exists
	if _, err := a.fs.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrRepositoryPathNotExist, path)
	}

	return &Repository{
		path:   path,
		config: a.config,
		fs:     a.fs,
	}, nil
}

// Init is not supported for archive repositories.
func (a *Adapter) Init(_ context.Context, _ string, _ ports.InitOptions) (ports.GitRepository, error) { //nolint:ireturn
	return nil, ErrInitNotSupported
}

// Cleanup removes temporary files or performs cleanup for archive operations.
func (a *Adapter) Cleanup(_ context.Context, _ string) error {
	// For archive operations, we might want to clean up extracted files
	// is a no-op for now, but could be extended for temporary extractions
	return nil
}

// SupportsURL checks if the adapter supports the given URL.
func (a *Adapter) SupportsURL(url string) bool {
	if !strings.HasPrefix(url, "file://") {
		return false
	}

	// Check if it's an archive file
	archivePath := strings.TrimPrefix(url, "file://")

	return strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz")
}

// GetName returns the name of this git implementation.
func (a *Adapter) GetName() string {
	return "archive"
}

// CreateTmpDir implements the ports.GitOperations interface.
func (a *Adapter) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	ctxWithTmp, err := filesystem.CreateTmpDir(ctx, a.fs, dir, prefix)
	if err != nil {
		return ctx, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return ctxWithTmp, nil
}

// GetTmpDirPath implements the ports.GitOperations interface.
func (a *Adapter) GetTmpDirPath(ctx context.Context) (string, error) {
	path, err := filesystem.GetTmpDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	return path, nil
}

// DeleteTmpDir implements the ports.GitOperations interface.
func (a *Adapter) DeleteTmpDir(ctx context.Context) error {
	if err := filesystem.DeleteTmpDir(ctx, a.fs); err != nil {
		return fmt.Errorf("failed to delete temporary directory: %w", err)
	}

	return nil
}

// Helper methods

// ExtractArchive extracts a tar.gz archive to the specified destination.
func (a *Adapter) extractArchive(archivePath, destPath string) error {
	tarReader, cleanupFunc, err := a.openArchiveForReading(archivePath)
	if err != nil {
		return err
	}
	defer cleanupFunc()

	return a.extractTarContents(tarReader, destPath)
}

// OpenArchiveForReading opens a tar.gz archive and returns the tar reader and cleanup function.
func (a *Adapter) openArchiveForReading(archivePath string) (*tar.Reader, func(), error) {
	// Archive path is from controlled repository operations
	file, err := a.fs.Open(archivePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open archive: %w", err)
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		if err := file.Close(); err != nil {
			_ = err // Log close error
		}

		return nil, nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	cleanupFunc := func() {
		if err := gzipReader.Close(); err != nil {
			_ = err // Log close error but don't fail
		}

		if err := file.Close(); err != nil {
			_ = err // Log close error
		}
	}

	return tar.NewReader(gzipReader), cleanupFunc, nil
}

// ExtractTarContents extracts the contents of a tar reader to the destination path.
func (a *Adapter) extractTarContents(tarReader *tar.Reader, destPath string) error {
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		targetPath, err := a.validateAndBuildTargetPath(header.Name, destPath)
		if err != nil {
			return err
		}

		if err := a.extractEntry(tarReader, header, targetPath); err != nil {
			return err
		}
	}

	return nil
}

// ValidateAndBuildTargetPath validates the path and builds the target path.
func (a *Adapter) validateAndBuildTargetPath(name, destPath string) (string, error) {
	// Handle empty or current directory paths
	if strings.TrimSpace(name) == "" || name == "." {
		return destPath, nil
	}

	// Security check: validate and clean the header name to prevent path traversal
	if strings.Contains(name, "..") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("%w: %s", ErrUnsafePathInArchive, name)
	}

	targetPath := a.fs.Join(destPath, a.fs.Clean(name))

	// Additional security check: ensure target is within destination
	cleanDestPath := a.fs.Clean(destPath)
	if !strings.HasPrefix(targetPath, cleanDestPath+string(filepath.Separator)) && targetPath != cleanDestPath {
		return "", fmt.Errorf("%w: %s", ErrInvalidPathInArchive, name)
	}

	return targetPath, nil
}

// ExtractEntry extracts a single entry from the tar archive.
func (a *Adapter) extractEntry(tarReader *tar.Reader, header *tar.Header, targetPath string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return a.extractDirectory(targetPath)
	case tar.TypeReg:
		return a.extractFile(tarReader, header, targetPath)
	default:
		// Skip unsupported file types
		return nil
	}
}

// ExtractDirectory creates a directory with safe permissions.
func (a *Adapter) extractDirectory(targetPath string) error {
	err := a.fs.MkdirAll(targetPath, 0750)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// ExtractFile extracts a single file from the tar archive.
func (a *Adapter) extractFile(tarReader *tar.Reader, header *tar.Header, targetPath string) error {
	// Ensure parent directory exists
	err := a.fs.MkdirAll(a.fs.Dir(targetPath), 0750)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Security: Prevent decompression bomb attacks and memory exhaustion
	// 100MB limit chosen based on: typical git repos <50MB, Docker layer max ~100MB
	// Protects against malicious archives with small compressed/large uncompressed ratios
	if header.Size > 100*1024*1024 { // 100MB limit
		return fmt.Errorf("%w: %s (%d bytes)", ErrFileTooLargeInArchive, header.Name, header.Size)
	}

	// Target path is constructed with security checks above
	outFile, err := a.fs.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		if err := outFile.Close(); err != nil {
			_ = err // Log close error
		}
	}()

	// Use limited reader to prevent decompression bombs
	limitedReader := io.LimitReader(tarReader, header.Size)

	_, err = io.Copy(outFile, limitedReader)
	if err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	// Set modification time (non-critical, ignore errors)
	// Note: Chtimes not available in filesystem interface

	return nil
}

// CreateArchive creates a tar.gz archive from the specified source directory.
func (a *Adapter) createArchive(sourcePath, archivePath string) error {
	tarWriter, cleanupFunc, err := a.createArchiveWriters(archivePath)
	if err != nil {
		return err
	}
	defer cleanupFunc()

	return a.walkAndAddToArchive(sourcePath, tarWriter)
}

// CreateArchiveWriters creates the output file and writer chain for the archive.
func (a *Adapter) createArchiveWriters(archivePath string) (*tar.Writer, func(), error) {
	// Create the archive file
	// Archive path is from controlled operations
	outFile, err := a.fs.Create(archivePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create archive file: %w", err)
	}

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)

	cleanupFunc := func() {
		if err := tarWriter.Close(); err != nil {
			_ = err // Log close error
		}

		if err := gzipWriter.Close(); err != nil {
			_ = err // Log close error
		}

		if err := outFile.Close(); err != nil {
			_ = err // Log close error
		}
	}

	return tarWriter, cleanupFunc, nil
}

// WalkAndAddToArchive walks the source directory and adds files to the archive.
func (a *Adapter) walkAndAddToArchive(sourcePath string, tarWriter *tar.Writer) error {
	if err := a.fs.Walk(sourcePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// FileInfo is already provided by Walk
		if info == nil {
			return fmt.Errorf("file info is nil for %s", path)
		}

		return a.addFileToArchive(sourcePath, path, info, tarWriter)
	}); err != nil {
		return fmt.Errorf("failed to walk source directory: %w", err)
	}

	return nil
}

// AddFileToArchive adds a single file or directory to the archive.
func (a *Adapter) addFileToArchive(sourcePath, path string, info fs.FileInfo, tarWriter *tar.Writer) error {
	// Skip the root directory itself
	if path == sourcePath {
		return nil
	}

	// Get relative path
	relPath, err := a.fs.Rel(sourcePath, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Validate that the relative path doesn't try to escape the source directory
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("failed to get relative path: path %s is outside source directory %s", path, sourcePath)
	}

	// Create tar header
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header: %w", err)
	}

	// Use relative path in archive
	header.Name = relPath

	// Write header
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	// If it's a regular file, write its content
	if info.Mode().IsRegular() {
		return a.writeFileContent(path, tarWriter)
	}

	return nil
}

// WriteFileContent writes a file's content to the tar archive.
func (a *Adapter) writeFileContent(path string, tarWriter *tar.Writer) error {
	// Path comes from controlled directory walking
	file, err := a.fs.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			_ = err // Log close error
		}
	}()

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}
