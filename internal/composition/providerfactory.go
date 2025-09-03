// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/adapters/providers/gitea"
	"itiquette/git-provider-sync/internal/adapters/providers/github"
	"itiquette/git-provider-sync/internal/adapters/providers/gitlab"
	"itiquette/git-provider-sync/internal/adapters/transport"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProviderFactory creates provider instances with simplified configuration.
type ProviderFactory struct {
	httpFactory *transport.HTTPFactory
	logger      ports.Logger
}

// ProviderConfig contains simplified provider configuration.
type ProviderConfig struct {
	ProviderType string
	Domain       string
	Owner        string
	Token        string
	Username     string
	SSHKeyPath   string
	SSHKey       string
	BaseURL      string
	APIVersion   string
	DisableSSL   bool
	Timeout      time.Duration
	MaxRetries   int
	UserAgent    string
	Options      map[string]any // Provider-specific options
}

// NewProviderFactory creates a new provider factory.
func NewProviderFactory(httpFactory *transport.HTTPFactory, logger ports.Logger) *ProviderFactory {
	return &ProviderFactory{
		httpFactory: httpFactory,
		logger:      logger,
	}
}

// CreateProvider creates a provider instance based on configuration.
//
//nolint:ireturn // Factory method returns interface
func (pf *ProviderFactory) CreateProvider(
	ctx context.Context,
	config ProviderConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Info(ctx, "Creating provider client", map[string]any{
		"provider_type": config.ProviderType,
		"domain":        config.Domain,
		"owner":         config.Owner,
	})

	// Validate configuration
	if err := pf.validateConfig(config); err != nil {
		return nil, fmt.Errorf("provider configuration validation failed for %s: %w", config.ProviderType, err)
	}

	// Create HTTP client
	httpClient, err := pf.createHTTPClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("HTTP client creation failed for %s provider: %w", config.ProviderType, err)
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
		return nil, fmt.Errorf("%w: %s", domain.ErrUnsupportedProviderType, config.ProviderType)
	}
}

// CreateProviderFromConfig creates a provider from ports.ProviderConfig.
//
//nolint:ireturn // Factory method returns interface
func (pf *ProviderFactory) CreateProviderFromConfig(
	ctx context.Context,
	config ports.ProviderConfig,
) (ports.RepositoryProvider, error) {
	clientConfig := ProviderConfig{
		ProviderType: config.ProviderType,
		Domain:       config.Domain,
		Owner:        config.Owner,
		Token:        config.AuthConfig.Token,
		Username:     config.AuthConfig.Username,
		SSHKeyPath:   config.AuthConfig.SSHKeyPath,
		SSHKey:       config.AuthConfig.SSHKey,
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		UserAgent:    "git-provider-sync/1.0",
	}

	return pf.CreateProvider(ctx, clientConfig)
}

// ensureHTTPSScheme returns domain with proper HTTPS scheme.
func ensureHTTPSScheme(domain string) string {
	scheme := "https"

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

	return fmt.Errorf("%w: %s, supported: %v", domain.ErrUnsupportedProviderType, providerType, supportedProviders)
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

// validateConfig validates the provider configuration.
func (pf *ProviderFactory) validateConfig(config ProviderConfig) error {
	if config.ProviderType == "" {
		return domain.ErrProviderTypeRequired
	}

	if config.Owner == "" {
		return domain.ErrOwnerRequired
	}

	// Validate authentication
	if config.Token == "" && config.SSHKey == "" && config.SSHKeyPath == "" {
		return domain.ErrAuthenticationRequired
	}

	// Validate domain format
	if config.Domain != "" {
		if !strings.Contains(config.Domain, ".") {
			return fmt.Errorf("%w: %s", domain.ErrInvalidDomainFormat, config.Domain)
		}
	}

	return nil
}

// createHTTPClient creates an HTTP client with proper configuration.
func (pf *ProviderFactory) createHTTPClient(_ context.Context, config ProviderConfig) (*http.Client, error) {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = "git-provider-sync/1.0"
	}

	options := transport.HTTPClientOptions{
		Provider:      config.ProviderType,
		Authenticated: config.Token != "",
		Custom: map[string]any{
			"timeout":     timeout,
			"disable_ssl": config.DisableSSL,
			"max_retries": maxRetries,
			"user_agent":  userAgent,
		},
	}

	client, err := pf.httpFactory.CreateClient(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return client, nil
}

// createGitHubProvider creates a GitHub provider.
//
//nolint:ireturn // Factory helper method returns interface
func (pf *ProviderFactory) createGitHubProvider(
	ctx context.Context,
	httpClient *http.Client,
	config ProviderConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Debug(ctx, "Creating GitHub provider", map[string]any{
		"domain":   config.Domain,
		"base_url": config.BaseURL,
	})

	// Create GitHub adapter
	adapter := github.NewWithConfig(ctx, github.Config{
		Token:      config.Token,
		HTTPClient: httpClient,
		BaseURL:    pf.determineGitHubBaseURL(config),
		UserAgent:  config.UserAgent,
	})

	return adapter, nil
}

// createGitLabProvider creates a GitLab provider.
//
//nolint:ireturn // Factory helper method returns interface
func (pf *ProviderFactory) createGitLabProvider(
	ctx context.Context,
	httpClient *http.Client,
	config ProviderConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Debug(ctx, "Creating GitLab provider", map[string]any{
		"domain":   config.Domain,
		"base_url": config.BaseURL,
	})

	// Create GitLab adapter
	adapter, err := gitlab.NewWithConfig(ctx, gitlab.Config{
		Token:      config.Token,
		HTTPClient: httpClient,
		BaseURL:    pf.determineGitLabBaseURL(config),
		UserAgent:  config.UserAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab adapter: %w", err)
	}

	return adapter, nil
}

// createGiteaProvider creates a Gitea provider.
//
//nolint:ireturn // Factory helper method returns interface
func (pf *ProviderFactory) createGiteaProvider(
	ctx context.Context,
	httpClient *http.Client,
	config ProviderConfig,
) (ports.RepositoryProvider, error) {
	pf.logger.Debug(ctx, "Creating Gitea provider", map[string]any{
		"domain":   config.Domain,
		"base_url": config.BaseURL,
	})

	// Create Gitea adapter
	adapter, err := gitea.NewWithConfig(ctx, gitea.Config{
		Token:         config.Token,
		HTTPClient:    httpClient,
		BaseURL:       pf.determineGiteaBaseURL(config),
		UserAgent:     config.UserAgent,
		SkipVerifySSL: config.DisableSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea adapter: %w", err)
	}

	return adapter, nil
}

// Helper methods for URL determination

func (pf *ProviderFactory) determineGitHubBaseURL(config ProviderConfig) string {
	if config.BaseURL != "" {
		return config.BaseURL
	}

	if config.Domain != "" && config.Domain != "github.com" {
		return pf.buildURL("https", config.Domain, "/api/v3/")
	}

	return "https://api.github.com/"
}

func (pf *ProviderFactory) determineGitLabBaseURL(config ProviderConfig) string {
	if config.BaseURL != "" {
		return config.BaseURL
	}

	if config.Domain != "" {
		return pf.buildURL("https", config.Domain, "/")
	}

	return "https://gitlab.com/"
}

func (pf *ProviderFactory) determineGiteaBaseURL(config ProviderConfig) string {
	if config.BaseURL != "" {
		return config.BaseURL
	}

	if config.Domain != "" {
		return pf.buildURL("https", config.Domain, "/")
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
