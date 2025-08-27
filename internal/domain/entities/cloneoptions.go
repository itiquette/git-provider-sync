// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/domain"
)

// CloneOptions represents immutable clone operation configuration as domain value object.
// This replaces model.CloneOption with proper domain design and immutability.
type CloneOptions struct {
	repositoryName string
	sourceURL      string
	targetPath     string
	isMirror       bool
	isNonBare      bool
	useASCIIName   bool
	authConfig     AuthConfig
	protocol       string
}

// CloneOptionsBuilder provides functional approach to building clone options.
type CloneOptionsBuilder struct {
	options CloneOptions
}

// NewCloneOptionsBuilder creates a new clone options builder with defaults.
func NewCloneOptionsBuilder() CloneOptionsBuilder {
	return CloneOptionsBuilder{
		options: CloneOptions{
			isMirror:     false,
			isNonBare:    false,
			useASCIIName: false,
			protocol:     "https",
		},
	}
}

// WithRepositoryName specifies the repository identifier for clone operations.
func (b CloneOptionsBuilder) WithRepositoryName(name string) CloneOptionsBuilder {
	b.options.repositoryName = strings.TrimSpace(name)

	return b
}

// WithSourceURL specifies the source repository location for cloning.
func (b CloneOptionsBuilder) WithSourceURL(url string) CloneOptionsBuilder {
	b.options.sourceURL = strings.TrimSpace(url)

	return b
}

// WithTargetPath specifies where the repository should be cloned locally.
func (b CloneOptionsBuilder) WithTargetPath(path string) CloneOptionsBuilder {
	b.options.targetPath = strings.TrimSpace(path)

	return b
}

// WithMirror sets whether this is a mirror clone.
func (b CloneOptionsBuilder) WithMirror(mirror bool) CloneOptionsBuilder {
	b.options.isMirror = mirror

	return b
}

// WithNonBare sets whether to create a non-bare repository (with working tree).
func (b CloneOptionsBuilder) WithNonBare(nonBare bool) CloneOptionsBuilder {
	b.options.isNonBare = nonBare

	return b
}

// WithASCIIName sets whether to use ASCII-cleaned name.
func (b CloneOptionsBuilder) WithASCIIName(ascii bool) CloneOptionsBuilder {
	b.options.useASCIIName = ascii

	return b
}

// WithAuthentication sets the authentication configuration.
func (b CloneOptionsBuilder) WithAuthentication(auth AuthConfig) CloneOptionsBuilder {
	b.options.authConfig = auth

	return b
}

// WithProtocol sets the protocol (https, ssh).
func (b CloneOptionsBuilder) WithProtocol(protocol string) CloneOptionsBuilder {
	b.options.protocol = strings.ToLower(strings.TrimSpace(protocol))

	return b
}

// Build creates immutable clone options after validation.
func (b CloneOptionsBuilder) Build() (CloneOptions, error) {
	if b.options.repositoryName == "" {
		return CloneOptions{}, domain.ErrRepositoryNameRequired
	}

	if b.options.sourceURL == "" {
		return CloneOptions{}, domain.ErrSourceURLRequired
	}

	return b.options, nil
}

// Factory functions for common patterns

// NewCloneOptionsFromRepository creates clone options from repository entity.
func NewCloneOptionsFromRepository(repo Repository, auth AuthConfig, mirror bool) (CloneOptions, error) {
	builder := NewCloneOptionsBuilder().
		WithRepositoryName(repo.Name()).
		WithMirror(mirror).
		WithAuthentication(auth)

	// Prefer HTTPS URL if available
	if repo.HTTPSURL() != "" { //nolint:gocritic // if-else chain is more readable for URL preference logic
		builder = builder.WithSourceURL(repo.HTTPSURL()).WithProtocol("https")
	} else if repo.SSHURL() != "" {
		builder = builder.WithSourceURL(repo.SSHURL()).WithProtocol("ssh")
	} else {
		return CloneOptions{}, domain.ErrNoCloneURLs
	}

	return builder.Build()
}

// NewMirrorCloneOptions creates clone options for mirror operations.
func NewMirrorCloneOptions(repoName, sourceURL string, auth AuthConfig) (CloneOptions, error) {
	return NewCloneOptionsBuilder().
		WithRepositoryName(repoName).
		WithSourceURL(sourceURL).
		WithMirror(true).
		WithNonBare(false).
		WithAuthentication(auth).
		Build()
}

// NewRegularCloneOptions creates clone options for regular repository clones.
func NewRegularCloneOptions(repoName, sourceURL, targetPath string, auth AuthConfig) (CloneOptions, error) {
	return NewCloneOptionsBuilder().
		WithRepositoryName(repoName).
		WithSourceURL(sourceURL).
		WithTargetPath(targetPath).
		WithMirror(false).
		WithNonBare(true).
		WithAuthentication(auth).
		Build()
}

// Immutable accessor methods

// RepositoryName returns the repository name.
func (co CloneOptions) RepositoryName() string {
	return co.repositoryName
}

// SourceURL returns the source repository URL.
func (co CloneOptions) SourceURL() string {
	return co.sourceURL
}

// TargetPath returns the target clone path.
func (co CloneOptions) TargetPath() string {
	return co.targetPath
}

// IsMirror returns whether this is a mirror clone.
func (co CloneOptions) IsMirror() bool {
	return co.isMirror
}

// IsNonBare returns whether to create non-bare repository.
func (co CloneOptions) IsNonBare() bool {
	return co.isNonBare
}

// UseASCIIName returns whether to use ASCII-cleaned name.
func (co CloneOptions) UseASCIIName() bool {
	return co.useASCIIName
}

// AuthConfig returns the authentication configuration.
func (co CloneOptions) AuthConfig() AuthConfig {
	return co.authConfig
}

// Protocol returns the protocol (https, ssh).
func (co CloneOptions) Protocol() string {
	return co.protocol
}

// Domain behavior methods

// EffectiveName returns the repository name to use (ASCII-cleaned if enabled).
func (co CloneOptions) EffectiveName() string {
	if co.useASCIIName && co.repositoryName != "" {
		return CleanRepositoryName(co.repositoryName)
	}

	return co.repositoryName
}

// IsSecureProtocol returns whether the protocol is secure (https or ssh).
func (co CloneOptions) IsSecureProtocol() bool {
	return co.protocol == "https" || co.protocol == "ssh"
}

// RequiresAuthentication returns whether authentication is required.
func (co CloneOptions) RequiresAuthentication() bool {
	return co.authConfig.Type() != AuthTypeNone
}

// GetCloneURL returns the appropriate clone URL based on protocol preference.
func (co CloneOptions) GetCloneURL(httpsURL, sshURL string) string {
	if co.protocol == "ssh" && sshURL != "" {
		return sshURL
	}

	return httpsURL
}

// WithUpdatedAuth returns new clone options with updated authentication.
func (co CloneOptions) WithUpdatedAuth(auth AuthConfig) CloneOptions {
	updated := co
	updated.authConfig = auth

	return updated
}

// WithUpdatedProtocol returns new clone options with updated protocol.
func (co CloneOptions) WithUpdatedProtocol(protocol string) CloneOptions {
	updated := co
	updated.protocol = strings.ToLower(strings.TrimSpace(protocol))

	return updated
}

// String provides string representation for debugging.
func (co CloneOptions) String() string {
	return fmt.Sprintf("CloneOptions{Name: %s, URL: %s, Mirror: %t, NonBare: %t, Protocol: %s, Auth: %s}",
		co.repositoryName,
		co.sourceURL,
		co.isMirror,
		co.isNonBare,
		co.protocol,
		co.authConfig.String())
}
