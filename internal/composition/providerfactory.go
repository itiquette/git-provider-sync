// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/adapters/providers/gitea"
	"itiquette/git-provider-sync/internal/adapters/providers/github"
	"itiquette/git-provider-sync/internal/adapters/providers/gitlab"
	"itiquette/git-provider-sync/internal/adapters/transport"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProviderFactory creates provider instances with sophisticated configuration.
// This ports the provider client creation logic from main branch to hexagonal architecture.
type ProviderFactory struct {
	httpClientFactory HTTPClientFactory
	logger            ports.Logger
}

// HTTPClientFactory provides HTTP clients with proper configuration.
type HTTPClientFactory interface {
	CreateHTTPClient(ctx context.Context, config HTTPClientConfig) (*http.Client, error)
}

// HTTPClientFactoryAdapter adapts transport.HTTPFactory to HTTPClientFactory interface.
type HTTPClientFactoryAdapter struct {
	httpFactory *transport.HTTPFactory
}

// NewHTTPClientFactoryAdapter creates a new adapter.
func NewHTTPClientFactoryAdapter(httpFactory *transport.HTTPFactory) HTTPClientFactory {
	return &HTTPClientFactoryAdapter{
		httpFactory: httpFactory,
	}
}

// HTTPClientConfig contains configuration for HTTP client creation.
type HTTPClientConfig struct {
	Timeout         time.Duration
	DisableSSL      bool
	TrustDomains    []string
	MaxRetries      int
	RetryWaitMin    time.Duration
	RetryWaitMax    time.Duration
	UserAgent       string
	ProxyURL        string
	CustomHeaders   map[string]string
	RateLimitConfig RateLimitConfig
}

// RateLimitConfig contains rate limiting configuration.
type RateLimitConfig struct {
	Enabled         bool
	RequestsPerHour int
	BurstSize       int
	RetryAfter      time.Duration
}

// ProviderClientConfig contains advanced provider configuration.
// This ports the GitProviderClientOption functionality from main branch.
type ProviderClientConfig struct {
	// Basic configuration
	ProviderType string
	Domain       string
	Owner        string
	AuthConfig   AuthenticationConfig

	// Advanced configuration
	HTTPScheme    string
	UploadURL     string
	APIVersion    string
	DisableSSL    bool
	CustomHeaders map[string]string
	Timeout       time.Duration
	MaxRetries    int
	UserAgent     string
	ProxyURL      string

	// Provider-specific options
	GitHubOptions GitHubProviderOptions
	GitLabOptions GitLabProviderOptions
	GiteaOptions  GiteaProviderOptions
}

// AuthenticationConfig contains authentication settings.
type AuthenticationConfig struct {
	Token      string
	Username   string
	SSHKeyPath string
	SSHKey     string
	HTTPScheme string
}

// GitHubProviderOptions contains GitHub-specific options.
type GitHubProviderOptions struct {
	BaseURL            string
	UploadURL          string
	EnabledRateLimit   bool
	SecondaryRateLimit bool
	AppID              int64
	InstallationID     int64
	PrivateKeyPath     string
}

// GitLabProviderOptions contains GitLab-specific options.
type GitLabProviderOptions struct {
	BaseURL         string
	APIVersion      string
	GroupsEnabled   bool
	ProjectsEnabled bool
	IssuesEnabled   bool
	MergeEnabled    bool
}

// GiteaProviderOptions contains Gitea-specific options.
type GiteaProviderOptions struct {
	BaseURL       string
	APIVersion    string
	AdminToken    string
	OrgMode       bool
	SkipVerifySSL bool
}

// NewProviderFactory creates a new provider factory.
func NewProviderFactory(httpClientFactory HTTPClientFactory, logger ports.Logger) *ProviderFactory {
	return &ProviderFactory{
		httpClientFactory: httpClientFactory,
		logger:            logger,
	}
}

// NewProviderFactoryWithTransport creates a new provider factory using transport.HTTPFactory.
func NewProviderFactoryWithTransport(httpFactory *transport.HTTPFactory, logger ports.Logger) *ProviderFactory {
	adapter := NewHTTPClientFactoryAdapter(httpFactory)

	return NewProviderFactory(adapter, logger)
}

// CreateProvider creates a provider instance based on configuration.
// This ports the sophisticated provider creation logic from main branch.
func (pf *ProviderFactory) CreateProvider(
	ctx context.Context,
	config ProviderClientConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Info(ctx, "Creating provider client", map[string]interface{}{
		"provider_type": config.ProviderType,
		"domain":        config.Domain,
		"owner":         config.Owner,
	})

	// Validate configuration
	if err := pf.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid provider configuration: %w", err)
	}

	// Create HTTP client with advanced configuration
	httpClient, err := pf.createHTTPClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create provider-specific client
	switch strings.ToLower(config.ProviderType) {
	case "github":
		return pf.createGitHubProvider(ctx, httpClient, config)
	case "gitlab":
		return pf.createGitLabProvider(ctx, httpClient, config)
	case "gitea":
		return pf.createGiteaProvider(ctx, httpClient, config)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.ProviderType)
	}
}

// CreateProviderFromConfig creates a provider from ports.ProviderConfig.
func (pf *ProviderFactory) CreateProviderFromConfig(
	ctx context.Context,
	config ports.ProviderConfig,
) (ports.RepositoryProvider, error) {
	clientConfig := ProviderClientConfig{
		ProviderType: config.ProviderType,
		Domain:       config.Domain,
		Owner:        config.Owner,
		AuthConfig: AuthenticationConfig{
			Token:      config.AuthConfig.Token,
			Username:   config.AuthConfig.Username,
			SSHKeyPath: config.AuthConfig.SSHKeyPath,
			SSHKey:     config.AuthConfig.SSHKey,
		},
		HTTPScheme: "https",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		UserAgent:  "git-provider-sync/1.0",
	}

	return pf.CreateProvider(ctx, clientConfig)
}

// validateConfig validates the provider configuration.
func (pf *ProviderFactory) validateConfig(config ProviderClientConfig) error {
	if config.ProviderType == "" {
		return errors.New("provider type is required")
	}

	if config.Owner == "" {
		return errors.New("owner is required")
	}

	// Validate authentication
	if config.AuthConfig.Token == "" && config.AuthConfig.SSHKey == "" && config.AuthConfig.SSHKeyPath == "" {
		return errors.New("authentication is required (token, SSH key, or SSH key path)")
	}

	// Validate domain format
	if config.Domain != "" {
		if !strings.Contains(config.Domain, ".") {
			return fmt.Errorf("invalid domain format: %s", config.Domain)
		}
	}

	return nil
}

// CreateHTTPClient implements HTTPClientFactory interface.
func (a *HTTPClientFactoryAdapter) CreateHTTPClient(ctx context.Context, config HTTPClientConfig) (*http.Client, error) {
	// Convert HTTPClientConfig to transport.HTTPClientOptions
	options := transport.HTTPClientOptions{
		Provider:      "",    // Will be set by provider-specific methods
		Authenticated: false, // Will be handled by provider adapters
		Custom: map[string]interface{}{
			"timeout":        config.Timeout,
			"disable_ssl":    config.DisableSSL,
			"max_retries":    config.MaxRetries,
			"user_agent":     config.UserAgent,
			"proxy_url":      config.ProxyURL,
			"custom_headers": config.CustomHeaders,
		},
	}

	client, err := a.httpFactory.CreateClient(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return client, nil
}

// createHTTPClient creates an HTTP client with advanced configuration.
func (pf *ProviderFactory) createHTTPClient(ctx context.Context, config ProviderClientConfig) (*http.Client, error) {
	httpConfig := HTTPClientConfig{
		Timeout:       config.Timeout,
		DisableSSL:    config.DisableSSL,
		MaxRetries:    config.MaxRetries,
		UserAgent:     config.UserAgent,
		ProxyURL:      config.ProxyURL,
		CustomHeaders: config.CustomHeaders,
	}

	if httpConfig.Timeout == 0 {
		httpConfig.Timeout = 30 * time.Second
	}

	if httpConfig.UserAgent == "" {
		httpConfig.UserAgent = "git-provider-sync/1.0"
	}

	client, err := pf.httpClientFactory.CreateHTTPClient(ctx, httpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return client, nil
}

// createGitHubProvider creates a GitHub provider with advanced configuration.
func (pf *ProviderFactory) createGitHubProvider(
	ctx context.Context,
	httpClient *http.Client,
	config ProviderClientConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Debug(ctx, "Creating GitHub provider", map[string]interface{}{
		"domain":     config.Domain,
		"base_url":   config.GitHubOptions.BaseURL,
		"upload_url": config.GitHubOptions.UploadURL,
	})

	// Use enhanced GitHub adapter creation
	adapter := github.NewWithConfig(ctx, github.Config{
		Token:      config.AuthConfig.Token,
		HTTPClient: httpClient,
		BaseURL:    pf.determineGitHubBaseURL(config),
		UploadURL:  pf.determineGitHubUploadURL(config),
		UserAgent:  config.UserAgent,
	})

	return adapter, nil
}

// createGitLabProvider creates a GitLab provider with advanced configuration.
func (pf *ProviderFactory) createGitLabProvider(
	ctx context.Context,
	httpClient *http.Client,
	config ProviderClientConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Debug(ctx, "Creating GitLab provider", map[string]interface{}{
		"domain":   config.Domain,
		"base_url": config.GitLabOptions.BaseURL,
	})

	// Use enhanced GitLab adapter creation
	adapter, err := gitlab.NewWithConfig(ctx, gitlab.Config{
		Token:      config.AuthConfig.Token,
		HTTPClient: httpClient,
		BaseURL:    pf.determineGitLabBaseURL(config),
		UserAgent:  config.UserAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab adapter: %w", err)
	}

	return adapter, nil
}

// createGiteaProvider creates a Gitea provider with advanced configuration.
func (pf *ProviderFactory) createGiteaProvider(
	ctx context.Context,
	httpClient *http.Client,
	config ProviderClientConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Debug(ctx, "Creating Gitea provider", map[string]interface{}{
		"domain":   config.Domain,
		"base_url": config.GiteaOptions.BaseURL,
	})

	// Use enhanced Gitea adapter creation
	adapter, err := gitea.NewWithConfig(ctx, gitea.Config{
		Token:         config.AuthConfig.Token,
		HTTPClient:    httpClient,
		BaseURL:       pf.determineGiteaBaseURL(config),
		UserAgent:     config.UserAgent,
		SkipVerifySSL: config.GiteaOptions.SkipVerifySSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea adapter: %w", err)
	}

	return adapter, nil
}

// Helper methods for URL determination

func (pf *ProviderFactory) determineGitHubBaseURL(config ProviderClientConfig) string {
	if config.GitHubOptions.BaseURL != "" {
		return config.GitHubOptions.BaseURL
	}

	if config.Domain != "" && config.Domain != "github.com" {
		return pf.buildURL(config.HTTPScheme, config.Domain, "/api/v3/")
	}

	return "https://api.github.com/"
}

func (pf *ProviderFactory) determineGitHubUploadURL(config ProviderClientConfig) string {
	if config.GitHubOptions.UploadURL != "" {
		return config.GitHubOptions.UploadURL
	}

	if config.Domain != "" && config.Domain != "github.com" {
		return pf.buildURL(config.HTTPScheme, config.Domain, "/api/uploads/")
	}

	return "https://uploads.github.com/"
}

func (pf *ProviderFactory) determineGitLabBaseURL(config ProviderClientConfig) string {
	if config.GitLabOptions.BaseURL != "" {
		return config.GitLabOptions.BaseURL
	}

	if config.Domain != "" {
		return pf.buildURL(config.HTTPScheme, config.Domain, "/")
	}

	return "https://gitlab.com/"
}

func (pf *ProviderFactory) determineGiteaBaseURL(config ProviderClientConfig) string {
	if config.GiteaOptions.BaseURL != "" {
		return config.GiteaOptions.BaseURL
	}

	if config.Domain != "" {
		return pf.buildURL(config.HTTPScheme, config.Domain, "/")
	}

	return "https://gitea.com/"
}

func (pf *ProviderFactory) buildURL(scheme, domain, path string) string {
	if scheme == "" {
		scheme = "https"
	}

	parsedURL := &url.URL{
		Scheme: scheme,
		Host:   domain,
		Path:   path,
	}

	return parsedURL.String()
}

// DomainWithScheme returns domain with proper HTTP scheme.
// This ports the DomainWithScheme functionality from main branch.
func (ac AuthenticationConfig) DomainWithScheme(domain string) string {
	scheme := ac.HTTPScheme
	if scheme == "" {
		scheme = "https"
	}

	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return domain
	}

	return fmt.Sprintf("%s://%s", scheme, domain)
}

// GetSupportedProviders returns a list of supported provider types.
func (pf *ProviderFactory) GetSupportedProviders() []string {
	return []string{"github", "gitlab", "gitea"}
}

// ValidateProviderType validates if a provider type is supported.
func (pf *ProviderFactory) ValidateProviderType(providerType string) error {
	supportedProviders := pf.GetSupportedProviders()
	for _, supported := range supportedProviders {
		if strings.EqualFold(providerType, supported) {
			return nil
		}
	}

	return fmt.Errorf("unsupported provider type: %s, supported: %v", providerType, supportedProviders)
}

// CreateProvidersFromMirrorsWithContext creates provider instances from mirror configurations.
func (pf *ProviderFactory) CreateProvidersFromMirrorsWithContext(
	ctx context.Context,
	mirrors map[string]ports.MirrorConfiguration,
) (map[string]ports.RepositoryProvider, error) {
	providers := make(map[string]ports.RepositoryProvider)

	for name, mirror := range mirrors {
		providerConfig := ports.ProviderConfig{
			ProviderType: mirror.ProviderType,
			Domain:       mirror.Domain,
			Owner:        mirror.Owner,
			AuthConfig: ports.AuthenticationConfig{
				Token:      mirror.Authentication.Token,
				Username:   mirror.Authentication.Username,
				SSHKeyPath: mirror.Authentication.SSHKeyPath,
				SSHKey:     mirror.Authentication.SSHKey,
			},
		}

		provider, err := pf.CreateProviderFromConfig(ctx, providerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider for mirror %s: %w", name, err)
		}

		providers[name] = provider
	}

	return providers, nil
}
