// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package baseoption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestExtractRootInputOptions_FormatPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFlags     func(*cli.Command)
		expectedFormat string
	}{
		{
			name: "explicit format takes precedence",
			setupFlags: func(cmd *cli.Command) {
				cmd.Flags = []cli.Flag{
					&cli.StringFlag{Name: "format", Value: "yaml"},
					&cli.BoolFlag{Name: "plain", Value: true},
					&cli.BoolFlag{Name: "json", Value: true},
				}
			},
			expectedFormat: "yaml",
		},
		{
			name: "plain flag when no explicit format",
			setupFlags: func(cmd *cli.Command) {
				cmd.Flags = []cli.Flag{
					&cli.StringFlag{Name: "format", Value: ""},
					&cli.BoolFlag{Name: "plain", Value: true},
					&cli.BoolFlag{Name: "json", Value: false},
				}
			},
			expectedFormat: "plain",
		},
		{
			name: "json flag when no plain or format",
			setupFlags: func(cmd *cli.Command) {
				cmd.Flags = []cli.Flag{
					&cli.StringFlag{Name: "format", Value: ""},
					&cli.BoolFlag{Name: "plain", Value: false},
					&cli.BoolFlag{Name: "json", Value: true},
				}
			},
			expectedFormat: "json",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config-file"},
					&cli.BoolFlag{Name: "config-file-only"},
					&cli.BoolFlag{Name: "log-caller"},
					&cli.StringFlag{Name: "color"},
				},
			}
			testCase.setupFlags(cmd)

			config, err := ExtractRootInputOptions(cmd)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFormat, config.OutputFormat())
		})
	}
}

func TestExtractRootInputOptions_LogLevelConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFlags  func(*cli.Command)
		expectError bool
	}{
		{
			name: "quiet and verbose conflict",
			setupFlags: func(cmd *cli.Command) {
				cmd.Flags = append(cmd.Flags,
					&cli.BoolFlag{Name: "quiet", Value: true},
					&cli.BoolFlag{Name: "verbose", Value: true},
				)
			},
			expectError: true,
		},
		{
			name: "debug and log-level conflict",
			setupFlags: func(cmd *cli.Command) {
				cmd.Flags = append(cmd.Flags,
					&cli.BoolFlag{Name: "debug", Value: true},
					&cli.StringFlag{Name: "log-level", Value: "info"},
				)
			},
			expectError: true,
		},
		{
			name: "single log flag is valid",
			setupFlags: func(cmd *cli.Command) {
				cmd.Flags = append(cmd.Flags,
					&cli.BoolFlag{Name: "verbose", Value: true},
				)
			},
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config-file"},
					&cli.BoolFlag{Name: "config-file-only"},
					&cli.BoolFlag{Name: "log-caller"},
					&cli.StringFlag{Name: "format"},
					&cli.BoolFlag{Name: "plain"},
					&cli.BoolFlag{Name: "json"},
					&cli.StringFlag{Name: "color"},
				},
			}
			testCase.setupFlags(cmd)

			_, err := ExtractRootInputOptions(cmd)
			if testCase.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "cannot use multiple log level flags")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
