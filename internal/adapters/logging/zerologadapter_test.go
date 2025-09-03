// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestNewZerologAdapter(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(nil)
	adapter := NewZerologAdapter(&logger)

	assert.NotNil(t, adapter)
	// Verify it implements the interface
	var _ = adapter
}

func TestZerologAdapterLogLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		logFunc func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]any)
		level   string
	}{
		{
			name: "info level",
			logFunc: func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]any) {
				adapter.Info(ctx, msg, fields)
			},
			level: "info",
		},
		{
			name: "error level",
			logFunc: func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]any) {
				adapter.Error(ctx, msg, fields)
			},
			level: "error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := zerolog.New(&buf).With().Timestamp().Logger()
			adapter := NewZerologAdapter(&logger)
			ctx := context.Background()

			testCase.logFunc(adapter, ctx, "test message", nil)

			output := buf.String()

			var logEntry map[string]any

			err := json.Unmarshal([]byte(output), &logEntry)
			require.NoError(t, err)

			assert.Equal(t, testCase.level, logEntry["level"])
			assert.Equal(t, "test message", logEntry["message"])
		})
	}
}

func TestZerologAdapterWithFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := zerolog.New(&buf)
	adapter := NewZerologAdapter(&logger)
	ctx := context.Background()

	fields := map[string]any{
		"component": "test",
		"count":     42,
	}

	adapter.Info(ctx, "test message", fields)

	output := buf.String()

	var logEntry map[string]any

	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test", logEntry["component"])
	assert.InEpsilon(t, float64(42), logEntry["count"], 0.001)
}
