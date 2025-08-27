// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package baseoption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const jsonFormat = "json"

const (
	outputFormatFlag = "format"
	plainFormat      = "plain"
)

// Helper functions to create test commands.
func createDefaultCommand() *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Value: ""},
			&cli.BoolFlag{Name: "config-file-only", Value: false},
			&cli.BoolFlag{Name: "log-caller", Value: false},
			&cli.StringFlag{Name: "format", Value: "console"},
			&cli.BoolFlag{Name: plainFormat, Value: false},
		},
	}
}

func createAllFlagsSetCommand() *cli.Command {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Value: ""},
			&cli.BoolFlag{Name: "config-file-only", Value: false},
			&cli.BoolFlag{Name: "log-caller", Value: false},
			&cli.StringFlag{Name: outputFormatFlag, Value: jsonFormat},
			&cli.BoolFlag{Name: plainFormat, Value: false},
		},
	}
	// Simulate flag values being set
	for _, flag := range cmd.Flags {
		switch currentFlag := flag.(type) {
		case *cli.StringFlag:
			switch currentFlag.Name {
			case "config-file":
				currentFlag.Value = "/path/to/config.yaml"
			case outputFormatFlag:
				currentFlag.Value = jsonFormat
			}
		case *cli.BoolFlag:
			if currentFlag.Name == "config-file-only" || currentFlag.Name == "log-caller" {
				currentFlag.Value = true
			}
		}
	}

	return cmd
}

func createPlainOverrideCommand() *cli.Command {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config-file", Value: ""},
			&cli.BoolFlag{Name: "config-file-only", Value: false},
			&cli.BoolFlag{Name: "log-caller", Value: false},
			&cli.StringFlag{Name: outputFormatFlag, Value: ""}, // No explicit format
			&cli.BoolFlag{Name: plainFormat, Value: true},
			&cli.BoolFlag{Name: "json", Value: true}, // Both json and plain set, plain should win
		},
	}
	// Simulate plain flag being set
	for _, flag := range cmd.Flags {
		switch currentFlag := flag.(type) {
		case *cli.StringFlag:
			if currentFlag.Name == outputFormatFlag {
				currentFlag.Value = "" // Keep empty so flags can work
			}
		case *cli.BoolFlag:
			switch currentFlag.Name {
			case plainFormat:
				currentFlag.Value = true
			case "json":
				currentFlag.Value = true // Both set, plain should win
			}
		}
	}

	return cmd
}

func TestExtractRootInputOptions_FromCLIFlags_ReturnsCorrectOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		setupCmd           func() *cli.Command
		expectedConfigPath string
		expectedFileOnly   bool
		expectedCaller     bool
		expectedFormat     string
		expectError        bool
	}{
		{
			name:               "default values",
			setupCmd:           createDefaultCommand,
			expectedConfigPath: "",
			expectedFileOnly:   false,
			expectedCaller:     false,
			expectedFormat:     "console",
			expectError:        false,
		},
		{
			name:               "all flags set",
			setupCmd:           createAllFlagsSetCommand,
			expectedConfigPath: "/path/to/config.yaml",
			expectedFileOnly:   true,
			expectedCaller:     true,
			expectedFormat:     jsonFormat,
			expectError:        false,
		},
		{
			name:               "plain flag overrides output format",
			setupCmd:           createPlainOverrideCommand,
			expectedConfigPath: "",
			expectedFileOnly:   false,
			expectedCaller:     false,
			expectedFormat:     "plain", // Should override json
			expectError:        false,
		},
		{
			name: "custom config path",
			setupCmd: func() *cli.Command {
				cmd := &cli.Command{
					Flags: []cli.Flag{
						&cli.StringFlag{Name: "config-file", Value: "custom.yaml"},
						&cli.BoolFlag{Name: "config-file-only", Value: false},
						&cli.BoolFlag{Name: "log-caller", Value: false},
						&cli.StringFlag{Name: "format", Value: "console"},
						&cli.BoolFlag{Name: plainFormat, Value: false},
					},
				}
				for _, flag := range cmd.Flags {
					if f, ok := flag.(*cli.StringFlag); ok && f.Name == "config-file" {
						f.Value = "custom.yaml"
					}
				}

				return cmd
			},
			expectedConfigPath: "custom.yaml",
			expectedFileOnly:   false,
			expectedCaller:     false,
			expectedFormat:     "console",
			expectError:        false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := testCase.setupCmd()
			config, err := ExtractRootInputOptions(cmd)

			if testCase.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedConfigPath, config.ConfigFilePath())
			assert.Equal(t, testCase.expectedFileOnly, config.ConfigFileOnly())
			assert.Equal(t, testCase.expectedCaller, config.VerbosityWithCaller())
			assert.Equal(t, testCase.expectedFormat, config.OutputFormat())
		})
	}
}

func TestExtractRootOptions_NoFlags(t *testing.T) {
	t.Parallel()

	// Test behavior when command has no flags (edge case)
	cmd := &cli.Command{
		Name:  "test",
		Flags: []cli.Flag{},
	}

	config, err := ExtractRootInputOptions(cmd)

	// Should not error, but will use empty/default values
	require.NoError(t, err)
	assert.Empty(t, config.ConfigFilePath())
	assert.False(t, config.ConfigFileOnly())
	assert.False(t, config.VerbosityWithCaller())
	// Output format will be "plain" since terminal.IsOutput() returns false in tests
	assert.Equal(t, "plain", config.OutputFormat())
}

func TestExtractRootOptions_OutputFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		outputFormat   string
		plainFlag      bool
		expectedFormat string
	}{
		{
			name:           "console format",
			outputFormat:   "console",
			plainFlag:      false,
			expectedFormat: "console",
		},
		{
			name:           "json format",
			outputFormat:   jsonFormat,
			plainFlag:      false,
			expectedFormat: jsonFormat,
		},
		{
			name:           "plain format via flag",
			outputFormat:   "", // No explicit format, plain flag should take effect
			plainFlag:      true,
			expectedFormat: plainFormat,
		},
		{
			name:           "plain format explicit",
			outputFormat:   plainFormat,
			plainFlag:      false,
			expectedFormat: plainFormat,
		},
		{
			name:           "empty format",
			outputFormat:   "",
			plainFlag:      false,
			expectedFormat: "plain", // Default when not TTY
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config-file", Value: ""},
					&cli.BoolFlag{Name: "config-file-only", Value: false},
					&cli.BoolFlag{Name: "log-caller", Value: false},
					&cli.StringFlag{Name: outputFormatFlag, Value: testCase.outputFormat},
					&cli.BoolFlag{Name: plainFormat, Value: testCase.plainFlag},
					&cli.BoolFlag{Name: "json", Value: false},
				},
			}

			// Set flag values
			for _, flag := range cmd.Flags {
				switch currentFlag := flag.(type) {
				case *cli.StringFlag:
					if currentFlag.Name == outputFormatFlag {
						currentFlag.Value = testCase.outputFormat
					}
				case *cli.BoolFlag:
					if currentFlag.Name == plainFormat {
						currentFlag.Value = testCase.plainFlag
					}
				}
			}

			config, err := ExtractRootInputOptions(cmd)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFormat, config.OutputFormat())
		})
	}
}

func TestExtractRootOptions_FlagCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		configFile          string
		configFileOnly      bool
		verbosityWithCaller bool
		outputFormat        string
		plain               bool
	}{
		{
			name:                "minimal config",
			configFile:          "",
			configFileOnly:      false,
			verbosityWithCaller: false,
			outputFormat:        "console",
			plain:               false,
		},
		{
			name:                "file only mode",
			configFile:          "config.yaml",
			configFileOnly:      true,
			verbosityWithCaller: false,
			outputFormat:        "console",
			plain:               false,
		},
		{
			name:                "verbose mode",
			configFile:          "",
			configFileOnly:      false,
			verbosityWithCaller: true,
			outputFormat:        "console",
			plain:               false,
		},
		{
			name:                "all options enabled",
			configFile:          "/full/path/config.yaml",
			configFileOnly:      true,
			verbosityWithCaller: true,
			outputFormat:        jsonFormat,
			plain:               false,
		},
		{
			name:                "plain overrides json",
			configFile:          "test.yaml",
			configFileOnly:      false,
			verbosityWithCaller: true,
			outputFormat:        "", // No explicit format so plain flag can take effect
			plain:               true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config-file", Value: testCase.configFile},
					&cli.BoolFlag{Name: "config-file-only", Value: testCase.configFileOnly},
					&cli.BoolFlag{Name: "log-caller", Value: testCase.verbosityWithCaller},
					&cli.StringFlag{Name: outputFormatFlag, Value: testCase.outputFormat},
					&cli.BoolFlag{Name: plainFormat, Value: testCase.plain},
				},
			}

			config, err := ExtractRootInputOptions(cmd)
			require.NoError(t, err)

			assert.Equal(t, testCase.configFile, config.ConfigFilePath())
			assert.Equal(t, testCase.configFileOnly, config.ConfigFileOnly())
			assert.Equal(t, testCase.verbosityWithCaller, config.VerbosityWithCaller())

			expectedFormat := testCase.outputFormat
			if testCase.plain {
				expectedFormat = plainFormat
			}

			assert.Equal(t, expectedFormat, config.OutputFormat())
		})
	}
}

func TestExtractRootInputOptions_JSONFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		outputFormat   string
		jsonFlag       bool
		plainFlag      bool
		expectedFormat string
	}{
		{
			name:           "json flag sets format to json",
			outputFormat:   "",
			jsonFlag:       true,
			plainFlag:      false,
			expectedFormat: jsonFormat,
		},
		{
			name:           "explicit format overrides json flag",
			outputFormat:   "plain",
			jsonFlag:       true,
			plainFlag:      false,
			expectedFormat: "plain",
		},
		{
			name:           "plain flag takes precedence over json when no explicit format",
			outputFormat:   "",
			jsonFlag:       true,
			plainFlag:      true,
			expectedFormat: "plain",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config-file", Value: ""},
					&cli.BoolFlag{Name: "config-file-only", Value: false},
					&cli.BoolFlag{Name: "log-caller", Value: false},
					&cli.StringFlag{Name: "format", Value: testCase.outputFormat},
					&cli.BoolFlag{Name: "plain", Value: testCase.plainFlag},
					&cli.BoolFlag{Name: jsonFormat, Value: testCase.jsonFlag},
				},
			}

			config, err := ExtractRootInputOptions(cmd)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFormat, config.OutputFormat())
		})
	}
}
