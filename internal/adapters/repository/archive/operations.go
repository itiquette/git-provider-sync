// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"fmt"
	"os"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Operations implements the ports.ArchiveOperations interface.
type Operations struct {
	logger     ports.Logger
	tempDir    string
	archiveDir string
}

// NewOperations creates a new archive operations adapter.
func NewOperations(logger ports.Logger, tempDir, archiveDir string) *Operations {
	return &Operations{
		logger:     logger,
		tempDir:    tempDir,
		archiveDir: archiveDir,
	}
}

// CreateMirror creates an archive mirror of a source repository.
func (o *Operations) CreateMirror(ctx context.Context, request ports.ArchiveMirrorRequest) error {
	// Create the mirror service
	mirrorService := NewMirrorService(o.logger, o.tempDir, o.archiveDir)

	// Convert domain request to adapter request
	adapterRequest := MirrorRequest{
		SourceRepository:   request.SourceRepository,
		TargetRepository:   request.TargetRepository,
		ArchiveFormat:      request.ArchiveFormat,
		CompressionLevel:   request.CompressionLevel,
		IncludeMetadata:    request.IncludeMetadata,
		IncludeHistory:     request.IncludeHistory,
		PreservePaths:      request.PreservePaths,
		ExcludePatterns:    request.ExcludePatterns,
		IncludePatterns:    request.IncludePatterns,
		ArchiveNamePattern: request.ArchiveNamePattern,
		DryRun:             request.DryRun,
	}

	// Create archive directory if it doesn't exist
	if err := os.MkdirAll(o.archiveDir, 0750); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Execute the mirroring
	_, err := mirrorService.Mirror(ctx, adapterRequest)

	return err
}
