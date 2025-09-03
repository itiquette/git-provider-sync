// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Service validates configurations using pure functions.
type Service struct {
	connectivityValidator ConnectivityValidator
	fileSystemValidator   FileSystemValidator
	config                Config
}

// Config contains configuration for validation service.
type Config struct {
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
	config Config,
) *Service {
	return &Service{
		connectivityValidator: connectivityValidator,
		fileSystemValidator:   fileSystemValidator,
		config:                config,
	}
}

// ComprehensiveResult represents the result of complete validation.
type ComprehensiveResult struct {
	ConfigurationResults Results
	ConnectivityResults  []ConnectivityResult
	FileSystemResults    []FileSystemResult
	RepositoryResults    []RepositoryResult
	OverallSuccess       bool
	TotalErrors          int
	TotalWarnings        int
	Duration             time.Duration
	Summary              Summary
}

// Summary provides a high-level summary of validation results.
type Summary struct {
	ConfigurationValid   bool
	ConnectivityValid    bool
	FileSystemValid      bool
	RepositoryNamesValid bool
	CriticalIssues       []string
	Warnings             []string
	Suggestions          []string
}

// RepositoryResult represents repository-specific validation results.
type RepositoryResult struct {
	RepositoryName string
	ProviderType   string
	Results        []Result
	Valid          bool
}

// NewDefaultValidationService creates a validation service with default configuration.
func NewDefaultValidationService(
	connectivityValidator ConnectivityValidator,
	fileSystemValidator FileSystemValidator,
) *Service {
	config := Config{
		EnableConnectivityTests: true,
		EnableFileSystemTests:   true,
		ConnectivityTimeout:     30 * time.Second,
		SkipOptionalTests:       false,
		MaxConcurrentTests:      5,
	}

	return NewService(connectivityValidator, fileSystemValidator, config)
}

// NewQuickValidationService creates a validation service for quick validation.
func NewQuickValidationService() *Service {
	config := Config{
		EnableConnectivityTests: false,
		EnableFileSystemTests:   false,
		ConnectivityTimeout:     5 * time.Second,
		SkipOptionalTests:       true,
		MaxConcurrentTests:      1,
	}

	return NewService(nil, nil, config)
}

// NewFullValidationService creates a validation service with all validators enabled.
func NewFullValidationService(
	connectivityValidator ConnectivityValidator,
	fileSystemValidator FileSystemValidator,
) *Service {
	config := Config{
		EnableConnectivityTests: true,
		EnableFileSystemTests:   true,
		ConnectivityTimeout:     60 * time.Second,
		SkipOptionalTests:       false,
		MaxConcurrentTests:      10,
	}

	return NewService(connectivityValidator, fileSystemValidator, config)
}

// ValidateConfiguration validates the entire configuration.
//
//nolint:cyclop // Multiple validation types and error paths
func (s *Service) ValidateConfiguration(ctx context.Context, config ports.AppConfiguration) ComprehensiveResult {
	start := time.Now()

	result := ComprehensiveResult{
		ConfigurationResults: Results{Valid: true},
		ConnectivityResults:  []ConnectivityResult{},
		FileSystemResults:    []FileSystemResult{},
		RepositoryResults:    []RepositoryResult{},
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
	result.Summary = s.buildSummary(result)

	return result
}

// ValidateEnvironment validates a single environment configuration.
func (s *Service) ValidateEnvironment(_ context.Context, env ports.EnvironmentConfiguration) Results {
	return ValidateEnvironment(env)
}

// ValidateRepositoryName validates a repository name for a specific provider.
func (s *Service) ValidateRepositoryName(name, providerType string) Result {
	return ValidateRepositoryName(name, providerType)
}

// ValidateURL validates a URL format.
func (s *Service) ValidateURL(url string) Result {
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

func (s *Service) validateRepositoryNames(config ports.AppConfiguration) []RepositoryResult {
	results := make([]RepositoryResult, 0, len(config.Environments)*2)

	for envName, env := range config.Environments {
		// Validate source repositories (if we can enumerate them)
		// This is placeholder logic - in practice, you'd get actual repo names from the provider
		sourceResult := RepositoryResult{
			RepositoryName: "source-repos-" + envName,
			ProviderType:   env.Source.ProviderType,
			Results:        []Result{},
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
			mirrorResult := RepositoryResult{
				RepositoryName: mirrorName,
				ProviderType:   mirror.ProviderType,
				Results:        []Result{},
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

//nolint:cyclop // Complex summary building logic with multiple validation result types
func (s *Service) buildSummary(result ComprehensiveResult) Summary {
	summary := Summary{
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

// Utility functions for validation results

// HasCriticalErrors checks if validation results contain critical errors.
func HasCriticalErrors(result ComprehensiveResult) bool {
	return !result.OverallSuccess || result.TotalErrors > 0
}

// GetValidationErrors extracts all error messages from validation results.
//
//nolint:cyclop // Complex error extraction logic with multiple validation result types
func GetValidationErrors(result ComprehensiveResult) []string {
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
func GetValidationWarnings(result ComprehensiveResult) []string {
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
