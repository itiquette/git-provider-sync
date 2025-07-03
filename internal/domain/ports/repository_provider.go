// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package ports defines interfaces (ports) for the hexagonal architecture.
// These interfaces represent the boundaries between the domain and external systems.
package ports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// RepositoryDiscovery provides repository discovery capabilities.
type RepositoryDiscovery interface {
	ListRepositories(ctx context.Context, config ProviderConfig) ([]entities.Repository, error)
	GetRepository(ctx context.Context, config ProviderConfig, name string) (entities.Repository, error)
	RepositoryExists(ctx context.Context, request RepositoryExistsRequest) (bool, string, error)
}

// RepositoryLifecycle provides repository creation, update, and deletion.
type RepositoryLifecycle interface {
	CreateRepository(ctx context.Context, config ProviderConfig, options CreateRepositoryOptions) (entities.Repository, error)
	UpdateRepository(ctx context.Context, config ProviderConfig, name string, options UpdateRepositoryOptions) error
	DeleteRepository(ctx context.Context, config ProviderConfig, name string) error
	SetDefaultBranch(ctx context.Context, owner, name, branch string) error
}

// RepositoryValidation provides name validation and transformation.
type RepositoryValidation interface {
	ValidateRepositoryName(name string) error
	IsValidProjectName(ctx context.Context, name string) bool
	TransformRepositoryName(name string, options NameTransformOptions) string
}

// BranchProtectionManager provides branch protection management.
type BranchProtectionManager interface {
	GetBranchProtection(ctx context.Context, config ProviderConfig, repoName, branch string) (BranchProtection, error)
	SetBranchProtection(ctx context.Context, config ProviderConfig, repoName, branch string, protection BranchProtection) error
	RemoveBranchProtection(ctx context.Context, config ProviderConfig, repoName, branch string) error
	ListProtectedBranches(ctx context.Context, config ProviderConfig, repoName string) ([]string, error)
}

// ProviderPushOperations provides push-to-provider functionality from main branch.
type ProviderPushOperations interface {
	CreateRepositoryForPush(ctx context.Context, request CreateRepositoryRequest) (string, error)
	ProjectExists(ctx context.Context, owner, repo string) (bool, string, error)
	Protect(ctx context.Context, owner string, defaultBranch string, projectIDstr string) error
	Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error
}

// ProviderCapabilitiesPort provides information about provider capabilities.
type ProviderCapabilitiesPort interface {
	GetProviderInfo() ProviderInfo
	SupportsFeature(feature ProviderFeature) bool
}

// RepositoryProvider defines the interface for git provider operations (secondary port).
// This port is implemented by adapters that connect to external git providers like GitHub, GitLab, Gitea.
// It composes smaller, focused interfaces following the Interface Segregation Principle.
type RepositoryProvider interface {
	RepositoryDiscovery
	RepositoryLifecycle
	RepositoryValidation
	BranchProtectionManager
	ProviderCapabilitiesPort
	ProviderPushOperations
}

// ProviderConfig represents configuration for connecting to a git provider.
type ProviderConfig struct {
	ProviderType string
	Domain       string
	Owner        string
	AuthConfig   AuthenticationConfig
}

// AuthenticationConfig represents authentication configuration for providers.
type AuthenticationConfig struct {
	Token      string
	Username   string
	SSHKeyPath string
	SSHKey     string
}

// CreateRepositoryOptions contains options for creating repositories.
type CreateRepositoryOptions struct {
	Name              string
	Description       string
	Visibility        string
	DefaultBranch     string
	AutoInit          bool
	LicenseTemplate   string
	GitignoreTemplate string
	Topics            []string
}

// CreateRepositoryRequest contains parameters for creating a repository.
type CreateRepositoryRequest struct {
	Name          string
	Description   string
	Visibility    string
	DefaultBranch string
	Private       bool
}

// RepositoryExistsRequest contains parameters for checking repository existence.
type RepositoryExistsRequest struct {
	Owner string
	Name  string
}

// UpdateRepositoryOptions contains options for updating repositories.
type UpdateRepositoryOptions struct {
	Description   *string
	Visibility    *string
	DefaultBranch *string
	Topics        []string
}

// FilterOptions contains options for filtering repositories.
type FilterOptions struct {
	// Pattern matching
	IncludePatterns []string
	ExcludePatterns []string

	// Repository characteristics
	IncludeForks    bool
	IncludeArchived bool
	IncludePrivate  bool
	IncludePublic   bool

	// Content filtering
	Languages []string
	Topics    []string

	// Size filtering
	MinSize int64
	MaxSize int64

	// Activity filtering
	ActiveSince   *time.Time
	InactiveSince *time.Time

	// Metadata filtering
	HasDescription *bool
	HasWiki        *bool
	HasIssues      *bool
	HasProjects    *bool
}

// NameTransformOptions contains options for transforming repository names.
type NameTransformOptions struct {
	// Transformations
	Prefix           string
	Suffix           string
	Replacements     map[string]string
	ToLowercase      bool
	ToUppercase      bool
	AlphaNumericOnly bool

	// Constraints
	MaxLength int
	MinLength int

	// Validation
	ValidateForProvider string
}

// BranchProtection represents branch protection settings.
type BranchProtection struct {
	Protected                      bool
	RequiredStatusChecks           RequiredStatusChecks
	RequiredPullRequestReviews     RequiredPullRequestReviews
	Restrictions                   BranchRestrictions
	EnforceAdmins                  bool
	RequiredLinearHistory          bool
	AllowForcePushes               bool
	AllowDeletions                 bool
	RequiredConversationResolution bool
}

// RequiredStatusChecks represents required status check settings.
type RequiredStatusChecks struct {
	Strict   bool
	Contexts []string
}

// RequiredPullRequestReviews represents pull request review requirements.
type RequiredPullRequestReviews struct {
	RequiredApprovingReviewCount int
	DismissStaleReviews          bool
	RequireCodeOwnerReviews      bool
	DismissalRestrictions        UserRestrictions
}

// BranchRestrictions represents who can push to protected branches.
type BranchRestrictions struct {
	Users []string
	Teams []string
	Apps  []string
}

// UserRestrictions represents user-based restrictions.
type UserRestrictions struct {
	Users []string
	Teams []string
}

// ProviderInfo contains information about a git provider.
type ProviderInfo struct {
	Name         string
	Type         string
	Domain       string
	APIVersion   string
	Features     []ProviderFeature
	Capabilities ProviderCapabilities
}

// ProviderFeature represents a feature supported by a provider.
type ProviderFeature string

const (
	// Repository management.
	FeatureRepositoryCreation  ProviderFeature = "repository_creation"
	FeatureRepositoryDeletion  ProviderFeature = "repository_deletion"
	FeatureRepositoryTransfer  ProviderFeature = "repository_transfer"
	FeatureRepositoryTemplates ProviderFeature = "repository_templates"

	// Branch protection.
	FeatureBranchProtection     ProviderFeature = "branch_protection"
	FeatureRequiredStatusChecks ProviderFeature = "required_status_checks"
	FeatureRequiredReviews      ProviderFeature = "required_reviews"
	FeatureBranchRestrictions   ProviderFeature = "branch_restrictions"

	// Collaboration.
	FeatureIssueTracking ProviderFeature = "issue_tracking"
	FeaturePullRequests  ProviderFeature = "pull_requests"
	FeatureMergeRequests ProviderFeature = "merge_requests"
	FeatureWikis         ProviderFeature = "wikis"
	FeatureProjects      ProviderFeature = "projects"

	// CI/CD.
	FeatureContinuousIntegration ProviderFeature = "continuous_integration"
	FeatureActions               ProviderFeature = "actions"
	FeaturePipelines             ProviderFeature = "pipelines"
	FeaturePackageRegistry       ProviderFeature = "package_registry"

	// Security.
	FeatureSecurityAdvisories  ProviderFeature = "security_advisories"
	FeatureDependencyGraph     ProviderFeature = "dependency_graph"
	FeatureVulnerabilityAlerts ProviderFeature = "vulnerability_alerts"
	FeatureCodeScanning        ProviderFeature = "code_scanning"
	FeatureSecretScanning      ProviderFeature = "secret_scanning"

	// Organization management.
	FeatureOrganizations   ProviderFeature = "organizations"
	FeatureTeams           ProviderFeature = "teams"
	FeatureGroupManagement ProviderFeature = "group_management"
	FeatureUserManagement  ProviderFeature = "user_management"

	// Advanced features.
	FeatureWebhooks   ProviderFeature = "webhooks"
	FeatureDeployKeys ProviderFeature = "deploy_keys"
	FeatureGitLFS     ProviderFeature = "git_lfs"
	FeatureTopics     ProviderFeature = "topics"
	FeatureReleases   ProviderFeature = "releases"
)

// ProviderCapabilities represents the capabilities of a provider.
type ProviderCapabilities struct {
	MaxRepositoriesPerRequest int
	MaxFileSize               int64
	MaxRepositorySize         int64
	RateLimitPerHour          int
	SupportsSSH               bool
	SupportsHTTPS             bool
	SupportsPrivateRepos      bool
	SupportsOrganizations     bool
}

// RepositoryFilter provides functional filtering of repositories.
type RepositoryFilter interface {
	// Filter methods that can be chained
	ByName(patterns []string) RepositoryFilter
	ByVisibility(visibility string) RepositoryFilter
	ByActivity(since time.Time) RepositoryFilter
	BySize(minSize, maxSize int64) RepositoryFilter
	ByLanguage(languages []string) RepositoryFilter
	ByTopics(topics []string) RepositoryFilter
	ExcludeForks() RepositoryFilter
	ExcludeArchived() RepositoryFilter

	// Terminal operation
	Apply(repositories []entities.Repository) []entities.Repository
}

// ProviderFactory creates provider instances based on configuration.
type ProviderFactory interface {
	CreateProvider(config ProviderConfig) (RepositoryProvider, error)
	SupportedProviders() []string
	GetProviderDefaults(providerType string) ProviderConfig
}

// RepositorySync coordinates repository synchronization between providers.
type RepositorySync interface {
	// Sync operations
	SyncRepository(ctx context.Context, source RepositoryProvider, target RepositoryProvider, repo entities.Repository, options SyncOptions) (SyncResult, error)
	SyncRepositories(ctx context.Context, source RepositoryProvider, target RepositoryProvider, repos []entities.Repository, options SyncOptions) ([]SyncResult, error)

	// Validation
	ValidateSync(source RepositoryProvider, target RepositoryProvider, repos []entities.Repository) ([]ValidationError, error)

	// Status
	GetSyncStatus(ctx context.Context, target RepositoryProvider, repo entities.Repository) (SyncStatus, error)
}

// SyncOptions contains options for repository synchronization.
type SyncOptions struct {
	DryRun               bool
	ForceUpdate          bool
	SyncDescription      bool
	SyncTopics           bool
	SyncVisibility       bool
	SyncDefaultBranch    bool
	SyncBranchProtection bool
	CreateIfNotExists    bool
	UpdateIfExists       bool
	DeleteExtraBranches  bool
	PreservePullRequests bool
	PreserveIssues       bool
	TransformName        NameTransformOptions
}

// SyncResult represents the result of a repository sync operation.
type SyncResult struct {
	Repository       entities.Repository
	Success          bool
	Created          bool
	Updated          bool
	Errors           []error
	Warnings         []string
	Duration         time.Duration
	BytesTransferred int64
	Metadata         map[string]interface{}
}

// SyncStatus represents the current sync status of a repository.
type SyncStatus struct {
	Repository   entities.Repository
	InSync       bool
	LastSyncTime time.Time
	NextSyncTime time.Time
	SyncErrors   []error
	Differences  []SyncDifference
}

// SyncDifference represents a difference between source and target repositories.
type SyncDifference struct {
	Field          string
	SourceValue    interface{}
	TargetValue    interface{}
	DifferenceType DifferenceType
	CanAutoResolve bool
}

// DifferenceType represents the type of difference found.
type DifferenceType string

const (
	DifferenceTypeMetadata    DifferenceType = "metadata"
	DifferenceTypeBranches    DifferenceType = "branches"
	DifferenceTypeProtection  DifferenceType = "protection"
	DifferenceTypeContent     DifferenceType = "content"
	DifferenceTypePermissions DifferenceType = "permissions"
)

// ValidationError represents a validation error during sync preparation.
type ValidationError struct {
	Repository string
	Field      string
	Err        error
	Severity   ValidationSeverity
}

// ValidationSeverity represents the severity of a validation error.
type ValidationSeverity string

const (
	ValidationSeverityError   ValidationSeverity = "error"
	ValidationSeverityWarning ValidationSeverity = "warning"
	ValidationSeverityInfo    ValidationSeverity = "info"
)

// Error implements the error interface for ValidationError.
func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s validation %s for repository %s in field %s: %v",
		ve.Severity, "error", ve.Repository, ve.Field, ve.Err)
}

// Helper functions for working with providers

// NewProviderConfig creates a new provider configuration.
func NewProviderConfig(providerType, domain, owner string) ProviderConfig {
	return ProviderConfig{
		ProviderType: providerType,
		Domain:       domain,
		Owner:        owner,
	}
}

// WithAuth adds authentication to provider config.
func (pc ProviderConfig) WithAuth(auth AuthenticationConfig) ProviderConfig {
	pc.AuthConfig = auth

	return pc
}

// WithToken adds token authentication to provider config.
func (pc ProviderConfig) WithToken(token, username string) ProviderConfig {
	pc.AuthConfig = AuthenticationConfig{
		Token:    token,
		Username: username,
	}

	return pc
}

// WithSSHKey adds SSH key authentication to provider config.
func (pc ProviderConfig) WithSSHKey(keyPath, username string) ProviderConfig {
	pc.AuthConfig = AuthenticationConfig{
		SSHKeyPath: keyPath,
		Username:   username,
	}

	return pc
}

// Validate validates the provider configuration.
func (pc ProviderConfig) Validate() error {
	if pc.ProviderType == "" {
		return errors.New("provider type is required")
	}

	if pc.Owner == "" {
		return errors.New("owner is required")
	}

	return nil
}
