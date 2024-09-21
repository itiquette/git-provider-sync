// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"

	"github.com/rs/zerolog"
)

// PushOption represents options for a git push operation.
// It encapsulates the target repository, reference specifications,
// and flags for pruning and force pushing.
type PushOption struct {
	Target     string   // The URL of the target repository
	RefSpecs   []string // The reference specifications to push
	Prune      bool     // Whether to prune remote branches that no longer exist locally
	Force      bool     // Whether to force push (overwrite remote history)
	HTTPClient HTTPClientOption
}

// String provides a string representation of PushOption.
// It uses ScrubTarget to ensure sensitive information is not exposed.
//
// Returns:
//   - A string representation of the PushOption struct.
func (po PushOption) String() string {
	return fmt.Sprintf("PushOption{Target: %s, RefSpecs: %v, Prune: %t, Force: %t}",
		po.Target, po.RefSpecs, po.Prune, po.Force)
}

// DebugLog creates a debug log event with repository metadata.
// This method is useful for debugging and tracing Push option operations.
func (po PushOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Str("target", po.Target).
				Strs("refspecs", po.RefSpecs).
				Bool("prune", po.Prune).
				Bool("force", po.Force)
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
func NewPushOption(target string, prune, force bool, httpClient HTTPClientOption) PushOption {
	refSpecs := []string{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"}
	if force {
		for i, spec := range refSpecs {
			refSpecs[i] = "+" + spec
		}
	}

	return PushOption{
		Target:     target,
		RefSpecs:   refSpecs,
		Prune:      prune,
		Force:      force,
		HTTPClient: httpClient,
	}
}
