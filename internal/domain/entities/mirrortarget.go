// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain"
)

// Domain errors for mirror targets.
var (
	ErrInvalidMirrorTarget    = errors.New("invalid mirror target")
	ErrMissingMirrorPath      = errors.New("path is required for directory/archive mirrors")
	ErrInvalidProviderType    = errors.New("invalid provider type")
	ErrMirrorTargetExists     = errors.New("mirror target already exists")
	ErrIncompatibleMirrorType = errors.New("incompatible mirror type for operation")
)

// MirrorTarget represents a destination for repository mirroring with validation and behavior.
type MirrorTarget struct {
	name         string
	providerType ProviderType
	domain       string
	owner        string
	path         string
	useGitBinary bool
	authConfig   AuthConfig
	options      MirrorOptions
}

// ProviderType represents the type of mirror provider.
type ProviderType string

const (
	// ProviderTypeGitHub represents GitHub provider.
	ProviderTypeGitHub ProviderType = "github"
	// ProviderTypeGitLab represents GitLab provider.
	ProviderTypeGitLab ProviderType = "gitlab"
	// ProviderTypeGitea represents Gitea provider.
	ProviderTypeGitea ProviderType = "gitea"
	// ProviderTypeDirectory represents directory provider.
	ProviderTypeDirectory ProviderType = "directory"
	// ProviderTypeArchive represents archive provider.
	ProviderTypeArchive ProviderType = "archive"
)

// AuthConfig represents authentication configuration for a mirror target.
type AuthConfig struct {
	authType   AuthType
	token      string
	username   string
	sshKeyPath string
	sshKey     string
}

// MirrorOptions contains options that control mirror behavior.
type MirrorOptions struct {
	createIfNotExists    bool
	updateDescription    bool
	setDefaultBranch     bool
	syncBranchProtection bool
	preserveVisibility   bool
	enableLFS            bool
	disableProtection    bool // Restore main branch Settings.Disabled functionality
}

// MirrorTargetBuilder builds mirror targets.
type MirrorTargetBuilder struct {
	target MirrorTarget
}

// NewMirrorTargetBuilder creates a new mirror target builder.
func NewMirrorTargetBuilder() MirrorTargetBuilder {
	return MirrorTargetBuilder{
		target: MirrorTarget{
			options: MirrorOptions{
				createIfNotExists:    true,  // Convenient default
				updateDescription:    true,  // Keep descriptions in sync
				setDefaultBranch:     true,  // Keep default branches in sync
				syncBranchProtection: false, // Secure default
				preserveVisibility:   true,  // Maintain source visibility
				enableLFS:            false, // Disable by default
			},
		},
	}
}

// WithName sets the mirror target name.
func (b MirrorTargetBuilder) WithName(name string) (MirrorTargetBuilder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return b, domain.ErrMirrorNameEmpty
	}

	b.target.name = name

	return b, nil
}

// WithProvider sets the provider type and validates it.
func (b MirrorTargetBuilder) WithProvider(providerType string) (MirrorTargetBuilder, error) {
	pt, err := ParseProviderType(providerType)
	if err != nil {
		return b, err
	}

	b.target.providerType = pt

	return b, nil
}

// WithDomain sets the provider domain.
func (b MirrorTargetBuilder) WithDomain(domain string) MirrorTargetBuilder {
	b.target.domain = strings.TrimSpace(domain)

	return b
}

// WithOwner sets the owner (user/organization).
func (b MirrorTargetBuilder) WithOwner(owner string) (MirrorTargetBuilder, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" && b.target.providerType.RequiresOwner() {
		return b, fmt.Errorf("%w: %s", domain.ErrOwnerRequiredForProvider, b.target.providerType)
	}

	b.target.owner = owner

	return b, nil
}

// WithPath sets the path for directory/archive providers.
func (b MirrorTargetBuilder) WithPath(path string) (MirrorTargetBuilder, error) {
	path = strings.TrimSpace(path)

	if b.target.providerType.RequiresPath() && path == "" {
		return b, ErrMissingMirrorPath
	}

	if path != "" {
		// Validate path format
		if !filepath.IsAbs(path) && b.target.providerType == ProviderTypeDirectory {
			return b, fmt.Errorf("%w: %s", domain.ErrDirectoryPathMustBeAbsolute, path)
		}
	}

	b.target.path = path

	return b, nil
}

// WithGitBinary sets whether to use git binary instead of go-git.
func (b MirrorTargetBuilder) WithGitBinary(use bool) MirrorTargetBuilder {
	b.target.useGitBinary = use

	return b
}

// WithAuth sets the authentication configuration.
func (b MirrorTargetBuilder) WithAuth(auth AuthConfig) MirrorTargetBuilder {
	b.target.authConfig = auth

	return b
}

// WithOptions sets the mirror options.
func (b MirrorTargetBuilder) WithOptions(options MirrorOptions) MirrorTargetBuilder {
	b.target.options = options

	return b
}

// WithDisableProtection sets whether to temporarily disable protection during push.
func (b MirrorTargetBuilder) WithDisableProtection(disable bool) MirrorTargetBuilder {
	b.target.options.disableProtection = disable

	return b
}

// Build creates the final mirror target after validation.
func (b MirrorTargetBuilder) Build() (MirrorTarget, error) {
	if err := b.target.Validate(); err != nil {
		return MirrorTarget{}, err
	}

	return b.target, nil
}

// MirrorTarget accessor methods

// NewMirrorTarget creates a new mirror target
// is a convenience function for use case code.
func NewMirrorTarget(name string, providerType ProviderType, domain, owner, path string, authConfig AuthConfig, enabled bool) MirrorTarget {
	return MirrorTarget{
		name:         name,
		providerType: providerType,
		domain:       domain,
		owner:        owner,
		path:         path,
		authConfig:   authConfig,
		options: MirrorOptions{
			createIfNotExists:    enabled,
			updateDescription:    true,
			setDefaultBranch:     true,
			syncBranchProtection: false,
			preserveVisibility:   true,
			enableLFS:            false,
		},
	}
}

// Name returns the mirror target name.
func (mt MirrorTarget) Name() string {
	return mt.name
}

// ProviderType returns the provider type.
func (mt MirrorTarget) ProviderType() ProviderType {
	return mt.providerType
}

// Domain returns the provider domain.
func (mt MirrorTarget) Domain() string {
	return mt.domain
}

// Owner returns the owner.
func (mt MirrorTarget) Owner() string {
	return mt.owner
}

// Path returns the path for directory/archive providers.
func (mt MirrorTarget) Path() string {
	return mt.path
}

// UseGitBinary returns whether to use git binary.
func (mt MirrorTarget) UseGitBinary() bool {
	return mt.useGitBinary
}

// AuthConfig returns the authentication configuration.
func (mt MirrorTarget) AuthConfig() AuthConfig {
	return mt.authConfig
}

// Options returns the mirror options.
func (mt MirrorTarget) Options() MirrorOptions {
	return mt.options
}

// MirrorTarget direct option access methods

// CreateIfNotExists returns whether to create repositories if they don't exist.
func (mt MirrorTarget) CreateIfNotExists() bool {
	return mt.options.createIfNotExists
}

// UpdateDescription returns whether to update repository descriptions.
func (mt MirrorTarget) UpdateDescription() bool {
	return mt.options.updateDescription
}

// SetDefaultBranch returns whether to set the default branch.
func (mt MirrorTarget) SetDefaultBranch() bool {
	return mt.options.setDefaultBranch
}

// SyncBranchProtection returns whether to sync branch protection rules.
func (mt MirrorTarget) SyncBranchProtection() bool {
	return mt.options.syncBranchProtection
}

// PreserveVisibility returns whether to preserve repository visibility.
func (mt MirrorTarget) PreserveVisibility() bool {
	return mt.options.preserveVisibility
}

// EnableLFS returns whether to enable Git LFS.
func (mt MirrorTarget) EnableLFS() bool {
	return mt.options.enableLFS
}

// DisableProtection returns whether to temporarily disable protection during push.
func (mt MirrorTarget) DisableProtection() bool {
	return mt.options.disableProtection
}

// MirrorTarget behavior methods

// Validate validates the mirror target configuration.
func (mt MirrorTarget) Validate() error {
	if mt.name == "" {
		return domain.ErrMirrorNameRequired
	}

	if mt.providerType == "" {
		return ErrInvalidProviderType
	}

	if mt.providerType.RequiresOwner() && mt.owner == "" {
		return fmt.Errorf("%w: %s", domain.ErrOwnerRequiredForProvider, mt.providerType)
	}

	if mt.providerType.RequiresPath() && mt.path == "" {
		return ErrMissingMirrorPath
	}

	if mt.providerType.RequiresAuth() && !mt.authConfig.HasValidAuth() {
		return fmt.Errorf("%w: %s", domain.ErrAuthenticationRequiredForProvider, mt.providerType)
	}

	return nil
}

// GetDefaultDomain returns the default domain for the provider type.
func (mt MirrorTarget) GetDefaultDomain() string {
	if mt.domain != "" {
		return mt.domain
	}

	return mt.providerType.DefaultDomain()
}

// GetMirrorURL constructs the full URL for the mirror target.
func (mt MirrorTarget) GetMirrorURL(repository Repository) string {
	switch mt.providerType {
	case ProviderTypeGitHub:
		return fmt.Sprintf("https://%s/%s/%s", mt.GetDefaultDomain(), mt.owner, repository.Name())
	case ProviderTypeGitLab:
		return fmt.Sprintf("https://%s/%s/%s", mt.GetDefaultDomain(), mt.owner, repository.Name())
	case ProviderTypeGitea:
		return fmt.Sprintf("https://%s/%s/%s", mt.GetDefaultDomain(), mt.owner, repository.Name())
	case ProviderTypeDirectory:
		return filepath.Join(mt.path, repository.Name())
	case ProviderTypeArchive:
		return filepath.Join(mt.path, repository.Name()+".tar.gz")
	default:
		return ""
	}
}

// SupportsOperation checks if the mirror target supports a specific operation.
func (mt MirrorTarget) SupportsOperation(operation MirrorOperation) bool {
	switch operation {
	case MirrorOperationClone, MirrorOperationPush:
		return true // All providers support clone and push
	case MirrorOperationCreateRepository:
		return mt.providerType.SupportsRepositoryManagement()
	case MirrorOperationBranchProtection:
		return mt.providerType.SupportsBranchProtection()
	case MirrorOperationVisibilityManagement:
		return mt.providerType.SupportsVisibilityManagement()
	default:
		return false
	}
}

// GetRepositoryNameForTarget returns the appropriate repository name for this target.
func (mt MirrorTarget) GetRepositoryNameForTarget(repo Repository, useCleanName bool) string {
	if useCleanName {
		return repo.CleanName()
	}

	// Validate name against target provider
	if err := repo.ValidateForProvider(string(mt.providerType)); err != nil {
		// If validation fails, use clean name as fallback
		return repo.CleanName()
	}

	return repo.Name()
}

// CanMirrorRepository checks if a repository can be mirrored to this target.
func (mt MirrorTarget) CanMirrorRepository(repo Repository) (bool, error) {
	// Check if repository name is valid for target provider
	if err := repo.ValidateForProvider(string(mt.providerType)); err != nil {
		return false, fmt.Errorf("repository name invalid for %s: %w", mt.providerType, err)
	}

	// Check authentication requirements
	if mt.providerType.RequiresAuth() && !mt.authConfig.HasValidAuth() {
		return false, fmt.Errorf("%w: %s", domain.ErrAuthenticationRequiredForProvider, mt.providerType)
	}

	return true, nil
}

// EstimateMirrorTime estimates the time needed to mirror a repository to this target.
func (mt MirrorTarget) EstimateMirrorTime(_ Repository) time.Duration {
	baseTime := 30 * time.Second // Base mirror time

	// Adjust based on provider type
	switch mt.providerType {
	case ProviderTypeDirectory:
		return baseTime / 2 // Directory operations are faster
	case ProviderTypeArchive:
		return baseTime / 3 // Archive operations are fastest
	case ProviderTypeGitHub, ProviderTypeGitLab, ProviderTypeGitea:
		return baseTime // Standard network operations
	default:
		return baseTime
	}
}

// WithUpdatedAuth returns a new mirror target with updated authentication.
func (mt MirrorTarget) WithUpdatedAuth(auth AuthConfig) MirrorTarget {
	updated := mt
	updated.authConfig = auth

	return updated
}

// WithUpdatedOptions returns a new mirror target with updated options.
func (mt MirrorTarget) WithUpdatedOptions(options MirrorOptions) MirrorTarget {
	updated := mt
	updated.options = options

	return updated
}

// ProviderType methods

// RequiresOwner returns true if the provider type requires an owner.
func (pt ProviderType) RequiresOwner() bool {
	switch pt {
	case ProviderTypeGitHub, ProviderTypeGitLab, ProviderTypeGitea:
		return true
	case ProviderTypeDirectory, ProviderTypeArchive:
		return false
	default:
		return false
	}
}

// RequiresPath returns true if the provider type requires a path.
func (pt ProviderType) RequiresPath() bool {
	switch pt {
	case ProviderTypeDirectory, ProviderTypeArchive:
		return true
	case ProviderTypeGitHub, ProviderTypeGitLab, ProviderTypeGitea:
		return false
	default:
		return false
	}
}

// RequiresAuth returns true if the provider type requires authentication.
func (pt ProviderType) RequiresAuth() bool {
	switch pt {
	case ProviderTypeGitHub, ProviderTypeGitLab, ProviderTypeGitea:
		return true
	case ProviderTypeDirectory, ProviderTypeArchive:
		return false
	default:
		return false
	}
}

// DefaultDomain returns the default domain for the provider type.
func (pt ProviderType) DefaultDomain() string {
	switch pt {
	case ProviderTypeGitHub:
		return "github.com"
	case ProviderTypeGitLab:
		return "gitlab.com"
	case ProviderTypeGitea:
		return "gitea.com"
	case ProviderTypeDirectory, ProviderTypeArchive:
		return "" // Local providers don't have domains
	default:
		return ""
	}
}

// SupportsRepositoryManagement returns true if the provider supports repository creation/management.
func (pt ProviderType) SupportsRepositoryManagement() bool {
	switch pt {
	case ProviderTypeGitHub, ProviderTypeGitLab, ProviderTypeGitea:
		return true
	case ProviderTypeDirectory, ProviderTypeArchive:
		return false
	default:
		return false
	}
}

// SupportsBranchProtection returns true if the provider supports branch protection.
func (pt ProviderType) SupportsBranchProtection() bool {
	switch pt {
	case ProviderTypeGitHub, ProviderTypeGitLab, ProviderTypeGitea:
		return true
	case ProviderTypeDirectory, ProviderTypeArchive:
		return false
	default:
		return false
	}
}

// SupportsVisibilityManagement returns true if the provider supports visibility management.
func (pt ProviderType) SupportsVisibilityManagement() bool {
	switch pt {
	case ProviderTypeGitHub, ProviderTypeGitLab, ProviderTypeGitea:
		return true
	case ProviderTypeDirectory, ProviderTypeArchive:
		return false
	default:
		return false
	}
}

// AuthConfig constructors

// NewAuthConfigWithToken creates auth config with token authentication.
func NewAuthConfigWithToken(token, username string) AuthConfig {
	return AuthConfig{
		token:    strings.TrimSpace(token),
		username: strings.TrimSpace(username),
	}
}

// NewAuthConfigWithSSH creates auth config with SSH key authentication.
func NewAuthConfigWithSSH(keyPath, username string) AuthConfig {
	return AuthConfig{
		sshKeyPath: strings.TrimSpace(keyPath),
		username:   strings.TrimSpace(username),
	}
}

// NewAuthConfigWithSSHKey creates auth config with SSH key content.
func NewAuthConfigWithSSHKey(keyContent, username string) AuthConfig {
	return AuthConfig{
		sshKey:   strings.TrimSpace(keyContent),
		username: strings.TrimSpace(username),
	}
}

// NewAuthenticationConfig creates a new authentication config.
func NewAuthenticationConfig(authType AuthType, token, username, sshKeyPath, sshKey string) AuthConfig {
	return AuthConfig{
		authType:   authType,
		token:      strings.TrimSpace(token),
		username:   strings.TrimSpace(username),
		sshKeyPath: strings.TrimSpace(sshKeyPath),
		sshKey:     strings.TrimSpace(sshKey),
	}
}

// MirrorTarget direct auth access methods

// HasValidAuth returns true if the auth config has valid authentication.
func (mt MirrorTarget) HasValidAuth() bool {
	return mt.authConfig.token != "" || mt.authConfig.sshKeyPath != "" || mt.authConfig.sshKey != ""
}

// Token returns the authentication token.
func (mt MirrorTarget) Token() string {
	return mt.authConfig.token
}

// Username returns the username.
func (mt MirrorTarget) Username() string {
	return mt.authConfig.username
}

// SSHKeyPath returns the SSH key path.
func (mt MirrorTarget) SSHKeyPath() string {
	return mt.authConfig.sshKeyPath
}

// SSHKey returns the SSH key content.
func (mt MirrorTarget) SSHKey() string {
	return mt.authConfig.sshKey
}

// AuthType returns the authentication type.
func (mt MirrorTarget) AuthType() AuthType {
	return mt.authConfig.authType
}

// AuthConfig methods

// HasValidAuth returns true if the auth config has valid authentication.
func (ac AuthConfig) HasValidAuth() bool {
	return ac.token != "" || ac.sshKeyPath != "" || ac.sshKey != ""
}

// Token returns the authentication token.
func (ac AuthConfig) Token() string {
	return ac.token
}

// Username returns the username.
func (ac AuthConfig) Username() string {
	return ac.username
}

// SSHKeyPath returns the SSH key path.
func (ac AuthConfig) SSHKeyPath() string {
	return ac.sshKeyPath
}

// SSHKey returns the SSH key content.
func (ac AuthConfig) SSHKey() string {
	return ac.sshKey
}

// Supporting types and enums

// MirrorOperation represents different types of mirror operations.
type MirrorOperation string

const (
	// MirrorOperationClone represents repository cloning operation.
	MirrorOperationClone MirrorOperation = "clone"
	// MirrorOperationPush represents repository push operation.
	MirrorOperationPush MirrorOperation = "push"
	// MirrorOperationCreateRepository represents repository creation operation.
	MirrorOperationCreateRepository MirrorOperation = "create_repository"
	// MirrorOperationBranchProtection represents branch protection operation.
	MirrorOperationBranchProtection MirrorOperation = "branch_protection"
	// MirrorOperationVisibilityManagement represents visibility management operation.
	MirrorOperationVisibilityManagement MirrorOperation = "visibility_management"
)

// Helper functions

// ParseProviderType parses a string into a ProviderType.
func ParseProviderType(providerStr string) (ProviderType, error) {
	switch strings.ToLower(strings.TrimSpace(providerStr)) {
	case "github":
		return ProviderTypeGitHub, nil
	case "gitlab":
		return ProviderTypeGitLab, nil
	case "gitea":
		return ProviderTypeGitea, nil
	case "directory":
		return ProviderTypeDirectory, nil
	case "archive":
		return ProviderTypeArchive, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidProviderType, providerStr)
	}
}

// Type returns the authentication type.
func (ac AuthConfig) Type() AuthType {
	return ac.authType
}

// String provides string representation for debugging.
func (ac AuthConfig) String() string {
	return fmt.Sprintf("AuthConfig{Type: %s, Username: %s, HasToken: %t, HasSSHKey: %t}",
		ac.authType, ac.username, ac.token != "", ac.sshKey != "")
}

// AuthType represents the type of authentication.
type AuthType string

const (
	// AuthTypeNone represents no authentication.
	AuthTypeNone AuthType = "none"
	// AuthTypeToken represents token-based authentication.
	AuthTypeToken AuthType = "token"
	// AuthTypeSSH represents SSH key authentication.
	AuthTypeSSH AuthType = "ssh"
	// AuthTypeBasic represents basic authentication.
	AuthTypeBasic AuthType = "basic"
)

// Enabled returns true if the mirror target is enabled
// For now, all mirror targets are considered enabled if they have valid configuration.
func (mt MirrorTarget) Enabled() bool {
	return mt.name != "" && mt.providerType != ""
}
