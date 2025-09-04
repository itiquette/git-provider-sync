// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import "time"

// ImmutableResults represents aggregated results from a complete sync operation.
type ImmutableResults struct {
	startTime         time.Time
	endTime           time.Time
	durationSeconds   float64
	totalSources      int
	totalMirrors      int
	totalRepositories int
	successfulSyncs   int
	failedSyncs       int
	skippedSyncs      int
	dryRun            bool
	results           []Result
}

// NewImmutableResults creates a new immutable results with start time set.
func NewImmutableResults(dryRun bool) ImmutableResults {
	return ImmutableResults{
		startTime: time.Now(),
		dryRun:    dryRun,
		results:   []Result{},
	}
}

// WithResult returns a new ImmutableResults with the result added.
func (ir ImmutableResults) WithResult(result Result) ImmutableResults {
	// Create a new slice to maintain immutability
	newResults := make([]Result, len(ir.results)+1)
	copy(newResults, ir.results)
	newResults[len(ir.results)] = result

	// Update counters based on status
	newIR := ir
	newIR.results = newResults

	switch result.Status {
	case "SUCCESS":
		newIR.successfulSyncs++
	case "FAILED":
		newIR.failedSyncs++
	case "SKIPPED":
		newIR.skippedSyncs++
	}

	return newIR
}

// WithResults returns a new ImmutableResults with multiple results added
// is more efficient than calling WithResult multiple times.
func (ir ImmutableResults) WithResults(results ...Result) ImmutableResults {
	if len(results) == 0 {
		return ir
	}

	// Create a new slice to maintain immutability
	newResults := make([]Result, len(ir.results)+len(results))
	copy(newResults, ir.results)
	copy(newResults[len(ir.results):], results)

	// Calculate new counters
	newIR := ir
	newIR.results = newResults

	for _, result := range results {
		switch result.Status {
		case "SUCCESS":
			newIR.successfulSyncs++
		case "FAILED":
			newIR.failedSyncs++
		case "SKIPPED":
			newIR.skippedSyncs++
		}
	}

	return newIR
}

// WithCompletion returns a new ImmutableResults with end time and duration set.
func (ir ImmutableResults) WithCompletion() ImmutableResults {
	now := time.Now()
	newIR := ir
	newIR.endTime = now
	newIR.durationSeconds = now.Sub(ir.startTime).Seconds()

	return newIR
}

// WithSourceCount returns a new ImmutableResults with source count updated.
func (ir ImmutableResults) WithSourceCount(count int) ImmutableResults {
	newIR := ir
	newIR.totalSources = count

	return newIR
}

// WithMirrorCount returns a new ImmutableResults with mirror count updated.
func (ir ImmutableResults) WithMirrorCount(count int) ImmutableResults {
	newIR := ir
	newIR.totalMirrors = count

	return newIR
}

// WithRepositoryCount returns a new ImmutableResults with repository count updated.
func (ir ImmutableResults) WithRepositoryCount(count int) ImmutableResults {
	newIR := ir
	newIR.totalRepositories = count

	return newIR
}

// Getters for accessing immutable fields

// StartTime returns the start time.
func (ir ImmutableResults) StartTime() time.Time {
	return ir.startTime
}

// EndTime returns the end time.
func (ir ImmutableResults) EndTime() time.Time {
	return ir.endTime
}

// DurationSeconds returns the duration in seconds.
func (ir ImmutableResults) DurationSeconds() float64 {
	return ir.durationSeconds
}

// TotalSources returns the total number of sources.
func (ir ImmutableResults) TotalSources() int {
	return ir.totalSources
}

// TotalMirrors returns the total number of mirrors.
func (ir ImmutableResults) TotalMirrors() int {
	return ir.totalMirrors
}

// TotalRepositories returns the total number of repositories.
func (ir ImmutableResults) TotalRepositories() int {
	return ir.totalRepositories
}

// SuccessfulSyncs returns the number of successful syncs.
func (ir ImmutableResults) SuccessfulSyncs() int {
	return ir.successfulSyncs
}

// FailedSyncs returns the number of failed syncs.
func (ir ImmutableResults) FailedSyncs() int {
	return ir.failedSyncs
}

// SkippedSyncs returns the number of skipped syncs.
func (ir ImmutableResults) SkippedSyncs() int {
	return ir.skippedSyncs
}

// DryRun returns whether this was a dry run.
func (ir ImmutableResults) DryRun() bool {
	return ir.dryRun
}

// Results returns a copy of the results slice to maintain immutability.
func (ir ImmutableResults) Results() []Result {
	// Return a copy to prevent external modification
	resultsCopy := make([]Result, len(ir.results))
	copy(resultsCopy, ir.results)

	return resultsCopy
}

// ToMutable converts to the mutable Results type for compatibility.
func (ir ImmutableResults) ToMutable() *Results {
	return &Results{
		StartTime:         ir.startTime,
		EndTime:           ir.endTime,
		DurationSeconds:   ir.durationSeconds,
		TotalSources:      ir.totalSources,
		TotalMirrors:      ir.totalMirrors,
		TotalRepositories: ir.totalRepositories,
		SuccessfulSyncs:   ir.successfulSyncs,
		FailedSyncs:       ir.failedSyncs,
		SkippedSyncs:      ir.skippedSyncs,
		DryRun:            ir.dryRun,
		Results:           ir.Results(), // Use getter for safe copy
	}
}

// MergeResults combines multiple ImmutableResults into one.
func MergeResults(results ...ImmutableResults) ImmutableResults {
	if len(results) == 0 {
		return NewImmutableResults(false)
	}

	// Start with the first result as base
	merged := results[0]

	// Collect all individual results
	var allResults []Result
	for _, r := range results {
		allResults = append(allResults, r.results...)
	}

	// Create new immutable result with merged data
	merged = merged.WithResults(allResults...)

	// Sum up the counts
	totalSources := 0
	totalMirrors := 0
	totalRepos := 0

	for _, r := range results {
		totalSources += r.totalSources
		totalMirrors += r.totalMirrors
		totalRepos += r.totalRepositories
	}

	return merged.
		WithSourceCount(totalSources).
		WithMirrorCount(totalMirrors).
		WithRepositoryCount(totalRepos)
}
