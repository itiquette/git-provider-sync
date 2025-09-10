// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package ports

// SyncInfo represents the sync results information to be persisted.
// This is a simple DTO to avoid import cycles with the sync package.
type SyncInfo struct {
	Timestamp         int64
	TotalRepositories int
	SuccessfulSyncs   int
	FailedSyncs       int
	SkippedSyncs      int
}

// SyncInfoWriter defines the interface for persisting sync information.
// Following hexagonal architecture, this port allows the domain to save
// sync results without knowing about the storage implementation.
type SyncInfoWriter interface {
	// WriteSyncInfo persists the sync results information.
	WriteSyncInfo(info SyncInfo) error
}
