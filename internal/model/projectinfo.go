// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"itiquette/git-provider-sync/internal/shared"
)

// ProjectInfo holds metadata about a repository.
type ProjectInfo struct {
	// OriginalName is the repository's name as it appears in the source system
	OriginalName string

	CleanName string

	// HTTPSURL is the HTTPS URL for cloning the repository
	HTTPSURL string

	// SSHURL is the SSH URL for cloning the repository
	SSHURL string

	// DefaultBranch is the name of the repository's default branch
	DefaultBranch string

	// Description is a brief summary of the repository's purpose or contents
	Description string

	// Visibility indicates whether the repository is public or private
	Visibility string

	// LastActivityAt is a pointer to the time of the last activity in the repository
	// It's a pointer to allow for nil values, indicating no activity data is available
	LastActivityAt *time.Time

	ProjectID string

	ASCIIName bool
}

// SetASCIIName enables or disables ASCII name conversion.
func (rm *ProjectInfo) SetASCIIName(name bool) {
	rm.ASCIIName = name
}

// SetCleanName sets the cleaned repository name.
func (rm *ProjectInfo) SetCleanName(name string) {
	rm.CleanName = name
}

// Name returns the repository name, optionally cleaned up based on CLI options
// If ASCIIName is enabled, returns the cleaned name, otherwise returns the original name.
func (rm ProjectInfo) Name(_ context.Context) string {
	if rm.ASCIIName {
		return rm.CleanName
	}

	return rm.OriginalName
}

// DebugLog creates a debug log event with repository metadata.
func (rm ProjectInfo) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("defaultBranch", rm.DefaultBranch).
				Str("description", shared.RemoveLinebreaks(rm.Description)).
				Str("url", rm.HTTPSURL).
				Str("visibility", rm.Visibility).
				Time("lastActivity", rm.Time())
}

// Time returns the LastActivityAt time or the zero time if it's nil.
func (rm ProjectInfo) Time() time.Time {
	if rm.LastActivityAt == nil {
		return time.Time{}
	}

	return *rm.LastActivityAt
}
