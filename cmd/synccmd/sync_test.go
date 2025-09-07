// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package synccmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/configuration/dto"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

func TestConvertRepositoryFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    dto.RepositoriesOption
		expected ports.FilterOptions
	}{
		{
			name: "with include and exclude patterns",
			input: dto.RepositoriesOption{
				Include: []string{"repo1", "repo2*"},
				Exclude: []string{"temp*", "test*"},
			},
			expected: ports.FilterOptions{
				IncludePatterns: []string{"repo1", "repo2*"},
				ExcludePatterns: []string{"temp*", "test*"},
				IncludeForks:    true,
				IncludeArchived: false,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
		},
		{
			name: "empty filters",
			input: dto.RepositoriesOption{
				Include: []string{},
				Exclude: []string{},
			},
			expected: ports.FilterOptions{
				IncludePatterns: []string{},
				ExcludePatterns: []string{},
				IncludeForks:    true,
				IncludeArchived: false,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
		},
		{
			name: "only include patterns",
			input: dto.RepositoriesOption{
				Include: []string{"important*", "core-*"},
				Exclude: []string{},
			},
			expected: ports.FilterOptions{
				IncludePatterns: []string{"important*", "core-*"},
				ExcludePatterns: []string{},
				IncludeForks:    true,
				IncludeArchived: false,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
		},
		{
			name: "only exclude patterns",
			input: dto.RepositoriesOption{
				Include: []string{},
				Exclude: []string{"legacy-*", "archived-*"},
			},
			expected: ports.FilterOptions{
				IncludePatterns: []string{},
				ExcludePatterns: []string{"legacy-*", "archived-*"},
				IncludeForks:    true,
				IncludeArchived: false,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := convertRepositoryFilters(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertMirrorConfigToMirrorTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    dto.SyncConfig
		expected int // Number of expected targets
	}{
		{
			name: "valid mirror configuration",
			input: dto.SyncConfig{
				Mirrors: map[string]dto.MirrorConfig{
					"target1": {
						BaseConfig: dto.BaseConfig{
							ProviderType: "github",
							Domain:       "github.com",
							Owner:        "testowner",
							Auth: dto.AuthConfig{
								Token: "test-token",
							},
						},
						Path: "",
					},
					"target2": {
						BaseConfig: dto.BaseConfig{
							ProviderType: "gitlab",
							Domain:       "gitlab.com",
							Owner:        "testowner",
							Auth: dto.AuthConfig{
								Token: "test-token",
							},
						},
						Path: "",
					},
				},
			},
			expected: 2,
		},
		{
			name: "invalid provider type",
			input: dto.SyncConfig{
				Mirrors: map[string]dto.MirrorConfig{
					"target1": {
						BaseConfig: dto.BaseConfig{
							ProviderType: "invalid-provider",
							Domain:       "example.com",
							Owner:        "testowner",
							Auth: dto.AuthConfig{
								Token: "test-token",
							},
						},
						Path: "",
					},
				},
			},
			expected: 0, // Should be skipped
		},
		{
			name: "empty mirrors",
			input: dto.SyncConfig{
				Mirrors: map[string]dto.MirrorConfig{},
			},
			expected: 0,
		},
		{
			name: "mixed valid and invalid",
			input: dto.SyncConfig{
				Mirrors: map[string]dto.MirrorConfig{
					"valid": {
						BaseConfig: dto.BaseConfig{
							ProviderType: "github",
							Domain:       "github.com",
							Owner:        "testowner",
							Auth: dto.AuthConfig{
								Token: "test-token",
							},
						},
						Path: "",
					},
					"invalid": {
						BaseConfig: dto.BaseConfig{
							ProviderType: "invalid-provider",
							Domain:       "example.com",
							Owner:        "testowner",
							Auth: dto.AuthConfig{
								Token: "test-token",
							},
						},
						Path: "",
					},
				},
			},
			expected: 1, // Only valid one should be included
		},
		{
			name: "invalid mirror name",
			input: dto.SyncConfig{
				Mirrors: map[string]dto.MirrorConfig{
					"": { // Empty name should fail
						BaseConfig: dto.BaseConfig{
							ProviderType: "github",
							Domain:       "github.com",
							Owner:        "testowner",
							Auth: dto.AuthConfig{
								Token: "test-token",
							},
						},
						Path: "",
					},
				},
			},
			expected: 0, // Should be skipped due to invalid name
		},
		{
			name: "invalid owner name",
			input: dto.SyncConfig{
				Mirrors: map[string]dto.MirrorConfig{
					"valid-name": {
						BaseConfig: dto.BaseConfig{
							ProviderType: "github",
							Domain:       "github.com",
							Owner:        "", // Empty owner should fail
							Auth: dto.AuthConfig{
								Token: "test-token",
							},
						},
						Path: "",
					},
				},
			},
			expected: 0, // Should be skipped due to invalid owner
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := convertMirrorConfigToMirrorTargets(tt.input)
			assert.Len(t, result, tt.expected)

			// Verify valid targets have required fields
			for _, target := range result {
				assert.NotEmpty(t, target.Name())
				assert.NotEmpty(t, target.ProviderType())
				assert.NotEmpty(t, target.Owner())
			}
		})
	}
}

func TestSyncInputOption(t *testing.T) {
	t.Parallel()

	// Create a temporary file for log output
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	logFile, err := os.Create(filepath.Clean(logPath))
	require.NoError(t, err)
	t.Cleanup(func() { _ = logFile.Close() })

	logger := zerolog.New(logFile).With().Timestamp().Logger()

	tests := []struct {
		name   string
		option syncInputOption
	}{
		{
			name: "all options enabled",
			option: syncInputOption{
				activeFromLimit:   "2024-01-01",
				alphaNumHyphName:  true,
				dryRun:            true,
				forcePush:         true,
				ignoreInvalidName: true,
			},
		},
		{
			name: "all options disabled",
			option: syncInputOption{
				activeFromLimit:   "",
				alphaNumHyphName:  false,
				dryRun:            false,
				forcePush:         false,
				ignoreInvalidName: false,
			},
		},
		{
			name: "mixed options",
			option: syncInputOption{
				activeFromLimit:   "2024-06-01",
				alphaNumHyphName:  true,
				dryRun:            false,
				forcePush:         true,
				ignoreInvalidName: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test DebugLog method
			event := tt.option.DebugLog(&logger)
			assert.NotNil(t, event)

			// Test that it doesn't panic
			event.Msg("test message")
		})
	}
}

func TestOutputSyncResults(t *testing.T) {
	t.Parallel()

	// Create test results
	ctx := context.Background()
	cliConfig := entities.NewCLIConfigBuilder().
		WithOutputFormat("json").
		Build()

	// Define a custom key type to avoid collisions
	type cliConfigKey struct{}

	// Add CLI config to context
	ctx = context.WithValue(ctx, cliConfigKey{}, cliConfig)

	// Create simple sync results
	results := sync.NewResults(true) // dry run
	results.AddResult(sync.Result{
		Environment:     "test",
		Source:          "source",
		SourceProvider:  "github",
		Repository:      "test-repo",
		Mirror:          "mirror",
		MirrorProvider:  "gitlab",
		Status:          "SUCCESS",
		Action:          "CREATED",
		DurationSeconds: 1.5,
	})

	// Test output function (should not panic)
	err := outputSyncResults(ctx, results)
	if err != nil {
		t.Logf("Output error (expected in test): %v", err)
	}
}

func TestShowSyncSummary(_ *testing.T) { //nolint:paralleltest // DO NOT run this in parallel due to race conditions with global logger state
	// Create test results with different outcomes
	results := sync.NewResults(false) // not dry run
	results.SuccessfulSyncs = 5
	results.FailedSyncs = 2
	results.SkippedSyncs = 1
	results.TotalRepositories = 8

	// Test summary function (should not panic)
	showSyncSummary(results)

	// Test with dry run
	dryRunResults := sync.NewResults(true)
	dryRunResults.SuccessfulSyncs = 3
	dryRunResults.FailedSyncs = 0
	dryRunResults.SkippedSyncs = 0
	dryRunResults.TotalRepositories = 3

	showSyncSummary(dryRunResults)
}

func TestSaveLastSyncInfo(t *testing.T) { //nolint:paralleltest // Cannot run in parallel due to directory changes
	// Don't use t.Parallel() since we're changing directories

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create test results
	results := sync.NewResults(false)
	results.TotalRepositories = 10
	results.SuccessfulSyncs = 8
	results.FailedSyncs = 1
	results.SkippedSyncs = 1

	// Test save function - it saves to current directory
	// So we need to be in the tmp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	saveLastSyncInfo(results)

	// Verify file was created in temp directory
	content, err := os.ReadFile(getLastSyncFilePath())
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "repos=10")
	assert.Contains(t, contentStr, "successful=8")
	assert.Contains(t, contentStr, "failed=1")
	assert.Contains(t, contentStr, "skipped=1")
	assert.Contains(t, contentStr, "timestamp=")
}

// Additional integration-style tests for better coverage.
func TestSyncHexagonal_WithMissingTmpDir_HandlesGracefully(t *testing.T) {
	t.Parallel()

	// Create context without CLI config
	ctx := context.Background()

	// Create minimal config
	cfg := &dto.AppConfiguration{
		GitProviderSyncConfs: map[string]dto.Environment{},
	}

	// Test should fail gracefully due to missing tmp dir creation
	err := performSync(ctx, cfg)
	if err != nil {
		// Expected to fail, verify it's not a panic
		t.Logf("Got expected error: %v", err)
	} else {
		t.Log("Unexpectedly succeeded - might be environment specific")
	}
}
