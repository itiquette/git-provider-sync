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
	Name             string // The name of the remote (e.g., "origin")
	URL              string // The URL of the remote repository
	GitOption        model.GitOption
	HTTPClientOption model.HTTPClientOption
	SSHClient        model.SSHClientOption
}

func (po PullOption) String() string {
	return fmt.Sprintf("PullOption{Name: %s, URL: %s, GitOption: %s, HTTPClient: %s, SSHClient: %s}",
		po.Name,
		po.URL,
		po.GitOption.String(),
		po.HTTPClientOption.String(),
		po.SSHClient.String())
}

func (po PullOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("name", po.Name).
				Str("url", po.URL).
				Str("git_option", po.GitOption.String()).
				Str("http_client", po.HTTPClientOption.String()).
				Str("ssh_client", po.SSHClient.String())
}

// NewPullOption creates a new PullOption.
func NewPullOption(name, url string, gitInfo model.GitOption, httpClient model.HTTPClientOption, sshClient model.SSHClientOption) PullOption {
	return PullOption{
		Name:             name,
		URL:              url,
		GitOption:        gitInfo,
		HTTPClientOption: httpClient,
		SSHClient:        sshClient,
	}
}
