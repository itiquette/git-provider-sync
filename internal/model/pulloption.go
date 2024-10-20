// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	model "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/rs/zerolog"
)

// PullOption represents options for a git pull operation.
// It includes the name of the remote, its URL, and the local target path.
type PullOption struct {
	Name             string                 // The name of the remote (e.g., "origin")
	URL              string                 // The URL of the remote repository
	GitOption        model.GitOption        // GitOption options
	HTTPClientOption model.HTTPClientOption // GitOption options
	SSHClient        model.SSHClientOption  // GitOption options
}

// DebugLog creates a debug log event with repository metadata.
// This method is useful for debugging and tracing Pull option operations.
func (po PullOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("target", po.Name).
				Str("gitoption.type", po.GitOption.Type).
				Bool("gitoption.includeforks", po.GitOption.IncludeForks).
				Str("prune", po.URL)
}

// NewPullOption creates a new PullOption.
//
// Parameters:
//   - name: The name of the remote.
//   - url: The URL of the remote repository.
//   - targetPath: The local path where the repository will be pulled.
//
// Returns:
//   - A new PullOption struct configured with the provided options.
func NewPullOption(name, url string, gitInfo model.GitOption, httpClient model.HTTPClientOption) PullOption {
	return PullOption{Name: name, URL: url, GitOption: gitInfo, HTTPClientOption: httpClient}
}

// String provides a string representation of PullOption.
//
// Returns:
//   - A string representation of the PullOption struct.
func (po PullOption) String() string {
	return fmt.Sprintf("PullOption{Name: %v, URL: %q, GitOption: %v}",
		po.Name,
		po.URL,
		po.GitOption)
}
