// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	model "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/shared"
)

// PushOption represents options for a git push operation.
// It encapsulates the target repository, reference specifications,
// and flags for pruning and force pushing.
type PushOption struct {
	Force    bool // Whether to force push (overwrite remote history)
	AuthCfg  model.AuthConfig
	Prune    bool     // Whether to prune remote branches that no longer exist locally
	RefSpecs []string // The reference specifications to push
	Target   string   // The URL of the target repository
}

// NewPushOption creates a new PushOption with appropriate RefSpecs.
// Automatically sets up correct reference specifications based on force push setting.
func NewPushOption(target string, prune, force bool, authCfg model.AuthConfig) PushOption {
	refSpecs := []string{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"} // Standard refspecs for push
	if force {
		for i, spec := range refSpecs {
			if !strings.HasPrefix(spec, "^") {
				refSpecs[i] = "+" + spec
			}
		}
	}

	return PushOption{
		Force:    force,
		AuthCfg:  authCfg,
		Prune:    prune,
		RefSpecs: refSpecs,
		Target:   target,
	}
}

func (po PushOption) String() string {
	return fmt.Sprintf("PushOption{Target: %s, RefSpecs: %v, Prune: %t, Force: %t, AuthCfg: %s}",
		po.Target,
		po.RefSpecs,
		po.Prune,
		po.Force,
		po.AuthCfg.String(),
	)
}

// DebugLog logs the push option details for debugging purposes.
func (po PushOption) DebugLog(_ context.Context, logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("target", shared.RemoveBasicAuthFromURL(po.Target, true)).
				Strs("refspecs", po.RefSpecs).
				Bool("prune", po.Prune).
				Bool("force", po.Force).
				Str("auth_confg", po.AuthCfg.String())
}
