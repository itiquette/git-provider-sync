// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "dry run enabled",
			dryRun: true,
		},
		{
			name:   "dry run disabled",
			dryRun: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			start := time.Now()
			results := NewResults(test.dryRun)
			end := time.Now()

			require.NotNil(t, results)
			assert.Equal(t, test.dryRun, results.DryRun)
			assert.NotNil(t, results.Results)
			assert.Empty(t, results.Results)

			// Start time should be between test start and end
			assert.True(t, results.StartTime.After(start) || results.StartTime.Equal(start))
			assert.True(t, results.StartTime.Before(end) || results.StartTime.Equal(end))
		})
	}
}

func TestResults_Complete(t *testing.T) {
	t.Parallel()

	// Create results with a specific start time in the past
	results := NewResults(false)

	// Manually set start time to ensure measurable duration without sleep
	results.StartTime = time.Now().Add(-100 * time.Millisecond)

	beforeComplete := time.Now()

	results.Complete()

	afterComplete := time.Now()

	// Verify completion time is properly set
	assert.True(t, results.EndTime.After(beforeComplete) || results.EndTime.Equal(beforeComplete))
	assert.True(t, results.EndTime.Before(afterComplete) || results.EndTime.Equal(afterComplete))
	assert.True(t, results.EndTime.After(results.StartTime))

	// Duration should be positive
	assert.Greater(t, results.DurationSeconds, 0.0)

	// Duration should match the time difference
	expectedDuration := results.EndTime.Sub(results.StartTime).Seconds()
	assert.InDelta(t, expectedDuration, results.DurationSeconds, 0.001)
}

func TestResults_AddResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		results         []Result
		expectedSuccess int
		expectedFailed  int
		expectedSkipped int
	}{
		{
			name: "single successful result",
			results: []Result{
				{
					Environment:     "test",
					Source:          "source1",
					SourceProvider:  "github",
					Repository:      "repo1",
					Mirror:          "mirror1",
					MirrorProvider:  "gitlab",
					Status:          "SUCCESS",
					Action:          "CREATED",
					StartTime:       time.Now(),
					EndTime:         time.Now(),
					DurationSeconds: 1.5,
				},
			},
			expectedSuccess: 1,
			expectedFailed:  0,
			expectedSkipped: 0,
		},
		{
			name: "single failed result",
			results: []Result{
				{
					Environment:     "test",
					Source:          "source1",
					SourceProvider:  "github",
					Repository:      "repo1",
					Mirror:          "mirror1",
					MirrorProvider:  "gitlab",
					Status:          "FAILED",
					Action:          "NO_CHANGE",
					Error:           "connection timeout",
					StartTime:       time.Now(),
					EndTime:         time.Now(),
					DurationSeconds: 0.5,
				},
			},
			expectedSuccess: 0,
			expectedFailed:  1,
			expectedSkipped: 0,
		},
		{
			name: "single skipped result",
			results: []Result{
				{
					Environment:     "test",
					Source:          "source1",
					SourceProvider:  "github",
					Repository:      "repo1",
					Mirror:          "mirror1",
					MirrorProvider:  "gitlab",
					Status:          "SKIPPED",
					Action:          "NO_CHANGE",
					StartTime:       time.Now(),
					EndTime:         time.Now(),
					DurationSeconds: 0.1,
				},
			},
			expectedSuccess: 0,
			expectedFailed:  0,
			expectedSkipped: 1,
		},
		{
			name: "mixed results",
			results: []Result{
				{Status: "SUCCESS", Action: "CREATED"},
				{Status: "SUCCESS", Action: "UPDATED"},
				{Status: "FAILED", Action: "NO_CHANGE", Error: "auth failed"},
				{Status: "SKIPPED", Action: "NO_CHANGE"},
				{Status: "SUCCESS", Action: "UPDATED"},
				{Status: "FAILED", Action: "NO_CHANGE", Error: "network error"},
			},
			expectedSuccess: 3,
			expectedFailed:  2,
			expectedSkipped: 1,
		},
		{
			name: "unknown status ignored",
			results: []Result{
				{Status: "SUCCESS", Action: "CREATED"},
				{Status: "UNKNOWN", Action: "NO_CHANGE"}, // Should not increment any counter
				{Status: "FAILED", Action: "NO_CHANGE"},
			},
			expectedSuccess: 1,
			expectedFailed:  1,
			expectedSkipped: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			results := NewResults(false)

			for _, result := range test.results {
				results.AddResult(result)
			}

			assert.Len(t, results.Results, len(test.results))
			assert.Equal(t, test.expectedSuccess, results.SuccessfulSyncs)
			assert.Equal(t, test.expectedFailed, results.FailedSyncs)
			assert.Equal(t, test.expectedSkipped, results.SkippedSyncs)

			// Verify results are stored correctly
			for i, expected := range test.results {
				actual := results.Results[i]
				assert.Equal(t, expected.Status, actual.Status)
				assert.Equal(t, expected.Action, actual.Action)
				assert.Equal(t, expected.Error, actual.Error)
				assert.Equal(t, expected.Environment, actual.Environment)
				assert.Equal(t, expected.Repository, actual.Repository)
			}
		})
	}
}

func TestResults_AddAndComplete_TracksAllOperations(t *testing.T) {
	t.Parallel()

	// Test a complete workflow from creation to completion
	results := NewResults(true) // dry run
	initialTime := results.StartTime

	// Add some results
	results.AddResult(Result{
		Environment:     "prod",
		Source:          "github-source",
		SourceProvider:  "github",
		Repository:      "my-repo",
		Mirror:          "gitlab-mirror",
		MirrorProvider:  "gitlab",
		Status:          "SUCCESS",
		Action:          "CREATED",
		StartTime:       time.Now(),
		EndTime:         time.Now(),
		DurationSeconds: 2.5,
	})

	results.AddResult(Result{
		Environment:     "prod",
		Source:          "github-source",
		SourceProvider:  "github",
		Repository:      "another-repo",
		Mirror:          "gitlab-mirror",
		MirrorProvider:  "gitlab",
		Status:          "FAILED",
		Action:          "NO_CHANGE",
		Error:           "repository already exists",
		StartTime:       time.Now(),
		EndTime:         time.Now(),
		DurationSeconds: 1.0,
	})

	// Complete the sync
	// Set start time in the past to ensure measurable duration
	results.StartTime = time.Now().Add(-10 * time.Millisecond)
	results.Complete()

	// Verify final state
	assert.True(t, results.DryRun)
	assert.Len(t, results.Results, 2)
	assert.Equal(t, 1, results.SuccessfulSyncs)
	assert.Equal(t, 1, results.FailedSyncs)
	assert.Equal(t, 0, results.SkippedSyncs)
	assert.True(t, results.EndTime.After(initialTime))
	assert.Greater(t, results.DurationSeconds, 0.0)
}

func TestResults_EmptyResults(t *testing.T) {
	t.Parallel()

	results := NewResults(false)
	results.Complete()

	assert.False(t, results.DryRun)
	assert.Empty(t, results.Results)
	assert.Zero(t, results.SuccessfulSyncs)
	assert.Zero(t, results.FailedSyncs)
	assert.Zero(t, results.SkippedSyncs)
	assert.Greater(t, results.DurationSeconds, 0.0) // Should still have measurable duration
}
