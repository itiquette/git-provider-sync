// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// QuietFormatter provides minimal output - only errors and final summary.
// Ideal for scripts and automation.
type QuietFormatter struct {
	writer io.Writer
}

// NewQuietFormatter creates a new quiet formatter.
func NewQuietFormatter(writer io.Writer) ports.SyncOutputFormatter {
	if writer == nil {
		writer = os.Stdout
	}

	return &QuietFormatter{
		writer: writer,
	}
}

// StartEnvironment - silent in quiet mode.
func (f *QuietFormatter) StartEnvironment(env string) {
	// Silent
}

// StartSync - silent in quiet mode.
func (f *QuietFormatter) StartSync(env, config, sourceProvider string) {
	// Silent
}

// SourceFetching - silent in quiet mode.
func (f *QuietFormatter) SourceFetching(provider, domain, owner string) {
	// Silent
}

// SourceFetched - silent in quiet mode.
func (f *QuietFormatter) SourceFetched(count int, duration time.Duration) {
	// Silent
}

// MirrorSyncing - silent in quiet mode.
func (f *QuietFormatter) MirrorSyncing(name, provider, domain, owner string) {
	// Silent
}

// MirrorSynced - silent in quiet mode.
func (f *QuietFormatter) MirrorSynced(name string, synced, failed, skipped int, duration time.Duration) {
	// Silent
}

// RepositorySynced - silent in quiet mode.
func (f *QuietFormatter) RepositorySynced(name string, success bool, message string) {
	// Silent
}

// SyncCompleted shows only the essential summary.
func (f *QuietFormatter) SyncCompleted(results ports.SyncResults) {
	if results.DryRun {
		fmt.Fprintf(f.writer, "DRY RUN: Would sync %d repositories\n", results.TotalRepositories)
	} else {
		if results.FailedSyncs > 0 {
			fmt.Fprintf(f.writer, "FAILED: %d/%d repositories\n",
				results.FailedSyncs, results.TotalRepositories)
		} else {
			fmt.Fprintf(f.writer, "SUCCESS: %d repositories synced\n", results.SuccessfulSyncs)
		}
	}
}

// Error always outputs in quiet mode.
func (f *QuietFormatter) Error(message string, err error) {
	fmt.Fprintf(f.writer, "ERROR: %s", message)
	if err != nil {
		fmt.Fprintf(f.writer, " - %v", err)
	}
	fmt.Fprintln(f.writer)
}

// Warning - silent in quiet mode.
func (f *QuietFormatter) Warning(message string) {
	// Silent - only errors in quiet mode
}

// Info - silent in quiet mode.
func (f *QuietFormatter) Info(message string) {
	// Silent
}

// Debug - silent in quiet mode.
func (f *QuietFormatter) Debug(message string) {
	// Silent
}

// Writer returns the underlying writer.
func (f *QuietFormatter) Writer() io.Writer {
	return f.writer
}
