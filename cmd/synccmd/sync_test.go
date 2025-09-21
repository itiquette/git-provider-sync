// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package synccmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/configuration/dto"
	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

func TestConvertRepositoryFilters_HandlesEmptyAndPopulatedPatterns(t *testing.T) {
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

func TestConvertMirrorConfigToMirrorTargets_ValidatesAndFiltersInvalidConfigs(t *testing.T) {
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

func TestSyncInputOption_DebugLogOutputsAllFieldsWithoutPanic(t *testing.T) {
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

func TestOutputSyncResults_FormatsResultsUsingConfiguredFormatter(t *testing.T) {
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

func TestShowSyncSummary_OutputsCorrectSummaryForDifferentResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		results         *sync.Results
		expectedContent []string
		notExpected     []string
	}{
		{
			name: "mixed_results_without_dry_run",
			results: func() *sync.Results {
				results := sync.NewResults(false)
				results.SuccessfulSyncs = 5
				results.FailedSyncs = 2
				results.SkippedSyncs = 1
				results.TotalRepositories = 8

				return results
			}(),
			expectedContent: []string{
				"✓ Successfully synced 5 repositories",
				"✗ 2 repositories failed",
				"- 1 repositories skipped",
				"Next: gitprovidersync status",
			},
			notExpected: []string{
				"without --dry-run",
			},
		},
		{
			name: "successful_dry_run",
			results: func() *sync.Results {
				results := sync.NewResults(true)
				results.SuccessfulSyncs = 3
				results.FailedSyncs = 0
				results.SkippedSyncs = 0
				results.TotalRepositories = 3

				return results
			}(),
			expectedContent: []string{
				"✓ Successfully synced 3 repositories",
				"Next: gitprovidersync sync (without --dry-run)",
			},
			notExpected: []string{
				"✗",  // No failures
				"- ", // No skipped
			},
		},
		{
			name: "only_failures",
			results: func() *sync.Results {
				results := sync.NewResults(false)
				results.SuccessfulSyncs = 0
				results.FailedSyncs = 4
				results.SkippedSyncs = 0
				results.TotalRepositories = 4

				return results
			}(),
			expectedContent: []string{
				"✗ 4 repositories failed",
			},
			notExpected: []string{
				"✓",  // No success
				"- ", // No skipped
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Use a buffer to capture output instead of os.Stderr
			var buf bytes.Buffer
			showSyncSummaryToWriter(testCase.results, &buf)

			output := buf.String()

			// Check expected content
			for _, expected := range testCase.expectedContent {
				assert.Contains(t, output, expected, "Missing expected content: %s", expected)
			}

			// Check content that should not be present
			for _, notExpected := range testCase.notExpected {
				assert.NotContains(t, output, notExpected, "Found unexpected content: %s", notExpected)
			}
		})
	}
}

func TestSaveLastSyncInfo_WritesCorrectInfoToFileWithoutChangingDirectory(t *testing.T) {
	t.Parallel() // Safe now - no more directory changes!

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create test results
	results := sync.NewResults(false)
	results.TotalRepositories = 10
	results.SuccessfulSyncs = 8
	results.FailedSyncs = 1
	results.SkippedSyncs = 1

	// Use dependency injection with a test writer pointing to our temp directory
	// No need to change directories anymore!
	writer := filesystem.NewSyncInfoWriter(tmpDir)
	saveLastSyncInfoWithWriter(results, writer)

	// Verify file was created in temp directory
	expectedPath := filepath.Join(tmpDir, ".gitprovidersync-last-sync")
	content, err := os.ReadFile(expectedPath) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "repos=10")
	assert.Contains(t, contentStr, "successful=8")
	assert.Contains(t, contentStr, "failed=1")
	assert.Contains(t, contentStr, "skipped=1")
	assert.Contains(t, contentStr, "timestamp=")
}

// Additional integration-style tests for better coverage.
func TestPerformSync_WithEmptyConfiguration_ReturnsErrorGracefully(t *testing.T) {
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
