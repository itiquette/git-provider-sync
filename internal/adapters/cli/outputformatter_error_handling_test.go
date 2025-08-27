// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/sync"
	model "itiquette/git-provider-sync/internal/model/configuration"
)

// errorWriter is a writer that always returns an error.
type errorWriter struct{}

func (w errorWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestOutputFormatter_ErrorHandling(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	t.Run("writeSyncConfigHeader error handling", func(t *testing.T) {
		t.Parallel()

		err := formatter.writeSyncConfigHeader(errorWriter{}, "  ", "test-config")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write sync config name")
	})

	t.Run("writeMirrorConfigHeader error handling", func(t *testing.T) {
		t.Parallel()

		err := formatter.writeMirrorConfigHeader(errorWriter{}, "  ", "test-mirror")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write mirror name")
	})

	t.Run("writeMirrorSettingsHeader error handling", func(t *testing.T) {
		t.Parallel()

		err := formatter.writeMirrorSettingsHeader(errorWriter{}, "  ")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write settings header")
	})

	t.Run("writeSyncProgress error handling", func(t *testing.T) {
		t.Parallel()

		syncResults := &sync.Results{
			DurationSeconds: 45.67,
		}

		err := formatter.writeSyncProgress(syncResults, errorWriter{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write progress")
	})

	t.Run("writeSyncSummaryHeader error handling", func(t *testing.T) {
		t.Parallel()

		err := formatter.writeSyncSummaryHeader(errorWriter{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write summary header")
	})

	t.Run("writeSyncDetailedResultsHeader error handling", func(t *testing.T) {
		t.Parallel()

		err := formatter.writeSyncDetailedResultsHeader(errorWriter{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write detailed results header")
	})

	t.Run("writeSyncDetailedResult error handling", func(t *testing.T) {
		t.Parallel()

		result := sync.Result{
			Environment:     "test",
			Source:          "source",
			Mirror:          "mirror",
			Repository:      "repo",
			Action:          "sync",
			Status:          "success",
			DurationSeconds: 1.23,
		}

		err := formatter.writeSyncDetailedResult(result, errorWriter{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write detailed result")
	})

	t.Run("writeAuthHeader error handling", func(t *testing.T) {
		t.Parallel()

		err := formatter.writeAuthHeader(errorWriter{}, "  ")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write authentication header")
	})

	t.Run("writeAuthMandatoryFields error handling", func(t *testing.T) {
		t.Parallel()

		authCfg := model.AuthConfig{
			Protocol: "https",
		}

		err := formatter.writeAuthMandatoryFields(errorWriter{}, "  ", authCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write protocol")
	})

	t.Run("writeSSHHeader error handling", func(t *testing.T) {
		t.Parallel()

		err := formatter.writeSSHHeader(errorWriter{}, "  ")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write SSH configuration header")
	})
}

func TestOutputFormatter_ComplexMirrorConfig(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	t.Run("writeMirrorConfigMandatoryFields with all fields", func(t *testing.T) {
		t.Parallel()

		mirrorCfg := model.MirrorConfig{
			BaseConfig: model.BaseConfig{
				ProviderType: "gitea",
				Domain:       "gitea.example.com",
				Owner:        "testorg",
				OwnerType:    "organization",
			},
		}

		var buffer bytes.Buffer

		err := formatter.writeMirrorConfigMandatoryFields(&buffer, "  ", mirrorCfg)

		require.NoError(t, err)

		output := buffer.String()

		assert.Contains(t, output, "Type: gitea")
		assert.Contains(t, output, "Domain: gitea.example.com")
		assert.Contains(t, output, "Owner: testorg")
		assert.Contains(t, output, "Owner Type: organization")
	})

	t.Run("writeMirrorConfigMandatoryFields with empty domain", func(t *testing.T) {
		t.Parallel()

		mirrorCfg := model.MirrorConfig{
			BaseConfig: model.BaseConfig{
				ProviderType: "github",
				Domain:       "",
				Owner:        "",
				OwnerType:    "user",
			},
		}

		var buffer bytes.Buffer

		err := formatter.writeMirrorConfigMandatoryFields(&buffer, "  ", mirrorCfg)

		require.NoError(t, err)

		output := buffer.String()

		assert.Contains(t, output, "Type: github")
		assert.Contains(t, output, "Owner Type: user")
		// Should not contain domain or owner when empty
		assert.NotContains(t, output, "Domain:")
		assert.NotContains(t, output, "Owner:")
	})

	t.Run("printAuthConfig complete", func(t *testing.T) {
		t.Parallel()

		authCfg := model.AuthConfig{
			Protocol:          "https",
			HTTPScheme:        "https",
			Token:             "secret-token",
			ProxyURL:          "http://proxy.com",
			CertDirPath:       "/certs",
			SSHCommand:        "ssh -i key",
			SSHURLRewriteFrom: "git@github.com:",
			SSHURLRewriteTo:   "https://github.com/",
		}

		var buffer bytes.Buffer

		err := formatter.printAuthConfig(authCfg, &buffer, 1, 2)

		require.NoError(t, err)

		output := buffer.String()

		assert.Contains(t, output, "Authentication:")
		assert.Contains(t, output, "Protocol: https")
		assert.Contains(t, output, "HTTP Scheme: https")
		assert.Contains(t, output, "Token: <*****>")
		assert.Contains(t, output, "Proxy URL: http://proxy.com")
		assert.Contains(t, output, "Certificate Directory: /certs")
		assert.Contains(t, output, "SSH Configuration:")
		assert.Contains(t, output, "Command: ssh -i key")
		assert.Contains(t, output, "URL Rewrite From: git@github.com:")
		assert.Contains(t, output, "URL Rewrite To: https://github.com/")
	})

	t.Run("printMirrorConfig complete", func(t *testing.T) {
		t.Parallel()

		mirrorCfg := model.MirrorConfig{
			BaseConfig: model.BaseConfig{
				ProviderType: "gitlab",
				Domain:       "gitlab.example.com",
				Owner:        "myorg",
				OwnerType:    "organization",
				UseGitBinary: true,
				Auth: model.AuthConfig{
					Protocol: "https",
					Token:    "secret",
				},
			},
			Path: "/custom/path",
			Settings: model.MirrorSettings{
				AlphaNumHyphName:  true,
				DescriptionPrefix: "Mirror: ",
				Disabled:          false,
				ForcePush:         true,
			},
		}

		var buffer bytes.Buffer

		err := formatter.printMirrorConfig("test-mirror", mirrorCfg, &buffer, 1, 2)

		require.NoError(t, err)

		output := buffer.String()

		assert.Contains(t, output, "Mirror: test-mirror")
		assert.Contains(t, output, "Type: gitlab")
		assert.Contains(t, output, "Domain: gitlab.example.com")
		assert.Contains(t, output, "Owner: myorg")
		assert.Contains(t, output, "Owner Type: organization")
		assert.Contains(t, output, "Use Git Binary: true")
		assert.Contains(t, output, "Path: /custom/path")
		assert.Contains(t, output, "Settings:")
		assert.Contains(t, output, "ASCII Name: true")
		assert.Contains(t, output, "Description Prefix: Mirror: ")
		assert.Contains(t, output, "Force Push: true")
		assert.Contains(t, output, "Authentication:")
		assert.Contains(t, output, "Protocol: https")
		assert.Contains(t, output, "Token: <*****>")
	})
}

func TestOutputFormatter_SyncDetailedResults(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	t.Run("writeSyncDetailedResults with results", func(t *testing.T) {
		t.Parallel()

		syncResults := &sync.Results{
			Results: []sync.Result{
				{
					Environment:     "prod",
					Source:          "github-source",
					SourceProvider:  "github",
					Repository:      "main-repo",
					Mirror:          "backup",
					MirrorProvider:  "gitea",
					Status:          "SUCCESS",
					Action:          "UPDATED",
					DurationSeconds: 2.34,
					Error:           "",
				},
				{
					Environment:     "prod",
					Source:          "github-source",
					SourceProvider:  "github",
					Repository:      "test-repo",
					Mirror:          "backup",
					MirrorProvider:  "gitea",
					Status:          "FAILED",
					Action:          "SKIPPED",
					DurationSeconds: 0.15,
					Error:           "connection timeout",
				},
			},
		}

		var buffer bytes.Buffer

		err := formatter.writeSyncDetailedResults(syncResults, &buffer)

		require.NoError(t, err)

		output := buffer.String()

		assert.Contains(t, output, "To mirrors:")
		assert.Contains(t, output, "main-repo -> backup:")
		assert.Contains(t, output, "test-repo -> backup:")
	})

	t.Run("writeSyncDetailedResults empty results", func(t *testing.T) {
		t.Parallel()

		syncResults := &sync.Results{
			Results: []sync.Result{},
		}

		var buffer bytes.Buffer

		err := formatter.writeSyncDetailedResults(syncResults, &buffer)

		require.NoError(t, err)

		output := buffer.String()

		// Should be empty for no results
		assert.Empty(t, output)
	})
}

func TestOutputFormatter_WriteSSHFields(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	tests := []struct {
		name         string
		authCfg      model.AuthConfig
		expectedText []string
		notExpected  []string
	}{
		{
			name: "all SSH fields set",
			authCfg: model.AuthConfig{
				SSHCommand:        "ssh -i /path/to/key",
				SSHURLRewriteFrom: "git@github.com:",
				SSHURLRewriteTo:   "https://github.com/",
			},
			expectedText: []string{
				"Command: ssh -i /path/to/key",
				"URL Rewrite From: git@github.com:",
				"URL Rewrite To: https://github.com/",
			},
		},
		{
			name: "only SSH command set",
			authCfg: model.AuthConfig{
				SSHCommand: "ssh -o StrictHostKeyChecking=no",
			},
			expectedText: []string{
				"Command: ssh -o StrictHostKeyChecking=no",
			},
			notExpected: []string{
				"URL Rewrite From:",
				"URL Rewrite To:",
			},
		},
		{
			name: "only URL rewrite fields set",
			authCfg: model.AuthConfig{
				SSHURLRewriteFrom: "git@gitlab.com:",
				SSHURLRewriteTo:   "https://gitlab.com/",
			},
			expectedText: []string{
				"URL Rewrite From: git@gitlab.com:",
				"URL Rewrite To: https://gitlab.com/",
			},
			notExpected: []string{
				"Command:",
			},
		},
		{
			name:         "no SSH fields set",
			authCfg:      model.AuthConfig{},
			expectedText: []string{},
			notExpected: []string{
				"Command:",
				"URL Rewrite From:",
				"URL Rewrite To:",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			err := formatter.writeSSHFields(&buffer, "  ", test.authCfg)

			require.NoError(t, err)

			output := buffer.String()

			// Check that all expected text is present
			for _, expectedText := range test.expectedText {
				assert.Contains(t, output, expectedText, "Output should contain: %s", expectedText)
			}

			// Check that unexpected text is not present
			for _, notExpectedText := range test.notExpected {
				assert.NotContains(t, output, notExpectedText, "Output should not contain: %s", notExpectedText)
			}
		})
	}
}
