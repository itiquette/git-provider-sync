// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// SyncInfoWriter implements the SyncInfoWriter port for filesystem storage.
// It follows hexagonal architecture by implementing the domain interface
// while keeping filesystem details isolated in the adapter layer.
type SyncInfoWriter struct {
	basePath string
}

// NewSyncInfoWriter creates a new filesystem-based sync info writer.
// The basePath parameter specifies where to write the sync info file.
func NewSyncInfoWriter(basePath string) *SyncInfoWriter {
	return &SyncInfoWriter{
		basePath: basePath,
	}
}

// WriteSyncInfo persists the sync results to a file in the configured directory.
func (w *SyncInfoWriter) WriteSyncInfo(info ports.SyncInfo) error {
	content := fmt.Sprintf("timestamp=%d\nrepos=%d\nsuccessful=%d\nfailed=%d\nskipped=%d\n",
		info.Timestamp,
		info.TotalRepositories,
		info.SuccessfulSyncs,
		info.FailedSyncs,
		info.SkippedSyncs)

	filePath := filepath.Join(w.basePath, ".gitprovidersync-last-sync")

	// Ensure the directory exists
	if err := os.MkdirAll(w.basePath, 0750); err != nil {
		return fmt.Errorf("failed to create sync info directory: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write sync info: %w", err)
	}

	return nil
}
