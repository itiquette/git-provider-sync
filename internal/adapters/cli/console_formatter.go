// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ConsoleFormatter provides rich, colored output for terminal users.
type ConsoleFormatter struct {
	writer      io.Writer
	verbosity   string
	noColor     bool
	currentSync string

	// Color functions - disabled if noColor is true
	blue   *color.Color
	green  *color.Color
	yellow *color.Color
	red    *color.Color
	bold   *color.Color
	dim    *color.Color
	cyan   *color.Color
}

// NewConsoleFormatter creates a new console formatter.
func NewConsoleFormatter(writer io.Writer, verbosity string, noColor bool) ports.SyncOutputFormatter {
	if writer == nil {
		writer = os.Stdout
	}

	// Initialize colors (automatically disabled if NO_COLOR env var is set)
	f := &ConsoleFormatter{
		writer:    writer,
		verbosity: verbosity,
		noColor:   noColor,
		blue:      color.New(color.FgBlue),
		green:     color.New(color.FgGreen),
		yellow:    color.New(color.FgYellow),
		red:       color.New(color.FgRed),
		bold:      color.New(color.Bold),
		dim:       color.New(color.Faint),
		cyan:      color.New(color.FgCyan),
	}

	// Disable colors if requested
	if noColor {
		color.NoColor = true
	}

	return f
}

// StartEnvironment indicates the start of syncing an environment.
func (f *ConsoleFormatter) StartEnvironment(env string) {
	f.bold.Fprintf(f.writer, "\n↻ Syncing environment: %s\n\n", env)
}

// StartSync indicates the start of a sync configuration.
func (f *ConsoleFormatter) StartSync(env, config, sourceProvider string) {
	f.currentSync = fmt.Sprintf("%s/%s", env, config)
	// Only show this in verbose mode
	if f.verbosity == "verbose" || f.verbosity == "debug" || f.verbosity == "trace" {
		f.dim.Fprintf(f.writer, "  Starting sync: %s (%s)\n", config, sourceProvider)
	}
}

// SourceFetching indicates fetching from source is starting.
func (f *ConsoleFormatter) SourceFetching(provider, domain, owner string) {
	f.blue.Fprintf(f.writer, "  Source: ")
	fmt.Fprintf(f.writer, "%s/%s", domain, owner)
	f.dim.Fprintf(f.writer, " (%s)\n", provider)
}

// SourceFetched indicates source fetch completed.
func (f *ConsoleFormatter) SourceFetched(count int, duration time.Duration) {
	// Don't show "Found X repositories" - it's confusing
	// The summary will show what actually matters
	f.dim.Fprintf(f.writer, "\n")
}

// MirrorSyncing indicates starting sync to a mirror.
func (f *ConsoleFormatter) MirrorSyncing(name, provider, domain, owner string) {
	f.cyan.Fprintf(f.writer, "  Mirror: ")
	fmt.Fprintf(f.writer, "%s/%s", domain, owner)
	f.dim.Fprintf(f.writer, " (%s)\n", name)
}

// MirrorSynced indicates mirror sync completed.
func (f *ConsoleFormatter) MirrorSynced(name string, synced, failed, skipped int, duration time.Duration) {
	// Don't show confusing "No changes" here
	// The repository list in the summary will show the actual status
	f.dim.Fprintf(f.writer, "\n")
}

// RepositorySynced indicates a single repository was synced.
func (f *ConsoleFormatter) RepositorySynced(name string, success bool, message string) {
	// Only show in verbose mode
	if f.verbosity != "verbose" && f.verbosity != "debug" && f.verbosity != "trace" {
		return
	}

	if success {
		f.green.Fprintf(f.writer, "        ✓ %s", name)
	} else {
		f.red.Fprintf(f.writer, "        ✗ %s", name)
	}

	if message != "" {
		f.dim.Fprintf(f.writer, " - %s", message)
	}
	fmt.Fprintln(f.writer)
}

// SyncCompleted shows the final summary.
func (f *ConsoleFormatter) SyncCompleted(results ports.SyncResults) {
	// Skip summary if nothing was done
	if results.TotalRepositories == 0 && !results.DryRun {
		return
	}

	// Draw separator
	f.drawSeparator()

	// Title
	if results.DryRun {
		f.bold.Fprintf(f.writer, "%s\n", f.centerText("DRY RUN SUMMARY", 60))
	} else {
		f.bold.Fprintf(f.writer, "%s\n", f.centerText("SYNC COMPLETE", 60))
	}
	f.drawSeparator()
	fmt.Fprintln(f.writer)

	// Summary stats
	if results.DryRun {
		if results.TotalRepositories == 1 {
			fmt.Fprintf(f.writer, "  Repository:  1\n")
		} else {
			fmt.Fprintf(f.writer, "  Repositories: %d\n", results.TotalRepositories)
		}
		fmt.Fprintf(f.writer, "  Action:      Would sync\n")
	} else {
		if results.SuccessfulSyncs > 0 {
			f.green.Fprintf(f.writer, "  ✓ Synced:    ")
			fmt.Fprintf(f.writer, "%d repositories\n", results.SuccessfulSyncs)
		}
		if results.FailedSyncs > 0 {
			f.red.Fprintf(f.writer, "  ✗ Failed:    ")
			fmt.Fprintf(f.writer, "%d repositories\n", results.FailedSyncs)
		}
		if results.SkippedSyncs > 0 {
			f.yellow.Fprintf(f.writer, "  ⚠ Skipped:   ")
			fmt.Fprintf(f.writer, "%d repositories\n", results.SkippedSyncs)
		}
	}

	f.dim.Fprintf(f.writer, "  Duration:    %s\n", f.formatDuration(results.Duration))

	if results.DryRun {
		fmt.Fprintln(f.writer)
		f.yellow.Fprintln(f.writer, "  ↓ DRY RUN MODE - No changes made")
	}
	fmt.Fprintln(f.writer)

	// Show repository details with their sync status
	if !results.DryRun && len(results.Repositories) > 0 {
		fmt.Fprintln(f.writer)
		f.bold.Fprintln(f.writer, "  Repositories:")

		if len(results.Repositories) <= 10 {
			// Show individual repos with their status
			for _, repo := range results.Repositories {
				if repo.Success {
					f.green.Fprintf(f.writer, "    ✓ %s", repo.Name)
					// Add status info if available
					if repo.ErrorMessage == "" {
						f.dim.Fprintf(f.writer, " (up to date)")
					}
					fmt.Fprintln(f.writer)
				} else if repo.Skipped {
					f.yellow.Fprintf(f.writer, "    - %s", repo.Name)
					f.dim.Fprintf(f.writer, " (skipped)")
					if repo.ErrorMessage != "" {
						f.dim.Fprintf(f.writer, " - %s", repo.ErrorMessage)
					}
					fmt.Fprintln(f.writer)
				} else {
					f.red.Fprintf(f.writer, "    ✗ %s", repo.Name)
					if repo.ErrorMessage != "" {
						f.dim.Fprintf(f.writer, " - %s", repo.ErrorMessage)
					}
					fmt.Fprintln(f.writer)
				}
			}
		} else {
			// For many repos, show summary
			successCount := 0
			failedCount := 0
			skippedCount := 0
			for _, repo := range results.Repositories {
				if repo.Success {
					successCount++
				} else if repo.Skipped {
					skippedCount++
				} else {
					failedCount++
				}
			}

			if successCount > 0 {
				f.green.Fprintf(f.writer, "    ✓ %d synced successfully\n", successCount)
			}
			if failedCount > 0 {
				f.red.Fprintf(f.writer, "    ✗ %d failed\n", failedCount)
			}
			if skippedCount > 0 {
				f.yellow.Fprintf(f.writer, "    - %d skipped\n", skippedCount)
			}
		}
		fmt.Fprintln(f.writer)
	}

	// Next steps
	if results.DryRun {
		f.bold.Fprintln(f.writer, "  Next steps:")
		fmt.Fprintln(f.writer, "    • Run without --dry-run to perform sync")
		fmt.Fprintln(f.writer, "    • Add --verbose to see detailed operations")
		fmt.Fprintln(f.writer)
		f.cyan.Fprintln(f.writer, "  $ gitprovidersync sync")
	} else if results.FailedSyncs > 0 {
		f.bold.Fprintln(f.writer, "  Next steps:")
		fmt.Fprintln(f.writer, "    • Check error messages above")
		fmt.Fprintln(f.writer, "    • Run with --verbose for more details")
		fmt.Fprintln(f.writer, "    • Fix issues and retry")
	}

	f.drawSeparator()
	fmt.Fprintln(f.writer)
}

// Error outputs an error message.
func (f *ConsoleFormatter) Error(message string, err error) {
	f.red.Fprintf(f.writer, "✗ Error: ")
	fmt.Fprintf(f.writer, "%s", message)
	if err != nil {
		f.dim.Fprintf(f.writer, " - %v", err)
	}
	fmt.Fprintln(f.writer)
}

// Warning outputs a warning message.
func (f *ConsoleFormatter) Warning(message string) {
	f.yellow.Fprintf(f.writer, "⚠ Warning: ")
	fmt.Fprintln(f.writer, message)
}

// Info outputs an info message (respects verbosity).
func (f *ConsoleFormatter) Info(message string) {
	if f.verbosity == "quiet" || f.verbosity == "brief" {
		return
	}
	f.blue.Fprintf(f.writer, "ℹ ")
	fmt.Fprintln(f.writer, message)
}

// Debug outputs a debug message (only with --debug).
func (f *ConsoleFormatter) Debug(message string) {
	if f.verbosity != "debug" && f.verbosity != "trace" {
		return
	}
	f.dim.Fprintf(f.writer, "[DEBUG] %s\n", message)
}

// Writer returns the underlying writer.
func (f *ConsoleFormatter) Writer() io.Writer {
	return f.writer
}

// Helper methods

func (f *ConsoleFormatter) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
}

func (f *ConsoleFormatter) formatTimeAgo(t time.Time) string {
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

func (f *ConsoleFormatter) drawSeparator() {
	line := strings.Repeat("═", 60)
	f.dim.Fprintln(f.writer, line)
}

func (f *ConsoleFormatter) centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}
