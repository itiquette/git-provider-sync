// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import "context"

// Logger defines the core interface for logging operations (secondary port).
// This port is implemented by adapters that handle logging to various destinations.
// Following ISP: separated core logging from configuration concerns.
type Logger interface {
	// Core logging methods
	Trace(ctx context.Context, msg string, fields map[string]interface{})
	Debug(ctx context.Context, msg string, fields map[string]interface{})
	Info(ctx context.Context, msg string, fields map[string]interface{})
	Warn(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, msg string, fields map[string]interface{})
	Fatal(ctx context.Context, msg string, fields map[string]interface{})

	// Level check for performance
	IsLevelEnabled(level LogLevel) bool
}

// LoggerLevelController provides level management capabilities.
type LoggerLevelController interface {
	SetLevel(level LogLevel)
	GetLevel() LogLevel
}

// LoggerWithEnrichment provides field enrichment capabilities.
// Note: These methods should return new logger instances (functional style).
type LoggerWithEnrichment interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithContext(ctx context.Context) Logger
}

// FullLogger composes all logging capabilities for implementations that need them.
type FullLogger interface {
	Logger
	LoggerLevelController
	LoggerWithEnrichment
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
