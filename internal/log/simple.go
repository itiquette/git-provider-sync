// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// SimpleLogger provides a simple implementation of the Logger port using zerolog.
// This adapter bridges the domain Logger interface with zerolog implementation.
type SimpleLogger struct {
	logger zerolog.Logger
}

// NewSimpleLogger creates a new simple logger with default configuration.
func NewSimpleLogger() *SimpleLogger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return &SimpleLogger{logger: logger}
}

// NewSimpleLoggerWithLevel creates a new simple logger with specified level.
func NewSimpleLoggerWithLevel(level ports.LogLevel) *SimpleLogger {
	zerologLevel := convertLogLevel(level)
	logger := zerolog.New(os.Stderr).Level(zerologLevel).With().Timestamp().Logger()

	return &SimpleLogger{logger: logger}
}

// Trace logs a trace message.
func (sl *SimpleLogger) Trace(ctx context.Context, msg string, fields map[string]interface{}) {
	event := sl.logger.Trace()
	if fields != nil {
		event = event.Fields(fields)
	}

	event.Msg(msg)
}

// Debug logs a debug message.
func (sl *SimpleLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	event := sl.logger.Debug()
	if fields != nil {
		event = event.Fields(fields)
	}

	event.Msg(msg)
}

// Info logs an info message.
func (sl *SimpleLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	event := sl.logger.Info()
	if fields != nil {
		event = event.Fields(fields)
	}

	event.Msg(msg)
}

// Warn logs a warning message.
func (sl *SimpleLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	event := sl.logger.Warn()
	if fields != nil {
		event = event.Fields(fields)
	}

	event.Msg(msg)
}

// Error logs an error message.
func (sl *SimpleLogger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	event := sl.logger.Error()
	if fields != nil {
		event = event.Fields(fields)
	}

	event.Msg(msg)
}

// Fatal logs a fatal message and exits.
func (sl *SimpleLogger) Fatal(ctx context.Context, msg string, fields map[string]interface{}) {
	event := sl.logger.Fatal()
	if fields != nil {
		event = event.Fields(fields)
	}

	event.Msg(msg)
}

// IsLevelEnabled checks if a log level is enabled.
func (sl *SimpleLogger) IsLevelEnabled(level ports.LogLevel) bool {
	zerologLevel := convertLogLevel(level)

	return sl.logger.GetLevel() <= zerologLevel
}

// convertLogLevel converts domain LogLevel to zerolog.Level.
func convertLogLevel(level ports.LogLevel) zerolog.Level {
	switch level {
	case ports.LogLevelTrace:
		return zerolog.TraceLevel
	case ports.LogLevelDebug:
		return zerolog.DebugLevel
	case ports.LogLevelInfo:
		return zerolog.InfoLevel
	case ports.LogLevelWarn:
		return zerolog.WarnLevel
	case ports.LogLevelError:
		return zerolog.ErrorLevel
	case ports.LogLevelFatal:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
