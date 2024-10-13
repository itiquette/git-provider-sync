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
)

// CloneOption represents options for a git clone operation.
// It includes flags for cleaning up names, mirroring, and specifies
// the source URL and target path.
type CloneOption struct {
	CleanupName bool   // Whether to clean up the repository name
	URL         string // The URL of the repository to clone
	Mirror      bool   // Whether to create a mirror clone
	Git         model.GitOption
	HTTPClient  model.HTTPClientOption
	PlainRepo   bool
}

// NewCloneOption creates a new CloneOption.
//
// Parameters:
//   - url: The URL of the repository to clone.
//   - mirror: Whether to create a mirror clone.
//   - targetPath: The local path where the repository will be cloned.
//
// Returns:
//   - A new CloneOption struct configured with the provided options.
func NewCloneOption(ctx context.Context, metainfo RepositoryMetainfo, mirror bool, providerConfig model.ProviderConfig) CloneOption {
	logger := log.Logger(ctx)

	var cloneURL string
	if strings.EqualFold(providerConfig.Git.Type, model.SSHAGENT) || strings.EqualFold(providerConfig.Git.Type, model.SSHKEY) {
		cloneURL = metainfo.SSHURL
	} else {
		cloneURL = metainfo.HTTPSURL
	}

	logger.Info().
		Str("url", cloneURL).
		Msg("Cloning repository")

	return CloneOption{URL: cloneURL, Mirror: mirror, Git: providerConfig.Git, HTTPClient: providerConfig.HTTPClient}
}

// String provides a string representation of CloneOption.
func (co CloneOption) String() string {
	return fmt.Sprintf("CloneOption{CleanupName: %v, URL: %q, Mirror: %v, GitOption: %+v}",
		co.CleanupName,
		co.URL,
		co.Mirror,
		co.Git)
}
