// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package logging provides adapters for logging operations.
package logging

import (
	"context"

	"github.com/rs/zerolog"

	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/shared"
)

// ZerologAdapter adapts zerolog.Logger to implement the domain ports.Logger interface.
type ZerologAdapter struct {
	logger *zerolog.Logger
}

// NewZerologAdapter creates a new zerolog adapter.
func NewZerologAdapter(logger *zerolog.Logger) ports.Logger { //nolint:ireturn // Factory function returning interface
	return &ZerologAdapter{logger: logger}
}

// Trace logs a trace level message with the provided fields.
func (z *ZerologAdapter) Trace(_ /* ctx */ context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Trace()
	z.addFields(event, fields)
	event.Msg(z.sanitizeMessage(msg))
}

// Debug logs a debug level message with the provided fields.
func (z *ZerologAdapter) Debug(_ context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Debug()
	z.addFields(event, fields)
	event.Msg(z.sanitizeMessage(msg))
}

// Info logs an info level message with the provided fields.
func (z *ZerologAdapter) Info(_ context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Info()
	z.addFields(event, fields)
	event.Msg(z.sanitizeMessage(msg))
}

// Warn logs a warn level message with the provided fields.
func (z *ZerologAdapter) Warn(_ context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Warn()
	z.addFields(event, fields)
	event.Msg(z.sanitizeMessage(msg))
}

func (z *ZerologAdapter) Error(_ context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Error()
	z.addFields(event, fields)
	event.Msg(z.sanitizeMessage(msg))
}

// Fatal logs a fatal message and terminates the program.
func (z *ZerologAdapter) Fatal(_ context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Fatal()
	z.addFields(event, fields)
	event.Msg(z.sanitizeMessage(msg))
}

// WithField returns a new logger instance with the specified field added.
func (z *ZerologAdapter) WithField(key string, value interface{}) ports.Logger { //nolint:ireturn // Interface implementation
	newLogger := z.logger.With().Interface(key, value).Logger()

	return &ZerologAdapter{logger: &newLogger}
}

// WithFields creates a new logger with additional fields.
func (z *ZerologAdapter) WithFields(fields map[string]interface{}) ports.Logger { //nolint:ireturn // Interface implementation
	ctx := z.logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}

	newLogger := ctx.Logger()

	return &ZerologAdapter{logger: &newLogger}
}

// WithContext creates a new logger with the provided context.
func (z *ZerologAdapter) WithContext(ctx context.Context) ports.Logger { //nolint:ireturn // Interface implementation
	newLogger := z.logger.WithContext(ctx)

	return &ZerologAdapter{logger: zerolog.Ctx(newLogger)}
}

// SetLevel sets the minimum log level for this logger.
func (z *ZerologAdapter) SetLevel(level ports.LogLevel) {
	*z.logger = z.logger.Level(z.convertLogLevel(level))
}

// GetLevel returns the current logging level.
func (z *ZerologAdapter) GetLevel() ports.LogLevel {
	return z.convertZerologLevel(z.logger.GetLevel())
}

// IsLevelEnabled checks if the given log level is enabled.
func (z *ZerologAdapter) IsLevelEnabled(level ports.LogLevel) bool {
	return z.logger.GetLevel() <= z.convertLogLevel(level)
}

// Helper methods.
func (z *ZerologAdapter) addFields(event *zerolog.Event, fields map[string]interface{}) {
	// Sanitize the fields before logging to prevent credential leaks
	sanitized := shared.SanitizeStringMap(fields)
	for key, value := range sanitized {
		event.Interface(key, value)
	}
}

// sanitizeMessage sanitizes URLs and credentials in log messages.
func (z *ZerologAdapter) sanitizeMessage(msg string) string {
	// Check if message contains URL patterns
	if shared.ContainsURL(msg) {
		return shared.SanitizeURL(msg)
	}

	return msg
}

func (z *ZerologAdapter) convertLogLevel(level ports.LogLevel) zerolog.Level {
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

func (z *ZerologAdapter) convertZerologLevel(level zerolog.Level) ports.LogLevel {
	switch level {
	case zerolog.TraceLevel:
		return ports.LogLevelTrace
	case zerolog.DebugLevel:
		return ports.LogLevelDebug
	case zerolog.InfoLevel:
		return ports.LogLevelInfo
	case zerolog.WarnLevel:
		return ports.LogLevelWarn
	case zerolog.ErrorLevel:
		return ports.LogLevelError
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return ports.LogLevelFatal
	case zerolog.NoLevel, zerolog.Disabled:
		return ports.LogLevelInfo
	default:
		return ports.LogLevelInfo
	}
}
