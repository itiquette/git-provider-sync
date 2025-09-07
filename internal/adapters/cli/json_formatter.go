// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// JSONFormatter provides structured JSON output for programmatic consumption.
type JSONFormatter struct {
	writer    io.Writer
	verbosity string
	events    []Event
}

// Event represents a sync event for JSON output.
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(writer io.Writer, verbosity string) ports.SyncOutputFormatter {
	if writer == nil {
		writer = os.Stdout
	}

	return &JSONFormatter{
		writer:    writer,
		verbosity: verbosity,
		events:    make([]Event, 0),
	}
}

// StartEnvironment indicates the start of syncing an environment.
func (f *JSONFormatter) StartEnvironment(env string) {
	f.emitEvent("environment_start", map[string]interface{}{
		"environment": env,
	})
}

// StartSync indicates the start of a sync configuration.
func (f *JSONFormatter) StartSync(env, config, sourceProvider string) {
	f.emitEvent("sync_start", map[string]interface{}{
		"environment": env,
		"config":      config,
		"provider":    sourceProvider,
	})
}

// SourceFetching indicates fetching from source is starting.
func (f *JSONFormatter) SourceFetching(provider, domain, owner string) {
	f.emitEvent("source_fetching", map[string]interface{}{
		"provider": provider,
		"domain":   domain,
		"owner":    owner,
	})
}

// SourceFetched indicates source fetch completed.
func (f *JSONFormatter) SourceFetched(count int, duration time.Duration) {
	f.emitEvent("source_fetched", map[string]interface{}{
		"repository_count": count,
		"duration_ms":      duration.Milliseconds(),
	})
}

// MirrorSyncing indicates starting sync to a mirror.
func (f *JSONFormatter) MirrorSyncing(name, provider, domain, owner string) {
	f.emitEvent("mirror_syncing", map[string]interface{}{
		"name":     name,
		"provider": provider,
		"domain":   domain,
		"owner":    owner,
	})
}

// MirrorSynced indicates mirror sync completed.
func (f *JSONFormatter) MirrorSynced(name string, synced, failed, skipped int, duration time.Duration) {
	f.emitEvent("mirror_synced", map[string]interface{}{
		"name":        name,
		"synced":      synced,
		"failed":      failed,
		"skipped":     skipped,
		"duration_ms": duration.Milliseconds(),
	})
}

// RepositorySynced indicates a single repository was synced.
func (f *JSONFormatter) RepositorySynced(name string, success bool, message string) {
	data := map[string]interface{}{
		"repository": name,
		"success":    success,
	}
	if message != "" {
		data["message"] = message
	}

	f.emitEvent("repository_synced", data)
}

// SyncCompleted shows the final summary.
func (f *JSONFormatter) SyncCompleted(results ports.SyncResults) {
	// Build repository list for JSON
	repos := make([]map[string]interface{}, 0, len(results.Repositories))
	for _, repo := range results.Repositories {
		repoData := map[string]interface{}{
			"name":    repo.Name,
			"success": repo.Success,
			"skipped": repo.Skipped,
		}
		if repo.ErrorMessage != "" {
			repoData["error"] = repo.ErrorMessage
		}

		if !repo.LastUpdated.IsZero() {
			repoData["last_updated"] = repo.LastUpdated.Format(time.RFC3339)
		}

		if repo.Size > 0 {
			repoData["size_bytes"] = repo.Size
		}

		repos = append(repos, repoData)
	}

	f.emitEvent("sync_completed", map[string]interface{}{
		"total_sources":      results.TotalSources,
		"total_mirrors":      results.TotalMirrors,
		"total_repositories": results.TotalRepositories,
		"successful_syncs":   results.SuccessfulSyncs,
		"failed_syncs":       results.FailedSyncs,
		"skipped_syncs":      results.SkippedSyncs,
		"duration_ms":        results.Duration.Milliseconds(),
		"dry_run":            results.DryRun,
		"repositories":       repos,
	})
}

// Error outputs an error message.
func (f *JSONFormatter) Error(message string, err error) {
	data := map[string]interface{}{
		"message": message,
	}
	if err != nil {
		data["error"] = err.Error()
	}

	f.emitEvent("error", data)
}

// Warning outputs a warning message.
func (f *JSONFormatter) Warning(message string) {
	f.emitEvent("warning", map[string]interface{}{
		"message": message,
	})
}

// Info outputs an info message (respects verbosity).
func (f *JSONFormatter) Info(message string) {
	if f.verbosity == "quiet" || f.verbosity == "brief" {
		return
	}

	f.emitEvent("info", map[string]interface{}{
		"message": message,
	})
}

// Debug outputs a debug message (only with --debug).
func (f *JSONFormatter) Debug(message string) {
	if f.verbosity != VerbosityDebug && f.verbosity != VerbosityTrace {
		return
	}

	f.emitEvent("debug", map[string]interface{}{
		"message": message,
	})
}

// Writer returns the underlying writer.
func (f *JSONFormatter) Writer() io.Writer {
	return f.writer
}

// emitEvent outputs a JSON event.
func (f *JSONFormatter) emitEvent(eventType string, data map[string]interface{}) {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Store event for potential batch output
	f.events = append(f.events, event)

	// For now, emit immediately (could batch later)
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	// Explicitly ignore error as we're writing to a writer that doesn't typically fail
	_ = encoder.Encode(event) //nolint:errchkjson
}
