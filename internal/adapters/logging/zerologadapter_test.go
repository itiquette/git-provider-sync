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
	assert.IsType(t, &ZerologAdapter{}, adapter)
}

func TestZerologAdapterLogLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		logFunc func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]interface{})
		level   string
		message string
		fields  map[string]interface{}
	}{
		{
			name: "trace level",
			logFunc: func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]interface{}) {
				adapter.Trace(ctx, msg, fields)
			},
			level:   "trace",
			message: "trace message",
			fields:  map[string]interface{}{"key": "value"},
		},
		{
			name: "debug level",
			logFunc: func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]interface{}) {
				adapter.Debug(ctx, msg, fields)
			},
			level:   "debug",
			message: "debug message",
			fields:  map[string]interface{}{"debug": true},
		},
		{
			name: "info level",
			logFunc: func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]interface{}) {
				adapter.Info(ctx, msg, fields)
			},
			level:   "info",
			message: "info message",
			fields:  map[string]interface{}{"count": float64(42)},
		},
		{
			name: "warn level",
			logFunc: func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]interface{}) {
				adapter.Warn(ctx, msg, fields)
			},
			level:   "warn",
			message: "warning message",
			fields:  map[string]interface{}{"warning": "test"},
		},
		{
			name: "error level",
			logFunc: func(adapter ports.Logger, ctx context.Context, msg string, fields map[string]interface{}) {
				adapter.Error(ctx, msg, fields)
			},
			level:   "error",
			message: "error message",
			fields:  map[string]interface{}{"error": "test error"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := zerolog.New(&buf).With().Timestamp().Logger()
			adapter := NewZerologAdapter(&logger)
			ctx := context.Background()

			testCase.logFunc(adapter, ctx, testCase.message, testCase.fields)

			output := buf.String()

			// Parse JSON log output
			var logEntry map[string]interface{}

			err := json.Unmarshal([]byte(output), &logEntry)
			require.NoError(t, err)

			assert.Equal(t, testCase.level, logEntry["level"])
			assert.Equal(t, testCase.message, logEntry["message"])

			// Check fields are present
			for key, expectedValue := range testCase.fields {
				assert.Equal(t, expectedValue, logEntry[key])
			}
		})
	}
}

func TestZerologAdapterWithField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "string field",
			key:   "component",
			value: "test-component",
		},
		{
			name:  "number field",
			key:   "count",
			value: float64(123),
		},
		{
			name:  "boolean field",
			key:   "enabled",
			value: true,
		},
		{
			name:  "nil field",
			key:   "nullable",
			value: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := zerolog.New(&buf)
			adapter, ok := NewZerologAdapter(&logger).(*ZerologAdapter)
			require.True(t, ok)

			newAdapter := adapter.WithField(testCase.key, testCase.value)
			assert.NotNil(t, newAdapter)
			assert.IsType(t, &ZerologAdapter{}, newAdapter)

			// Log with the new adapter to verify field is included
			newAdapter.Info(context.Background(), "test message", nil)

			output := buf.String()

			var logEntry map[string]interface{}

			err := json.Unmarshal([]byte(output), &logEntry)
			require.NoError(t, err)

			assert.Equal(t, "test message", logEntry["message"])
			assert.Equal(t, testCase.value, logEntry[testCase.key])
		})
	}
}

func TestZerologAdapterWithFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields map[string]interface{}
	}{
		{
			name: "single field",
			fields: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "multiple fields",
			fields: map[string]interface{}{
				"component": "test",
				"version":   "1.0.0",
				"debug":     true,
				"count":     float64(42),
			},
		},
		{
			name:   "empty fields",
			fields: map[string]interface{}{},
		},
		{
			name:   "nil fields",
			fields: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := zerolog.New(&buf)
			adapter, ok := NewZerologAdapter(&logger).(*ZerologAdapter)
			require.True(t, ok)

			newAdapter := adapter.WithFields(testCase.fields)
			assert.NotNil(t, newAdapter)
			assert.IsType(t, &ZerologAdapter{}, newAdapter)

			// Log with the new adapter to verify fields are included
			newAdapter.Info(context.Background(), "test message", nil)

			output := buf.String()

			var logEntry map[string]interface{}

			err := json.Unmarshal([]byte(output), &logEntry)
			require.NoError(t, err)

			assert.Equal(t, "test message", logEntry["message"])

			// Check all fields are present
			for key, expectedValue := range testCase.fields {
				assert.Equal(t, expectedValue, logEntry[key])
			}
		})
	}
}

func TestZerologAdapterLogLevelConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		domainLevel     ports.LogLevel
		expectedZerolog zerolog.Level
	}{
		{
			name:            "trace level",
			domainLevel:     ports.LogLevelTrace,
			expectedZerolog: zerolog.TraceLevel,
		},
		{
			name:            "debug level",
			domainLevel:     ports.LogLevelDebug,
			expectedZerolog: zerolog.DebugLevel,
		},
		{
			name:            "info level",
			domainLevel:     ports.LogLevelInfo,
			expectedZerolog: zerolog.InfoLevel,
		},
		{
			name:            "warn level",
			domainLevel:     ports.LogLevelWarn,
			expectedZerolog: zerolog.WarnLevel,
		},
		{
			name:            "error level",
			domainLevel:     ports.LogLevelError,
			expectedZerolog: zerolog.ErrorLevel,
		},
		{
			name:            "fatal level",
			domainLevel:     ports.LogLevelFatal,
			expectedZerolog: zerolog.FatalLevel,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			adapter := &ZerologAdapter{}
			zerologLevel := adapter.convertLogLevel(testCase.domainLevel)
			assert.Equal(t, testCase.expectedZerolog, zerologLevel)
		})
	}
}

func TestZerologAdapterZerologLevelConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		zerologLevel   zerolog.Level
		expectedDomain ports.LogLevel
	}{
		{
			name:           "trace level",
			zerologLevel:   zerolog.TraceLevel,
			expectedDomain: ports.LogLevelTrace,
		},
		{
			name:           "debug level",
			zerologLevel:   zerolog.DebugLevel,
			expectedDomain: ports.LogLevelDebug,
		},
		{
			name:           "info level",
			zerologLevel:   zerolog.InfoLevel,
			expectedDomain: ports.LogLevelInfo,
		},
		{
			name:           "warn level",
			zerologLevel:   zerolog.WarnLevel,
			expectedDomain: ports.LogLevelWarn,
		},
		{
			name:           "error level",
			zerologLevel:   zerolog.ErrorLevel,
			expectedDomain: ports.LogLevelError,
		},
		{
			name:           "fatal level",
			zerologLevel:   zerolog.FatalLevel,
			expectedDomain: ports.LogLevelFatal,
		},
		{
			name:           "panic level",
			zerologLevel:   zerolog.PanicLevel,
			expectedDomain: ports.LogLevelFatal,
		},
		{
			name:           "no level",
			zerologLevel:   zerolog.NoLevel,
			expectedDomain: ports.LogLevelInfo,
		},
		{
			name:           "disabled level",
			zerologLevel:   zerolog.Disabled,
			expectedDomain: ports.LogLevelInfo,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			adapter := &ZerologAdapter{}
			domainLevel := adapter.convertZerologLevel(testCase.zerologLevel)
			assert.Equal(t, testCase.expectedDomain, domainLevel)
		})
	}
}

func TestZerologAdapterSetGetLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level ports.LogLevel
	}{
		{
			name:  "trace level",
			level: ports.LogLevelTrace,
		},
		{
			name:  "debug level",
			level: ports.LogLevelDebug,
		},
		{
			name:  "info level",
			level: ports.LogLevelInfo,
		},
		{
			name:  "warn level",
			level: ports.LogLevelWarn,
		},
		{
			name:  "error level",
			level: ports.LogLevelError,
		},
		{
			name:  "fatal level",
			level: ports.LogLevelFatal,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			logger := zerolog.New(nil)
			adapter, ok := NewZerologAdapter(&logger).(*ZerologAdapter)
			require.True(t, ok)

			adapter.SetLevel(testCase.level)
			retrievedLevel := adapter.GetLevel()

			assert.Equal(t, testCase.level, retrievedLevel)
		})
	}
}

func TestZerologAdapterIsLevelEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setLevel      ports.LogLevel
		checkLevel    ports.LogLevel
		expectEnabled bool
	}{
		{
			name:          "trace enabled for trace",
			setLevel:      ports.LogLevelTrace,
			checkLevel:    ports.LogLevelTrace,
			expectEnabled: true,
		},
		{
			name:          "info enabled for debug when set to debug",
			setLevel:      ports.LogLevelDebug,
			checkLevel:    ports.LogLevelInfo,
			expectEnabled: true,
		},
		{
			name:          "debug disabled when set to info",
			setLevel:      ports.LogLevelInfo,
			checkLevel:    ports.LogLevelDebug,
			expectEnabled: false,
		},
		{
			name:          "error enabled when set to warn",
			setLevel:      ports.LogLevelWarn,
			checkLevel:    ports.LogLevelError,
			expectEnabled: true,
		},
		{
			name:          "trace disabled when set to error",
			setLevel:      ports.LogLevelError,
			checkLevel:    ports.LogLevelTrace,
			expectEnabled: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			logger := zerolog.New(nil)
			adapter, ok := NewZerologAdapter(&logger).(*ZerologAdapter)
			require.True(t, ok)

			adapter.SetLevel(testCase.setLevel)
			isEnabled := adapter.IsLevelEnabled(testCase.checkLevel)

			assert.Equal(t, testCase.expectEnabled, isEnabled)
		})
	}
}

func TestZerologAdapterWithContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := zerolog.New(&buf)
	adapter := NewZerologAdapter(&logger)

	ctx := context.Background()
	zerologAdapter, ok := adapter.(*ZerologAdapter)
	require.True(t, ok)

	newAdapter := zerologAdapter.WithContext(ctx)

	assert.NotNil(t, newAdapter)
	assert.IsType(t, &ZerologAdapter{}, newAdapter)

	// Test that the new adapter works
	newAdapter.Info(ctx, "test message", map[string]interface{}{
		"key": "value",
	})

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "value")
}

func TestZerologAdapterEmptyFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields map[string]interface{}
	}{
		{
			name:   "nil fields",
			fields: nil,
		},
		{
			name:   "empty fields",
			fields: map[string]interface{}{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := zerolog.New(&buf)
			adapter := NewZerologAdapter(&logger)
			ctx := context.Background()

			adapter.Info(ctx, "test message", testCase.fields)

			output := buf.String()
			assert.Contains(t, output, "test message")

			// Should not error with nil or empty fields
			var logEntry map[string]interface{}

			err := json.Unmarshal([]byte(output), &logEntry)
			require.NoError(t, err)
			assert.Equal(t, "test message", logEntry["message"])
		})
	}
}

func TestZerologAdapterFieldTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		field interface{}
	}{
		{
			name:  "string",
			field: "test string",
		},
		{
			name:  "integer",
			field: 42,
		},
		{
			name:  "float",
			field: 3.14,
		},
		{
			name:  "boolean",
			field: true,
		},
		{
			name:  "slice",
			field: []string{"a", "b", "c"},
		},
		{
			name:  "map",
			field: map[string]string{"key": "value"},
		},
		{
			name:  "struct",
			field: struct{ Name string }{Name: "test"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := zerolog.New(&buf)
			adapter := NewZerologAdapter(&logger)

			fields := map[string]interface{}{
				"testfield": testCase.field,
			}

			adapter.Info(context.Background(), "test message", fields)

			output := buf.String()
			assert.Contains(t, output, "test message")

			// Should handle all field types without error
			assert.NotEmpty(t, output)

			// Parse to ensure valid JSON
			var logEntry map[string]interface{}

			err := json.Unmarshal([]byte(output), &logEntry)
			require.NoError(t, err)
		})
	}
}

func TestZerologAdapterDefaultLevel(t *testing.T) {
	t.Parallel()

	adapter := &ZerologAdapter{}

	// Test unknown level defaults to info
	unknownLevel := ports.LogLevel("unknown")
	zerologLevel := adapter.convertLogLevel(unknownLevel)
	assert.Equal(t, zerolog.InfoLevel, zerologLevel)

	// Test unknown zerolog level defaults to info
	unknownZerologLevel := zerolog.Level(127) // Use valid but unmapped level
	domainLevel := adapter.convertZerologLevel(unknownZerologLevel)
	assert.Equal(t, ports.LogLevelInfo, domainLevel)
}
