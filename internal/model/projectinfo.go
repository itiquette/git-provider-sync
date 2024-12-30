// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"time"

	"itiquette/git-provider-sync/internal/provider/stringconvert"

	"github.com/rs/zerolog"
)

// ProjectInfo holds metadata about a repository.
// It encapsulates various attributes that describe a repository's
// properties and state.
type ProjectInfo struct {
	// OriginalName is the repository's name as it appears in the source system.
	OriginalName string

	CleanName string

	// HTTPSURL is the HTTPS URL for cloning the repository.
	HTTPSURL string

	// SSHURL is the SSH URL for cloning the repository.
	SSHURL string

	// DefaultBranch is the name of the repository's default branch.
	DefaultBranch string

	// Description is a brief summary of the repository's purpose or contents.
	Description string

	// Visibility indicates whether the repository is public or private.
	Visibility string

	// LastActivityAt is a pointer to the time of the last activity in the repository.
	// It's a pointer to allow for nil values, indicating no activity data is available.
	LastActivityAt *time.Time

	ProjectID string

	ASCIIName bool
}

func (rm *ProjectInfo) SetCleanName(name string) {
	rm.CleanName = name
}

// Name returns the repository name, optionally cleaned up based on CLI options.
// If the ASCIIName option is set in the context, it removes non-alphanumeric
// characters from the original name.
//
// Parameters:
//   - ctx: A context.Context that may contain CLI options.
//
// Returns:
//   - A string representing the (possibly cleaned) repository name.
func (rm ProjectInfo) Name(_ context.Context) string {
	if rm.CleanName != "" {
		return rm.CleanName
	}

	return rm.OriginalName
}

// DebugLog creates a debug log event with repository metadata.
// This method is useful for debugging and tracing repository operations.
//
// Parameters:
//   - logger: A pointer to a zerolog.Logger to log the debug information.
//
// Returns:
//   - A pointer to a zerolog.Event containing the repository metadata.
//
// Note: This method uses the Time() method to safely handle the LastActivityAt field.
func (rm ProjectInfo) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("defaultBranch", rm.DefaultBranch).
				Str("description", stringconvert.RemoveLinebreaks(rm.Description)).
				Str("url", rm.HTTPSURL).
				Str("visibility", rm.Visibility).
				Time("lastActivity", rm.Time())
}

// Time returns the LastActivityAt time or the zero time if it's nil.
// This method provides a safe way to access the LastActivityAt field,
// ensuring that a valid time.Time is always returned.
//
// Returns:
//   - A time.Time representing the last activity time, or a zero time if not set.
func (rm ProjectInfo) Time() time.Time {
	if rm.LastActivityAt == nil {
		return time.Time{}
	}

	return *rm.LastActivityAt
}
