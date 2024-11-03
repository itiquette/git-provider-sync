// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	model "itiquette/git-provider-sync/internal/model/configuration"
	"strings"

	"github.com/rs/zerolog"
)

// CloneOption represents options for a git clone operation.
// It includes flags for cleaning up names, mirroring, and specifies
// the source URL and target path.
type CloneOption struct {
	CleanupName bool                   // Whether to clean up the repository name
	URL         string                 // The URL of the repository to clone
	Mirror      bool                   // Whether to create a mirror clone
	Git         model.GitOption        // Git configuration options
	HTTPClient  model.HTTPClientOption // HTTP client options
	SSHClient   model.SSHClientOption  // SSH client options
	PlainRepo   bool                   // Whether to clone as a plain repository
	Name        string                 // Repository name
}

// String provides a string representation of CloneOption.
func (co CloneOption) String() string {
	return fmt.Sprintf("CloneOption{Name: %s, URL: %s, CleanupName: %t, Mirror: %t, PlainRepo: %t, Git: %s, HTTPClient: %s, SSHClient: %s}",
		co.Name,
		co.URL,
		co.CleanupName,
		co.Mirror,
		co.PlainRepo,
		co.Git.String(),
		co.HTTPClient.String(),
		co.SSHClient.String())
}

// DebugLog creates a debug log event with clone options.
func (co CloneOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("name", co.Name).
				Str("url", co.URL).
				Bool("cleanup_name", co.CleanupName).
				Bool("mirror", co.Mirror).
				Bool("plain_repo", co.PlainRepo).
				Str("git", co.Git.String()).
				Str("http_client", co.HTTPClient.String()).
				Str("ssh_client", co.SSHClient.String())
}

// NewCloneOption creates a new CloneOption.
func NewCloneOption(ctx context.Context, metainfo ProjectInfo, mirror bool, providerConfig model.ProviderConfig) CloneOption {
	logger := log.Logger(ctx)

	cloneURL := metainfo.HTTPSURL
	if strings.EqualFold(providerConfig.Git.Type, model.SSHAGENT) {
		cloneURL = metainfo.SSHURL
	}

	logger.Info().
		Str("url", cloneURL).
		Msg("Cloning repository:")

	return CloneOption{
		Name:       metainfo.Name(ctx),
		URL:        cloneURL,
		Mirror:     mirror,
		Git:        providerConfig.Git,
		HTTPClient: providerConfig.HTTPClient,
		SSHClient:  providerConfig.SSHClient,
	}
}
