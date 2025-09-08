// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"io"
	"os"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FormatterFactory creates the appropriate formatter based on CLI configuration.
type FormatterFactory struct{}

// NewFormatterFactory creates a new formatter factory.
func NewFormatterFactory() *FormatterFactory {
	return &FormatterFactory{}
}

// CreateFormatter creates the appropriate formatter based on the CLI config.
// It also accepts the verbosity level as a separate parameter since it's not stored in CLIConfig.
func (f *FormatterFactory) CreateFormatter(config entities.CLIConfig, verbosity string, writer io.Writer) ports.SyncOutputFormatter {
	if writer == nil {
		writer = os.Stdout // Default to stdout for user-facing output
	}

	// If quiet/error mode is explicitly set, use quiet formatter
	if config.Quiet() || verbosity == VerbosityError {
		return NewQuietFormatter(writer)
	}

	// Select formatter based on output format
	switch config.OutputFormat() {
	case FormatJSON:
		return NewJSONFormatter(writer, verbosity)

	case FormatPlain:
		return NewPlainFormatter(writer, verbosity)

	case FormatDefault, "":
		// Default formatter with colors and icons for human-friendly output
		noColor := os.Getenv("NO_COLOR") != "" || config.ColorMode() == ColorModeNever

		return NewConsoleFormatter(writer, verbosity, noColor)

	default:
		// Unknown format - use default formatter
		noColor := os.Getenv("NO_COLOR") != "" || config.ColorMode() == ColorModeNever

		return NewConsoleFormatter(writer, verbosity, noColor)
	}
}
