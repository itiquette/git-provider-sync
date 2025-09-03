// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import "time"

// Status constants for sync results.
const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusSkipped = "SKIPPED"
)

// Action constants for sync results.
const (
	ActionCreated  = "CREATED"
	ActionUpdated  = "UPDATED"
	ActionNoChange = "NO_CHANGE"
)

// Result represents the result of a single repository sync operation.
type Result struct {
	Environment     string    `json:"environment"`
	Source          string    `json:"source"`
	SourceProvider  string    `json:"source_provider"`
	Repository      string    `json:"repository"`
	Mirror          string    `json:"mirror"`
	MirrorProvider  string    `json:"mirror_provider"`
	Status          string    `json:"status"` // SUCCESS, FAILED, SKIPPED
	Action          string    `json:"action"` // CREATED, UPDATED, NO_CHANGE
	Error           string    `json:"error,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationSeconds float64   `json:"duration_seconds"`
}

// Results represents aggregated results from a complete sync operation.
type Results struct {
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	DurationSeconds   float64   `json:"duration_seconds"`
	TotalSources      int       `json:"total_sources"`
	TotalMirrors      int       `json:"total_mirrors"`
	TotalRepositories int       `json:"total_repositories"`
	SuccessfulSyncs   int       `json:"successful_syncs"`
	FailedSyncs       int       `json:"failed_syncs"`
	SkippedSyncs      int       `json:"skipped_syncs"`
	DryRun            bool      `json:"dry_run"`
	Results           []Result  `json:"results"`
}

// NewResults creates a new Results with start time set.
func NewResults(dryRun bool) *Results {
	return &Results{
		StartTime: time.Now(),
		DryRun:    dryRun,
		Results:   make([]Result, 0),
	}
}

// Complete finalizes the sync results with end time and duration.
func (r *Results) Complete() {
	r.EndTime = time.Now()
	r.DurationSeconds = r.EndTime.Sub(r.StartTime).Seconds()
}

// AddResult adds a sync result and updates counters.
func (r *Results) AddResult(result Result) {
	r.Results = append(r.Results, result)

	switch result.Status {
	case StatusSuccess:
		r.SuccessfulSyncs++
	case StatusFailed:
		r.FailedSyncs++
	case StatusSkipped:
		r.SkippedSyncs++
	}
}
