// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"context"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// ArchiveMirrorRequest contains parameters for archive-based mirroring.
type ArchiveMirrorRequest struct {
	SourceRepository   entities.Repository
	TargetRepository   entities.Repository
	ArchiveFormat      string // "tar.gz", "zip", "tar"
	CompressionLevel   int    // 0-9 for gzip
	IncludeMetadata    bool
	IncludeHistory     bool
	PreservePaths      bool
	ExcludePatterns    []string
	IncludePatterns    []string
	ArchiveNamePattern string
	DryRun             bool
}

// ArchiveOperations defines operations for creating and managing repository archives.
type ArchiveOperations interface {
	// CreateMirror creates an archive mirror of a source repository
	CreateMirror(ctx context.Context, request ArchiveMirrorRequest) error
}
