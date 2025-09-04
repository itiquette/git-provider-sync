// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// Test context key types for tests.
type testContextKey string
type testIntKey int
type testStructKey struct{}

// Test ConfigKey type

func TestConfigKey_Type(t *testing.T) {
	t.Parallel()

	var (
		key        ConfigKey
		anotherKey ConfigKey
	)

	// ConfigKey should be a zero-value struct

	assert.Equal(t, ConfigKey{}, key)
	assert.Equal(t, key, anotherKey)
}

// Test ConfigFromContext function

func TestConfigFromContext(t *testing.T) {
	t.Parallel()

	// Create a test CLI config for the valid case
	validConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/test/config.yaml").
		WithDryRun(true).
		WithOutputFormat("json").
		WithQuiet(true).
		WithForcePush(false).
		WithVerbosityWithCaller(true).
		Build()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedFound  bool
		expectedConfig entities.CLIConfig
	}{
		{
			name:           "empty context",
			setupContext:   context.Background,
			expectedFound:  false,
			expectedConfig: entities.CLIConfig{},
		},
		{
			name: "context with valid config",
			setupContext: func() context.Context {
				ctx := context.Background()

				return WithCLIConfig(ctx, validConfig)
			},
			expectedFound:  true,
			expectedConfig: validConfig,
		},
		{
			name: "context with wrong type",
			setupContext: func() context.Context {
				ctx := context.Background()

				return context.WithValue(ctx, ConfigKey{}, "not-a-config")
			},
			expectedFound:  false,
			expectedConfig: entities.CLIConfig{},
		},
		{
			name: "context with nil value",
			setupContext: func() context.Context {
				ctx := context.Background()

				return context.WithValue(ctx, ConfigKey{}, nil)
			},
			expectedFound:  false,
			expectedConfig: entities.CLIConfig{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := test.setupContext()

			// Act
			config, found := ConfigFromContext(ctx)

			// Assert
			if test.expectedFound {
				require.True(t, found)
			} else {
				assert.False(t, found)
			}

			assert.Equal(t, test.expectedConfig, config)
		})
	}
}

// Test WithCLIConfig function

func TestWithCLIConfig_NewContext(t *testing.T) {
	t.Parallel()

	originalCtx := context.Background()
	config := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/path/to/config.yaml").
		WithOutputFormat("plain").
		Build()

	newCtx := WithCLIConfig(originalCtx, config)

	// Should return a new context
	assert.NotEqual(t, originalCtx, newCtx)

	// Original context should not have the config
	_, found := ConfigFromContext(originalCtx)
	assert.False(t, found)

	// New context should have the config
	retrievedConfig, found := ConfigFromContext(newCtx)
	require.True(t, found)
	assert.Equal(t, config, retrievedConfig)
}

func TestWithCLIConfig_OverwriteExisting(t *testing.T) {
	t.Parallel()

	// Create initial context with config
	initialConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/initial/config.yaml").
		WithDryRun(false).
		Build()

	ctx := WithCLIConfig(context.Background(), initialConfig)

	// Verify initial config is set
	retrievedConfig, found := ConfigFromContext(ctx)
	require.True(t, found)
	assert.Equal(t, initialConfig, retrievedConfig)

	// Overwrite with new config
	newConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/new/config.yaml").
		WithDryRun(true).
		WithForcePush(true).
		Build()

	newCtx := WithCLIConfig(ctx, newConfig)

	// Should have the new config
	retrievedConfig, found = ConfigFromContext(newCtx)
	require.True(t, found)
	assert.Equal(t, newConfig, retrievedConfig)
	assert.NotEqual(t, initialConfig, retrievedConfig)

	// Verify specific overwritten fields
	assert.Equal(t, "/new/config.yaml", retrievedConfig.ConfigFilePath())
	assert.True(t, retrievedConfig.DryRun())
	assert.True(t, retrievedConfig.ForcePush())
}

func TestWithCLIConfig_DefaultConfig(t *testing.T) {
	t.Parallel()

	// Create default config using builder
	defaultConfig := entities.NewCLIConfigBuilder().Build()

	ctx := WithCLIConfig(context.Background(), defaultConfig)

	retrievedConfig, found := ConfigFromContext(ctx)
	require.True(t, found)
	assert.Equal(t, defaultConfig, retrievedConfig)

	// Verify default values
	assert.Empty(t, retrievedConfig.ConfigFilePath())
	assert.False(t, retrievedConfig.DryRun())
	assert.Equal(t, "console", retrievedConfig.OutputFormat())
	assert.False(t, retrievedConfig.Quiet())
	assert.False(t, retrievedConfig.ForcePush())
	assert.False(t, retrievedConfig.VerbosityWithCaller())
	assert.False(t, retrievedConfig.ConfigFileOnly())
	assert.False(t, retrievedConfig.AlphaNumHyphName())
	assert.False(t, retrievedConfig.IgnoreInvalidName())
	assert.Empty(t, retrievedConfig.ActiveFromLimit())
}

// Test context propagation and inheritance

func TestConfigContext_Propagation(t *testing.T) {
	t.Parallel()

	config := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/test/propagation.yaml").
		WithOutputFormat("json").
		WithQuiet(true).
		Build()

	// Create context chain: background -> with config -> derived
	ctx1 := context.Background()
	ctx2 := WithCLIConfig(ctx1, config)
	ctx3 := context.WithValue(ctx2, testContextKey("extra"), "value")

	// Test that config propagates through context chain correctly
	_, found1 := ConfigFromContext(ctx1)
	config2, found2 := ConfigFromContext(ctx2)
	config3, found3 := ConfigFromContext(ctx3)

	// Ctx1 (base) should not have config
	assert.False(t, found1)

	// Ctx2 (with config) should have config
	require.True(t, found2)
	assert.Equal(t, config.ConfigFilePath(), config2.ConfigFilePath())
	assert.Equal(t, config.OutputFormat(), config2.OutputFormat())
	assert.Equal(t, config.Quiet(), config2.Quiet())

	// Ctx3 (derived from ctx2) should inherit the same config
	require.True(t, found3)
	assert.Equal(t, config2.ConfigFilePath(), config3.ConfigFilePath())
	assert.Equal(t, config2.OutputFormat(), config3.OutputFormat())
	assert.Equal(t, config2.Quiet(), config3.Quiet())

	// Extra value should only be in ctx3
	assert.Nil(t, ctx1.Value(testContextKey("extra")))
	assert.Nil(t, ctx2.Value(testContextKey("extra")))
	assert.Equal(t, "value", ctx3.Value(testContextKey("extra")))
}

func TestConfigContext_Cancellation(t *testing.T) {
	t.Parallel()

	config := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/test/cancellation.yaml").
		Build()

	// Create cancellable context with config
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctxWithConfig := WithCLIConfig(ctx, config)

	// Should have config before cancellation
	retrievedConfig, found := ConfigFromContext(ctxWithConfig)
	require.True(t, found)
	assert.Equal(t, config, retrievedConfig)

	// Cancel the context
	cancel()

	// Should still have config after cancellation (config is independent of cancellation)
	retrievedConfig, found = ConfigFromContext(ctxWithConfig)
	require.True(t, found)
	assert.Equal(t, config, retrievedConfig)

	// But context should be cancelled
	select {
	case <-ctxWithConfig.Done():
		// Expected: context is cancelled
	default:
		t.Error("Context should be cancelled")
	}
}

// Test edge cases and error conditions

func TestConfigContext_WithOtherKeys_DoesNotInterfere(t *testing.T) {
	t.Parallel()

	config := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/test/multiple.yaml").
		Build()

	// Test that our ConfigKey doesn't interfere with other keys
	ctx := context.Background()
	ctx = context.WithValue(ctx, testContextKey("string-key"), "string-value")
	ctx = context.WithValue(ctx, testIntKey(42), "int-key-value")
	ctx = WithCLIConfig(ctx, config)
	ctx = context.WithValue(ctx, testStructKey{}, "struct-key-value")

	// All values should be retrievable
	assert.Equal(t, "string-value", ctx.Value(testContextKey("string-key")))
	assert.Equal(t, "int-key-value", ctx.Value(testIntKey(42)))
	assert.Equal(t, "struct-key-value", ctx.Value(testStructKey{}))

	// Our config should also be retrievable
	retrievedConfig, found := ConfigFromContext(ctx)
	require.True(t, found)
	assert.Equal(t, config, retrievedConfig)
}

func TestConfigContext_ZeroValueConfig(t *testing.T) {
	t.Parallel()

	// Test with zero-value config (not built with builder)
	var zeroConfig entities.CLIConfig

	ctx := WithCLIConfig(context.Background(), zeroConfig)

	retrievedConfig, found := ConfigFromContext(ctx)
	require.True(t, found)
	assert.Equal(t, zeroConfig, retrievedConfig)

	// Zero value should have empty/false defaults
	assert.Empty(t, retrievedConfig.ConfigFilePath())
	assert.False(t, retrievedConfig.DryRun())
	assert.Empty(t, retrievedConfig.OutputFormat()) // Zero value, not "console"
	assert.False(t, retrievedConfig.Quiet())
}

// Test concurrent access (context is safe for concurrent use)

func TestConfigContext_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	config := entities.NewCLIConfigBuilder().
		WithConfigFilePath("/test/concurrent.yaml").
		WithOutputFormat("json").
		Build()

	ctx := WithCLIConfig(context.Background(), config)

	// Launch multiple goroutines that read from context
	const numGoroutines = 100

	results := make(chan bool, numGoroutines)

	for range numGoroutines {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					results <- false

					return
				}
			}()

			retrievedConfig, found := ConfigFromContext(ctx)
			if !found {
				results <- false

				return
			}

			if retrievedConfig.ConfigFilePath() != "/test/concurrent.yaml" {
				results <- false

				return
			}

			if retrievedConfig.OutputFormat() != FormatJSON {
				results <- false

				return
			}

			results <- true
		}()
	}

	// Collect results
	successCount := 0

	for range numGoroutines {
		success := <-results
		if success {
			successCount++
		}
	}

	assert.Equal(t, numGoroutines, successCount, "All goroutines should successfully read from context")
}

// Test command-specific configuration patterns

func TestConfigContext_PrintCommandConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configFile     string
		outputFormat   string
		quiet          bool
		expectedConfig func(config entities.CLIConfig) bool
	}{
		{
			name:         "default print configuration",
			configFile:   "gitprovidersync.yaml",
			outputFormat: "console",
			quiet:        false,
			expectedConfig: func(config entities.CLIConfig) bool {
				return config.ConfigFilePath() == "gitprovidersync.yaml" &&
					config.OutputFormat() == "console" &&
					!config.Quiet()
			},
		},
		{
			name:         "quiet print configuration",
			configFile:   "/path/to/config.yaml",
			outputFormat: "json",
			quiet:        true,
			expectedConfig: func(config entities.CLIConfig) bool {
				return config.ConfigFilePath() == "/path/to/config.yaml" &&
					config.OutputFormat() == FormatJSON &&
					config.Quiet()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Build configuration for print command use case
			config := entities.NewCLIConfigBuilder().
				WithConfigFilePath(test.configFile).
				WithOutputFormat(test.outputFormat).
				WithQuiet(test.quiet).
				Build()

			ctx := WithCLIConfig(context.Background(), config)

			// Act: Extract config as a print command handler would
			retrievedConfig, found := ConfigFromContext(ctx)

			// Assert: Verify configuration extraction and values
			require.True(t, found, "Print command should access CLI config via context")
			assert.True(t, test.expectedConfig(retrievedConfig), "Configuration should match expected print command settings")
		})
	}
}

func TestConfigContext_SyncCommandConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configFile     string
		dryRun         bool
		forcePush      bool
		outputFormat   string
		quiet          bool
		expectedConfig func(config entities.CLIConfig) bool
	}{
		{
			name:         "standard sync configuration",
			configFile:   "/custom/config.yaml",
			dryRun:       true,
			forcePush:    false,
			outputFormat: "plain",
			quiet:        true,
			expectedConfig: func(config entities.CLIConfig) bool {
				return config.ConfigFilePath() == "/custom/config.yaml" &&
					config.DryRun() &&
					!config.ForcePush() &&
					config.OutputFormat() == "plain" &&
					config.Quiet()
			},
		},
		{
			name:         "production sync configuration",
			configFile:   "prod-sync.yaml",
			dryRun:       false,
			forcePush:    true,
			outputFormat: "json",
			quiet:        false,
			expectedConfig: func(config entities.CLIConfig) bool {
				return config.ConfigFilePath() == "prod-sync.yaml" &&
					!config.DryRun() &&
					config.ForcePush() &&
					config.OutputFormat() == "json" &&
					!config.Quiet()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Build configuration for sync command use case
			config := entities.NewCLIConfigBuilder().
				WithConfigFilePath(test.configFile).
				WithDryRun(test.dryRun).
				WithForcePush(test.forcePush).
				WithOutputFormat(test.outputFormat).
				WithQuiet(test.quiet).
				Build()

			ctx := WithCLIConfig(context.Background(), config)

			// Act: Extract config as a sync command handler would
			retrievedConfig, found := ConfigFromContext(ctx)

			// Assert: Verify configuration extraction and values
			require.True(t, found, "Sync command should access CLI config via context")
			assert.True(t, test.expectedConfig(retrievedConfig), "Configuration should match expected sync command settings")
		})
	}
}

func TestConfigContext_StatusCommandConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configFile     string
		outputFormat   string
		configFileOnly bool
		expectedConfig func(config entities.CLIConfig) bool
	}{
		{
			name:           "json status configuration",
			configFile:     "config.yaml",
			outputFormat:   "json",
			configFileOnly: true,
			expectedConfig: func(config entities.CLIConfig) bool {
				return config.ConfigFilePath() == "config.yaml" &&
					config.OutputFormat() == "json" &&
					config.ConfigFileOnly()
			},
		},
		{
			name:           "console status configuration",
			configFile:     "/full/path/status.yaml",
			outputFormat:   "console",
			configFileOnly: false,
			expectedConfig: func(config entities.CLIConfig) bool {
				return config.ConfigFilePath() == "/full/path/status.yaml" &&
					config.OutputFormat() == "console" &&
					!config.ConfigFileOnly()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Build configuration for status command use case
			config := entities.NewCLIConfigBuilder().
				WithConfigFilePath(test.configFile).
				WithOutputFormat(test.outputFormat).
				WithConfigFileOnly(test.configFileOnly).
				Build()

			ctx := WithCLIConfig(context.Background(), config)

			// Act: Extract config as a status command handler would
			retrievedConfig, found := ConfigFromContext(ctx)

			// Assert: Verify configuration extraction and values
			require.True(t, found, "Status command should access CLI config via context")
			assert.True(t, test.expectedConfig(retrievedConfig), "Configuration should match expected status command settings")
		})
	}
}
