// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"io"
	"time"
)

// SyncOutputFormatter defines the interface for formatting sync operation output.
// This allows different output styles (console, plain, json, quiet) while keeping
// the sync logic clean and testable.
type SyncOutputFormatter interface {
	// StartEnvironment indicates the start of syncing an environment
	StartEnvironment(env string)

	// StartSync indicates the start of a sync configuration
	StartSync(env, config, sourceProvider string)

	// SourceFetching indicates fetching from source is starting
	SourceFetching(provider, domain, owner string)

	// SourceFetched indicates source fetch completed
	SourceFetched(count int, duration time.Duration)

	// MirrorSyncing indicates starting sync to a mirror
	MirrorSyncing(name, provider, domain, owner string)

	// MirrorSynced indicates mirror sync completed
	MirrorSynced(name string, synced, failed, skipped int, duration time.Duration)

	// RepositorySynced indicates a single repository was synced
	RepositorySynced(name string, success bool, message string)

	// SyncCompleted shows the final summary
	SyncCompleted(results SyncResults)

	// Error outputs an error message
	Error(message string, err error)

	// Warning outputs a warning message
	Warning(message string)

	// Info outputs an info message (respects verbosity)
	Info(message string)

	// Debug outputs a debug message (only with --debug)
	Debug(message string)

	// Writer returns the underlying writer for direct output if needed
	Writer() io.Writer
}

// SyncResults contains the summary of a sync operation.
type SyncResults struct {
	TotalSources      int
	TotalMirrors      int
	TotalRepositories int
	SuccessfulSyncs   int
	FailedSyncs       int
	SkippedSyncs      int
	Duration          time.Duration
	DryRun            bool
	Repositories      []RepositoryResult
}

// RepositoryResult contains the result of syncing a single repository.
type RepositoryResult struct {
	Name         string
	Success      bool
	Skipped      bool
	ErrorMessage string
	LastUpdated  time.Time
	Size         int64 // in bytes
}
