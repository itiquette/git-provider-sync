// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/configuration/dto"
)

// TestFormatConfiguration_InvalidFormat tests error handling for unsupported formats.
func TestFormatConfiguration_InvalidFormat(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	var buffer bytes.Buffer

	err := formatter.FormatConfiguration(dto.AppConfiguration{}, "invalid", &buffer)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedOutputFormat)
}

// TestFormatConfiguration_JSONValidity ensures JSON output is valid JSON.
func TestFormatConfiguration_JSONValidity(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	config := dto.AppConfiguration{
		GitProviderSyncConfs: map[string]dto.Environment{
			"test": {
				"source": dto.SyncConfig{
					BaseConfig: dto.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testowner",
						OwnerType:    "user",
						Auth: dto.AuthConfig{
							Token: "secret-token",
						},
					},
					Mirrors: map[string]dto.MirrorConfig{
						"mirror1": {
							BaseConfig: dto.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.com",
								Owner:        "backup",
							},
						},
					},
				},
			},
		},
	}

	var buffer bytes.Buffer

	err := formatter.FormatConfiguration(config, "json", &buffer)
	require.NoError(t, err)

	// Verify it's valid JSON
	var result map[string]interface{}

	err = json.Unmarshal(buffer.Bytes(), &result)
	require.NoError(t, err)
	assert.NotNil(t, result["gitprovidersync"])
}

// TestFormatSyncResults_NilResults tests nil handling.
func TestFormatSyncResults_NilResults(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	var buffer bytes.Buffer

	err := formatter.FormatSyncResults(nil, "console", &buffer, &buffer)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSyncResultsType)
}

// TestFormatConfiguration_WriterError tests error propagation from writer.
func TestFormatConfiguration_WriterError(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	config := dto.AppConfiguration{
		GitProviderSyncConfs: map[string]dto.Environment{
			"test": {
				"source": dto.SyncConfig{
					BaseConfig: dto.BaseConfig{
						ProviderType: "github",
					},
				},
			},
		},
	}

	// Create a writer that always errors
	errorWriter := &erroringWriter{err: errors.New("write failed")}

	err := formatter.FormatConfiguration(config, "console", errorWriter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

// ErroringWriter is a test helper that always returns an error on Write.
type erroringWriter struct {
	err error
}

func (w *erroringWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}
