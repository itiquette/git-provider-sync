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
func (f *QuietFormatter) StartEnvironment(_ string) {
	// Silent in quiet mode
}

// StartSync - silent in quiet mode.
func (f *QuietFormatter) StartSync(_, _, _ string) {
	// Silent in quiet mode
}

// SourceFetching - silent in quiet mode.
func (f *QuietFormatter) SourceFetching(_, _, _ string) {
	// Silent in quiet mode
}

// SourceFetched - silent in quiet mode.
func (f *QuietFormatter) SourceFetched(_ int, _ time.Duration) {
	// Silent in quiet mode
}

// MirrorSyncing - silent in quiet mode.
func (f *QuietFormatter) MirrorSyncing(_, _, _, _ string) {
	// Silent in quiet mode
}

// MirrorSynced - silent in quiet mode.
func (f *QuietFormatter) MirrorSynced(_ string, _, _, _ int, _ time.Duration) {
	// Silent in quiet mode
}

// RepositorySynced - silent in quiet mode.
func (f *QuietFormatter) RepositorySynced(_ string, _ bool, _ string) {
	// Silent
}

// SyncCompleted shows only the essential summary.
func (f *QuietFormatter) SyncCompleted(results ports.SyncResults) {
	if results.DryRun {
		_, _ = fmt.Fprintf(f.writer, "DRY RUN: Would sync %d repositories\n", results.TotalRepositories)
	} else {
		if results.FailedSyncs > 0 {
			_, _ = fmt.Fprintf(f.writer, "FAILED: %d/%d repositories\n",
				results.FailedSyncs, results.TotalRepositories)
		} else {
			_, _ = fmt.Fprintf(f.writer, "SUCCESS: %d repositories synced\n", results.SuccessfulSyncs)
		}
	}
}

// Error always outputs in quiet mode.
func (f *QuietFormatter) Error(message string, err error) {
	_, _ = fmt.Fprintf(f.writer, "ERROR: %s", message)
	if err != nil {
		_, _ = fmt.Fprintf(f.writer, " - %v", err)
	}

	_, _ = fmt.Fprintln(f.writer)
}

// Warning - silent in quiet mode.
func (f *QuietFormatter) Warning(_ string) {
	// Silent in quiet mode
}

// Info - silent in quiet mode.
func (f *QuietFormatter) Info(_ string) {
	// Silent in quiet mode
}

// Debug - silent in quiet mode.
func (f *QuietFormatter) Debug(_ string) {
	// Silent
}

// Writer returns the underlying writer.
func (f *QuietFormatter) Writer() io.Writer {
	return f.writer
}
