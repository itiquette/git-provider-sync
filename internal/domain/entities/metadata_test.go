// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package entities_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
)

func TestNewSyncRunMetadata(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("github", "gitlab", "prod-config", "production")

	require.NotNil(t, metadata)
	require.Equal(t, "github", metadata.SourceProvider)
	require.Equal(t, "gitlab", metadata.TargetProvider)
	require.Equal(t, "prod-config", metadata.ConfigurationName)
	require.Equal(t, "production", metadata.EnvironmentName)
	require.Zero(t, metadata.ProcessedCount)
	require.Zero(t, metadata.SuccessCount)
	require.Zero(t, metadata.FailureCount)
	require.Zero(t, metadata.SkippedCount)
	require.False(t, metadata.IsFinished())
	require.NotNil(t, metadata.StartTime)
}

func TestSyncRunMetadata_AddSuccess(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")

	metadata.AddSuccess("clone", "repo-1")
	metadata.AddSuccess("clone", "repo-2")
	metadata.AddSuccess("push", "repo-1")

	require.Equal(t, 3, metadata.ProcessedCount)
	require.Equal(t, 3, metadata.SuccessCount)
	require.Equal(t, 0, metadata.FailureCount)

	successes := metadata.GetSuccessesByCategory()
	require.Len(t, successes["clone"], 2)
	require.Contains(t, successes["clone"], "repo-1")
	require.Contains(t, successes["clone"], "repo-2")
	require.Len(t, successes["push"], 1)
	require.Contains(t, successes["push"], "repo-1")
}

func TestSyncRunMetadata_AddFailure(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")

	metadata.AddFailure("clone", "repo-1")
	metadata.AddFailure("auth", "repo-2")
	metadata.AddFailure("clone", "repo-3")

	require.Equal(t, 3, metadata.ProcessedCount)
	require.Equal(t, 0, metadata.SuccessCount)
	require.Equal(t, 3, metadata.FailureCount)
	require.True(t, metadata.HasFailures())

	failures := metadata.GetFailuresByCategory()
	require.Len(t, failures["clone"], 2)
	require.Contains(t, failures["clone"], "repo-1")
	require.Contains(t, failures["clone"], "repo-3")
	require.Len(t, failures["auth"], 1)
	require.Contains(t, failures["auth"], "repo-2")
}

func TestSyncRunMetadata_AddSkipped(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")

	metadata.AddSkipped("filter", "repo-1")
	metadata.AddSkipped("archived", "repo-2")

	require.Equal(t, 0, metadata.ProcessedCount) // Skipped items don't count as processed
	require.Equal(t, 2, metadata.SkippedCount)

	skipped := metadata.GetSkippedByCategory()
	require.Len(t, skipped["filter"], 1)
	require.Contains(t, skipped["filter"], "repo-1")
	require.Len(t, skipped["archived"], 1)
	require.Contains(t, skipped["archived"], "repo-2")
}

func TestSyncRunMetadata_CalculateRates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		setupMetadata       func(*entities.SyncRunMetadata)
		expectedSuccessRate float64
		expectedFailureRate float64
	}{
		{
			name: "no_processing_returns_zero_rates",
			setupMetadata: func(_ *entities.SyncRunMetadata) {
				// No operations
			},
			expectedSuccessRate: 0.0,
			expectedFailureRate: 0.0,
		},
		{
			name: "all_successful_returns_100_percent_success",
			setupMetadata: func(m *entities.SyncRunMetadata) {
				m.AddSuccess("clone", "repo-1")
				m.AddSuccess("clone", "repo-2")
				m.AddSuccess("clone", "repo-3")
			},
			expectedSuccessRate: 100.0,
			expectedFailureRate: 0.0,
		},
		{
			name: "all_failed_returns_100_percent_failure",
			setupMetadata: func(m *entities.SyncRunMetadata) {
				m.AddFailure("clone", "repo-1")
				m.AddFailure("clone", "repo-2")
			},
			expectedSuccessRate: 0.0,
			expectedFailureRate: 100.0,
		},
		{
			name: "mixed_results_calculates_correct_rates",
			setupMetadata: func(m *entities.SyncRunMetadata) {
				m.AddSuccess("clone", "repo-1")
				m.AddSuccess("clone", "repo-2")
				m.AddSuccess("clone", "repo-3")
				m.AddFailure("clone", "repo-4")
				// 3 successes, 1 failure = 75% success, 25% failure
			},
			expectedSuccessRate: 75.0,
			expectedFailureRate: 25.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")
			test.setupMetadata(metadata)

			successRate := metadata.SuccessRate()
			failureRate := metadata.FailureRate()

			require.InDelta(t, test.expectedSuccessRate, successRate, 0.001)
			require.InDelta(t, test.expectedFailureRate, failureRate, 0.001)
		})
	}
}

func TestSyncRunMetadata_FinishAndDuration(t *testing.T) {
	t.Parallel()

	// Create metadata with a fixed start time for deterministic testing
	metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")

	// Capture the initial state
	startTime := metadata.StartTime
	require.False(t, metadata.IsFinished())

	// Duration should be calculable even before finishing
	duration1 := metadata.Duration()
	require.GreaterOrEqual(t, duration1, time.Duration(0))

	// Manually set the internal clock forward instead of sleeping
	// simulates time passing without actually waiting
	metadata.StartTime = startTime.Add(-100 * time.Millisecond)

	// Finish the metadata
	metadata.Finish()

	// After finishing
	require.True(t, metadata.IsFinished())
	require.NotZero(t, metadata.EndTime)

	// Duration should be frozen after finishing
	duration2 := metadata.Duration()
	require.Greater(t, duration2, time.Duration(0))

	// Verify duration is stable after finishing (no matter how much time passes)
	duration3 := metadata.Duration()
	require.Equal(t, duration2, duration3, "Duration should be frozen after Finish()")

	// End time should be after the adjusted start time
	require.True(t, metadata.EndTime.After(metadata.StartTime))
}

func TestSyncRunMetadata_GetSummary(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("github", "gitlab", "prod-config", "production")
	metadata.SetTotalRepositories(5)
	metadata.SetDryRun(false)

	metadata.AddSuccess("clone", "repo-1")
	metadata.AddSuccess("clone", "repo-2")
	metadata.AddFailure("clone", "repo-3")
	metadata.AddSkipped("filter", "repo-4")

	metadata.Finish()

	summary := metadata.GetSummary()

	require.Equal(t, 5, summary.TotalRepositories)
	require.Equal(t, 3, summary.ProcessedCount)
	require.Equal(t, 2, summary.SuccessCount)
	require.Equal(t, 1, summary.FailureCount)
	require.Equal(t, 1, summary.SkippedCount)
	require.InDelta(t, 66.67, summary.SuccessRate, 0.01) // 2/3 * 100
	require.InDelta(t, 33.33, summary.FailureRate, 0.01) // 1/3 * 100
	require.True(t, summary.IsFinished)
	require.Equal(t, "github", summary.SourceProvider)
	require.Equal(t, "gitlab", summary.TargetProvider)
	require.Equal(t, "prod-config", summary.ConfigurationName)
	require.Equal(t, "production", summary.EnvironmentName)
	require.False(t, summary.DryRun)
	require.Greater(t, summary.Duration, time.Duration(0))
}

func TestSyncRunMetadata_ContainsFailure(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")

	metadata.AddFailure("clone", "repo-1")
	metadata.AddFailure("auth", "repo-2")

	// Test direct method
	require.True(t, metadata.ContainsFailure("clone", "repo-1"))
	require.True(t, metadata.ContainsFailure("auth", "repo-2"))
	require.False(t, metadata.ContainsFailure("clone", "repo-2"))
	require.False(t, metadata.ContainsFailure("push", "repo-1"))
	require.False(t, metadata.ContainsFailure("nonexistent", "nonexistent"))
}

func TestSyncRunMetadata_ContextIntegration(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")
	metadata.AddFailure("invalid", "repo-1")

	// Test adding to context
	ctx := context.Background()
	ctxWithMetadata := entities.AddMetadataToContext(ctx, metadata)

	// Test retrieving from context
	retrievedMetadata, found := entities.GetMetadataFromContext(ctxWithMetadata)
	require.True(t, found)
	require.Equal(t, metadata, retrievedMetadata)

	// Test context-based failure checking
	containsFailure := entities.ContainsFailureInContext(ctxWithMetadata, "invalid", "repo-1")
	require.True(t, containsFailure)

	containsFailure = entities.ContainsFailureInContext(ctxWithMetadata, "invalid", "repo-2")
	require.False(t, containsFailure)

	// Test context without metadata
	emptyCtx := context.Background()
	_, found = entities.GetMetadataFromContext(emptyCtx)
	require.False(t, found)

	containsFailure = entities.ContainsFailureInContext(emptyCtx, "invalid", "repo-1")
	require.False(t, containsFailure)
}

func TestSyncRunMetadata_String(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("github", "gitlab", "prod-config", "production")
	metadata.SetTotalRepositories(5)

	metadata.AddSuccess("clone", "repo-1")
	metadata.AddFailure("auth", "repo-2")
	metadata.AddSkipped("filter", "repo-3")

	metadata.Finish()

	str := metadata.String()

	// Check that essential information is present
	require.Contains(t, str, "Source: github")
	require.Contains(t, str, "Target: gitlab")
	require.Contains(t, str, "Total: 5")
	require.Contains(t, str, "Processed: 2")
	require.Contains(t, str, "Successful: 1")
	require.Contains(t, str, "Failed: 1")
	require.Contains(t, str, "Skipped: 1")
	require.Contains(t, str, "Duration:")
	require.Contains(t, str, "Failures: {auth: repo-2}")
}

func TestSyncRunMetadata_NoFailures(t *testing.T) {
	t.Parallel()

	metadata := entities.NewSyncRunMetadata("source", "target", "config", "env")

	metadata.AddSuccess("clone", "repo-1")
	metadata.AddSuccess("clone", "repo-2")

	str := metadata.String()
	require.Contains(t, str, "No failures")
}
