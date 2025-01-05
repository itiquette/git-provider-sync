// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	model "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
	"strings"

	"github.com/rs/zerolog"
)

// CloneOption represents options for a git clone operation.
// It includes flags for cleaning up names, mirroring, and specifies
// the source URL and target path.
type CloneOption struct {
	ASCIIName   string           // Whether to clean up the repository name
	URL         string           // The URL of the repository to clone
	Mirror      bool             // Whether to create a mirror clone
	SourceCfg   model.SyncConfig // Git configuration options
	AuthCfg     model.AuthConfig // HTTP client options
	NonBareRepo bool             // Whether to clone as a nonbare (regular with worktree) repository
	Name        string           // Repository name
}

// String provides a string representation of CloneOption.
func (co CloneOption) String() string {
	return fmt.Sprintf("CloneOption{Name: %s, URL: %s, ASCIIName: %s, Mirror: %t, NonBareRepo: %t, SourceCfg: %s, AuthCfg: %s}",
		co.Name,
		co.URL,
		co.ASCIIName,
		co.Mirror,
		co.NonBareRepo,
		co.SourceCfg.String(),
		co.AuthCfg.String())
}

// DebugLog creates a debug log event with clone options.
func (co CloneOption) DebugLog(ctx context.Context, logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("name", co.Name).
				Str("url", stringconvert.RemoveBasicAuthFromURL(ctx, co.URL, false)).
				Str("ASCIIName", co.ASCIIName).
				Bool("mirror", co.Mirror).
				Bool("nonbare_repo", co.NonBareRepo).
				Str("sourceCfg", co.SourceCfg.String()).
				Str("authCfg", co.AuthCfg.String())
}

// NewCloneOption creates a new CloneOption.
func NewCloneOption(ctx context.Context, projectInfo ProjectInfo, mirror bool, syncCfg model.SyncConfig) CloneOption {
	logger := log.Logger(ctx)

	cloneURL := projectInfo.HTTPSURL
	if strings.EqualFold(syncCfg.Auth.Protocol, model.SSH) {
		cloneURL = projectInfo.SSHURL
	}

	logger.Info().
		Str("url", cloneURL).
		Msg("Cloning repository:")

	return CloneOption{
		Name:      projectInfo.Name(ctx),
		URL:       cloneURL,
		Mirror:    mirror,
		SourceCfg: syncCfg,
		AuthCfg:   syncCfg.Auth,
		ASCIIName: projectInfo.CleanName,
	}
}
