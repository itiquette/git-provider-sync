// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// MirrorService provides directory-specific mirror operations.
type MirrorService struct {
	adapter *Adapter
	logger  ports.Logger
	fs      ports.FileSystem
}

// NewMirrorService creates a new directory mirror service.
func NewMirrorService(adapter *Adapter, logger ports.Logger) *MirrorService {
	return &MirrorService{
		adapter: adapter,
		logger:  logger,
		fs:      adapter.fs,
	}
}

// MirrorRequest contains parameters for mirroring to directory.
type MirrorRequest struct {
	SourceRepository ports.GitRepository
	TargetPath       string
	Options          MirrorOptions
}

// MirrorOptions contains options for directory mirroring.
type MirrorOptions struct {
	Overwrite         bool
	PreserveStructure bool
	IncludeHidden     bool
	ExcludePatterns   []string
	IncludePatterns   []string
	CompressionLevel  int
	CreateArchive     bool
	ArchiveFormat     string // "tar", "zip", "tar.gz"
	Metadata          MirrorMetadata
}

// MirrorMetadata contains metadata about the mirror operation.
type MirrorMetadata struct {
	CreatedAt   time.Time
	CreatedBy   string
	Source      string
	Version     string
	Description string
	Tags        map[string]string
}

// MirrorResult contains the result of a mirror operation.
type MirrorResult struct {
	Success    bool
	TargetPath string
	FilesCount int
	TotalSize  int64
	Duration   time.Duration
	Errors     []error
	Warnings   []string
	Metadata   MirrorMetadata
}

// CreateMirror creates a directory mirror of the source repository.
func (ms *MirrorService) CreateMirror(ctx context.Context, request MirrorRequest) (*MirrorResult, error) {
	startTime := time.Now()

	ms.logger.Info(ctx, "Creating directory mirror", map[string]any{
		"source_path": request.SourceRepository.Path(),
		"target_path": request.TargetPath,
		"overwrite":   request.Options.Overwrite,
	})

	result := &MirrorResult{
		TargetPath: request.TargetPath,
		Metadata:   request.Options.Metadata,
	}

	// Check if target exists and handle overwrite option
	if ms.pathExists(request.TargetPath) {
		if !request.Options.Overwrite {
			return nil, fmt.Errorf("%w: %s", domain.ErrTargetPathAlreadyExists, request.TargetPath)
		}

		ms.logger.Warn(ctx, "Target path exists, removing", map[string]any{
			"target_path": request.TargetPath,
		})

		// Safety check before removing directory
		if err := validateDirectoryPath(ms.fs, request.TargetPath); err != nil {
			return nil, fmt.Errorf("cannot remove target directory: %w", err)
		}

		if err := ms.fs.RemoveAll(request.TargetPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing target: %w", err)
		}
	}

	// Create target directory
	if err := ms.fs.MkdirAll(request.TargetPath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Perform the actual mirroring
	if err := ms.performMirror(ctx, request, result); err != nil {
		result.Success = false
		result.Errors = append(result.Errors, err)

		return result, err
	}

	// Create metadata file
	if err := ms.createMetadataFile(ctx, request.TargetPath, request.Options.Metadata); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create metadata: %v", err))
	}

	// Create archive if requested
	if request.Options.CreateArchive {
		if err := ms.createArchive(ctx, request.TargetPath, request.Options); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create archive: %v", err))
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	ms.logger.Info(ctx, "Directory mirror created successfully", map[string]any{
		"target_path": request.TargetPath,
		"files_count": result.FilesCount,
		"total_size":  result.TotalSize,
		"duration":    result.Duration,
	})

	return result, nil
}

// UpdateMirror updates an existing directory mirror.
func (ms *MirrorService) UpdateMirror(ctx context.Context, request MirrorRequest) (*MirrorResult, error) {
	ms.logger.Info(ctx, "Updating directory mirror", map[string]any{
		"source_path": request.SourceRepository.Path(),
		"target_path": request.TargetPath,
	})

	// Check if target exists
	if !ms.pathExists(request.TargetPath) {
		return ms.CreateMirror(ctx, request)
	}

	// Update the mirror by recreating it
	return ms.CreateMirror(ctx, request)
}

// VerifyMirror verifies the integrity of a directory mirror.
func (ms *MirrorService) VerifyMirror(ctx context.Context, targetPath string) (*MirrorVerification, error) {
	ms.logger.Debug(ctx, "Verifying directory mirror", map[string]any{
		"target_path": targetPath,
	})

	verification := &MirrorVerification{
		Path:      targetPath,
		StartTime: time.Now(),
	}

	// Check if target exists
	if !ms.pathExists(targetPath) {
		verification.IsValid = false
		verification.Errors = append(verification.Errors, "target path does not exist")

		return verification, nil
	}

	// Count files and calculate size using injected filesystem
	err := ms.adapter.GetFileSystem().Walk(targetPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			verification.Warnings = append(verification.Warnings, fmt.Sprintf("Error accessing %s: %v", path, err))

			return nil
		}

		if !info.IsDir() {
			verification.FileCount++
			verification.TotalSize += info.Size()
		}

		return nil
	})
	if err != nil {
		verification.IsValid = false
		verification.Errors = append(verification.Errors, err.Error())

		return verification, nil
	}

	// Check metadata file
	metadataPath := filepath.Join(targetPath, ".mirror-metadata.json")
	if ms.pathExists(metadataPath) {
		verification.HasMetadata = true
	}

	verification.IsValid = len(verification.Errors) == 0
	verification.EndTime = time.Now()
	verification.Duration = verification.EndTime.Sub(verification.StartTime)

	ms.logger.Info(ctx, "Directory mirror verification completed", map[string]any{
		"target_path": targetPath,
		"is_valid":    verification.IsValid,
		"file_count":  verification.FileCount,
		"total_size":  verification.TotalSize,
	})

	return verification, nil
}

// MirrorVerification contains the result of mirror verification.
type MirrorVerification struct {
	Path        string
	IsValid     bool
	FileCount   int
	TotalSize   int64
	HasMetadata bool
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Errors      []string
	Warnings    []string
}

// DeleteMirror deletes a directory mirror.
func (ms *MirrorService) DeleteMirror(ctx context.Context, targetPath string) error {
	ms.logger.Info(ctx, "Deleting directory mirror", map[string]any{
		"target_path": targetPath,
	})

	if !ms.pathExists(targetPath) {
		return fmt.Errorf("%w: %s", domain.ErrTargetPathDoesNotExist, targetPath)
	}

	// Safety check before removing directory
	if err := validateDirectoryPath(ms.fs, targetPath); err != nil {
		return fmt.Errorf("cannot delete mirror directory: %w", err)
	}

	if err := ms.fs.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to delete mirror: %w", err)
	}

	ms.logger.Info(ctx, "Directory mirror deleted successfully", map[string]any{
		"target_path": targetPath,
	})

	return nil
}

// PerformMirror performs the actual mirroring operation.
func (ms *MirrorService) performMirror(_ /* ctx */ context.Context, request MirrorRequest, result *MirrorResult) error {
	sourcePath := request.SourceRepository.Path()

	if err := ms.fs.Walk(sourcePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Error accessing %s: %v", path, err))

			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip if it matches exclude patterns
		if ms.shouldExclude(relPath, request.Options) {
			return nil
		}

		// Skip if it doesn't match include patterns
		if !ms.shouldInclude(relPath, request.Options) {
			return nil
		}

		targetPath := filepath.Join(request.TargetPath, relPath)

		if info.IsDir() {
			return ms.fs.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		if err := ms.copyFile(path, targetPath); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to copy %s: %v", relPath, err))

			return nil
		}

		result.FilesCount++
		result.TotalSize += info.Size()

		return nil
	}); err != nil {
		return fmt.Errorf("failed to walk source directory: %w", err)
	}

	return nil
}

// ShouldExclude checks if a path should be excluded.
func (ms *MirrorService) shouldExclude(path string, options MirrorOptions) bool {
	if !options.IncludeHidden && ms.isHidden(path) {
		return true
	}

	for _, pattern := range options.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}

		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}

// ShouldInclude checks if a path should be included.
func (ms *MirrorService) shouldInclude(path string, options MirrorOptions) bool {
	if len(options.IncludePatterns) == 0 {
		return true
	}

	for _, pattern := range options.IncludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}

		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}

// IsHidden checks if a path represents a hidden file or directory.
func (ms *MirrorService) isHidden(path string) bool {
	return filepath.Base(path)[0] == '.'
}

// CopyFile copies a file from source to target.
func (ms *MirrorService) copyFile(sourcePath, targetPath string) error {
	// Read source file content
	sourceData, err := ms.fs.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Get source file info for permissions
	sourceInfo, err := ms.fs.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Ensure target directory exists
	if err := ms.fs.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Write target file with same permissions
	if err := ms.fs.WriteFile(targetPath, sourceData, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	return nil
}

// CreateMetadataFile creates a metadata file for the mirror.
func (ms *MirrorService) createMetadataFile(_ /* ctx */ context.Context, targetPath string, metadata MirrorMetadata) error {
	// would create a JSON metadata file
	// For now, we'll just create a simple text file
	metadataPath := filepath.Join(targetPath, ".mirror-metadata.txt")

	content := fmt.Sprintf("Mirror created: %s\nSource: %s\nCreated by: %s\n",
		metadata.CreatedAt.Format(time.RFC3339),
		metadata.Source,
		metadata.CreatedBy)

	if err := ms.fs.WriteFile(metadataPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// CreateArchive creates an archive of the mirrored directory.
func (ms *MirrorService) createArchive(ctx context.Context, targetPath string, options MirrorOptions) error {
	// would create an archive based on the specified format
	// For now, we'll skip the implementation and return an error
	ms.logger.Debug(ctx, "Archive creation requested but not implemented", map[string]any{
		"target_path": targetPath,
		"format":      options.ArchiveFormat,
	})

	return fmt.Errorf("archive creation not implemented for format: %s", options.ArchiveFormat)
}

// PathExists checks if a path exists.
func (ms *MirrorService) pathExists(path string) bool {
	exists, _ := ms.fs.Exists(path)

	return exists
}
