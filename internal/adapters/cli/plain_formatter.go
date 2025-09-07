// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// PlainFormatter provides simple text output without colors or unicode.
// Suitable for CI/CD environments and non-terminal output.
type PlainFormatter struct {
	writer      io.Writer
	verbosity   string
	currentSync string
}

// NewPlainFormatter creates a new plain text formatter.
func NewPlainFormatter(writer io.Writer, verbosity string) ports.SyncOutputFormatter {
	if writer == nil {
		writer = os.Stdout
	}

	return &PlainFormatter{
		writer:    writer,
		verbosity: verbosity,
	}
}

// StartEnvironment indicates the start of syncing an environment.
func (f *PlainFormatter) StartEnvironment(env string) {
	fmt.Fprintf(f.writer, "\nSyncing environment: %s\n\n", env)
}

// StartSync indicates the start of a sync configuration.
func (f *PlainFormatter) StartSync(env, config, sourceProvider string) {
	f.currentSync = fmt.Sprintf("%s/%s", env, config)
	if f.verbosity == "verbose" || f.verbosity == "debug" || f.verbosity == "trace" {
		fmt.Fprintf(f.writer, "  Starting sync: %s (%s)\n", config, sourceProvider)
	}
}

// SourceFetching indicates fetching from source is starting.
func (f *PlainFormatter) SourceFetching(provider, domain, owner string) {
	fmt.Fprintf(f.writer, "  [SOURCE] %s/%s (%s)\n", domain, owner, provider)
	fmt.Fprintf(f.writer, "    - Fetching repositories...\n")
}

// SourceFetched indicates source fetch completed.
func (f *PlainFormatter) SourceFetched(count int, duration time.Duration) {
	if count > 0 {
		fmt.Fprintf(f.writer, "    - Found %d repositories (%s)\n\n", count, f.formatDuration(duration))
	} else {
		fmt.Fprintf(f.writer, "    - No repositories found (%s)\n\n", f.formatDuration(duration))
	}
}

// MirrorSyncing indicates starting sync to a mirror.
func (f *PlainFormatter) MirrorSyncing(name, provider, domain, owner string) {
	fmt.Fprintf(f.writer, "  [MIRROR] %s/%s (%s)\n", domain, owner, name)
	fmt.Fprintf(f.writer, "    - Syncing repositories...\n")
}

// MirrorSynced indicates mirror sync completed.
func (f *PlainFormatter) MirrorSynced(name string, synced, failed, skipped int, duration time.Duration) {
	total := synced + failed + skipped

	if failed > 0 {
		fmt.Fprintf(f.writer, "    - FAILED: %d/%d repositories (%s)\n\n", failed, total, f.formatDuration(duration))
	} else if synced > 0 {
		fmt.Fprintf(f.writer, "    - SUCCESS: %d synced", synced)
		if skipped > 0 {
			fmt.Fprintf(f.writer, ", %d skipped", skipped)
		}
		fmt.Fprintf(f.writer, " (%s)\n\n", f.formatDuration(duration))
	} else if skipped > 0 {
		fmt.Fprintf(f.writer, "    - SKIPPED: %d repositories (%s)\n\n", skipped, f.formatDuration(duration))
	} else {
		fmt.Fprintf(f.writer, "    - No changes (%s)\n\n", f.formatDuration(duration))
	}
}

// RepositorySynced indicates a single repository was synced.
func (f *PlainFormatter) RepositorySynced(name string, success bool, message string) {
	if f.verbosity != "verbose" && f.verbosity != "debug" && f.verbosity != "trace" {
		return
	}

	if success {
		fmt.Fprintf(f.writer, "      [OK] %s", name)
	} else {
		fmt.Fprintf(f.writer, "      [FAIL] %s", name)
	}

	if message != "" {
		fmt.Fprintf(f.writer, " - %s", message)
	}
	fmt.Fprintln(f.writer)
}

// SyncCompleted shows the final summary.
func (f *PlainFormatter) SyncCompleted(results ports.SyncResults) {
	fmt.Fprintln(f.writer, strings.Repeat("-", 60))

	if results.DryRun {
		fmt.Fprintln(f.writer, "                    DRY RUN SUMMARY")
	} else {
		fmt.Fprintln(f.writer, "                     SYNC SUMMARY")
	}
	fmt.Fprintln(f.writer, strings.Repeat("-", 60))
	fmt.Fprintln(f.writer)

	// Summary stats
	if results.DryRun {
		fmt.Fprintf(f.writer, "  Would sync:  %d repositories\n", results.TotalRepositories)
	} else {
		if results.SuccessfulSyncs > 0 {
			fmt.Fprintf(f.writer, "  Synced:      %d repositories\n", results.SuccessfulSyncs)
		}
		if results.FailedSyncs > 0 {
			fmt.Fprintf(f.writer, "  Failed:      %d repositories\n", results.FailedSyncs)
		}
		if results.SkippedSyncs > 0 {
			fmt.Fprintf(f.writer, "  Skipped:     %d repositories\n", results.SkippedSyncs)
		}
	}

	fmt.Fprintf(f.writer, "  Sources:     %d\n", results.TotalSources)
	fmt.Fprintf(f.writer, "  Mirrors:     %d\n", results.TotalMirrors)

	if results.DryRun {
		fmt.Fprintf(f.writer, "  Mode:        DRY RUN (no changes made)\n")
	} else {
		fmt.Fprintf(f.writer, "  Mode:        LIVE SYNC\n")
	}

	fmt.Fprintf(f.writer, "  Time:        %s\n", f.formatDuration(results.Duration))
	fmt.Fprintln(f.writer)

	// Show repository details if available
	if len(results.Repositories) > 0 && len(results.Repositories) <= 10 {
		fmt.Fprintln(f.writer, "  Repositories:")
		for _, repo := range results.Repositories {
			if repo.Success {
				fmt.Fprintf(f.writer, "    [OK]   %s", repo.Name)
			} else if repo.Skipped {
				fmt.Fprintf(f.writer, "    [SKIP] %s", repo.Name)
			} else {
				fmt.Fprintf(f.writer, "    [FAIL] %s", repo.Name)
			}

			if !repo.LastUpdated.IsZero() {
				fmt.Fprintf(f.writer, " (updated %s)", f.formatTimeAgo(repo.LastUpdated))
			}
			fmt.Fprintln(f.writer)
		}

		if len(results.Repositories) > 10 {
			fmt.Fprintf(f.writer, "    ... and %d more\n", len(results.Repositories)-10)
		}
		fmt.Fprintln(f.writer)
	}

	// Next steps
	if results.DryRun {
		fmt.Fprintln(f.writer, "  Next: Run 'gitprovidersync sync' without --dry-run")
	} else if results.FailedSyncs > 0 {
		fmt.Fprintln(f.writer, "  Next: Check errors and retry failed syncs")
	}

	fmt.Fprintln(f.writer, strings.Repeat("-", 60))
	fmt.Fprintln(f.writer)
}

// Error outputs an error message.
func (f *PlainFormatter) Error(message string, err error) {
	fmt.Fprintf(f.writer, "[ERROR] %s", message)
	if err != nil {
		fmt.Fprintf(f.writer, " - %v", err)
	}
	fmt.Fprintln(f.writer)
}

// Warning outputs a warning message.
func (f *PlainFormatter) Warning(message string) {
	fmt.Fprintf(f.writer, "[WARNING] %s\n", message)
}

// Info outputs an info message (respects verbosity).
func (f *PlainFormatter) Info(message string) {
	if f.verbosity == "quiet" || f.verbosity == "brief" {
		return
	}
	fmt.Fprintf(f.writer, "[INFO] %s\n", message)
}

// Debug outputs a debug message (only with --debug).
func (f *PlainFormatter) Debug(message string) {
	if f.verbosity != "debug" && f.verbosity != "trace" {
		return
	}
	fmt.Fprintf(f.writer, "[DEBUG] %s\n", message)
}

// Writer returns the underlying writer.
func (f *PlainFormatter) Writer() io.Writer {
	return f.writer
}

// Helper methods

func (f *PlainFormatter) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
}

func (f *PlainFormatter) formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	ago := time.Since(t)
	if ago < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(ago.Minutes()))
	}
	if ago < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(ago.Hours()))
	}
	if ago < 7*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(ago.Hours()/24))
	}
	if ago < 30*24*time.Hour {
		return fmt.Sprintf("%d weeks ago", int(ago.Hours()/(24*7)))
	}
	return fmt.Sprintf("%d months ago", int(ago.Hours()/(24*30)))
}
