// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides pure functional validation logic.
package validation

import (
	"net/url"
	"regexp"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ValidationResult represents the result of a validation operation.
type ValidationResult struct {
	Valid      bool
	Field      string
	Code       string
	Message    string
	Value      interface{}
	Suggestion string
}

// ValidationResults represents multiple validation results.
type ValidationResults struct {
	Valid   bool
	Results []ValidationResult
}

// Configuration Validation (Pure Functions)

// ValidateAppConfiguration validates application configuration using pure functions.
func ValidateAppConfiguration(config ports.AppConfiguration) ValidationResults {
	var results []ValidationResult

	// Validate global settings
	globalResults := ValidateGlobalSettings(config.GlobalSettings)
	results = append(results, globalResults.Results...)

	// Validate each environment
	for name, env := range config.Environments {
		envResults := ValidateEnvironment(env)
		for _, result := range envResults.Results {
			result.Field = "environments." + name + "." + result.Field
			results = append(results, result)
		}
	}

	return ValidationResults{
		Valid:   len(results) == 0,
		Results: results,
	}
}

// ValidateGlobalSettings validates global settings.
func ValidateGlobalSettings(settings ports.GlobalSettings) ValidationResults {
	var results []ValidationResult

	// Validate log level
	if !isValidLogLevel(string(settings.LogLevel)) {
		results = append(results, ValidationResult{
			Valid:      false,
			Field:      "log_level",
			Code:       "INVALID_LOG_LEVEL",
			Message:    "Invalid log level",
			Value:      settings.LogLevel,
			Suggestion: "Use one of: trace, debug, info, warn, error, fatal",
		})
	}

	// Validate log format
	if !isValidLogFormat(string(settings.LogFormat)) {
		results = append(results, ValidationResult{
			Valid:      false,
			Field:      "log_format",
			Code:       "INVALID_LOG_FORMAT",
			Message:    "Invalid log format",
			Value:      settings.LogFormat,
			Suggestion: "Use one of: json, text, console",
		})
	}

	// Validate cache size
	if settings.MaxCacheSize < 0 {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "max_cache_size",
			Code:    "NEGATIVE_CACHE_SIZE",
			Message: "Cache size cannot be negative",
			Value:   settings.MaxCacheSize,
		})
	}

	// Validate cache TTL
	if settings.CacheTTL < 0 {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "cache_ttl",
			Code:    "NEGATIVE_CACHE_TTL",
			Message: "Cache TTL cannot be negative",
			Value:   settings.CacheTTL,
		})
	}

	return ValidationResults{
		Valid:   len(results) == 0,
		Results: results,
	}
}

// ValidateEnvironment validates environment configuration.
func ValidateEnvironment(env ports.EnvironmentConfiguration) ValidationResults {
	results := make([]ValidationResult, 0)

	// Validate environment name
	if strings.TrimSpace(env.Name) == "" {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "name",
			Code:    "EMPTY_ENVIRONMENT_NAME",
			Message: "Environment name cannot be empty",
		})
	}

	// Validate source configuration
	sourceResults := ValidateSourceConfiguration(env.Source)
	for _, result := range sourceResults.Results {
		result.Field = "source." + result.Field
		results = append(results, result)
	}

	// Validate mirror configurations
	for name, mirror := range env.Mirrors {
		mirrorResults := ValidateMirrorConfiguration(mirror)
		for _, result := range mirrorResults.Results {
			result.Field = "mirrors." + name + "." + result.Field
			results = append(results, result)
		}
	}

	// Validate environment options
	optionsResults := ValidateEnvironmentOptions(env.Options)
	for _, result := range optionsResults.Results {
		result.Field = "options." + result.Field
		results = append(results, result)
	}

	return ValidationResults{
		Valid:   len(results) == 0,
		Results: results,
	}
}

// ValidateSourceConfiguration validates source provider configuration.
func ValidateSourceConfiguration(source ports.SourceConfiguration) ValidationResults {
	results := make([]ValidationResult, 0)

	// Validate provider type
	if !isValidProviderType(source.ProviderType) {
		results = append(results, ValidationResult{
			Valid:      false,
			Field:      "provider_type",
			Code:       "INVALID_PROVIDER_TYPE",
			Message:    "Invalid provider type",
			Value:      source.ProviderType,
			Suggestion: "Use one of: github, gitlab, gitea",
		})
	}

	// Validate domain
	if source.Domain != "" && !isValidDomain(source.Domain) {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "domain",
			Code:    "INVALID_DOMAIN",
			Message: "Invalid domain format",
			Value:   source.Domain,
		})
	}

	// Validate owner
	if !isValidOwner(source.Owner) {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "owner",
			Code:    "INVALID_OWNER",
			Message: "Invalid owner name",
			Value:   source.Owner,
		})
	}

	// Validate authentication
	authResults := ValidateAuthentication(source.Authentication)
	for _, result := range authResults.Results {
		result.Field = "authentication." + result.Field
		results = append(results, result)
	}

	return ValidationResults{
		Valid:   len(results) == 0,
		Results: results,
	}
}

// ValidateMirrorConfiguration validates mirror target configuration.
func ValidateMirrorConfiguration(mirror ports.MirrorConfiguration) ValidationResults {
	var results []ValidationResult

	// Validate mirror name
	if strings.TrimSpace(mirror.Name) == "" {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "name",
			Code:    "EMPTY_MIRROR_NAME",
			Message: "Mirror name cannot be empty",
		})
	}

	// Validate provider type
	if !isValidProviderType(mirror.ProviderType) {
		results = append(results, ValidationResult{
			Valid:      false,
			Field:      "provider_type",
			Code:       "INVALID_PROVIDER_TYPE",
			Message:    "Invalid provider type",
			Value:      mirror.ProviderType,
			Suggestion: "Use one of: github, gitlab, gitea, directory, archive",
		})
	}

	// Validate domain for remote providers
	if isRemoteProvider(mirror.ProviderType) && mirror.Domain != "" && !isValidDomain(mirror.Domain) {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "domain",
			Code:    "INVALID_DOMAIN",
			Message: "Invalid domain format",
			Value:   mirror.Domain,
		})
	}

	// Validate owner for remote providers
	if isRemoteProvider(mirror.ProviderType) && !isValidOwner(mirror.Owner) {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "owner",
			Code:    "INVALID_OWNER",
			Message: "Invalid owner name",
			Value:   mirror.Owner,
		})
	}

	// Validate path for local providers
	if isLocalProvider(mirror.ProviderType) && strings.TrimSpace(mirror.Path) == "" {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "path",
			Code:    "EMPTY_PATH",
			Message: "Path is required for local providers",
		})
	}

	// Validate authentication for remote providers
	if isRemoteProvider(mirror.ProviderType) {
		authResults := ValidateAuthentication(mirror.Authentication)
		for _, result := range authResults.Results {
			result.Field = "authentication." + result.Field
			results = append(results, result)
		}
	}

	return ValidationResults{
		Valid:   len(results) == 0,
		Results: results,
	}
}

// ValidateAuthentication validates authentication configuration.
func ValidateAuthentication(auth ports.AuthenticationConfiguration) ValidationResults {
	var results []ValidationResult

	// Validate auth type
	if !isValidAuthType(string(auth.Type)) {
		results = append(results, ValidationResult{
			Valid:      false,
			Field:      "type",
			Code:       "INVALID_AUTH_TYPE",
			Message:    "Invalid authentication type",
			Value:      auth.Type,
			Suggestion: "Use one of: none, token, basic, ssh",
		})
	}

	// Validate token auth
	if auth.Type == ports.AuthenticationTypeToken && strings.TrimSpace(auth.Token) == "" {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "token",
			Code:    "EMPTY_TOKEN",
			Message: "Token is required for token authentication",
		})
	}

	// Validate basic auth
	if auth.Type == ports.AuthenticationTypeBasic {
		if strings.TrimSpace(auth.Username) == "" {
			results = append(results, ValidationResult{
				Valid:   false,
				Field:   "username",
				Code:    "EMPTY_USERNAME",
				Message: "Username is required for basic authentication",
			})
		}

		if strings.TrimSpace(auth.Password) == "" {
			results = append(results, ValidationResult{
				Valid:   false,
				Field:   "password",
				Code:    "EMPTY_PASSWORD",
				Message: "Password is required for basic authentication",
			})
		}
	}

	// Validate SSH auth
	if auth.Type == ports.AuthenticationTypeSSH {
		if strings.TrimSpace(auth.SSHKeyPath) == "" && strings.TrimSpace(auth.SSHKey) == "" {
			results = append(results, ValidationResult{
				Valid:   false,
				Field:   "ssh_key",
				Code:    "MISSING_SSH_KEY",
				Message: "SSH key path or key content is required for SSH authentication",
			})
		}
	}

	return ValidationResults{
		Valid:   len(results) == 0,
		Results: results,
	}
}

// ValidateEnvironmentOptions validates environment options.
func ValidateEnvironmentOptions(options ports.EnvironmentOptions) ValidationResults {
	var results []ValidationResult

	// Validate max concurrency (0 is treated as unset/default, which is acceptable)
	if options.MaxConcurrency < 0 {
		results = append(results, ValidationResult{
			Valid:      false,
			Field:      "max_concurrency",
			Code:       "INVALID_CONCURRENCY",
			Message:    "Max concurrency cannot be negative",
			Value:      options.MaxConcurrency,
			Suggestion: "Set to a positive integer or 0 for default (recommended: 5-10)",
		})
	}

	// Validate timeout
	if options.Timeout < 0 {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "timeout",
			Code:    "NEGATIVE_TIMEOUT",
			Message: "Timeout cannot be negative",
			Value:   options.Timeout,
		})
	}

	// Validate retry attempts
	if options.RetryAttempts < 0 {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "retry_attempts",
			Code:    "NEGATIVE_RETRY_ATTEMPTS",
			Message: "Retry attempts cannot be negative",
			Value:   options.RetryAttempts,
		})
	}

	// Validate retry delay
	if options.RetryDelay < 0 {
		results = append(results, ValidationResult{
			Valid:   false,
			Field:   "retry_delay",
			Code:    "NEGATIVE_RETRY_DELAY",
			Message: "Retry delay cannot be negative",
			Value:   options.RetryDelay,
		})
	}

	return ValidationResults{
		Valid:   len(results) == 0,
		Results: results,
	}
}

// Repository Validation (Pure Functions)

// ValidateRepositoryName validates repository name for all providers.
func ValidateRepositoryName(name, providerType string) ValidationResult {
	name = strings.TrimSpace(name)

	if name == "" {
		return ValidationResult{
			Valid:   false,
			Code:    "EMPTY_NAME",
			Message: "Repository name cannot be empty",
		}
	}

	switch providerType {
	case "github":
		return validateGitHubRepositoryName(name)
	case "gitlab":
		return validateGitLabRepositoryName(name)
	case "gitea":
		return validateGiteaRepositoryName(name)
	default:
		return validateGenericRepositoryName(name)
	}
}

// URL Validation (Pure Functions)

// ValidateURL validates a URL format.
func ValidateURL(rawURL string) ValidationResult {
	if strings.TrimSpace(rawURL) == "" {
		return ValidationResult{
			Valid:   false,
			Code:    "EMPTY_URL",
			Message: "URL cannot be empty",
		}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ValidationResult{
			Valid:   false,
			Code:    "INVALID_URL_FORMAT",
			Message: "Invalid URL format: " + err.Error(),
			Value:   rawURL,
		}
	}

	if parsedURL.Scheme == "" {
		return ValidationResult{
			Valid:      false,
			Code:       "MISSING_SCHEME",
			Message:    "URL must include a scheme (http/https/git)",
			Value:      rawURL,
			Suggestion: "Add http://, https://, or git:// prefix",
		}
	}

	if parsedURL.Host == "" {
		return ValidationResult{
			Valid:   false,
			Code:    "MISSING_HOST",
			Message: "URL must include a host",
			Value:   rawURL,
		}
	}

	return ValidationResult{Valid: true}
}

// Pure validation helper functions

func isValidLogLevel(level string) bool {
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal"}
	for _, valid := range validLevels {
		if level == valid {
			return true
		}
	}

	return false
}

func isValidLogFormat(format string) bool {
	validFormats := []string{"json", "text", "console"}
	for _, valid := range validFormats {
		if format == valid {
			return true
		}
	}

	return false
}

func isValidProviderType(providerType string) bool {
	validTypes := []string{"github", "gitlab", "gitea", "directory", "archive"}
	for _, valid := range validTypes {
		if providerType == valid {
			return true
		}
	}

	return false
}

func isValidAuthType(authType string) bool {
	validTypes := []string{"none", "token", "basic", "ssh", "oauth"}
	for _, valid := range validTypes {
		if authType == valid {
			return true
		}
	}

	return false
}

func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// Basic domain validation regex
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)

	return domainRegex.MatchString(domain) && len(domain) <= 253
}

func isValidOwner(owner string) bool {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return false
	}

	// Basic owner name validation (alphanumeric, hyphens, underscores)
	ownerRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	return ownerRegex.MatchString(owner) && len(owner) <= 100
}

func isRemoteProvider(providerType string) bool {
	remoteProviders := []string{"github", "gitlab", "gitea"}
	for _, remote := range remoteProviders {
		if providerType == remote {
			return true
		}
	}

	return false
}

func isLocalProvider(providerType string) bool {
	localProviders := []string{"directory", "archive"}
	for _, local := range localProviders {
		if providerType == local {
			return true
		}
	}

	return false
}

// Provider-specific repository name validation

func validateGitHubRepositoryName(name string) ValidationResult {
	if len(name) > 100 {
		return ValidationResult{
			Valid:   false,
			Code:    "NAME_TOO_LONG",
			Message: "GitHub repository name cannot exceed 100 characters",
			Value:   name,
		}
	}

	// GitHub allows alphanumeric, hyphens, underscores, and dots
	githubRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !githubRegex.MatchString(name) {
		return ValidationResult{
			Valid:      false,
			Code:       "INVALID_CHARACTERS",
			Message:    "GitHub repository name contains invalid characters",
			Value:      name,
			Suggestion: "Use only letters, numbers, hyphens, underscores, and dots",
		}
	}

	// Cannot start or end with special characters
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") ||
		strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return ValidationResult{
			Valid:   false,
			Code:    "INVALID_START_END",
			Message: "GitHub repository name cannot start or end with dots or hyphens",
			Value:   name,
		}
	}

	return ValidationResult{Valid: true}
}

func validateGitLabRepositoryName(name string) ValidationResult {
	if len(name) > 100 {
		return ValidationResult{
			Valid:   false,
			Code:    "NAME_TOO_LONG",
			Message: "GitLab repository name cannot exceed 100 characters",
			Value:   name,
		}
	}

	// GitLab has stricter rules
	gitlabRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !gitlabRegex.MatchString(name) {
		return ValidationResult{
			Valid:      false,
			Code:       "INVALID_CHARACTERS",
			Message:    "GitLab repository name contains invalid characters",
			Value:      name,
			Suggestion: "Use only letters, numbers, hyphens, underscores, and dots",
		}
	}

	return ValidationResult{Valid: true}
}

func validateGiteaRepositoryName(name string) ValidationResult {
	if len(name) > 100 {
		return ValidationResult{
			Valid:   false,
			Code:    "NAME_TOO_LONG",
			Message: "Gitea repository name cannot exceed 100 characters",
			Value:   name,
		}
	}

	// Gitea similar to GitHub
	giteaRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !giteaRegex.MatchString(name) {
		return ValidationResult{
			Valid:      false,
			Code:       "INVALID_CHARACTERS",
			Message:    "Gitea repository name contains invalid characters",
			Value:      name,
			Suggestion: "Use only letters, numbers, hyphens, underscores, and dots",
		}
	}

	return ValidationResult{Valid: true}
}

func validateGenericRepositoryName(name string) ValidationResult {
	if len(name) > 100 {
		return ValidationResult{
			Valid:   false,
			Code:    "NAME_TOO_LONG",
			Message: "Repository name cannot exceed 100 characters",
			Value:   name,
		}
	}

	// Generic validation - alphanumeric and basic special characters
	genericRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !genericRegex.MatchString(name) {
		return ValidationResult{
			Valid:      false,
			Code:       "INVALID_CHARACTERS",
			Message:    "Repository name contains invalid characters",
			Value:      name,
			Suggestion: "Use only letters, numbers, hyphens, underscores, and dots",
		}
	}

	return ValidationResult{Valid: true}
}
