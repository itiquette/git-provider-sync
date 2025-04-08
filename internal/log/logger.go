// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package log provides logging functionality
// using zerolog and cobra. It offers easy setup of leveled logging.
package log

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

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
	JSON    Format = "json"
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

// InitLogger initializes and returns a context with a configured logger.
// It sets up the logger based on the command line flags for verbosity and quiet mode.
//
// Parameters:
//   - ctx: The parent context
//   - cmd: The cobra.Command instance, used to retrieve flags
//
// Returns:
//   - context.Context with the configured logger
//
// The logger is set up with a console writer for human-readable output.
// If the log level is set to trace, it includes the caller information in the log output.
func InitLogger(ctx context.Context, cmd *cobra.Command, withVerbosityCaller bool, outputFormat string) context.Context {
	level := getLogLevel(cmd)

	var writer io.Writer
	if outputFormat == "json" {
		writer = os.Stdout
		zerolog.TimeFieldFormat = time.RFC3339
	} else {
		writer = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	var logger zerolog.Logger

	if withVerbosityCaller {
		if level == zerolog.TraceLevel {
			logger = zerolog.New(writer).
				Level(level).
				With().
				Caller().
				Timestamp().
				Logger()
		} else {
			logger = zerolog.New(writer).
				Level(level).
				With().
				Caller().
				Timestamp().
				Logger()
		}
	} else {
		if level == zerolog.TraceLevel {
			logger = zerolog.New(writer).
				Level(level).
				With().
				Timestamp().
				Logger()
		} else {
			logger = zerolog.New(writer).
				Level(level).
				With().
				Timestamp().Logger()
		}
	}

	return logger.WithContext(ctx)
}

// Logger retrieves the zerolog.Logger from the given context.
// This function should be used to obtain a logger instance for logging.
//
// Parameters:
//   - ctx: The context containing the logger
//
// Returns:
//   - *zerolog.Logger: A pointer to the logger instance
//
// Usage:
//
//	logger := log.Logger(ctx)
//	logger.Info().Msg("This is an info message")
func Logger(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// getLogLevel determines the log level based on command flags.
// It checks for the "quiet" flag first, then falls back to the "verbosity" flag.
//
// Parameters:
//   - cmd: The cobra.Command instance to retrieve flags from
//
// Returns:
//   - zerolog.Level: The determined log level
//
// Note: This function assumes that the "verbosity" flag is a string
// and the "quiet" flag is a boolean.
func getLogLevel(cmd *cobra.Command) zerolog.Level {
	level, _ := cmd.Flags().GetString("verbosity")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if quiet {
		return LevelQuiet.ToZerologLevel()
	}

	return Level(level).ToZerologLevel()
}

// Example usage:
//
//	func Execute() {
//		cmd := &cobra.Command{
//			Use:   "myapp",
//			Short: "A brief description of your application",
//			Run: func(cmd *cobra.Command, args []string) {
//				ctx := context.Background()
//				ctx = log.InitLogger(ctx, cmd)
//				logger := log.Logger(ctx)
//				logger.Info().Msg("Application started")
//				// Your application logic here
//			},
//		}
//
//		cmd.PersistentFlags().String("verbosity", "brief", "Log level (quiet, brief, debug, trace)")
//		cmd.PersistentFlags().Bool("quiet", false, "Suppress all output except errors")
//
//		if err := cmd.Execute(); err != nil {
//			os.Exit(1)
//		}
//	}
