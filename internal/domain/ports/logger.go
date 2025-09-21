// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package ports

import "context"

// Logger defines the core interface for logging operations (secondary port)
// port is implemented by adapters that handle logging to various destinations
// Following ISP: separated core logging from configuration concerns.
type Logger interface {
	// Core logging methods
	Trace(ctx context.Context, msg string, fields map[string]any)
	Debug(ctx context.Context, msg string, fields map[string]any)
	Info(ctx context.Context, msg string, fields map[string]any)
	Warn(ctx context.Context, msg string, fields map[string]any)
	Error(ctx context.Context, msg string, fields map[string]any)
	Fatal(ctx context.Context, msg string, fields map[string]any)

	// Level check for performance
	IsLevelEnabled(level LogLevel) bool
	GetLevel() LogLevel
}

// LoggerWithEnrichment provides field enrichment capabilities
// Note: These methods should return new logger instances (functional style).
type LoggerWithEnrichment interface {
	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger
	WithContext(ctx context.Context) Logger
}

// LogLevel and LogFormat are already defined in configuration.go

// LoggerConfig contains configuration for the logger.
type LoggerConfig struct {
	Level      LogLevel
	Format     LogFormat
	Output     string
	WithCaller bool
	TimeFormat string
}

// NewLoggerConfig creates a default logger configuration.
func NewLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:      LogLevelInfo,
		Format:     LogFormatJSON,
		Output:     "stdout",
		WithCaller: false,
		TimeFormat: "2006-01-02T15:04:05Z07:00",
	}
}

// LoggerKey is the context key for storing the logger.
type loggerKey struct{}

// LoggerFromContext retrieves the Logger from the context.
// If no logger is found, returns a no-op logger to prevent nil pointer issues.
func LoggerFromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey{}).(Logger); ok {
		return logger
	}
	// Return a no-op logger if none found
	return &noOpLogger{}
}

// WithLogger adds a Logger to the context.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// noOpLogger is a logger that does nothing.
type noOpLogger struct{}

func (n *noOpLogger) Trace(_ context.Context, _ string, _ map[string]any) {}
func (n *noOpLogger) Debug(_ context.Context, _ string, _ map[string]any) {}
func (n *noOpLogger) Info(_ context.Context, _ string, _ map[string]any)  {}
func (n *noOpLogger) Warn(_ context.Context, _ string, _ map[string]any)  {}
func (n *noOpLogger) Error(_ context.Context, _ string, _ map[string]any) {}
func (n *noOpLogger) Fatal(_ context.Context, _ string, _ map[string]any) {}
func (n *noOpLogger) IsLevelEnabled(_ LogLevel) bool                      { return false }
func (n *noOpLogger) GetLevel() LogLevel                                  { return LogLevelInfo }
