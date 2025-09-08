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
	formatter := &ConsoleFormatter{
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

	return formatter
}

// StartEnvironment indicates the start of syncing an environment.
func (f *ConsoleFormatter) StartEnvironment(env string) {
	_, _ = f.bold.Fprintf(f.writer, "\n↻ Syncing environment: %s\n\n", env)
}

// StartSync indicates the start of a sync configuration.
func (f *ConsoleFormatter) StartSync(env, config, sourceProvider string) {
	f.currentSync = fmt.Sprintf("%s/%s", env, config)
	// Only show this in info mode or higher
	if f.verbosity == VerbosityInfo || f.verbosity == VerbosityDebug || f.verbosity == VerbosityTrace {
		_, _ = f.dim.Fprintf(f.writer, "  Starting sync: %s (%s)\n", config, sourceProvider)
	}
}

// SourceFetching indicates fetching from source is starting.
func (f *ConsoleFormatter) SourceFetching(provider, domain, owner string) {
	_, _ = f.blue.Fprintf(f.writer, "  Source: ")
	_, _ = fmt.Fprintf(f.writer, "%s/%s", domain, owner)
	_, _ = f.dim.Fprintf(f.writer, " (%s)\n", provider)
}

// SourceFetched indicates source fetch completed.
func (f *ConsoleFormatter) SourceFetched(_ int, _ time.Duration) {
	// Don't show "Found X repositories" - it's confusing
	// The summary will show what actually matters
	_, _ = f.dim.Fprintf(f.writer, "\n")
}

// MirrorSyncing indicates starting sync to a mirror.
func (f *ConsoleFormatter) MirrorSyncing(name, _, domain, owner string) {
	_, _ = f.cyan.Fprintf(f.writer, "  Mirror: ")
	_, _ = fmt.Fprintf(f.writer, "%s/%s", domain, owner)
	_, _ = f.dim.Fprintf(f.writer, " (%s)\n", name)
}

// MirrorSynced indicates mirror sync completed.
func (f *ConsoleFormatter) MirrorSynced(_ string, _, _, _ int, _ time.Duration) {
	// Don't show confusing "No changes" here
	// The repository list in the summary will show the actual status
	_, _ = f.dim.Fprintf(f.writer, "\n")
}

// RepositorySynced indicates a single repository was synced.
func (f *ConsoleFormatter) RepositorySynced(name string, success bool, message string) {
	// Only show in info mode or higher
	if f.verbosity != VerbosityInfo && f.verbosity != VerbosityDebug && f.verbosity != VerbosityTrace {
		return
	}

	if success {
		_, _ = f.green.Fprintf(f.writer, "        ✓ %s", name)
	} else {
		_, _ = f.red.Fprintf(f.writer, "        ✗ %s", name)
	}

	if message != "" {
		_, _ = f.dim.Fprintf(f.writer, " - %s", message)
	}

	_, _ = fmt.Fprintln(f.writer)
}

// SyncCompleted shows the final summary.
//
// SyncCompleted shows the final summary of the sync operation.
func (f *ConsoleFormatter) SyncCompleted(results ports.SyncResults) {
	// Skip summary if nothing was done
	if results.TotalRepositories == 0 && !results.DryRun {
		return
	}

	f.printSyncHeader(results.DryRun)
	f.printSyncStats(results)
	_, _ = f.dim.Fprintf(f.writer, "  Duration:    %s\n", f.formatDuration(results.Duration))

	if results.DryRun {
		_, _ = fmt.Fprintln(f.writer)
		_, _ = f.yellow.Fprintln(f.writer, "  ↓ DRY RUN MODE - No changes made")
	}

	_, _ = fmt.Fprintln(f.writer)

	// Show repository details with their sync status
	f.printRepositoryDetails(results)

	// Next steps
	f.printNextSteps(results)

	f.drawSeparator()
	_, _ = fmt.Fprintln(f.writer)
}

// Error outputs an error message.
func (f *ConsoleFormatter) Error(message string, err error) {
	_, _ = f.red.Fprintf(f.writer, "✗ Error: ")

	_, _ = fmt.Fprintf(f.writer, "%s", message)
	if err != nil {
		_, _ = f.dim.Fprintf(f.writer, " - %v", err)
	}

	_, _ = fmt.Fprintln(f.writer)
}

// Warning outputs a warning message.
func (f *ConsoleFormatter) Warning(message string) {
	_, _ = f.yellow.Fprintf(f.writer, "⚠ Warning: ")
	_, _ = fmt.Fprintln(f.writer, message)
}

// Info outputs an info message (respects verbosity).
func (f *ConsoleFormatter) Info(message string) {
	// Only show info messages in info level or higher
	if f.verbosity == VerbosityError || f.verbosity == VerbosityWarn {
		return
	}

	_, _ = f.blue.Fprintf(f.writer, "ℹ ")
	_, _ = fmt.Fprintln(f.writer, message)
}

// Debug outputs a debug message (only with --debug).
func (f *ConsoleFormatter) Debug(message string) {
	if f.verbosity != VerbosityDebug && f.verbosity != VerbosityTrace {
		return
	}

	_, _ = f.dim.Fprintf(f.writer, "[DEBUG] %s\n", message)
}

// Writer returns the underlying writer.
func (f *ConsoleFormatter) Writer() io.Writer {
	return f.writer
}

// Helper methods

func (f *ConsoleFormatter) formatDuration(duration time.Duration) string {
	if duration < time.Second {
		return fmt.Sprintf("%.0fms", duration.Seconds()*1000)
	}

	if duration < time.Minute {
		return fmt.Sprintf("%.1fs", duration.Seconds())
	}

	return fmt.Sprintf("%dm %ds", int(duration.Minutes()), int(duration.Seconds())%60)
}

func (f *ConsoleFormatter) drawSeparator() {
	line := strings.Repeat("═", 60)
	_, _ = f.dim.Fprintln(f.writer, line)
}

func (f *ConsoleFormatter) centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}

	padding := (width - len(text)) / 2

	return strings.Repeat(" ", padding) + text
}

// printSyncHeader prints the header for sync completion.
func (f *ConsoleFormatter) printSyncHeader(isDryRun bool) {
	f.drawSeparator()

	if isDryRun {
		_, _ = f.bold.Fprintf(f.writer, "%s\n", f.centerText("DRY RUN SUMMARY", 60))
	} else {
		_, _ = f.bold.Fprintf(f.writer, "%s\n", f.centerText("SYNC COMPLETE", 60))
	}

	f.drawSeparator()
	_, _ = fmt.Fprintln(f.writer)
}

// printSyncStats prints the sync statistics.
func (f *ConsoleFormatter) printSyncStats(results ports.SyncResults) {
	if results.DryRun {
		f.printDryRunStats(results.TotalRepositories)
	} else {
		f.printRealSyncStats(results)
	}
}

// printDryRunStats prints stats for dry run mode.
func (f *ConsoleFormatter) printDryRunStats(totalRepos int) {
	if totalRepos == 1 {
		_, _ = fmt.Fprintf(f.writer, "  Repository:  1\n")
	} else {
		_, _ = fmt.Fprintf(f.writer, "  Repositories: %d\n", totalRepos)
	}

	_, _ = fmt.Fprintf(f.writer, "  Action:      Would sync\n")
}

// printRealSyncStats prints stats for actual sync.
func (f *ConsoleFormatter) printRealSyncStats(results ports.SyncResults) {
	if results.SuccessfulSyncs > 0 {
		_, _ = f.green.Fprintf(f.writer, "  ✓ Synced:    ")
		_, _ = fmt.Fprintf(f.writer, "%d repositories\n", results.SuccessfulSyncs)
	}

	if results.FailedSyncs > 0 {
		_, _ = f.red.Fprintf(f.writer, "  ✗ Failed:    ")
		_, _ = fmt.Fprintf(f.writer, "%d repositories\n", results.FailedSyncs)
	}

	if results.SkippedSyncs > 0 {
		_, _ = f.yellow.Fprintf(f.writer, "  ⚠ Skipped:   ")
		_, _ = fmt.Fprintf(f.writer, "%d repositories\n", results.SkippedSyncs)
	}
}

// printRepositoryDetails prints individual repository statuses.
func (f *ConsoleFormatter) printRepositoryDetails(results ports.SyncResults) {
	if results.DryRun || len(results.Repositories) == 0 {
		return
	}

	_, _ = fmt.Fprintln(f.writer)
	_, _ = f.bold.Fprintln(f.writer, "  Repositories:")

	if len(results.Repositories) <= 10 {
		f.printDetailedRepos(results.Repositories)
	} else {
		f.printRepoSummary(results.Repositories)
	}

	_, _ = fmt.Fprintln(f.writer)
}

// printDetailedRepos prints detailed info for each repository.
func (f *ConsoleFormatter) printDetailedRepos(repos []ports.RepositoryResult) {
	for _, repo := range repos {
		f.printSingleRepo(repo)
	}
}

// printSingleRepo prints a single repository status.
func (f *ConsoleFormatter) printSingleRepo(repo ports.RepositoryResult) {
	switch {
	case repo.Success:
		_, _ = f.green.Fprintf(f.writer, "    ✓ %s", repo.Name)
		if repo.ErrorMessage == "" {
			_, _ = f.dim.Fprintf(f.writer, " (up to date)")
		}

		_, _ = fmt.Fprintln(f.writer)
	case repo.Skipped:
		_, _ = f.yellow.Fprintf(f.writer, "    - %s", repo.Name)

		_, _ = f.dim.Fprintf(f.writer, " (skipped)")
		if repo.ErrorMessage != "" {
			_, _ = f.dim.Fprintf(f.writer, " - %s", repo.ErrorMessage)
		}

		_, _ = fmt.Fprintln(f.writer)
	default:
		_, _ = f.red.Fprintf(f.writer, "    ✗ %s", repo.Name)
		if repo.ErrorMessage != "" {
			_, _ = f.dim.Fprintf(f.writer, " - %s", repo.ErrorMessage)
		}

		_, _ = fmt.Fprintln(f.writer)
	}
}

// printRepoSummary prints a summary when there are many repositories.
func (f *ConsoleFormatter) printRepoSummary(repos []ports.RepositoryResult) {
	successCount, failedCount, skippedCount := f.countRepoStatuses(repos)

	if successCount > 0 {
		_, _ = f.green.Fprintf(f.writer, "    ✓ %d synced successfully\n", successCount)
	}

	if failedCount > 0 {
		_, _ = f.red.Fprintf(f.writer, "    ✗ %d failed\n", failedCount)
	}

	if skippedCount > 0 {
		_, _ = f.yellow.Fprintf(f.writer, "    - %d skipped\n", skippedCount)
	}
}

// countRepoStatuses counts repositories by status.
func (f *ConsoleFormatter) countRepoStatuses(repos []ports.RepositoryResult) (int, int, int) {
	var success, failed, skipped int

	for _, repo := range repos {
		switch {
		case repo.Success:
			success++
		case repo.Skipped:
			skipped++
		default:
			failed++
		}
	}

	return success, failed, skipped
}

// printNextSteps prints suggested next actions.
func (f *ConsoleFormatter) printNextSteps(results ports.SyncResults) {
	if results.DryRun {
		_, _ = f.bold.Fprintln(f.writer, "  Next steps:")
		_, _ = fmt.Fprintln(f.writer, "    • Run without --dry-run to perform sync")
		_, _ = fmt.Fprintln(f.writer, "    • Add --verbose to see detailed operations")
		_, _ = fmt.Fprintln(f.writer)
		_, _ = f.cyan.Fprintln(f.writer, "  $ gitprovidersync sync")
	} else if results.FailedSyncs > 0 {
		_, _ = f.bold.Fprintln(f.writer, "  Next steps:")
		_, _ = fmt.Fprintln(f.writer, "    • Check error messages above")
		_, _ = fmt.Fprintln(f.writer, "    • Run with --verbose for more details")
		_, _ = fmt.Fprintln(f.writer, "    • Fix issues and retry")
	}
}
