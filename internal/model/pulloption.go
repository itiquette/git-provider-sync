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
	Name      string // The name of the remote (e.g., "origin")
	URL       string // The URL of the remote repository
	SyncCfg   model.SyncConfig
	AuthCfg   model.AuthConfig
	Path      string // For Dir and Arch
	TargetDir string
}

func (po PullOption) String() string {
	return fmt.Sprintf("PullOption{Name: %s, URL: %s, SyncCfg: %s, AuthCfg: %s}",
		po.Name,
		po.URL,
		po.SyncCfg.String(),
		po.AuthCfg.String())
}

func (po PullOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("name", po.Name).
				Str("url", po.URL).
				Str("git_option", po.SyncCfg.String()).
				Str("ssh_client", po.AuthCfg.String())
}

// NewPullOption creates a new PullOption.
func NewPullOption(name, url string, syncCfg model.SyncConfig, authCfg model.AuthConfig, targetDir string, path string) PullOption {
	return PullOption{
		Name:      name,
		URL:       url,
		SyncCfg:   syncCfg,
		AuthCfg:   authCfg,
		TargetDir: targetDir,
		Path:      path,
	}
}
