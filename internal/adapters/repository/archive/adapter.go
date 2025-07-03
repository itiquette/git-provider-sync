// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Adapter implements the GitOperations interface for archive-based operations.
// This adapter provides functionality for creating and extracting tar.gz archives
// of repositories, useful for backup and distribution purposes.
type Adapter struct {
	config ports.GitConfig
}

// New creates a new archive adapter.
func New(config ports.GitConfig) *Adapter {
	return &Adapter{
		config: config,
	}
}

// Clone extracts an archive to a destination directory.
// For archive adapter, this means extracting a tar.gz archive.
func (a *Adapter) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) {
	// Ensure destination directory exists
	err := os.MkdirAll(options.Path, 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// For file:// URLs pointing to archives, extract the archive
	if strings.HasPrefix(options.URL, "file://") {
		archivePath := strings.TrimPrefix(options.URL, "file://")

		// Check if it's a tar.gz file
		if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
			err = a.extractArchive(archivePath, options.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to extract archive: %w", err)
			}
		} else {
			return nil, errors.New("unsupported archive format, expected .tar.gz or .tgz")
		}
	} else {
		return nil, errors.New("archive adapter only supports file:// URLs")
	}

	// Return a simple repository representation
	return &Repository{
		path:   options.Path,
		config: a.config,
	}, nil
}

// Push creates an archive from the repository.
// For archive adapter, this means creating a tar.gz archive.
func (a *Adapter) Push(ctx context.Context, repo ports.GitRepository, options ports.PushOptions) error {
	// Get the repository path
	repoPath := repo.Path()

	// For archive operations, we'll create an archive in the repository directory
	// Since archive adapter doesn't use remotes, we'll create a default archive name
	archivePath := filepath.Join(repoPath, "repository-archive.tar.gz")

	err := a.createArchive(repoPath, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	return nil
}

// Fetch is not applicable for archive operations.
func (a *Adapter) Fetch(ctx context.Context, repo ports.GitRepository, options ports.FetchOptions) error {
	return errors.New("fetch operation not supported for archive adapter")
}

// Open creates a repository instance for the given path.
func (a *Adapter) Open(ctx context.Context, path string) (ports.GitRepository, error) {
	// Verify the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository path does not exist: %s", path)
	}

	return &Repository{
		path:   path,
		config: a.config,
	}, nil
}

// Init is not supported for archive repositories.
func (a *Adapter) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) {
	return nil, errors.New("init is not supported for archive repositories")
}

// Cleanup removes temporary files or performs cleanup for archive operations.
func (a *Adapter) Cleanup(ctx context.Context, path string) error {
	// For archive operations, we might want to clean up extracted files
	// This is a no-op for now, but could be extended for temporary extractions
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

// Helper methods

// extractArchive extracts a tar.gz archive to the specified destination.
func (a *Adapter) extractArchive(archivePath, destPath string) error {
	// #nosec G304 - Archive path is from controlled repository operations
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}

	defer func() {
		if err := gzipReader.Close(); err != nil {
			// Log close error but don't fail
			_ = err
		}
	}()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Security check: validate and clean the header name to prevent path traversal
		if strings.Contains(header.Name, "..") || strings.HasPrefix(header.Name, "/") {
			return fmt.Errorf("unsafe path in archive: %s", header.Name)
		}

		targetPath := filepath.Join(destPath, filepath.Clean(header.Name))

		// Additional security check: ensure target is within destination
		if !strings.HasPrefix(targetPath, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Use safe directory permissions
			err = os.MkdirAll(targetPath, 0750)
			if err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			// Ensure parent directory exists
			err = os.MkdirAll(filepath.Dir(targetPath), 0750)
			if err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			// #nosec G304 - Target path is constructed with security checks above
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// Check file size to prevent decompression bomb attacks
			if header.Size > 100*1024*1024 { // 100MB limit
				return fmt.Errorf("file too large in archive: %s (%d bytes)", header.Name, header.Size)
			}
			// Use limited reader to prevent decompression bombs
			limitedReader := io.LimitReader(tarReader, header.Size)
			_, err = io.Copy(outFile, limitedReader)
			if err := outFile.Close(); err != nil {
				// Log close error
				_ = err
			}

			if err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}

			// Set modification time
			err = os.Chtimes(targetPath, time.Now(), header.ModTime)
			if err != nil {
				// Non-critical error, continue
				continue
			}

		default:
			// Skip unsupported file types
			continue
		}
	}

	return nil
}

// createArchive creates a tar.gz archive from the specified source directory.
func (a *Adapter) createArchive(sourcePath, archivePath string) error {
	// Create the archive file
	// #nosec G304 - Archive path is from controlled operations
	outFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}

	defer func() {
		if err := outFile.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer func() {
		if err := gzipWriter.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if err := tarWriter.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	// Walk the source directory
	err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == sourcePath {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
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
			// #nosec G304 - Path comes from controlled directory walking
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}

			defer func() {
				if err := file.Close(); err != nil {
					// Log close error
					_ = err
				}
			}()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk source directory: %w", err)
	}

	return nil
}
