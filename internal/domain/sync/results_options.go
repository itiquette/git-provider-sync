// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import "time"

// ResultsOption is a functional option for configuring Results.
// This follows the idiomatic Go functional options pattern.
type ResultsOption func(*Results) *Results

// NewFunctionalResults creates Results using functional options.
func NewFunctionalResults(opts ...ResultsOption) Results {
	results := Results{
		StartTime: time.Now(),
		Results:   make([]Result, 0),
	}

	// Apply all options
	for _, opt := range opts {
		opt(&results)
	}

	return results
}

// WithDryRun sets the dry run flag.
func WithDryRun(dryRun bool) ResultsOption {
	return func(r *Results) *Results {
		r.DryRun = dryRun

		return r
	}
}

// WithStartTime sets a custom start time.
func WithStartTime(t time.Time) ResultsOption {
	return func(r *Results) *Results {
		r.StartTime = t

		return r
	}
}

// WithResult adds a single result and updates counters.
// Returns a new Results instance (immutable pattern).
func (r Results) WithResult(result Result) Results {
	// Create a new slice to avoid mutation
	newResults := make([]Result, len(r.Results)+1)
	copy(newResults, r.Results)
	newResults[len(r.Results)] = result

	r.Results = newResults

	// Update counters
	switch result.Status {
	case StatusSuccess:
		r.SuccessfulSyncs++
	case StatusFailed:
		r.FailedSyncs++
	case StatusSkipped:
		r.SkippedSyncs++
	}

	return r
}

// WithResults adds multiple results at once.
// More efficient than calling WithResult multiple times.
func (r Results) WithResults(results ...Result) Results {
	if len(results) == 0 {
		return r
	}

	// Create new slice with exact capacity needed
	newResults := make([]Result, len(r.Results)+len(results))
	copy(newResults, r.Results)
	copy(newResults[len(r.Results):], results)

	r.Results = newResults

	// Update all counters
	for _, result := range results {
		switch result.Status {
		case StatusSuccess:
			r.SuccessfulSyncs++
		case StatusFailed:
			r.FailedSyncs++
		case StatusSkipped:
			r.SkippedSyncs++
		}
	}

	return r
}

// WithCompletion marks the results as complete with end time.
func (r Results) WithCompletion() Results {
	r.EndTime = time.Now()
	r.DurationSeconds = r.EndTime.Sub(r.StartTime).Seconds()

	return r
}

// WithTotalSources sets the total number of sources.
func (r Results) WithTotalSources(count int) Results {
	r.TotalSources = count

	return r
}

// WithTotalMirrors sets the total number of mirrors.
func (r Results) WithTotalMirrors(count int) Results {
	r.TotalMirrors = count

	return r
}

// WithTotalRepositories sets the total number of repositories.
func (r Results) WithTotalRepositories(count int) Results {
	r.TotalRepositories = count

	return r
}

// Apply applies a custom transformation to Results.
func (r Results) Apply(fn func(Results) Results) Results {
	return fn(r)
}

// ResultBuilder builds sync results.
type ResultBuilder struct {
	results Results
}

// NewResultBuilder creates a new builder.
func NewResultBuilder() *ResultBuilder {
	return &ResultBuilder{
		results: Results{
			StartTime: time.Now(),
			Results:   make([]Result, 0),
		},
	}
}

// DryRun sets the dry run flag.
func (b *ResultBuilder) DryRun(dryRun bool) *ResultBuilder {
	b.results.DryRun = dryRun

	return b
}

// AddResult adds a result to the builder.
func (b *ResultBuilder) AddResult(result Result) *ResultBuilder {
	b.results = b.results.WithResult(result)

	return b
}

// TotalSources sets the total sources.
func (b *ResultBuilder) TotalSources(count int) *ResultBuilder {
	b.results.TotalSources = count

	return b
}

// TotalMirrors sets the total mirrors.
func (b *ResultBuilder) TotalMirrors(count int) *ResultBuilder {
	b.results.TotalMirrors = count

	return b
}

// TotalRepositories sets the total repositories.
func (b *ResultBuilder) TotalRepositories(count int) *ResultBuilder {
	b.results.TotalRepositories = count

	return b
}

// Build returns the built Results.
func (b *ResultBuilder) Build() Results {
	return b.results
}

// BuildCompleted returns the built Results with completion time set.
func (b *ResultBuilder) BuildCompleted() Results {
	return b.results.WithCompletion()
}
