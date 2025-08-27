// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package log provides logging functionality using zerolog.
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"

	"itiquette/git-provider-sync/internal/adapters/logging"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/shared"
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// debugLogPathKey is the context key for storing the debug log file path.
const debugLogPathKey contextKey = "debugLogPath"

// Level represents the available log levels.
type Level string

// Predefined log levels.
const (
	LevelQuiet Level = "quiet" // Only error messages
	LevelBrief Level = "brief" // Info and above (default)
	LevelDebug Level = "debug" // Debug and above
	LevelTrace Level = "trace" // Trace and above (most verbose)
)

// Format represents the available log output formats.
type Format string

const (
	// JSON represents JSON log format.
	JSON Format = "json"
	// CONSOLE represents console log format.
	CONSOLE Format = "console"
)

// ToZerologLevel converts a LogLevel to the corresponding zerolog.Level.
func (l Level) ToZerologLevel() zerolog.Level {
	switch l {
	case LevelQuiet:
		return zerolog.ErrorLevel
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelTrace:
		return zerolog.TraceLevel
	case LevelBrief:
		return zerolog.InfoLevel
	default:
		return zerolog.InfoLevel // Default to Info for LevelBrief and unknown levels
	}
}

// setupConsoleWriter creates and configures a console writer with optional debug file tee.
func setupConsoleWriter(logLevel string) (io.Writer, string) {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	// Set up debug file tee for debug/trace levels
	if logLevel == "debug" || logLevel == "trace" {
		return logging.TeeWriter(consoleWriter, logLevel)
	}

	return consoleWriter, ""
}

// InitLogger initializes and returns a context with a configured logger.
// It sets up the logger based on the provided log level and quiet parameters.
//
// Parameters:
//   - ctx: The parent context
//   - logLevel: The log level string (quiet | brief | verbose | debug | trace)
//   - quiet: Whether quiet mode is enabled
//   - withCaller: Whether to include caller information in logs
//   - outputFormat: The output format ("json" or "console")
//
// The logger is set up with a console writer for human-readable output.
// If withCaller is true, it includes the caller information in the log output.
// For debug/trace levels, output is also written to a debug file.
func InitLogger(ctx context.Context, logLevel string, quiet bool, withCaller bool, outputFormat string) context.Context {
	level := getLogLevel(logLevel, quiet)

	var writer io.Writer

	var debugPath string

	if outputFormat == "json" {
		writer = os.Stdout
	} else {
		writer, debugPath = setupConsoleWriter(logLevel)
		if debugPath != "" {
			// Store debug path in context for later reference
			ctx = context.WithValue(ctx, debugLogPathKey, debugPath)
		}
	}

	var logger zerolog.Logger

	if withCaller {
		// Include caller information in logs
		logger = zerolog.New(writer).
			Level(level).
			With().
			Caller().
			Timestamp().
			Logger()
	} else {
		// Standard logging without caller information
		logger = zerolog.New(writer).
			Level(level).
			With().
			Timestamp().
			Logger()
	}

	return logger.WithContext(ctx)
}

// Logger retrieves the zerolog.Logger from the given context.
func Logger(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// CreateDomainLogger retrieves a domain ports.Logger from the context.
// This maintains hexagonal architecture by returning the domain interface,
// not the concrete zerolog implementation.
func CreateDomainLogger(ctx context.Context) ports.Logger {
	zerologInstance := zerolog.Ctx(ctx)

	return logging.NewZerologAdapter(zerologInstance)
}

// getLogLevel determines the log level based on command flags.
func getLogLevel(logLevel string, quiet bool) zerolog.Level {
	if quiet {
		return LevelQuiet.ToZerologLevel()
	}

	return Level(logLevel).ToZerologLevel()
}

// GetDebugLogPath retrieves the debug log file path from context.
func GetDebugLogPath(ctx context.Context) string {
	if path := ctx.Value(debugLogPathKey); path != nil {
		if pathStr, ok := path.(string); ok {
			return pathStr
		}
	}

	return ""
}

// SanitizeError creates a sanitized error that removes credentials from error messages.
// This should be used when logging errors that might contain URLs with embedded credentials.
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	sanitized := shared.SanitizeError(err)

	return fmt.Errorf("%s", sanitized)
}
