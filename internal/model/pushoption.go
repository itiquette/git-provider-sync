// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"fmt"
	model "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
	"strings"

	"github.com/rs/zerolog"
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

func (po PushOption) String() string {
	return fmt.Sprintf("PushOption{Target: %s, RefSpecs: %v, Prune: %t, Force: %t, AuthCfg: %s}",
		po.Target,
		po.RefSpecs,
		po.Prune,
		po.Force,
		po.AuthCfg.String(),
	)
}

func (po PushOption) DebugLog(ctx context.Context, logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("target", stringconvert.RemoveBasicAuthFromURL(ctx, po.Target, false)).
				Strs("refspecs", po.RefSpecs).
				Bool("prune", po.Prune).
				Bool("force", po.Force).
				Str("auth_confg", po.AuthCfg.String())
}

// NewPushOption creates a new PushOption with appropriate RefSpecs.
// It automatically sets up the correct reference specifications based on
// whether a force push is requested.
//
// Parameters:
//   - target: The URL of the target repository.
//   - prune: Whether to prune remote branches.
//   - force: Whether to force push.
//
// Returns:
//   - A new PushOption struct configured with the provided options.
func NewPushOption(target string, prune, force bool, authCfg model.AuthConfig) PushOption {
	refSpecs := []string{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"} //TODO: add bug report
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
