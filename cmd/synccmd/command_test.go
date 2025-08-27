// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package synccmd

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"itiquette/git-provider-sync/internal/domain/entities"
)

func TestNewSyncCommand_Constructor_CreatesCommandWithFlags(t *testing.T) {
	t.Parallel()

	cmd := NewSyncCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "sync", cmd.Name)
	assert.Equal(t, "Mirror repositories from a source Git provider to targets", cmd.Usage)
	assert.Contains(t, cmd.Description, "The 'sync' command mirrors your repositories")
	assert.NotNil(t, cmd.Action)
	assert.Len(t, cmd.Flags, 5)

	// Check flags
	flagNames := make([]string, 0, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		flagNames = append(flagNames, flag.Names()[0])
	}

	expectedFlags := []string{"dry-run", "force-push", "since", "sanitize-names", "skip-invalid"}
	assert.ElementsMatch(t, expectedFlags, flagNames)
}

func TestMergeSyncOptionsWithCLIConfig_MergesFlags(t *testing.T) {
	t.Parallel()

	// Create a CLI config with base values
	baseConfig := entities.NewCLIConfigBuilder().
		WithOutputFormat("json").
		WithQuiet(true).
		Build()

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

	// Test the function merges values correctly
	result := mergeSyncOptionsWithCLIConfig(baseConfig, cmd)
	assert.NotNil(t, result)

	// Base config values should be preserved
	assert.Equal(t, "json", result.OutputFormat())
	assert.True(t, result.Quiet())

	// Sync flags should have default values when not set
	assert.False(t, result.DryRun())
	assert.False(t, result.ForcePush())
	assert.Empty(t, result.ActiveFromLimit())
	assert.False(t, result.AlphaNumHyphName())
	assert.False(t, result.IgnoreInvalidName())
}

func TestRunSync_InvalidConfig(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create invalid config file
	invalidConfig := `
invalid_yaml_content:
  missing_required_fields: true
  invalid_structure:
`
	configPath := filepath.Join(tmpDir, "gitprovidersync.yaml")
	err := os.WriteFile(configPath, []byte(invalidConfig), 0600)
	require.NoError(t, err)

	ctx := context.Background()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run"},
			&cli.BoolFlag{Name: "force-push"},
			&cli.StringFlag{Name: "since"},
			&cli.BoolFlag{Name: "sanitize-names"},
			&cli.BoolFlag{Name: "skip-invalid"},
			&cli.StringFlag{Name: "config-file"},
			&cli.BoolFlag{Name: "config-file-only"},
			&cli.StringFlag{Name: "format"},
			&cli.IntFlag{Name: "verbosity-with-caller"},
			&cli.BoolFlag{Name: "quiet"},
		},
	}

	// Set the config file path
	_ = cmd.Set("config-file", configPath)

	// Test should return error but not exit (since testing.Testing() returns true)
	err = runSync(ctx, cmd)
	require.Error(t, err)
	// We expect some kind of configuration error
	t.Logf("Got expected error: %v", err)
}

func TestRunSync_ValidConfig_DryRun(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create minimal valid config
	validConfig := `
gitprovidersync:
  testenv:
    testsource:
      provider_type: github
      domain: github.com
      owner: testowner
      owner_type: user
      auth:
        token: test-token
      mirrors:
        testmirror:
          provider_type: github
          domain: github.com
          owner: testowner
          owner_type: user
          auth:
            token: test-token
`
	configPath := filepath.Join(tmpDir, "gitprovidersync.yaml")
	err := os.WriteFile(configPath, []byte(validConfig), 0600)
	require.NoError(t, err)

	ctx := context.Background()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run"},
			&cli.BoolFlag{Name: "force-push"},
			&cli.StringFlag{Name: "since"},
			&cli.BoolFlag{Name: "sanitize-names"},
			&cli.BoolFlag{Name: "skip-invalid"},
			&cli.StringFlag{Name: "config-file"},
			&cli.BoolFlag{Name: "config-file-only"},
			&cli.StringFlag{Name: "format"},
			&cli.IntFlag{Name: "verbosity-with-caller"},
			&cli.BoolFlag{Name: "quiet"},
		},
	}

	// Set the config file path
	_ = cmd.Set("config-file", configPath)

	// This test may still fail due to complex dependencies but should not panic
	err = runSync(ctx, cmd)
	// We expect this to fail due to missing external dependencies, but it should fail gracefully
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestMergeSyncOptionsWithCLIConfig_IntegrationWithFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                      string
		baseConfig                entities.CLIConfig
		dryRun                    bool
		forcePush                 bool
		activeFromLimit           string
		alphaNumHyphName          bool
		ignoreInvalidName         bool
		expectedDryRun            bool
		expectedForcePush         bool
		expectedActiveFromLimit   string
		expectedAlphaNumHyphName  bool
		expectedIgnoreInvalidName bool
	}{
		{
			name: "basic flags with base config",
			baseConfig: entities.NewCLIConfigBuilder().
				WithOutputFormat("json").
				WithQuiet(true).
				Build(),
			dryRun:                    true,
			forcePush:                 false,
			activeFromLimit:           "2024-01-01",
			alphaNumHyphName:          true,
			ignoreInvalidName:         false,
			expectedDryRun:            true,
			expectedForcePush:         false,
			expectedActiveFromLimit:   "2024-01-01",
			expectedAlphaNumHyphName:  true,
			expectedIgnoreInvalidName: false,
		},
		{
			name: "all flags enabled",
			baseConfig: entities.NewCLIConfigBuilder().
				WithOutputFormat("console").
				Build(),
			dryRun:                    true,
			forcePush:                 true,
			activeFromLimit:           "2024-06-01",
			alphaNumHyphName:          true,
			ignoreInvalidName:         true,
			expectedDryRun:            true,
			expectedForcePush:         true,
			expectedActiveFromLimit:   "2024-06-01",
			expectedAlphaNumHyphName:  true,
			expectedIgnoreInvalidName: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock CLI command with flags
			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dry-run"},
					&cli.BoolFlag{Name: "force-push"},
					&cli.StringFlag{Name: "since"},
					&cli.BoolFlag{Name: "sanitize-names"},
					&cli.BoolFlag{Name: "skip-invalid"},
				},
			}

			// Set flag values
			_ = cmd.Set("dry-run", strconv.FormatBool(testCase.dryRun))
			_ = cmd.Set("force-push", strconv.FormatBool(testCase.forcePush))
			_ = cmd.Set("since", testCase.activeFromLimit)
			_ = cmd.Set("sanitize-names", strconv.FormatBool(testCase.alphaNumHyphName))
			_ = cmd.Set("skip-invalid", strconv.FormatBool(testCase.ignoreInvalidName))

			// Test the merge function
			result := mergeSyncOptionsWithCLIConfig(testCase.baseConfig, cmd)

			// Verify sync-specific flags are merged correctly
			assert.Equal(t, testCase.expectedDryRun, result.DryRun())
			assert.Equal(t, testCase.expectedForcePush, result.ForcePush())
			assert.Equal(t, testCase.expectedActiveFromLimit, result.ActiveFromLimit())
			assert.Equal(t, testCase.expectedAlphaNumHyphName, result.AlphaNumHyphName())
			assert.Equal(t, testCase.expectedIgnoreInvalidName, result.IgnoreInvalidName())

			// Verify base config values are preserved
			assert.Equal(t, testCase.baseConfig.OutputFormat(), result.OutputFormat())
			assert.Equal(t, testCase.baseConfig.Quiet(), result.Quiet())
		})
	}
}

func TestInitLogger(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cmd := &cli.Command{}

	// Test that initLogger doesn't panic and returns a valid context
	resultCtx := initLogger(ctx, cmd)
	assert.NotNil(t, resultCtx)
}

func TestSyncInputOption_DebugLog(t *testing.T) { //nolint:paralleltest // DO NOT run this in parallel due to race conditions with logger
	// Create a temporary file for log output
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.log")
	file, err := os.Create(filepath.Clean(tmpFile))
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	// Create logger that writes to file
	logger := createTestLogger(file)

	sio := syncInputOption{
		activeFromLimit:   "2024-01-01",
		alphaNumHyphName:  true,
		dryRun:            false,
		forcePush:         true,
		ignoreInvalidName: false,
	}

	// Test DebugLog method
	event := sio.DebugLog(logger)
	assert.NotNil(t, event)
	event.Msg("test message")

	// Close and read the file
	_ = file.Close()
	content, err := os.ReadFile(filepath.Clean(tmpFile))
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "test message")
	assert.Contains(t, logContent, "2024-01-01")
	assert.Contains(t, logContent, "alphaNumHyphName")
	assert.Contains(t, logContent, "dryRun")
	assert.Contains(t, logContent, "forcePush")
	assert.Contains(t, logContent, "ignoreInvalidName")
}

// Helper function to create a test logger.
func createTestLogger(output *os.File) *zerolog.Logger {
	logger := zerolog.New(output).With().Timestamp().Logger()

	return &logger
}

func TestRunSync_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCmd    func() *cli.Command
		expectedErr bool
	}{
		{
			name: "invalid CLI options",
			setupCmd: func() *cli.Command {
				return &cli.Command{
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "config-file"},
						&cli.StringFlag{Name: "format"},
						&cli.BoolFlag{Name: "config-file-only"},
						&cli.BoolFlag{Name: "quiet"},
						&cli.IntFlag{Name: "verbosity-with-caller"},
					},
				}
			},
			expectedErr: true,
		},
		{
			name: "missing config file",
			setupCmd: func() *cli.Command {
				cmd := &cli.Command{
					Flags: []cli.Flag{
						&cli.BoolFlag{Name: "dry-run"},
						&cli.BoolFlag{Name: "force-push"},
						&cli.StringFlag{Name: "since"},
						&cli.BoolFlag{Name: "sanitize-names"},
						&cli.BoolFlag{Name: "skip-invalid"},
						&cli.StringFlag{Name: "config-file"},
						&cli.BoolFlag{Name: "config-file-only"},
						&cli.StringFlag{Name: "format"},
						&cli.IntFlag{Name: "verbosity-with-caller"},
						&cli.BoolFlag{Name: "quiet"},
					},
				}
				// Set config file to non-existent path
				_ = cmd.Set("config-file", "/nonexistent/config.yaml")

				return cmd
			},
			expectedErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			cmd := testCase.setupCmd()

			err := runSync(ctx, cmd)
			if testCase.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunSync_FlagVariations(t *testing.T) {
	t.Parallel()

	// Create a minimal valid config for testing
	tmpDir := t.TempDir()
	validConfig := `
gitprovidersync:
  testenv:
    testsource:
      provider_type: github
      domain: github.com
      owner: testowner
      owner_type: user
      auth:
        token: test-token
      mirrors:
        testmirror:
          provider_type: github
          domain: github.com
          owner: testowner
          owner_type: user
          auth:
            token: test-token
`
	configPath := filepath.Join(tmpDir, "gitprovidersync.yaml")
	err := os.WriteFile(configPath, []byte(validConfig), 0600)
	require.NoError(t, err)

	tests := []struct {
		name     string
		setFlags func(*cli.Command)
	}{
		{
			name: "dry-run enabled",
			setFlags: func(cmd *cli.Command) {
				_ = cmd.Set("dry-run", "true")
				_ = cmd.Set("config-file", configPath)
			},
		},
		{
			name: "force-push enabled",
			setFlags: func(cmd *cli.Command) {
				_ = cmd.Set("force-push", "true")
				_ = cmd.Set("config-file", configPath)
			},
		},
		{
			name: "with since",
			setFlags: func(cmd *cli.Command) {
				_ = cmd.Set("since", "2024-01-01")
				_ = cmd.Set("config-file", configPath)
			},
		},
		{
			name: "sanitize-names enabled",
			setFlags: func(cmd *cli.Command) {
				_ = cmd.Set("sanitize-names", "true")
				_ = cmd.Set("config-file", configPath)
			},
		},
		{
			name: "skip-invalid enabled",
			setFlags: func(cmd *cli.Command) {
				_ = cmd.Set("skip-invalid", "true")
				_ = cmd.Set("config-file", configPath)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dry-run"},
					&cli.BoolFlag{Name: "force-push"},
					&cli.StringFlag{Name: "since"},
					&cli.BoolFlag{Name: "sanitize-names"},
					&cli.BoolFlag{Name: "skip-invalid"},
					&cli.StringFlag{Name: "config-file"},
					&cli.BoolFlag{Name: "config-file-only"},
					&cli.StringFlag{Name: "format"},
					&cli.IntFlag{Name: "verbosity-with-caller"},
					&cli.BoolFlag{Name: "quiet"},
				},
			}

			testCase.setFlags(cmd)

			// This will likely fail due to missing dependencies in test environment
			// but should fail gracefully without panicking
			err := runSync(ctx, cmd)
			if err != nil {
				t.Logf("Expected error in test environment: %v", err)
			}
			// The test passes if it doesn't panic
		})
	}
}

func TestMergeSyncOptionsWithCLIConfig_WithValues(t *testing.T) {
	t.Parallel()

	// Create a base CLI config
	baseConfig := entities.NewCLIConfigBuilder().
		WithOutputFormat("console").
		WithConfigFilePath("/test/config.yaml").
		Build()

	// Test with actual flag values set
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "dry-run"},
			&cli.BoolFlag{Name: "force-push"},
			&cli.StringFlag{Name: "since"},
			&cli.BoolFlag{Name: "sanitize-names"},
			&cli.BoolFlag{Name: "skip-invalid"},
		},
	}

	// Set some values
	_ = cmd.Set("dry-run", "true")
	_ = cmd.Set("since", "2024-06-01")
	_ = cmd.Set("sanitize-names", "true")

	result := mergeSyncOptionsWithCLIConfig(baseConfig, cmd)

	assert.NotNil(t, result)

	// Base config values should be preserved
	assert.Equal(t, "console", result.OutputFormat())
	assert.Equal(t, "/test/config.yaml", result.ConfigFilePath())

	// Sync flag values should be merged
	assert.True(t, result.DryRun())
	assert.False(t, result.ForcePush()) // Not set
	assert.Equal(t, "2024-06-01", result.ActiveFromLimit())
	assert.True(t, result.AlphaNumHyphName())
	assert.False(t, result.IgnoreInvalidName()) // Not set
}
