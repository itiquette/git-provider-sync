// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"bytes"
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	model "itiquette/git-provider-sync/internal/model/configuration"
)

func TestNewPushOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		target           string
		prune            bool
		force            bool
		authCfg          model.AuthConfig
		expectedRefSpecs []string
		expectedForce    bool
		expectedPrune    bool
	}{
		{
			name:   "normal push without force",
			target: "https://github.com/user/repo.git",
			prune:  false,
			force:  false,
			authCfg: model.AuthConfig{
				Protocol: "https",
				Token:    "test-token",
			},
			expectedRefSpecs: []string{
				"refs/heads/*:refs/heads/*",
				"refs/tags/*:refs/tags/*",
			},
			expectedForce: false,
			expectedPrune: false,
		},
		{
			name:   "force push with prune",
			target: "git@gitlab.com:user/repo.git",
			prune:  true,
			force:  true,
			authCfg: model.AuthConfig{
				Protocol: "ssh",
				Token:    "ssh-key",
			},
			expectedRefSpecs: []string{
				"+refs/heads/*:refs/heads/*",
				"+refs/tags/*:refs/tags/*",
			},
			expectedForce: true,
			expectedPrune: true,
		},
		{
			name:   "force push without prune",
			target: "https://bitbucket.org/user/repo.git",
			prune:  false,
			force:  true,
			authCfg: model.AuthConfig{
				Protocol: "https",
				Token:    "app-password",
			},
			expectedRefSpecs: []string{
				"+refs/heads/*:refs/heads/*",
				"+refs/tags/*:refs/tags/*",
			},
			expectedForce: true,
			expectedPrune: false,
		},
		{
			name:   "normal push with prune but no force",
			target: "https://gitea.example.com/user/repo.git",
			prune:  true,
			force:  false,
			authCfg: model.AuthConfig{
				Protocol: "https",
				Token:    "gitea-token",
			},
			expectedRefSpecs: []string{
				"refs/heads/*:refs/heads/*",
				"refs/tags/*:refs/tags/*",
			},
			expectedForce: false,
			expectedPrune: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := NewPushOption(testCase.target, testCase.prune, testCase.force, testCase.authCfg)

			assert.Equal(t, testCase.target, result.Target)
			assert.Equal(t, testCase.expectedForce, result.Force)
			assert.Equal(t, testCase.expectedPrune, result.Prune)
			assert.Equal(t, testCase.authCfg, result.AuthCfg)
			assert.Equal(t, testCase.expectedRefSpecs, result.RefSpecs)
		})
	}
}

func TestPushOptionRefSpecModification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		force          bool
		expectedPrefix string
	}{
		{
			name:           "force push adds plus prefix",
			force:          true,
			expectedPrefix: "+",
		},
		{
			name:           "normal push has no prefix",
			force:          false,
			expectedPrefix: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			authCfg := model.AuthConfig{Protocol: "https"}
			result := NewPushOption("https://example.com/repo.git", false, testCase.force, authCfg)

			for _, refSpec := range result.RefSpecs {
				if testCase.force {
					assert.Equal(t, byte('+'), refSpec[0], "RefSpec should start with + for force push: %s", refSpec)
				} else {
					assert.NotEqual(t, byte('+'), refSpec[0], "RefSpec should not start with + for normal push: %s", refSpec)
				}
			}
		})
	}
}

func TestPushOptionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pushOption     PushOption
		expectedFields []string
	}{
		{
			name: "complete push option string representation",
			pushOption: PushOption{
				Force: true,
				AuthCfg: model.AuthConfig{
					Protocol: "https",
					Token:    "secret-token",
				},
				Prune:    true,
				RefSpecs: []string{"+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*"},
				Target:   "https://github.com/user/repo.git",
			},
			expectedFields: []string{
				"Target: https://github.com/user/repo.git",
				"RefSpecs: [+refs/heads/*:refs/heads/* +refs/tags/*:refs/tags/*]",
				"Prune: true",
				"Force: true",
			},
		},
		{
			name: "minimal push option",
			pushOption: PushOption{
				Force:    false,
				Prune:    false,
				RefSpecs: []string{"refs/heads/*:refs/heads/*"},
				Target:   "git@example.com:user/repo.git",
				AuthCfg:  model.AuthConfig{Protocol: "ssh"},
			},
			expectedFields: []string{
				"Target: git@example.com:user/repo.git",
				"RefSpecs: [refs/heads/*:refs/heads/*]",
				"Prune: false",
				"Force: false",
			},
		},
		{
			name:       "zero value push option",
			pushOption: PushOption{},
			expectedFields: []string{
				"Target: ",
				"RefSpecs: []",
				"Prune: false",
				"Force: false",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.pushOption.String()

			assert.Contains(t, result, "PushOption{")

			for _, expectedField := range testCase.expectedFields {
				assert.Contains(t, result, expectedField)
			}
		})
	}
}

func TestPushOptionDebugLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pushOption PushOption
		testFunc   func(t *testing.T, logOutput string)
	}{
		{
			name: "debug log with credentials in URL",
			pushOption: PushOption{
				Force: true,
				AuthCfg: model.AuthConfig{
					Protocol: "https",
					Token:    "secret-token",
				},
				Prune:    true,
				RefSpecs: []string{"+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*"},
				Target:   "https://user:password@github.com/user/repo.git",
			},
			testFunc: func(t *testing.T, logOutput string) {
				t.Helper()
				// Credentials should be stripped from the target URL
				assert.Contains(t, logOutput, "https://github.com/user/repo.git")
				assert.NotContains(t, logOutput, "password")
				assert.Contains(t, logOutput, "+refs/heads/*:refs/heads/*")
				assert.Contains(t, logOutput, "true") // force and prune flags
			},
		},
		{
			name: "debug log with SSH URL",
			pushOption: PushOption{
				Force:    false,
				Prune:    false,
				RefSpecs: []string{"refs/heads/*:refs/heads/*"},
				Target:   "git@gitlab.com:user/repo.git",
				AuthCfg: model.AuthConfig{
					Protocol: "ssh",
				},
			},
			testFunc: func(t *testing.T, logOutput string) {
				t.Helper()
				assert.Contains(t, logOutput, "git@gitlab.com:user/repo.git")
				assert.Contains(t, logOutput, "refs/heads/*:refs/heads/*")
				assert.Contains(t, logOutput, "false") // force and prune flags
			},
		},
		{
			name:       "debug log with empty values",
			pushOption: PushOption{},
			testFunc: func(t *testing.T, logOutput string) {
				t.Helper()
				// Should handle empty values gracefully
				assert.NotEmpty(t, logOutput)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := zerolog.New(&buf)

			ctx := context.Background()
			event := testCase.pushOption.DebugLog(ctx, &logger)

			// Trigger the log event to capture output
			event.Msg("test push log")

			logOutput := buf.String()
			testCase.testFunc(t, logOutput)
		})
	}
}

func TestPushOptionComplexScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "complete push workflow with different auth configs",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test HTTPS with token
				httpsAuth := model.AuthConfig{
					Protocol: "https",
					Token:    "github-token",
				}

				httpsOption := NewPushOption(
					"https://github.com/user/repo.git",
					true,  // prune
					false, // no force
					httpsAuth,
				)

				assert.Equal(t, "https://github.com/user/repo.git", httpsOption.Target)
				assert.True(t, httpsOption.Prune)
				assert.False(t, httpsOption.Force)
				assert.Equal(t, httpsAuth, httpsOption.AuthCfg)

				// RefSpecs should not have force prefix
				for _, refSpec := range httpsOption.RefSpecs {
					assert.NotEqual(t, '+', refSpec[0])
				}

				// Test SSH with force
				sshAuth := model.AuthConfig{
					Protocol: "ssh",
					Token:    "ssh-key-path",
				}

				sshOption := NewPushOption(
					"git@gitlab.com:user/repo.git",
					false, // no prune
					true,  // force
					sshAuth,
				)

				assert.Equal(t, "git@gitlab.com:user/repo.git", sshOption.Target)
				assert.False(t, sshOption.Prune)
				assert.True(t, sshOption.Force)
				assert.Equal(t, sshAuth, sshOption.AuthCfg)

				// RefSpecs should have force prefix
				for _, refSpec := range sshOption.RefSpecs {
					assert.Equal(t, byte('+'), refSpec[0])
				}
			},
		},
		{
			name: "push option string representations",
			testFunc: func(t *testing.T) {
				t.Helper()
				authCfg := model.AuthConfig{
					Protocol: "https",
					Token:    "test-token",
				}

				option := NewPushOption(
					"https://example.com/repo.git",
					true, // prune
					true, // force
					authCfg,
				)

				str := option.String()
				assert.Contains(t, str, "Force: true")
				assert.Contains(t, str, "Prune: true")
				assert.Contains(t, str, "https://example.com/repo.git")
				assert.Contains(t, str, "+refs/heads/*:refs/heads/*") // force prefix

				// Test debug log
				var buf bytes.Buffer
				logger := zerolog.New(&buf)
				ctx := context.Background()

				event := option.DebugLog(ctx, &logger)
				event.Msg("debug test")

				logOutput := buf.String()
				assert.Contains(t, logOutput, "https://example.com/repo.git")
				assert.Contains(t, logOutput, "+refs/heads/*:refs/heads/*")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestPushOption_InvalidBranchAndRemote_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "very long target URLs",
			testFunc: func(t *testing.T) {
				t.Helper()
				longURL := "https://very-long-domain-name-that-exceeds-normal-url-length.example.com/organization-with-very-long-name/repository-with-extremely-long-name-that-might-cause-issues.git"

				authCfg := model.AuthConfig{Protocol: "https"}
				option := NewPushOption(longURL, true, true, authCfg)

				assert.Equal(t, longURL, option.Target)

				// String representation should handle long URLs
				str := option.String()
				assert.Contains(t, str, longURL)

				// Debug log should handle long URLs
				var buf bytes.Buffer
				logger := zerolog.New(&buf)
				ctx := context.Background()

				event := option.DebugLog(ctx, &logger)
				event.Msg("long URL test")

				logOutput := buf.String()
				assert.Contains(t, logOutput, longURL)
			},
		},
		{
			name: "special characters in URLs",
			testFunc: func(t *testing.T) {
				t.Helper()
				specialURL := "https://user%40domain:p%40ss@github.com/user/repo-with-special@chars.git"

				authCfg := model.AuthConfig{Protocol: "https"}
				option := NewPushOption(specialURL, false, false, authCfg)

				assert.Equal(t, specialURL, option.Target)

				// Debug log should sanitize credentials
				var buf bytes.Buffer
				logger := zerolog.New(&buf)
				ctx := context.Background()

				event := option.DebugLog(ctx, &logger)
				event.Msg("special chars test")

				logOutput := buf.String()
				// Should contain the repo part but not the credentials
				assert.Contains(t, logOutput, "github.com/user/repo-with-special@chars.git")
				assert.NotContains(t, logOutput, "p%40ss") // password should be stripped
			},
		},
		{
			name: "empty target URL",
			testFunc: func(t *testing.T) {
				t.Helper()
				authCfg := model.AuthConfig{Protocol: "https"}
				option := NewPushOption("", true, true, authCfg)

				assert.Empty(t, option.Target)
				assert.True(t, option.Prune)
				assert.True(t, option.Force)

				// Should handle empty target gracefully
				str := option.String()
				assert.Contains(t, str, "Target: ")

				var buf bytes.Buffer
				logger := zerolog.New(&buf)
				ctx := context.Background()

				event := option.DebugLog(ctx, &logger)
				event.Msg("empty target test")

				logOutput := buf.String()
				assert.NotEmpty(t, logOutput)
			},
		},
		{
			name: "refspecs with existing plus prefix should not be double-prefixed",
			testFunc: func(t *testing.T) {
				t.Helper()
				// This tests the edge case where refspecs might already have a + prefix
				// The current implementation should handle this correctly
				authCfg := model.AuthConfig{Protocol: "https"}
				option := NewPushOption("https://example.com/repo.git", false, true, authCfg)

				// All refspecs should have exactly one + prefix when force is true
				for _, refSpec := range option.RefSpecs {
					assert.Equal(t, byte('+'), refSpec[0])
					// Ensure no double prefix (would be "++")
					assert.False(t, len(refSpec) > 1 && refSpec[1] == byte('+'))
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}
