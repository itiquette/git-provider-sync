// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Service provides comprehensive validation using pure functions.
type Service struct {
	connectivityValidator ConnectivityValidator
	fileSystemValidator   FileSystemValidator
	config                ValidationConfig
}

// ValidationConfig contains configuration for validation service.
type ValidationConfig struct {
	EnableConnectivityTests bool
	EnableFileSystemTests   bool
	ConnectivityTimeout     time.Duration
	SkipOptionalTests       bool
	MaxConcurrentTests      int
}

// FileSystemValidator executes file system validations.
type FileSystemValidator interface {
	ValidateFileSystem(ctx context.Context, validation FileSystemValidation) FileSystemResult
}

// NewService creates a new validation service.
func NewService(
	connectivityValidator ConnectivityValidator,
	fileSystemValidator FileSystemValidator,
	config ValidationConfig,
) *Service {
	return &Service{
		connectivityValidator: connectivityValidator,
		fileSystemValidator:   fileSystemValidator,
		config:                config,
	}
}

// ComprehensiveValidationResult represents the result of complete validation.
type ComprehensiveValidationResult struct {
	ConfigurationResults ValidationResults
	ConnectivityResults  []ConnectivityResult
	FileSystemResults    []FileSystemResult
	RepositoryResults    []RepositoryValidationResult
	OverallSuccess       bool
	TotalErrors          int
	TotalWarnings        int
	Duration             time.Duration
	Summary              ValidationSummary
}

// ValidationSummary provides a high-level summary of validation results.
type ValidationSummary struct {
	ConfigurationValid   bool
	ConnectivityValid    bool
	FileSystemValid      bool
	RepositoryNamesValid bool
	CriticalIssues       []string
	Warnings             []string
	Suggestions          []string
}

// RepositoryValidationResult represents repository-specific validation results.
type RepositoryValidationResult struct {
	RepositoryName string
	ProviderType   string
	Results        []ValidationResult
	Valid          bool
}

// ValidateConfiguration performs comprehensive configuration validation.
func (s *Service) ValidateConfiguration(ctx context.Context, config ports.AppConfiguration) ComprehensiveValidationResult {
	start := time.Now()

	result := ComprehensiveValidationResult{
		ConfigurationResults: ValidationResults{Valid: true},
		ConnectivityResults:  []ConnectivityResult{},
		FileSystemResults:    []FileSystemResult{},
		RepositoryResults:    []RepositoryValidationResult{},
		OverallSuccess:       true,
	}

	// 1. Pure configuration validation (always runs)
	result.ConfigurationResults = ValidateAppConfiguration(config)
	if !result.ConfigurationResults.Valid {
		result.OverallSuccess = false
		result.TotalErrors += len(result.ConfigurationResults.Results)
	}

	// 2. Repository name validation
	repoResults := s.validateRepositoryNames(config)
	result.RepositoryResults = repoResults

	for _, repoResult := range repoResults {
		if !repoResult.Valid {
			result.OverallSuccess = false
			result.TotalErrors += len(repoResult.Results)
		}
	}

	// 3. Connectivity validation (optional)
	if s.config.EnableConnectivityTests {
		connectivityValidations := PlanConnectivityValidations(config)
		if len(connectivityValidations) > 0 {
			result.ConnectivityResults = ValidateAllConnectivity(ctx, s.connectivityValidator, connectivityValidations)

			totalFailed, requiredFailed, _ := CountConnectivityFailures(result.ConnectivityResults)
			if requiredFailed > 0 {
				result.OverallSuccess = false
				result.TotalErrors += requiredFailed
			}

			result.TotalWarnings += (totalFailed - requiredFailed)
		}
	}

	// 4. File system validation (optional)
	if s.config.EnableFileSystemTests {
		fileSystemValidations := PlanFileSystemValidations(config)
		if len(fileSystemValidations) > 0 {
			result.FileSystemResults = s.validateAllFileSystem(ctx, fileSystemValidations)

			for _, fsResult := range result.FileSystemResults {
				if !fsResult.Success && fsResult.Validation.Required {
					result.OverallSuccess = false
					result.TotalErrors++
				} else if !fsResult.Success {
					result.TotalWarnings++
				}
			}
		}
	}

	result.Duration = time.Since(start)
	result.Summary = s.buildValidationSummary(result)

	return result
}

// ValidateEnvironment validates a single environment configuration.
func (s *Service) ValidateEnvironment(ctx context.Context, env ports.EnvironmentConfiguration) ValidationResults {
	return ValidateEnvironment(env)
}

// ValidateRepositoryName validates a repository name for a specific provider.
func (s *Service) ValidateRepositoryName(name, providerType string) ValidationResult {
	return ValidateRepositoryName(name, providerType)
}

// ValidateURL validates a URL format.
func (s *Service) ValidateURL(url string) ValidationResult {
	return ValidateURL(url)
}

// TestConnectivity tests connectivity to a specific provider.
func (s *Service) TestConnectivity(ctx context.Context, providerType, domain, owner string, auth ports.AuthenticationConfiguration) []ConnectivityResult {
	if !s.config.EnableConnectivityTests {
		return []ConnectivityResult{}
	}

	validations := PlanProviderConnectivity(providerType, domain, owner, auth)

	return ValidateAllConnectivity(ctx, s.connectivityValidator, validations)
}

// Private helper methods

func (s *Service) validateRepositoryNames(config ports.AppConfiguration) []RepositoryValidationResult {
	results := make([]RepositoryValidationResult, 0, len(config.Environments)*2)

	for envName, env := range config.Environments {
		// Validate source repositories (if we can enumerate them)
		// This is placeholder logic - in practice, you'd get actual repo names from the provider
		sourceResult := RepositoryValidationResult{
			RepositoryName: "source-repos-" + envName,
			ProviderType:   env.Source.ProviderType,
			Results:        []ValidationResult{},
			Valid:          true,
		}

		// For demonstration, validate a sample repository name pattern
		if env.Source.Owner != "" {
			nameResult := ValidateRepositoryName("example-repo", env.Source.ProviderType)
			if !nameResult.Valid {
				sourceResult.Results = append(sourceResult.Results, nameResult)
				sourceResult.Valid = false
			}
		}

		results = append(results, sourceResult)

		// Validate mirror repository names
		for mirrorName, mirror := range env.Mirrors {
			mirrorResult := RepositoryValidationResult{
				RepositoryName: mirrorName,
				ProviderType:   mirror.ProviderType,
				Results:        []ValidationResult{},
				Valid:          true,
			}

			// Validate the mirror name as a repository name
			nameResult := ValidateRepositoryName(mirrorName, mirror.ProviderType)
			if !nameResult.Valid {
				mirrorResult.Results = append(mirrorResult.Results, nameResult)
				mirrorResult.Valid = false
			}

			results = append(results, mirrorResult)
		}
	}

	return results
}

func (s *Service) validateAllFileSystem(ctx context.Context, validations []FileSystemValidation) []FileSystemResult {
	results := make([]FileSystemResult, len(validations))

	for i, validation := range validations {
		results[i] = s.fileSystemValidator.ValidateFileSystem(ctx, validation)
	}

	return results
}

func (s *Service) buildValidationSummary(result ComprehensiveValidationResult) ValidationSummary {
	summary := ValidationSummary{
		ConfigurationValid:   result.ConfigurationResults.Valid,
		ConnectivityValid:    true,
		FileSystemValid:      true,
		RepositoryNamesValid: true,
		CriticalIssues:       []string{},
		Warnings:             []string{},
		Suggestions:          []string{},
	}

	// Analyze configuration results
	for _, configResult := range result.ConfigurationResults.Results {
		if !configResult.Valid {
			summary.CriticalIssues = append(summary.CriticalIssues,
				"Configuration error in "+configResult.Field+": "+configResult.Message)
			if configResult.Suggestion != "" {
				summary.Suggestions = append(summary.Suggestions, configResult.Suggestion)
			}
		}
	}

	// Analyze connectivity results
	_, requiredFailed, optionalFailed := CountConnectivityFailures(result.ConnectivityResults)
	if requiredFailed > 0 {
		summary.ConnectivityValid = false
		summary.CriticalIssues = append(summary.CriticalIssues,
			"Critical connectivity issues found")
	}

	if optionalFailed > 0 {
		summary.Warnings = append(summary.Warnings,
			"Some optional connectivity tests failed")
	}

	// Analyze file system results
	for _, fsResult := range result.FileSystemResults {
		if !fsResult.Success {
			if fsResult.Validation.Required {
				summary.FileSystemValid = false
				summary.CriticalIssues = append(summary.CriticalIssues,
					"File system error: "+fsResult.Validation.Description)
			} else {
				summary.Warnings = append(summary.Warnings,
					"File system warning: "+fsResult.Validation.Description)
			}
		}
	}

	// Analyze repository name results
	for _, repoResult := range result.RepositoryResults {
		if !repoResult.Valid {
			summary.RepositoryNamesValid = false
			for _, nameResult := range repoResult.Results {
				summary.CriticalIssues = append(summary.CriticalIssues,
					"Repository name error in "+repoResult.RepositoryName+": "+nameResult.Message)
				if nameResult.Suggestion != "" {
					summary.Suggestions = append(summary.Suggestions, nameResult.Suggestion)
				}
			}
		}
	}

	return summary
}

// Factory functions

// NewDefaultValidationService creates a validation service with default configuration.
func NewDefaultValidationService(
	connectivityValidator ConnectivityValidator,
	fileSystemValidator FileSystemValidator,
) *Service {
	config := ValidationConfig{
		EnableConnectivityTests: true,
		EnableFileSystemTests:   true,
		ConnectivityTimeout:     30 * time.Second,
		SkipOptionalTests:       false,
		MaxConcurrentTests:      5,
	}

	return NewService(connectivityValidator, fileSystemValidator, config)
}

// NewQuickValidationService creates a validation service for fast validation (configuration only).
func NewQuickValidationService() *Service {
	config := ValidationConfig{
		EnableConnectivityTests: false,
		EnableFileSystemTests:   false,
		ConnectivityTimeout:     5 * time.Second,
		SkipOptionalTests:       true,
		MaxConcurrentTests:      1,
	}

	return NewService(nil, nil, config)
}

// NewFullValidationService creates a validation service with comprehensive testing enabled.
func NewFullValidationService(
	connectivityValidator ConnectivityValidator,
	fileSystemValidator FileSystemValidator,
) *Service {
	config := ValidationConfig{
		EnableConnectivityTests: true,
		EnableFileSystemTests:   true,
		ConnectivityTimeout:     60 * time.Second,
		SkipOptionalTests:       false,
		MaxConcurrentTests:      10,
	}

	return NewService(connectivityValidator, fileSystemValidator, config)
}

// Utility functions for validation results

// HasCriticalErrors checks if validation results contain critical errors.
func HasCriticalErrors(result ComprehensiveValidationResult) bool {
	return !result.OverallSuccess || result.TotalErrors > 0
}

// GetValidationErrors extracts all error messages from validation results.
func GetValidationErrors(result ComprehensiveValidationResult) []string {
	var errors []string

	// Configuration errors
	for _, configResult := range result.ConfigurationResults.Results {
		if !configResult.Valid {
			errors = append(errors, configResult.Field+": "+configResult.Message)
		}
	}

	// Connectivity errors
	for _, connResult := range result.ConnectivityResults {
		if !connResult.Success && connResult.Validation.Required {
			errors = append(errors, "Connectivity: "+connResult.Validation.Description+" failed")
		}
	}

	// File system errors
	for _, fsResult := range result.FileSystemResults {
		if !fsResult.Success && fsResult.Validation.Required {
			errors = append(errors, "File system: "+fsResult.Validation.Description+" failed")
		}
	}

	// Repository name errors
	for _, repoResult := range result.RepositoryResults {
		if !repoResult.Valid {
			for _, nameResult := range repoResult.Results {
				errors = append(errors, "Repository "+repoResult.RepositoryName+": "+nameResult.Message)
			}
		}
	}

	return errors
}

// GetValidationWarnings extracts all warning messages from validation results.
func GetValidationWarnings(result ComprehensiveValidationResult) []string {
	var warnings []string

	// Connectivity warnings
	for _, connResult := range result.ConnectivityResults {
		if !connResult.Success && !connResult.Validation.Required {
			warnings = append(warnings, "Connectivity: "+connResult.Validation.Description+" failed (optional)")
		}
	}

	// File system warnings
	for _, fsResult := range result.FileSystemResults {
		if !fsResult.Success && !fsResult.Validation.Required {
			warnings = append(warnings, "File system: "+fsResult.Validation.Description+" failed (optional)")
		}
	}

	return warnings
}
