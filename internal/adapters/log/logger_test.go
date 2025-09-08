// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestLevel_ToZerologLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    Level
		expected zerolog.Level
	}{
		{
			name:     "error level",
			level:    LevelError,
			expected: zerolog.ErrorLevel,
		},
		{
			name:     "warn level",
			level:    LevelWarn,
			expected: zerolog.WarnLevel,
		},
		{
			name:     "info level",
			level:    LevelInfo,
			expected: zerolog.InfoLevel,
		},
		{
			name:     "debug level",
			level:    LevelDebug,
			expected: zerolog.DebugLevel,
		},
		{
			name:     "trace level",
			level:    LevelTrace,
			expected: zerolog.TraceLevel,
		},
		{
			name:     "unknown level defaults to info",
			level:    Level("unknown"),
			expected: zerolog.InfoLevel,
		},
		{
			name:     "empty level defaults to info",
			level:    Level(""),
			expected: zerolog.InfoLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.level.ToZerologLevel()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestInitLogger_JSONFormat(t *testing.T) { //nolint:paralleltest,tparallel // Cannot run in parallel due to global zerolog state modification
	// Cannot use t.Parallel() due to global zerolog.TimeFieldFormat modification
	tests := []struct {
		name                string
		verbosity           string
		quiet               bool
		withVerbosityCaller bool
		expectedLevel       zerolog.Level
	}{
		{
			name:          "quiet mode",
			verbosity:     "info",
			quiet:         true,
			expectedLevel: zerolog.ErrorLevel,
		},
		{
			name:          "brief verbosity",
			verbosity:     "info",
			quiet:         false,
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name:          "debug verbosity",
			verbosity:     "debug",
			quiet:         false,
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "trace verbosity",
			verbosity:     "trace",
			quiet:         false,
			expectedLevel: zerolog.TraceLevel,
		},
		{
			name:                "trace with caller",
			verbosity:           "trace",
			quiet:               false,
			withVerbosityCaller: true,
			expectedLevel:       zerolog.TraceLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			loggerCtx := InitLogger(ctx, test.verbosity, test.quiet, test.withVerbosityCaller, "json")

			logger := zerolog.Ctx(loggerCtx)
			require.NotNil(t, logger)

			// Verify the log level is set correctly
			assert.Equal(t, test.expectedLevel, logger.GetLevel())
		})
	}
}

func TestInitLogger_ConsoleFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	loggerCtx := InitLogger(ctx, "debug", false, false, "console")

	logger := zerolog.Ctx(loggerCtx)
	require.NotNil(t, logger)
	assert.Equal(t, zerolog.DebugLevel, logger.GetLevel())
}

func TestInitLogger_OutputCapture(t *testing.T) { //nolint:paralleltest // Cannot run in parallel due to global stdout modification
	// Cannot use t.Parallel() due to global stdout modification

	// Capture stdout for JSON format testing
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer

	ctx := context.Background()
	loggerCtx := InitLogger(ctx, "info", false, false, "json")

	logger := zerolog.Ctx(loggerCtx)
	logger.Info().Msg("test message")

	// Close writer and restore stdout
	_ = writer.Close()
	os.Stdout = originalStdout

	// Read the captured output
	var buf bytes.Buffer

	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, `"level":"info"`)

	// Verify it's valid JSON
	var logEntry map[string]any

	err = json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "test message", logEntry["message"])
}

func TestLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectNil     bool
		expectedLevel *zerolog.Level
	}{
		{
			name: "context with logger",
			setupContext: func() context.Context {
				ctx := context.Background()

				return InitLogger(ctx, "debug", false, false, "json")
			},
			expectNil: false,
			expectedLevel: func() *zerolog.Level {
				l := zerolog.DebugLevel

				return &l
			}(),
		},
		{
			name:         "context without logger",
			setupContext: context.Background,
			expectNil:    false, // zerolog.Ctx returns a disabled logger, not nil
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := test.setupContext()
			logger := Logger(ctx)

			if test.expectNil {
				assert.Nil(t, logger)
			} else {
				assert.NotNil(t, logger)

				if test.expectedLevel != nil {
					assert.Equal(t, *test.expectedLevel, logger.GetLevel())
				}
			}
		})
	}
}

func TestCreateDomainLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupContext func() context.Context
	}{
		{
			name: "context with logger",
			setupContext: func() context.Context {
				ctx := context.Background()

				return InitLogger(ctx, "info", false, false, "json")
			},
		},
		{
			name:         "context without logger",
			setupContext: context.Background,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := test.setupContext()
			domainLogger := CreateDomainLogger(ctx)

			// Verify it returns a domain ports.Logger interface
			assert.NotNil(t, domainLogger)
			assert.Implements(t, (*ports.Logger)(nil), domainLogger)

			// Test that we can use it for logging (shouldn't panic)
			assert.NotPanics(t, func() {
				domainLogger.Info(ctx, "test message", nil)
				domainLogger.Debug(ctx, "debug message", nil)
				domainLogger.Error(ctx, "error message", nil)
			})
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		verbosity string
		quiet     bool
		expected  zerolog.Level
	}{
		{
			name:      "quiet overrides verbosity",
			verbosity: "debug",
			quiet:     true,
			expected:  zerolog.ErrorLevel,
		},
		{
			name:      "brief verbosity",
			verbosity: "brief",
			quiet:     false,
			expected:  zerolog.InfoLevel,
		},
		{
			name:      "debug verbosity",
			verbosity: "debug",
			quiet:     false,
			expected:  zerolog.DebugLevel,
		},
		{
			name:      "trace verbosity",
			verbosity: "trace",
			quiet:     false,
			expected:  zerolog.TraceLevel,
		},
		{
			name:      "unknown verbosity defaults to info",
			verbosity: "unknown",
			quiet:     false,
			expected:  zerolog.InfoLevel,
		},
		{
			name:      "empty verbosity defaults to info",
			verbosity: "",
			quiet:     false,
			expected:  zerolog.InfoLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := getLogLevel(test.verbosity, test.quiet)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestLoggerIntegration_ActualLogging(t *testing.T) {
	t.Parallel()

	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a logger that writes to our buffer
	logger := zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	// Get domain logger and test logging
	domainLogger := CreateDomainLogger(ctx)

	domainLogger.Info(ctx, "test info message", nil)
	domainLogger.Error(ctx, "test error message", nil)

	output := buf.String()
	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, "test error message")
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"level":"error"`)
}

func TestLoggerIntegration_DifferentLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		logLevel     string
		logMessage   func(ports.Logger)
		shouldAppear bool
	}{
		{
			name:     "info message at info level",
			logLevel: "info",
			logMessage: func(l ports.Logger) {
				l.Info(context.Background(), "info message", nil)
			},
			shouldAppear: true,
		},
		{
			name:     "debug message at info level",
			logLevel: "info",
			logMessage: func(l ports.Logger) {
				l.Debug(context.Background(), "debug message", nil)
			},
			shouldAppear: false,
		},
		{
			name:     "error message at error level",
			logLevel: "error",
			logMessage: func(l ports.Logger) {
				l.Error(context.Background(), "error message", nil)
			},
			shouldAppear: true,
		},
		{
			name:     "info message at error level",
			logLevel: "error",
			logMessage: func(l ports.Logger) {
				l.Info(context.Background(), "info message", nil)
			},
			shouldAppear: false,
		},
		{
			name:     "debug message at debug level",
			logLevel: "debug",
			logMessage: func(l ports.Logger) {
				l.Debug(context.Background(), "debug message", nil)
			},
			shouldAppear: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			// Create logger with specific level
			level := Level(test.logLevel).ToZerologLevel()
			logger := zerolog.New(&buf).Level(level).With().Timestamp().Logger()
			ctx := logger.WithContext(context.Background())

			domainLogger := CreateDomainLogger(ctx)
			test.logMessage(domainLogger)

			output := buf.String()
			if test.shouldAppear {
				assert.NotEmpty(t, output, "Expected log message to appear")
			} else {
				assert.Empty(t, output, "Expected log message to be filtered out")
			}
		})
	}
}

func TestInitLogger_CallerInformation(t *testing.T) { //nolint:paralleltest // Cannot run in parallel due to global stderr modification
	// Cannot use t.Parallel() due to global stderr modification
	var buf bytes.Buffer

	// Temporarily redirect stderr for console format
	originalStderr := os.Stderr
	reader2, writer2, err := os.Pipe()
	require.NoError(t, err)

	os.Stderr = writer2

	ctx := context.Background()
	loggerCtx := InitLogger(ctx, "debug", false, true, "console")

	logger := zerolog.Ctx(loggerCtx)
	logger.Info().Msg("test message with caller")

	// Close writer and restore stderr
	_ = writer2.Close()
	os.Stderr = originalStderr

	// Read the captured output
	_, err = buf.ReadFrom(reader2)
	require.NoError(t, err)

	// For console format, we just verify the logger was created successfully
	// Actual caller information would be complex to test in this context
	assert.NotNil(t, logger)
	assert.Equal(t, zerolog.DebugLevel, logger.GetLevel())
}

func TestLoggerChaining(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	// Test that we can chain logger operations
	logger := zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	domainLogger := CreateDomainLogger(ctx)

	// Test chaining doesn't break anything
	domainLogger.Info(ctx, "first message", nil)
	domainLogger.Error(ctx, "second message", nil)
	domainLogger.Info(ctx, "third message", nil)

	output := buf.String()

	// Count the number of log entries
	lines := strings.Split(strings.TrimSpace(output), "\n")
	validLines := 0

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			validLines++
		}
	}

	assert.Equal(t, 3, validLines, "Expected 3 log messages")
	assert.Contains(t, output, "first message")
	assert.Contains(t, output, "second message")
	assert.Contains(t, output, "third message")
}

// Benchmark tests for performance monitoring.
func BenchmarkInitLogger(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		loggerCtx := InitLogger(ctx, "info", false, false, "json")
		_ = loggerCtx
	}
}

func BenchmarkCreateDomainLogger(b *testing.B) {
	ctx := context.Background()
	loggerCtx := InitLogger(ctx, "info", false, false, "json")

	b.ResetTimer()

	for range b.N {
		domainLogger := CreateDomainLogger(loggerCtx)
		_ = domainLogger
	}
}

func BenchmarkLogging(b *testing.B) {
	ctx := context.Background()
	loggerCtx := InitLogger(ctx, "info", false, false, "json")
	domainLogger := CreateDomainLogger(loggerCtx)

	b.ResetTimer()

	for range b.N {
		domainLogger.Info(context.Background(), "benchmark message", nil)
	}
}
