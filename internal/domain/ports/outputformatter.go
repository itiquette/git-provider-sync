// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"io"

	model "itiquette/git-provider-sync/internal/model/configuration"
)

// OutputFormatter defines the contract for formatting and outputting data.
// This port enables machine-readable output following UNIX philosophy.
type OutputFormatter interface {
	// FormatConfiguration formats application configuration for output.
	// Supports console (human-readable), json (structured), and plain (tabular text) formats.
	FormatConfiguration(appCfg model.AppConfiguration, format string, writer io.Writer) error

	// FormatSyncResults formats sync operation results for output.
	// Progress and status information should go to stderr, data to stdout.
	FormatSyncResults(results interface{}, format string, dataWriter, progressWriter io.Writer) error

	// SupportedFormats returns the list of supported output formats.
	SupportedFormats() []string
}
