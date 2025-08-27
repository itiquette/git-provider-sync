// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ValidateSyncUseCase validates synchronization operations before execution.
// This is a pure functional validator that ensures sync operations will succeed.
type ValidateSyncUseCase struct {
	repositoryProvider ports.RepositoryProvider
	config             ports.Configuration
}

// NewValidateSyncUseCase creates a new validate sync use case with explicit dependencies.
func NewValidateSyncUseCase(
	repoProvider ports.RepositoryProvider,
	cfg ports.Configuration,
) ValidateSyncUseCase {
	return ValidateSyncUseCase{
		repositoryProvider: repoProvider,
		config:             cfg,
	}
}

// ValidateSyncRequest represents the input for sync validation.
type ValidateSyncRequest struct {
	SourceConfig  ports.ProviderConfig
	MirrorTargets []entities.MirrorTarget
	Options       ValidationOptions
}

// ValidationOptions contains options that control validation behavior.
type ValidationOptions struct {
	CheckConnectivity     bool
	CheckAuthentication   bool
	CheckRepositoryAccess bool
	CheckNamingConflicts  bool
	CheckQuotaLimits      bool
	StrictMode            bool
}

// ValidateSyncResponse represents the result of sync validation.
type ValidateSyncResponse struct {
	Valid              bool
	Errors             []ValidationError
	Warnings           []ValidationWarning
	RepositoryCount    int
	EstimatedDuration  string
	RecommendedActions []string
}

// ValidationError represents a validation error that will prevent sync.
type ValidationError struct {
	Type          ValidationErrorType
	Component     string
	Message       string
	Field         string
	Value         interface{}
	Severity      ValidationSeverity
	CanAutoFix    bool
	FixSuggestion string
}

// ValidationWarning represents a validation warning that may affect sync.
type ValidationWarning struct {
	Type           ValidationWarningType
	Component      string
	Message        string
	Impact         string
	Recommendation string
}

// ValidationErrorType represents different types of validation errors.
type ValidationErrorType string

const (
	// ErrorTypeConfiguration indicates configuration validation errors.
	ErrorTypeConfiguration ValidationErrorType = "configuration"
	// ErrorTypeAuthentication indicates authentication validation errors.
	ErrorTypeAuthentication ValidationErrorType = "authentication"
	// ErrorTypeConnectivity indicates connectivity validation errors.
	ErrorTypeConnectivity ValidationErrorType = "connectivity"
	// ErrorTypePermissions indicates permissions validation errors.
	ErrorTypePermissions ValidationErrorType = "permissions"
	// ErrorTypeNaming indicates naming validation errors.
	ErrorTypeNaming ValidationErrorType = "naming"
	// ErrorTypeQuota indicates quota validation errors.
	ErrorTypeQuota ValidationErrorType = "quota"
	// ErrorTypeCompatibility indicates compatibility validation errors.
	ErrorTypeCompatibility ValidationErrorType = "compatibility"
)

// ValidationWarningType represents different types of validation warnings.
type ValidationWarningType string

const (
	// WarningTypePerformance indicates performance-related warnings.
	WarningTypePerformance ValidationWarningType = "performance"
	// WarningTypeNaming indicates naming-related warnings.
	WarningTypeNaming ValidationWarningType = "naming"
	// WarningTypeQuota indicates quota-related warnings.
	WarningTypeQuota ValidationWarningType = "quota"
	// WarningTypeCompatibility indicates compatibility warnings.
	WarningTypeCompatibility ValidationWarningType = "compatibility"
	// WarningTypeRecommendation indicates recommendation warnings.
	WarningTypeRecommendation ValidationWarningType = "recommendation"
)

// ValidationSeverity represents the severity of validation issues.
type ValidationSeverity string

const (
	// SeverityCritical indicates critical severity level.
	SeverityCritical ValidationSeverity = "critical"
	// SeverityHigh indicates high severity level.
	SeverityHigh ValidationSeverity = "high"
	// SeverityMedium indicates medium severity level.
	SeverityMedium ValidationSeverity = "medium"
	// SeverityLow indicates low severity level.
	SeverityLow ValidationSeverity = "low"
)

// Execute performs comprehensive validation of the sync operation.
func (uc ValidateSyncUseCase) Execute(
	ctx context.Context,
	request ValidateSyncRequest,
) (ValidateSyncResponse, error) {
	var errors []ValidationError

	var warnings []ValidationWarning

	var recommendations []string

	// Validate source configuration
	sourceErrors := uc.validateSourceConfig(request.SourceConfig)
	errors = append(errors, sourceErrors...)

	// Validate mirror targets
	mirrorErrors, mirrorWarnings := uc.validateMirrorTargets(request.MirrorTargets)
	errors = append(errors, mirrorErrors...)
	warnings = append(warnings, mirrorWarnings...)

	// Validate connectivity if requested
	if request.Options.CheckConnectivity {
		connErrors := uc.validateConnectivity(ctx, request.SourceConfig, request.MirrorTargets)
		errors = append(errors, connErrors...)
	}

	// Validate authentication if requested
	if request.Options.CheckAuthentication {
		authErrors := uc.validateAuthentication(ctx, request.SourceConfig, request.MirrorTargets)
		errors = append(errors, authErrors...)
	}

	// Get repository count estimate
	repoCount := 0

	if len(errors) == 0 || !request.Options.StrictMode {
		count, err := uc.estimateRepositoryCount(ctx, request.SourceConfig)
		if err != nil {
			warnings = append(warnings, ValidationWarning{
				Type:    WarningTypePerformance,
				Message: "Could not estimate repository count: " + err.Error(),
				Impact:  "Duration estimate may be inaccurate",
			})
		} else {
			repoCount = count
		}
	}

	// Generate recommendations
	recommendations = uc.generateRecommendations(errors, warnings, request.MirrorTargets, repoCount)

	// Calculate estimated duration
	estimatedDuration := uc.calculateEstimatedDuration(repoCount, len(request.MirrorTargets))

	response := ValidateSyncResponse{
		Valid:              len(errors) == 0,
		Errors:             errors,
		Warnings:           warnings,
		RepositoryCount:    repoCount,
		EstimatedDuration:  estimatedDuration,
		RecommendedActions: recommendations,
	}

	return response, nil
}

// validateSourceConfig validates the source provider configuration.
func (uc ValidateSyncUseCase) validateSourceConfig(config ports.ProviderConfig) []ValidationError {
	var errors []ValidationError

	// Validate required fields
	if err := config.Validate(); err != nil {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeConfiguration,
			Component: "source",
			Message:   err.Error(),
			Severity:  SeverityCritical,
		})
	}

	// Validate provider type
	if !isValidProviderType(config.ProviderType) {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeConfiguration,
			Component: "source.provider_type",
			Message:   "unsupported provider type: " + config.ProviderType,
			Field:     "provider_type",
			Value:     config.ProviderType,
			Severity:  SeverityCritical,
		})
	}

	// Validate domain format
	if config.Domain != "" && !isValidDomain(config.Domain) {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeConfiguration,
			Component: "source.domain",
			Message:   "invalid domain format",
			Field:     "domain",
			Value:     config.Domain,
			Severity:  SeverityHigh,
		})
	}

	// Validate owner format
	if !isValidOwnerName(config.Owner) {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeConfiguration,
			Component: "source.owner",
			Message:   "invalid owner name format",
			Field:     "owner",
			Value:     config.Owner,
			Severity:  SeverityHigh,
		})
	}

	// Check for authentication
	if !hasValidAuthentication(config.AuthConfig) {
		errors = append(errors, ValidationError{
			Type:          ErrorTypeAuthentication,
			Component:     "source.auth",
			Message:       "no valid authentication provided",
			Severity:      SeverityCritical,
			CanAutoFix:    false,
			FixSuggestion: "provide a valid token or SSH key",
		})
	}

	return errors
}

// validateMirrorTargets validates all mirror target configurations.
func (uc ValidateSyncUseCase) validateMirrorTargets(mirrors []entities.MirrorTarget) ([]ValidationError, []ValidationWarning) {
	var errors []ValidationError

	var warnings []ValidationWarning

	if len(mirrors) == 0 {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeConfiguration,
			Component: "mirrors",
			Message:   "no mirror targets specified",
			Severity:  SeverityCritical,
		})

		return errors, warnings
	}

	// Performance warning for too many mirrors
	if len(mirrors) > 10 {
		warnings = append(warnings, ValidationWarning{
			Type:           WarningTypePerformance,
			Component:      "mirrors",
			Message:        fmt.Sprintf("High number of mirror targets (%d) may impact performance", len(mirrors)),
			Impact:         "Increased sync time and resource usage",
			Recommendation: "Consider grouping mirrors or using staged sync approach",
		})
	}

	// Check for duplicate mirror names and gather stats
	names := make(map[string]bool)
	providerTypes := make(map[entities.ProviderType]int)

	for index, mirror := range mirrors {
		if names[mirror.Name()] {
			errors = append(errors, ValidationError{
				Type:      ErrorTypeConfiguration,
				Component: fmt.Sprintf("mirrors[%d]", index),
				Message:   "duplicate mirror name: " + mirror.Name(),
				Field:     "name",
				Value:     mirror.Name(),
				Severity:  SeverityHigh,
			})
		}

		names[mirror.Name()] = true
		providerTypes[mirror.ProviderType()]++

		// Validate individual mirror
		mirrorErrors := uc.validateSingleMirror(mirror, index)
		errors = append(errors, mirrorErrors...)
	}

	// Warning for single provider type (lack of diversity)
	if len(providerTypes) == 1 && len(mirrors) > 1 {
		for providerType := range providerTypes {
			warnings = append(warnings, ValidationWarning{
				Type:           WarningTypeRecommendation,
				Component:      "mirrors",
				Message:        fmt.Sprintf("All mirrors use the same provider type (%s)", providerType),
				Impact:         "Single point of failure if provider becomes unavailable",
				Recommendation: "Consider adding mirrors with different provider types for redundancy",
			})
		}
	}

	return errors, warnings
}

// validateSingleMirror validates a single mirror target.
func (uc ValidateSyncUseCase) validateSingleMirror(mirror entities.MirrorTarget, index int) []ValidationError {
	var errors []ValidationError

	component := fmt.Sprintf("mirrors[%d]", index)

	// Validate mirror configuration
	if err := mirror.Validate(); err != nil {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeConfiguration,
			Component: component,
			Message:   err.Error(),
			Severity:  SeverityHigh,
		})
	}

	// Validate provider-specific requirements
	switch mirror.ProviderType() {
	case entities.ProviderTypeGitHub, entities.ProviderTypeGitLab, entities.ProviderTypeGitea:
		if !hasValidAuthentication(convertMirrorAuth(mirror.AuthConfig())) {
			errors = append(errors, ValidationError{
				Type:      ErrorTypeAuthentication,
				Component: component + ".auth",
				Message:   "authentication required for git provider",
				Severity:  SeverityCritical,
			})
		}

		if !isValidOwnerName(mirror.Owner()) {
			errors = append(errors, ValidationError{
				Type:      ErrorTypeConfiguration,
				Component: component + ".owner",
				Message:   "invalid owner name format",
				Field:     "owner",
				Value:     mirror.Owner(),
				Severity:  SeverityHigh,
			})
		}

	case entities.ProviderTypeDirectory:
		if !isValidDirectoryPath(mirror.Path()) {
			errors = append(errors, ValidationError{
				Type:      ErrorTypeConfiguration,
				Component: component + ".path",
				Message:   "invalid directory path",
				Field:     "path",
				Value:     mirror.Path(),
				Severity:  SeverityHigh,
			})
		}

	case entities.ProviderTypeArchive:
		if !isValidArchivePath(mirror.Path()) {
			errors = append(errors, ValidationError{
				Type:      ErrorTypeConfiguration,
				Component: component + ".path",
				Message:   "invalid archive path",
				Field:     "path",
				Value:     mirror.Path(),
				Severity:  SeverityHigh,
			})
		}
	}

	return errors
}

// validateConnectivity validates connectivity to source and mirror providers.
func (uc ValidateSyncUseCase) validateConnectivity(
	ctx context.Context,
	source ports.ProviderConfig,
	mirrors []entities.MirrorTarget,
) []ValidationError {
	var errors []ValidationError

	// Test source connectivity
	if !uc.testProviderConnectivity(ctx, source) {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeConnectivity,
			Component: "source",
			Message:   "cannot connect to source provider",
			Severity:  SeverityCritical,
		})
	}

	// Test mirror connectivity
	for index, mirror := range mirrors {
		if mirror.ProviderType() == entities.ProviderTypeDirectory ||
			mirror.ProviderType() == entities.ProviderTypeArchive {
			continue // Skip connectivity tests for local targets
		}

		mirrorConfig := convertMirrorToProviderConfig(mirror)
		if !uc.testProviderConnectivity(ctx, mirrorConfig) {
			errors = append(errors, ValidationError{
				Type:      ErrorTypeConnectivity,
				Component: fmt.Sprintf("mirrors[%d]", index),
				Message:   "cannot connect to mirror provider",
				Severity:  SeverityCritical,
			})
		}
	}

	return errors
}

// validateAuthentication validates authentication for all providers.
func (uc ValidateSyncUseCase) validateAuthentication(
	ctx context.Context,
	source ports.ProviderConfig,
	mirrors []entities.MirrorTarget,
) []ValidationError {
	var errors []ValidationError

	// Test source authentication
	if !uc.testProviderAuthentication(ctx, source) {
		errors = append(errors, ValidationError{
			Type:      ErrorTypeAuthentication,
			Component: "source",
			Message:   "authentication failed for source provider",
			Severity:  SeverityCritical,
		})
	}

	// Test mirror authentication
	for index, mirror := range mirrors {
		if mirror.ProviderType() == entities.ProviderTypeDirectory ||
			mirror.ProviderType() == entities.ProviderTypeArchive {
			continue // Skip auth tests for local targets
		}

		mirrorConfig := convertMirrorToProviderConfig(mirror)
		if !uc.testProviderAuthentication(ctx, mirrorConfig) {
			errors = append(errors, ValidationError{
				Type:      ErrorTypeAuthentication,
				Component: fmt.Sprintf("mirrors[%d]", index),
				Message:   "authentication failed for mirror provider",
				Severity:  SeverityCritical,
			})
		}
	}

	return errors
}

// estimateRepositoryCount estimates the number of repositories to sync.
func (uc ValidateSyncUseCase) estimateRepositoryCount(
	ctx context.Context,
	config ports.ProviderConfig,
) (int, error) {
	repositories, err := uc.repositoryProvider.ListRepositories(ctx, config)
	if err != nil {
		return 0, fmt.Errorf("failed to list repositories for validation: %w", err)
	}

	return len(repositories), nil
}

// generateRecommendations generates actionable recommendations.
//
//nolint:cyclop // Complex recommendation generation logic with multiple validation types
func (uc ValidateSyncUseCase) generateRecommendations(
	errors []ValidationError,
	_ []ValidationWarning,
	mirrors []entities.MirrorTarget,
	repoCount int,
) []string {
	var recommendations []string

	// Recommendations based on errors
	for _, err := range errors {
		if err.CanAutoFix && err.FixSuggestion != "" {
			recommendations = append(recommendations, err.FixSuggestion)
		}
	}

	// Performance recommendations
	if repoCount > 100 && len(mirrors) > 1 {
		recommendations = append(recommendations,
			"Consider using parallel execution for large repository sets")
	}

	if repoCount > 1000 {
		recommendations = append(recommendations,
			"Consider filtering repositories to reduce sync time")
	}

	// Security recommendations
	hasGitProvider := false

	for _, mirror := range mirrors {
		if mirror.ProviderType() == entities.ProviderTypeGitHub ||
			mirror.ProviderType() == entities.ProviderTypeGitLab ||
			mirror.ProviderType() == entities.ProviderTypeGitea {
			hasGitProvider = true

			break
		}
	}

	if hasGitProvider {
		recommendations = append(recommendations,
			"Ensure authentication tokens have minimal required permissions")
	}

	return recommendations
}

// calculateEstimatedDuration calculates estimated sync duration.
func (uc ValidateSyncUseCase) calculateEstimatedDuration(repoCount, mirrorCount int) string {
	// Simple estimation: 30 seconds per repository per mirror
	totalOperations := repoCount * mirrorCount
	estimatedSeconds := totalOperations * 30

	if estimatedSeconds < 60 {
		return fmt.Sprintf("%d seconds", estimatedSeconds)
	}

	if estimatedSeconds < 3600 {
		return fmt.Sprintf("%d minutes", estimatedSeconds/60)
	}

	hours := estimatedSeconds / 3600
	minutes := (estimatedSeconds % 3600) / 60

	return fmt.Sprintf("%d hours %d minutes", hours, minutes)
}

// Helper functions

// isValidProviderType checks if a provider type is valid.
func isValidProviderType(providerType string) bool {
	validTypes := map[string]bool{
		"github": true,
		"gitlab": true,
		"gitea":  true,
	}

	return validTypes[strings.ToLower(providerType)]
}

// isValidDomain checks if a domain is valid.
func isValidDomain(domain string) bool {
	// Simple domain validation
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\.([a-zA-Z]{2,}|[a-zA-Z]{2,}\.[a-zA-Z]{2,})$`)

	return domainRegex.MatchString(domain)
}

// isValidOwnerName checks if an owner name is valid.
func isValidOwnerName(owner string) bool {
	if owner == "" {
		return false
	}

	// GitHub/GitLab style username validation
	ownerRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

	return ownerRegex.MatchString(owner) && len(owner) <= 39
}

// isValidDirectoryPath checks if a directory path is valid.
func isValidDirectoryPath(path string) bool {
	return path != "" && !strings.Contains(path, "..")
}

// isValidArchivePath checks if an archive path is valid.
func isValidArchivePath(path string) bool {
	return path != "" && !strings.Contains(path, "..")
}

// hasValidAuthentication checks if authentication configuration is valid.
func hasValidAuthentication(auth ports.AuthenticationConfig) bool {
	return auth.Token != "" || auth.SSHKeyPath != "" || auth.SSHKey != ""
}

// testProviderConnectivity tests connectivity to a provider.
func (uc ValidateSyncUseCase) testProviderConnectivity(_ context.Context, _ ports.ProviderConfig) bool {
	// This would implement actual connectivity testing
	// For now, return true as a placeholder
	return true
}

// testProviderAuthentication tests authentication to a provider.
func (uc ValidateSyncUseCase) testProviderAuthentication(_ context.Context, _ ports.ProviderConfig) bool {
	// This would implement actual authentication testing
	// For now, return true as a placeholder
	return true
}

// convertMirrorAuth converts mirror auth config to provider auth config.
func convertMirrorAuth(auth entities.AuthConfig) ports.AuthenticationConfig {
	return ports.AuthenticationConfig{
		Token:      auth.Token(),
		Username:   auth.Username(),
		SSHKeyPath: auth.SSHKeyPath(),
		SSHKey:     auth.SSHKey(),
	}
}

// convertMirrorToProviderConfig converts mirror target to provider config.
func convertMirrorToProviderConfig(mirror entities.MirrorTarget) ports.ProviderConfig {
	return ports.ProviderConfig{
		ProviderType: string(mirror.ProviderType()),
		Domain:       mirror.Domain(),
		Owner:        mirror.Owner(),
		AuthConfig:   convertMirrorAuth(mirror.AuthConfig()),
	}
}
