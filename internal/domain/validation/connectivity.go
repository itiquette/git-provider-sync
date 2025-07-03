// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	// Provider type constants.
	ProviderTypeGitHub = "github"
	ProviderTypeGitLab = "gitlab"
	ProviderTypeGitea  = "gitea"

	// Default domain constants.
	DefaultGitHubDomain = "github.com"
	DefaultGitLabDomain = "gitlab.com"
)

// ConnectivityValidation represents a connectivity validation plan.
type ConnectivityValidation struct {
	Type        ConnectivityType
	Target      string
	Timeout     time.Duration
	Description string
	Required    bool
}

// ConnectivityType defines types of connectivity validations.
type ConnectivityType string

const (
	ConnectivityTypeHTTP     ConnectivityType = "http"
	ConnectivityTypeGit      ConnectivityType = "git"
	ConnectivityTypeSSH      ConnectivityType = "ssh"
	ConnectivityTypeProvider ConnectivityType = "provider"
)

// ConnectivityResult represents the result of a connectivity validation.
type ConnectivityResult struct {
	Validation ConnectivityValidation
	Success    bool
	Error      error
	Duration   time.Duration
	Details    map[string]interface{}
}

// ConnectivityValidator executes connectivity validations.
type ConnectivityValidator interface {
	ValidateConnectivity(ctx context.Context, validation ConnectivityValidation) ConnectivityResult
}

// Pure Functions for Connectivity Validation Planning

// PlanConnectivityValidations creates pure validation plans for configuration.
func PlanConnectivityValidations(config ports.AppConfiguration) []ConnectivityValidation {
	var validations []ConnectivityValidation

	for envName, env := range config.Environments {
		if !env.Enabled {
			continue
		}

		// Plan source connectivity validation
		sourceValidations := planSourceConnectivity(env.Source, envName)
		validations = append(validations, sourceValidations...)

		// Plan mirror connectivity validations
		for mirrorName, mirror := range env.Mirrors {
			if !mirror.Enabled {
				continue
			}

			mirrorValidations := planMirrorConnectivity(mirror, envName, mirrorName)
			validations = append(validations, mirrorValidations...)
		}
	}

	return validations
}

// PlanProviderConnectivity creates connectivity validation plans for a provider.
func PlanProviderConnectivity(providerType, domain, owner string, auth ports.AuthenticationConfiguration) []ConnectivityValidation {
	var validations []ConnectivityValidation

	baseURL := buildProviderURL(providerType, domain)

	if baseURL == "" {
		return validations
	}

	// HTTP connectivity
	validations = append(validations, ConnectivityValidation{
		Type:        ConnectivityTypeHTTP,
		Target:      baseURL,
		Timeout:     10 * time.Second,
		Description: "HTTP connectivity to " + providerType + " provider",
		Required:    true,
	})

	// Provider API connectivity
	apiURL := buildProviderAPIURL(providerType, domain, owner)
	if apiURL != "" {
		validations = append(validations, ConnectivityValidation{
			Type:        ConnectivityTypeProvider,
			Target:      apiURL,
			Timeout:     30 * time.Second,
			Description: "API connectivity to " + providerType + " provider",
			Required:    true,
		})
	}

	// Git connectivity (if we have auth)
	if auth.Type != ports.AuthenticationTypeNone {
		gitURL := buildGitURL(providerType, domain, owner, "test-repo")
		if gitURL != "" {
			validations = append(validations, ConnectivityValidation{
				Type:        ConnectivityTypeGit,
				Target:      gitURL,
				Timeout:     30 * time.Second,
				Description: "Git connectivity to " + providerType + " provider",
				Required:    false, // Optional since test-repo might not exist
			})
		}
	}

	// SSH connectivity (if using SSH auth)
	if auth.Type == ports.AuthenticationTypeSSH {
		sshHost := buildSSHHost(providerType, domain)
		if sshHost != "" {
			validations = append(validations, ConnectivityValidation{
				Type:        ConnectivityTypeSSH,
				Target:      sshHost,
				Timeout:     10 * time.Second,
				Description: "SSH connectivity to " + providerType + " provider",
				Required:    true,
			})
		}
	}

	return validations
}

// PlanFileSystemValidations creates validation plans for file system access.
func PlanFileSystemValidations(config ports.AppConfiguration) []FileSystemValidation {
	var validations []FileSystemValidation

	// Validate global directories
	if config.GlobalSettings.TempDirectory != "" {
		validations = append(validations, FileSystemValidation{
			Type:        FileSystemTypeDirectory,
			Path:        config.GlobalSettings.TempDirectory,
			Required:    true,
			Writable:    true,
			Description: "Temporary directory access",
		})
	}

	if config.GlobalSettings.CacheDirectory != "" {
		validations = append(validations, FileSystemValidation{
			Type:        FileSystemTypeDirectory,
			Path:        config.GlobalSettings.CacheDirectory,
			Required:    true,
			Writable:    true,
			Description: "Cache directory access",
		})
	}

	// Validate mirror paths
	for envName, env := range config.Environments {
		for mirrorName, mirror := range env.Mirrors {
			if isLocalProvider(mirror.ProviderType) && mirror.Path != "" {
				validations = append(validations, FileSystemValidation{
					Type:        getFileSystemType(mirror.ProviderType),
					Path:        mirror.Path,
					Required:    true,
					Writable:    mirror.ProviderType == "directory",
					Description: "Mirror path for " + envName + "." + mirrorName,
				})
			}
		}
	}

	return validations
}

// FileSystemValidation represents a file system validation plan.
type FileSystemValidation struct {
	Type        FileSystemType
	Path        string
	Required    bool
	Writable    bool
	Description string
}

// FileSystemType defines types of file system validations.
type FileSystemType string

const (
	FileSystemTypeFile      FileSystemType = "file"
	FileSystemTypeDirectory FileSystemType = "directory"
	FileSystemTypeArchive   FileSystemType = "archive"
)

// FileSystemResult represents the result of a file system validation.
type FileSystemResult struct {
	Validation FileSystemValidation
	Success    bool
	Error      error
	Exists     bool
	Readable   bool
	Writable   bool
	Details    map[string]interface{}
}

// Helper functions for planning connectivity validations

func planSourceConnectivity(source ports.SourceConfiguration, envName string) []ConnectivityValidation {
	if isLocalProvider(source.ProviderType) {
		return []ConnectivityValidation{} // Local providers don't need connectivity
	}

	validations := PlanProviderConnectivity(
		source.ProviderType,
		source.Domain,
		source.Owner,
		source.Authentication,
	)

	// Add environment context to descriptions
	for index := range validations {
		validations[index].Description += " (source for " + envName + ")"
	}

	return validations
}

func planMirrorConnectivity(mirror ports.MirrorConfiguration, envName, mirrorName string) []ConnectivityValidation {
	if isLocalProvider(mirror.ProviderType) {
		return []ConnectivityValidation{} // Local providers don't need connectivity
	}

	validations := PlanProviderConnectivity(
		mirror.ProviderType,
		mirror.Domain,
		mirror.Owner,
		mirror.Authentication,
	)

	// Add environment and mirror context to descriptions
	for index := range validations {
		validations[index].Description += " (mirror " + mirrorName + " for " + envName + ")"
	}

	return validations
}

func buildProviderURL(providerType, domain string) string {
	if domain == "" {
		switch providerType {
		case ProviderTypeGitHub:
			return "https://github.com"
		case ProviderTypeGitLab:
			return "https://gitlab.com"
		default:
			return ""
		}
	}

	return "https://" + domain
}

func buildProviderAPIURL(providerType, domain, owner string) string {
	switch providerType {
	case ProviderTypeGitHub:
		if domain == "" || domain == DefaultGitHubDomain {
			return "https://api.github.com/users/" + owner
		}

		return "https://" + domain + "/api/v3/users/" + owner
	case ProviderTypeGitLab:
		if domain == "" || domain == DefaultGitLabDomain {
			return "https://gitlab.com/api/v4/users?username=" + owner
		}

		return "https://" + domain + "/api/v4/users?username=" + owner
	case ProviderTypeGitea:
		if domain == "" {
			return ""
		}

		return "https://" + domain + "/api/v1/users/" + owner
	default:
		return ""
	}
}

func buildGitURL(providerType, domain, owner, repo string) string {
	switch providerType {
	case ProviderTypeGitHub:
		if domain == "" || domain == DefaultGitHubDomain {
			return "https://github.com/" + owner + "/" + repo + ".git"
		}

		return "https://" + domain + "/" + owner + "/" + repo + ".git"
	case ProviderTypeGitLab:
		if domain == "" || domain == DefaultGitLabDomain {
			return "https://gitlab.com/" + owner + "/" + repo + ".git"
		}

		return "https://" + domain + "/" + owner + "/" + repo + ".git"
	case ProviderTypeGitea:
		if domain == "" {
			return ""
		}

		return "https://" + domain + "/" + owner + "/" + repo + ".git"
	default:
		return ""
	}
}

func buildSSHHost(providerType, domain string) string {
	switch providerType {
	case ProviderTypeGitHub:
		if domain == "" || domain == DefaultGitHubDomain {
			return "git@github.com"
		}

		return "git@" + domain
	case ProviderTypeGitLab:
		if domain == "" || domain == DefaultGitLabDomain {
			return "git@gitlab.com"
		}

		return "git@" + domain
	case ProviderTypeGitea:
		if domain == "" {
			return ""
		}

		return "git@" + domain
	default:
		return ""
	}
}

func getFileSystemType(providerType string) FileSystemType {
	switch providerType {
	case "directory":
		return FileSystemTypeDirectory
	case "archive":
		return FileSystemTypeArchive
	default:
		return FileSystemTypeFile
	}
}

// Validation composition functions

// ValidateAllConnectivity combines multiple connectivity validations.
func ValidateAllConnectivity(ctx context.Context, validator ConnectivityValidator, validations []ConnectivityValidation) []ConnectivityResult {
	results := make([]ConnectivityResult, len(validations))

	for index, validation := range validations {
		results[index] = validator.ValidateConnectivity(ctx, validation)

		// Early exit on required validation failure
		if !results[index].Success && validation.Required {
			// Still execute remaining validations but mark them as skipped
			for remainingIndex := index + 1; remainingIndex < len(validations); remainingIndex++ {
				results[remainingIndex] = ConnectivityResult{
					Validation: validations[remainingIndex],
					Success:    false,
					Error:      ErrValidationSkipped,
					Details:    map[string]interface{}{"reason": "previous required validation failed"},
				}
			}

			break
		}
	}

	return results
}

// CountConnectivityFailures counts failed connectivity validations.
func CountConnectivityFailures(results []ConnectivityResult) (int, int, int) {
	var total, required, optional int

	for _, result := range results {
		if !result.Success {
			total++

			if result.Validation.Required {
				required++
			} else {
				optional++
			}
		}
	}

	return total, required, optional
}

// FilterConnectivityResults filters results by type and success.
func FilterConnectivityResults(results []ConnectivityResult, filterType ConnectivityType, successOnly bool) []ConnectivityResult {
	filtered := make([]ConnectivityResult, 0, len(results))

	for _, result := range results {
		if filterType != "" && result.Validation.Type != filterType {
			continue
		}

		if successOnly && !result.Success {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// Common validation errors.
var (
	ErrValidationSkipped = ValidationError{
		Code:    "VALIDATION_SKIPPED",
		Message: "Validation was skipped due to previous failure",
	}
)

// ValidationError represents a validation error.
type ValidationError struct {
	Code    string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
