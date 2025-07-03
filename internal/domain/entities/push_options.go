// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"fmt"
	"strings"
)

// PushOptions represents immutable push operation configuration as domain value object.
// This replaces model.PushOption with proper domain design and immutability.
type PushOptions struct {
	targetURL    string
	refSpecs     []string
	forcePush    bool
	pruneRemotes bool
	authConfig   AuthConfig
	remoteName   string
}

// PushOptionsBuilder provides functional approach to building push options.
type PushOptionsBuilder struct {
	options PushOptions
}

// NewPushOptionsBuilder creates a new push options builder with defaults.
func NewPushOptionsBuilder() PushOptionsBuilder {
	return PushOptionsBuilder{
		options: PushOptions{
			forcePush:    false,
			pruneRemotes: false,
			remoteName:   "origin",
			refSpecs:     []string{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"},
		},
	}
}

// WithTargetURL sets the target repository URL.
func (b PushOptionsBuilder) WithTargetURL(url string) PushOptionsBuilder {
	b.options.targetURL = strings.TrimSpace(url)
	return b
}

// WithRefSpecs sets the reference specifications for push.
func (b PushOptionsBuilder) WithRefSpecs(specs []string) PushOptionsBuilder {
	b.options.refSpecs = make([]string, len(specs))
	copy(b.options.refSpecs, specs)
	return b
}

// WithForcePush sets whether to force push (overwrite remote history).
func (b PushOptionsBuilder) WithForcePush(force bool) PushOptionsBuilder {
	b.options.forcePush = force

	// Automatically update refspecs for force push
	if force {
		for i, spec := range b.options.refSpecs {
			if !strings.HasPrefix(spec, "+") && !strings.HasPrefix(spec, "^") {
				b.options.refSpecs[i] = "+" + spec
			}
		}
	} else {
		for i, spec := range b.options.refSpecs {
			if strings.HasPrefix(spec, "+") {
				b.options.refSpecs[i] = strings.TrimPrefix(spec, "+")
			}
		}
	}

	return b
}

// WithPruneRemotes sets whether to prune remote branches that no longer exist locally.
func (b PushOptionsBuilder) WithPruneRemotes(prune bool) PushOptionsBuilder {
	b.options.pruneRemotes = prune
	return b
}

// WithAuthentication sets the authentication configuration.
func (b PushOptionsBuilder) WithAuthentication(auth AuthConfig) PushOptionsBuilder {
	b.options.authConfig = auth
	return b
}

// WithRemoteName sets the remote name.
func (b PushOptionsBuilder) WithRemoteName(name string) PushOptionsBuilder {
	b.options.remoteName = strings.TrimSpace(name)
	return b
}

// Build creates immutable push options after validation.
func (b PushOptionsBuilder) Build() (PushOptions, error) {
	if b.options.targetURL == "" {
		return PushOptions{}, fmt.Errorf("target URL is required")
	}

	if len(b.options.refSpecs) == 0 {
		return PushOptions{}, fmt.Errorf("at least one refspec is required")
	}

	if b.options.remoteName == "" {
		return PushOptions{}, fmt.Errorf("remote name is required")
	}

	return b.options, nil
}

// Immutable accessor methods

// TargetURL returns the target repository URL.
func (po PushOptions) TargetURL() string {
	return po.targetURL
}

// RefSpecs returns a copy of the reference specifications.
func (po PushOptions) RefSpecs() []string {
	specs := make([]string, len(po.refSpecs))
	copy(specs, po.refSpecs)
	return specs
}

// ForcePush returns whether to force push.
func (po PushOptions) ForcePush() bool {
	return po.forcePush
}

// PruneRemotes returns whether to prune remote branches.
func (po PushOptions) PruneRemotes() bool {
	return po.pruneRemotes
}

// AuthConfig returns the authentication configuration.
func (po PushOptions) AuthConfig() AuthConfig {
	return po.authConfig
}

// RemoteName returns the remote name.
func (po PushOptions) RemoteName() string {
	return po.remoteName
}

// Domain behavior methods

// IsDestructive returns whether this push operation could be destructive.
func (po PushOptions) IsDestructive() bool {
	return po.forcePush || po.pruneRemotes
}

// HasForceRefSpecs returns whether any refspecs use force (+).
func (po PushOptions) HasForceRefSpecs() bool {
	for _, spec := range po.refSpecs {
		if strings.HasPrefix(spec, "+") {
			return true
		}
	}
	return false
}

// RequiresAuthentication returns whether authentication is required.
func (po PushOptions) RequiresAuthentication() bool {
	return po.authConfig.Type() != AuthTypeNone
}

// GetSanitizedURL returns the target URL with auth information removed for logging.
func (po PushOptions) GetSanitizedURL() string {
	url := po.targetURL

	// Remove basic auth from URL for logging
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			protocolPart := parts[0]
			if strings.Contains(protocolPart, "://") {
				protocolSplit := strings.Split(protocolPart, "://")
				if len(protocolSplit) == 2 {
					return protocolSplit[0] + "://" + strings.Join(parts[1:], "@")
				}
			}
		}
	}

	return url
}

// WithUpdatedAuth returns new push options with updated authentication.
func (po PushOptions) WithUpdatedAuth(auth AuthConfig) PushOptions {
	updated := po
	updated.authConfig = auth
	return updated
}

// WithUpdatedForce returns new push options with updated force setting.
func (po PushOptions) WithUpdatedForce(force bool) PushOptions {
	builder := NewPushOptionsBuilder().
		WithTargetURL(po.targetURL).
		WithRefSpecs(po.refSpecs).
		WithPruneRemotes(po.pruneRemotes).
		WithAuthentication(po.authConfig).
		WithRemoteName(po.remoteName).
		WithForcePush(force)

	// Safe to ignore error since we're using valid existing data
	updated, _ := builder.Build()
	return updated
}

// String provides string representation for debugging.
func (po PushOptions) String() string {
	return fmt.Sprintf("PushOptions{Target: %s, RefSpecs: %v, Force: %t, Prune: %t, Remote: %s, Auth: %s}",
		po.GetSanitizedURL(),
		po.refSpecs,
		po.forcePush,
		po.pruneRemotes,
		po.remoteName,
		po.authConfig.String())
}

// Factory functions for common patterns

// NewStandardPushOptions creates push options for standard repository push.
func NewStandardPushOptions(targetURL string, auth AuthConfig) (PushOptions, error) {
	return NewPushOptionsBuilder().
		WithTargetURL(targetURL).
		WithAuthentication(auth).
		Build()
}

// NewForcePushOptions creates push options for force push operations.
func NewForcePushOptions(targetURL string, auth AuthConfig) (PushOptions, error) {
	return NewPushOptionsBuilder().
		WithTargetURL(targetURL).
		WithForcePush(true).
		WithAuthentication(auth).
		Build()
}

// NewMirrorPushOptions creates push options for mirror push operations.
func NewMirrorPushOptions(targetURL string, auth AuthConfig) (PushOptions, error) {
	return NewPushOptionsBuilder().
		WithTargetURL(targetURL).
		WithRefSpecs([]string{"refs/*:refs/*"}).
		WithForcePush(true).
		WithPruneRemotes(true).
		WithAuthentication(auth).
		Build()
}

// NewCustomPushOptions creates push options with custom refspecs.
func NewCustomPushOptions(targetURL string, refspecs []string, auth AuthConfig) (PushOptions, error) {
	return NewPushOptionsBuilder().
		WithTargetURL(targetURL).
		WithRefSpecs(refspecs).
		WithAuthentication(auth).
		Build()
}
