// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package synccmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	cliAdapters "itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/domain/entities"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

func TestSyncHexagonalTmpDirCreation(t *testing.T) {
	t.Parallel()

	// Test performSync function with proper CLI context
	ctx := context.Background()

	// Create minimal CLI config
	cliConfig := entities.NewCLIConfigBuilder().
		WithDryRun(true).
		Build()
	ctx = cliAdapters.WithCLIConfig(ctx, cliConfig)

	// Create minimal app configuration
	cfg := &gpsconfig.AppConfiguration{
		GitProviderSyncConfs: map[string]gpsconfig.Environment{
			"test": {
				"source": gpsconfig.SyncConfig{
					BaseConfig: gpsconfig.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testuser",
						Auth: gpsconfig.AuthConfig{
							Token: "test-token",
						},
					},
					Mirrors: map[string]gpsconfig.MirrorConfig{
						"mirror1": {
							BaseConfig: gpsconfig.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "testuser",
								Auth: gpsconfig.AuthConfig{
									Token: "test-token",
								},
							},
						},
					},
				},
			},
		},
	}

	// This will fail at container creation but should successfully create temp dir
	err := performSync(ctx, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize application services")
}

func TestCreateContainer_WithCLIOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cliOpts  entities.CLIConfig
		expected bool // whether we expect success
	}{
		{
			name: "dry run enabled",
			cliOpts: entities.NewCLIConfigBuilder().
				WithDryRun(true).
				Build(),
			expected: false, // Will fail without proper config
		},
		{
			name: "force push enabled",
			cliOpts: entities.NewCLIConfigBuilder().
				WithForcePush(true).
				Build(),
			expected: false, // Will fail without proper config
		},
		{
			name: "with verbosity",
			cliOpts: entities.NewCLIConfigBuilder().
				WithVerbosityWithCaller(true).
				Build(),
			expected: false, // Will fail without proper config
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ctx = cliAdapters.WithCLIConfig(ctx, testCase.cliOpts)

			cfg := &gpsconfig.AppConfiguration{}

			container, err := createContainerWithConfig(ctx, cfg)
			if testCase.expected {
				require.NoError(t, err)
				assert.NotNil(t, container)

				if container != nil {
					_ = container.Close()
				}
			} else {
				require.Error(t, err)
				assert.Nil(t, container)
			}
		})
	}
}

func TestSyncInputOption_DebugLogWithEmptyValues(t *testing.T) { //nolint:paralleltest // DO NOT run this in parallel due to race conditions with logger
	// Test with empty values
	sio := syncInputOption{
		activeFromLimit:   "",
		alphaNumHyphName:  false,
		dryRun:            false,
		forcePush:         false,
		ignoreInvalidName: false,
	}

	// Create a temporary file for log output
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.log")
	file, err := os.Create(filepath.Clean(tmpFile))
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	// Create logger that writes to file
	logger := createTestLogger(file)

	// Test DebugLog method with empty values
	event := sio.DebugLog(logger)
	assert.NotNil(t, event)
	event.Msg("empty values test")

	// Close and read the file
	_ = file.Close()
	content, err := os.ReadFile(filepath.Clean(tmpFile))
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "empty values test")
	assert.Contains(t, logContent, "alphaNumHyphName")
	assert.Contains(t, logContent, "dryRun")
	assert.Contains(t, logContent, "forcePush")
	assert.Contains(t, logContent, "ignoreInvalidName")
}

func TestSyncHexagonal_WithoutCLIConfig_ReturnsError(t *testing.T) {
	t.Parallel()

	// Test performSync without CLI config in context
	ctx := context.Background()
	// Note: Not adding CLI config to context

	cfg := &gpsconfig.AppConfiguration{
		GitProviderSyncConfs: map[string]gpsconfig.Environment{},
	}

	// This should fail when trying to access CLI options
	err := performSync(ctx, cfg)
	require.Error(t, err)
	// Should fail early when trying to create temp dir or access CLI options
}

func TestCreateContainerWithoutCLIConfig(t *testing.T) {
	t.Parallel()

	// Test createContainer without CLI config in context
	ctx := context.Background()
	// Note: Not adding CLI config to context

	cfg := &gpsconfig.AppConfiguration{}

	// This should fail when trying to access CLI options
	container, err := createContainerWithConfig(ctx, cfg)
	require.Error(t, err)
	assert.Nil(t, container)
	// Error may be about CLI options or configuration - both are expected
}

func TestCLIConfigMerging_WithExistingConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		baseConfig     entities.CLIConfig
		expectSuccess  bool
		expectedQuiet  bool
		expectedFormat string
	}{
		{
			name: "default base config",
			baseConfig: entities.NewCLIConfigBuilder().
				WithOutputFormat("console").
				WithQuiet(false).
				Build(),
			expectSuccess:  true,
			expectedQuiet:  false,
			expectedFormat: "console",
		},
		{
			name: "configured base config",
			baseConfig: entities.NewCLIConfigBuilder().
				WithOutputFormat("json").
				WithQuiet(true).
				WithConfigFilePath("/test/path").
				WithConfigFileOnly(true).
				Build(),
			expectSuccess:  true,
			expectedQuiet:  true,
			expectedFormat: "json",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a command with sync flags
			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dry-run"},
					&cli.BoolFlag{Name: "force-push"},
					&cli.StringFlag{Name: "since"},
					&cli.BoolFlag{Name: "sanitize-names"},
					&cli.BoolFlag{Name: "skip-invalid"},
				},
			}

			// Set some sync-specific values
			_ = cmd.Set("dry-run", "true")
			_ = cmd.Set("since", "2024-01-01")

			// Test merging
			result := mergeSyncOptionsWithCLIConfig(testCase.baseConfig, cmd)

			// Verify base config values are preserved
			assert.Equal(t, testCase.expectedQuiet, result.Quiet())
			assert.Equal(t, testCase.expectedFormat, result.OutputFormat())

			// Verify sync flags are set correctly
			assert.True(t, result.DryRun())
			assert.Equal(t, "2024-01-01", result.ActiveFromLimit())
		})
	}
}

func TestInitLogger_WithCLIConfigurations(t *testing.T) { //nolint:paralleltest // DO NOT run this in parallel due to race conditions with global logger state
	tests := []struct {
		name     string
		setupCtx func() context.Context
	}{
		{
			name:     "context without CLI config",
			setupCtx: context.Background,
		},
		{
			name: "context with CLI config - verbosity with caller",
			setupCtx: func() context.Context {
				ctx := context.Background()
				cliConfig := entities.NewCLIConfigBuilder().
					WithVerbosityWithCaller(true).
					WithQuiet(false).
					WithOutputFormat("json").
					Build()

				return cliAdapters.WithCLIConfig(ctx, cliConfig)
			},
		},
		{
			name: "context with CLI config - quiet mode",
			setupCtx: func() context.Context {
				ctx := context.Background()
				cliConfig := entities.NewCLIConfigBuilder().
					WithVerbosityWithCaller(false).
					WithQuiet(true).
					WithOutputFormat("plain").
					Build()

				return cliAdapters.WithCLIConfig(ctx, cliConfig)
			},
		},
	}

	for _, tt := range tests { //nolint:paralleltest // subtests cannot run in parallel due to shared state
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()

			// Test that initLogger doesn't panic and returns a valid context
			resultCtx := initLogger(ctx, nil)
			assert.NotNil(t, resultCtx)
		})
	}
}
