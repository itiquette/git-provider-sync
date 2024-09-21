// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strings"
)

// CloneOption represents options for a git clone operation.
// It includes flags for cleaning up names, mirroring, and specifies
// the source URL and target path.
type CloneOption struct {
	CleanupName bool   // Whether to clean up the repository name
	URL         string // The URL of the repository to clone
	Mirror      bool   // Whether to create a mirror clone
	TargetPath  string // The path where the repository will be cloned
	Git         GitOption
	HTTPClient  HTTPClientOption
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
func NewCloneOption(ctx context.Context, info RepositoryMetainfo, mirror bool, targetPath string, gitInfo GitOption, httpClient HTTPClientOption) CloneOption {
	logger := log.Logger(ctx)

	var cloneURL string
	if strings.EqualFold(gitInfo.Type, SSHAGENT) || strings.EqualFold(gitInfo.Type, SSHKEY) {
		cloneURL = info.SSHURL
	} else {
		cloneURL = info.HTTPSURL
	}

	logger.Info().
		Str("url", cloneURL).
		Str("target", targetPath).
		Msg("Cloning repository")

	return CloneOption{URL: cloneURL, Mirror: mirror, TargetPath: targetPath, Git: gitInfo, HTTPClient: httpClient}
}

// String provides a string representation of CloneOption.
func (co CloneOption) String() string {
	return fmt.Sprintf("CloneOption{CleanupName: %v, URL: %q, Mirror: %v, TargetPath: %q, GitOption: %+v}",
		co.CleanupName,
		co.URL,
		co.Mirror,
		co.TargetPath,
		co.Git)
}
