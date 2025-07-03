// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

// SyncRunMetadata tracks metadata about a synchronization run.
// This restores the sophisticated sync tracking functionality from main branch.
type SyncRunMetadata struct {
	StartTime         time.Time
	EndTime           time.Time
	TotalRepositories int
	ProcessedCount    int
	SuccessCount      int
	FailureCount      int
	SkippedCount      int
	Success           map[string][]string
	Failures          map[string][]string
	Skipped           map[string][]string
	SourceProvider    string
	TargetProvider    string
	ConfigurationName string
	EnvironmentName   string
	DryRun            bool
	Options           SyncRunOptions
}

// SyncRunOptions contains options for synchronization runs.
// Note: Using different name to avoid conflict with existing SyncOptions.
type SyncRunOptions struct {
	AlphaNumHyphName  bool
	ActiveFromLimit   string
	ForcePush         bool
	IgnoreInvalidName bool
	IncludeForks      bool
	IncludeArchived   bool
	UseGitBinary      bool
}

// NewSyncRunMetadata creates a new sync run metadata tracker.
func NewSyncRunMetadata(sourceProvider, targetProvider, configName, envName string) *SyncRunMetadata {
	return &SyncRunMetadata{
		StartTime:         time.Now(),
		Success:           make(map[string][]string),
		Failures:          make(map[string][]string),
		Skipped:           make(map[string][]string),
		SourceProvider:    sourceProvider,
		TargetProvider:    targetProvider,
		ConfigurationName: configName,
		EnvironmentName:   envName,
	}
}

// AddSuccess records a successful repository operation.
func (srm *SyncRunMetadata) AddSuccess(category, repositoryName string) {
	if srm.Success[category] == nil {
		srm.Success[category] = []string{}
	}

	srm.Success[category] = append(srm.Success[category], repositoryName)
	srm.SuccessCount++
	srm.ProcessedCount++
}

// AddFailure records a failed repository operation.
func (srm *SyncRunMetadata) AddFailure(category, repositoryName string) {
	if srm.Failures[category] == nil {
		srm.Failures[category] = []string{}
	}

	srm.Failures[category] = append(srm.Failures[category], repositoryName)
	srm.FailureCount++
	srm.ProcessedCount++
}

// AddSkipped records a skipped repository operation.
func (srm *SyncRunMetadata) AddSkipped(category, repositoryName string) {
	if srm.Skipped[category] == nil {
		srm.Skipped[category] = []string{}
	}

	srm.Skipped[category] = append(srm.Skipped[category], repositoryName)
	srm.SkippedCount++
}

// Finish marks the sync run as completed.
func (srm *SyncRunMetadata) Finish() {
	srm.EndTime = time.Now()
}

// Duration returns the total duration of the sync run.
func (srm *SyncRunMetadata) Duration() time.Duration {
	if srm.EndTime.IsZero() {
		return time.Since(srm.StartTime)
	}

	return srm.EndTime.Sub(srm.StartTime)
}

// IsFinished returns true if the sync run has finished.
func (srm *SyncRunMetadata) IsFinished() bool {
	return !srm.EndTime.IsZero()
}

// SuccessRate returns the success rate as a percentage.
func (srm *SyncRunMetadata) SuccessRate() float64 {
	if srm.ProcessedCount == 0 {
		return 0.0
	}

	return (float64(srm.SuccessCount) / float64(srm.ProcessedCount)) * 100.0
}

// FailureRate returns the failure rate as a percentage.
func (srm *SyncRunMetadata) FailureRate() float64 {
	if srm.ProcessedCount == 0 {
		return 0.0
	}

	return (float64(srm.FailureCount) / float64(srm.ProcessedCount)) * 100.0
}

// GetSummary returns a summary of the sync run.
func (srm *SyncRunMetadata) GetSummary() SyncSummary {
	return SyncSummary{
		TotalRepositories: srm.TotalRepositories,
		ProcessedCount:    srm.ProcessedCount,
		SuccessCount:      srm.SuccessCount,
		FailureCount:      srm.FailureCount,
		SkippedCount:      srm.SkippedCount,
		SuccessRate:       srm.SuccessRate(),
		FailureRate:       srm.FailureRate(),
		Duration:          srm.Duration(),
		IsFinished:        srm.IsFinished(),
		SourceProvider:    srm.SourceProvider,
		TargetProvider:    srm.TargetProvider,
		ConfigurationName: srm.ConfigurationName,
		EnvironmentName:   srm.EnvironmentName,
		DryRun:            srm.DryRun,
	}
}

// SyncSummary contains a summary of sync run results.
type SyncSummary struct {
	TotalRepositories int
	ProcessedCount    int
	SuccessCount      int
	FailureCount      int
	SkippedCount      int
	SuccessRate       float64
	FailureRate       float64
	Duration          time.Duration
	IsFinished        bool
	SourceProvider    string
	TargetProvider    string
	ConfigurationName string
	EnvironmentName   string
	DryRun            bool
}

// HasFailures returns true if there were any failures during the sync.
func (srm *SyncRunMetadata) HasFailures() bool {
	return srm.FailureCount > 0
}

// GetFailuresByCategory returns failures grouped by category.
func (srm *SyncRunMetadata) GetFailuresByCategory() map[string][]string {
	result := make(map[string][]string)
	for category, repos := range srm.Failures {
		result[category] = make([]string, len(repos))
		copy(result[category], repos)
	}

	return result
}

// GetSuccessesByCategory returns successes grouped by category.
func (srm *SyncRunMetadata) GetSuccessesByCategory() map[string][]string {
	result := make(map[string][]string)
	for category, repos := range srm.Success {
		result[category] = make([]string, len(repos))
		copy(result[category], repos)
	}

	return result
}

// GetSkippedByCategory returns skipped repositories grouped by category.
func (srm *SyncRunMetadata) GetSkippedByCategory() map[string][]string {
	result := make(map[string][]string)
	for category, repos := range srm.Skipped {
		result[category] = make([]string, len(repos))
		copy(result[category], repos)
	}

	return result
}

// SetTotalRepositories sets the total number of repositories to be processed.
func (srm *SyncRunMetadata) SetTotalRepositories(total int) {
	srm.TotalRepositories = total
}

// SetOptions sets the sync options used for this run.
func (srm *SyncRunMetadata) SetOptions(options SyncRunOptions) {
	srm.Options = options
}

// SetDryRun marks this as a dry run.
func (srm *SyncRunMetadata) SetDryRun(dryRun bool) {
	srm.DryRun = dryRun
}

// GetProgressPercentage returns the progress as a percentage.
func (srm *SyncRunMetadata) GetProgressPercentage() float64 {
	if srm.TotalRepositories == 0 {
		return 0.0
	}

	return (float64(srm.ProcessedCount) / float64(srm.TotalRepositories)) * 100.0
}

// Context-based functionality - restores missing functionality from main branch

// SyncRunMetadataKey is used as a key for context values.
// This restores the context-based metadata tracking from main branch.
type SyncRunMetadataKey struct{}

// GetMetadataFromContext retrieves sync metadata from context.
// This restores the context-based metadata access from main branch.
func GetMetadataFromContext(ctx context.Context) (*SyncRunMetadata, bool) {
	if metadata, ok := ctx.Value(SyncRunMetadataKey{}).(*SyncRunMetadata); ok {
		return metadata, true
	}

	return nil, false
}

// AddMetadataToContext adds sync metadata to context.
// This enables the context-based tracking pattern from main branch.
func AddMetadataToContext(ctx context.Context, metadata *SyncRunMetadata) context.Context {
	return context.WithValue(ctx, SyncRunMetadataKey{}, metadata)
}

// ContainsFailure checks if a specific failure exists in this metadata.
// This restores the failure checking functionality from main branch.
func (srm *SyncRunMetadata) ContainsFailure(category, repositoryName string) bool {
	if failureList, exists := srm.Failures[category]; exists {
		return slices.Contains(failureList, repositoryName)
	}

	return false
}

// ContainsFailureInContext checks if a specific failure exists in context metadata.
// This restores the context-based failure checking from main branch.
func ContainsFailureInContext(ctx context.Context, category, repositoryName string) bool {
	if metadata, ok := GetMetadataFromContext(ctx); ok {
		return metadata.ContainsFailure(category, repositoryName)
	}

	return false
}

// String provides a comprehensive string representation of sync metadata.
// This restores and enhances the string formatting from main branch.
func (srm *SyncRunMetadata) String() string {
	var parts []string

	// Basic info
	parts = append(parts, "Source: "+srm.SourceProvider)
	parts = append(parts, "Target: "+srm.TargetProvider)
	parts = append(parts, fmt.Sprintf("Total: %d", srm.TotalRepositories))
	parts = append(parts, fmt.Sprintf("Processed: %d", srm.ProcessedCount))

	// Results
	parts = append(parts, fmt.Sprintf("Successful: %d", srm.SuccessCount))
	parts = append(parts, fmt.Sprintf("Failed: %d", srm.FailureCount))
	parts = append(parts, fmt.Sprintf("Skipped: %d", srm.SkippedCount))

	// Timing
	if !srm.EndTime.IsZero() {
		parts = append(parts, fmt.Sprintf("Duration: %v", srm.Duration()))
	}

	// Failure details (similar to main branch format)
	if len(srm.Failures) > 0 {
		var failureDetails []string
		for category, details := range srm.Failures {
			failureDetails = append(failureDetails, fmt.Sprintf("%s: %s", category, strings.Join(details, ", ")))
		}

		parts = append(parts, fmt.Sprintf("Failures: {%s}", strings.Join(failureDetails, "; ")))
	} else {
		parts = append(parts, "No failures")
	}

	return fmt.Sprintf("SyncRunMetadata{%s}", strings.Join(parts, ", "))
}
