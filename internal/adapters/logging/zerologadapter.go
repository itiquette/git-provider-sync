// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package logging provides adapters for logging operations.
package logging

import (
	"context"

	"github.com/rs/zerolog"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ZerologAdapter adapts zerolog.Logger to implement the domain ports.Logger interface.
type ZerologAdapter struct {
	logger *zerolog.Logger
}

// NewZerologAdapter creates a new zerolog adapter.
func NewZerologAdapter(logger *zerolog.Logger) ports.Logger {
	return &ZerologAdapter{logger: logger}
}

// Trace logs a trace level message with the provided fields.
func (z *ZerologAdapter) Trace(ctx context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Trace()
	z.addFields(event, fields)
	event.Msg(msg)
}

func (z *ZerologAdapter) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Debug()
	z.addFields(event, fields)
	event.Msg(msg)
}

func (z *ZerologAdapter) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Info()
	z.addFields(event, fields)
	event.Msg(msg)
}

func (z *ZerologAdapter) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Warn()
	z.addFields(event, fields)
	event.Msg(msg)
}

func (z *ZerologAdapter) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Error()
	z.addFields(event, fields)
	event.Msg(msg)
}

func (z *ZerologAdapter) Fatal(ctx context.Context, msg string, fields map[string]interface{}) {
	event := z.logger.Fatal()
	z.addFields(event, fields)
	event.Msg(msg)
}

// WithField returns a new logger instance with the specified field added.
func (z *ZerologAdapter) WithField(key string, value interface{}) ports.Logger {
	newLogger := z.logger.With().Interface(key, value).Logger()

	return &ZerologAdapter{logger: &newLogger}
}

func (z *ZerologAdapter) WithFields(fields map[string]interface{}) ports.Logger {
	ctx := z.logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}

	newLogger := ctx.Logger()

	return &ZerologAdapter{logger: &newLogger}
}

func (z *ZerologAdapter) WithContext(ctx context.Context) ports.Logger {
	newLogger := z.logger.WithContext(ctx)

	return &ZerologAdapter{logger: zerolog.Ctx(newLogger)}
}

// SetLevel sets the minimum log level for this logger.
func (z *ZerologAdapter) SetLevel(level ports.LogLevel) {
	z.logger.Level(z.convertLogLevel(level))
}

func (z *ZerologAdapter) GetLevel() ports.LogLevel {
	return z.convertZerologLevel(z.logger.GetLevel())
}

func (z *ZerologAdapter) IsLevelEnabled(level ports.LogLevel) bool {
	return z.logger.GetLevel() <= z.convertLogLevel(level)
}

// Helper methods.
func (z *ZerologAdapter) addFields(event *zerolog.Event, fields map[string]interface{}) {
	for key, value := range fields {
		event.Interface(key, value)
	}
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
