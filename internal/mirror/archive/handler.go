// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archives"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) CreateArchive(ctx context.Context, sourceDir, targetPath, name string) error {
	files, err := h.mapFilesToArchive(ctx, sourceDir, name)
	if err != nil {
		return err
	}

	return h.compress(ctx, targetPath, files)
}

// ArchiveTargetPath generates the full path for the target archive file.
// nowString returns a string representation of the current time.
// The format is _yearmonthday_hourminutesecond_unixmilli.
// This is used to create unique timestamps for archive file names.
func nowString() string {
	now := time.Now()

	return fmt.Sprintf("_%d%02d%02d_%02d%02d%02d_%d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.UnixMilli())
}

func TargetPath(name, targetDir string) string {
	tarArchive := fmt.Sprintf("%s%s.tar.gz", name, nowString())

	return filepath.Join(targetDir, tarArchive)
}

func (h *Handler) mapFilesToArchive(ctx context.Context, sourceDir, targetName string) ([]archives.FileInfo, error) {
	files, err := archives.FilesFromDisk(ctx, nil, map[string]string{
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

func (h *Handler) compress(ctx context.Context, targetPath string, files []archives.FileInfo) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrArchiveCreation, targetPath, err)
	}
	defer file.Close()

	if err := os.Chmod(targetPath, 0o644); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", targetPath, err)
	}

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Archival:    archives.Tar{},
	}

	if err := format.Archive(ctx, file, files); err != nil {
		return fmt.Errorf("%w: %w", ErrArchiveCompression, err)
	}

	return nil
}
