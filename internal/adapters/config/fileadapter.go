// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration file loading and parsing adapters.
package config

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	koanfpkg "github.com/knadh/koanf/v2"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Static errors for err113 compliance.
var (
	ErrAtLeastOneEnvironmentRequired = errors.New("at least one environment is required")
	ErrMirrorMustBeObject            = errors.New("mirror must be an object")
	ErrEnvironmentMustBeObject       = errors.New("environment must be an object")
	ErrAuthMustBeObject              = errors.New("authentication must be an object")
	ErrEnvironmentValidationFailed   = errors.New("environment validation failed")
	ErrSourceTypeNotImplemented      = errors.New("source type not yet implemented")
	ErrUnsupportedSourceType         = errors.New("unsupported source type")
	ErrRequiredConfigFileNotFound    = errors.New("required configuration file not found")
	ErrOptionalFileNotFound          = errors.New("optional file not found")
	ErrConfigFormatNotImplemented    = errors.New("configuration format not yet implemented")
	ErrUnsupportedConfigFormat       = errors.New("unsupported configuration format")
	ErrSourceMustBeObject            = errors.New("source must be an object")
	ErrEnvironmentNameEmpty          = errors.New("environment name cannot be empty")
	ErrOwnerRequired                 = errors.New("owner is required")
	ErrAtLeastOneMirrorRequired      = errors.New("at least one mirror is required")
	ErrPathRequired                  = errors.New("path is required")
	ErrOwnerRequiredForProvider      = errors.New("owner is required for provider")
	ErrProviderTypeRequired          = errors.New("provider type is required")
	ErrInvalidLogLevel               = errors.New("invalid log level")
	ErrMirrorsMustBeObject           = errors.New("mirrors must be an object")
)

// FileAdapter implements the Configuration interface using file-based configuration.
type FileAdapter struct {
	koanf        *koanfpkg.Koanf
	sources      []ports.ConfigurationSource
	lastModified time.Time
	version      string
	mu           sync.RWMutex
}

// New creates a new file-based configuration adapter.
func New() *FileAdapter {
	return &FileAdapter{
		koanf:   koanfpkg.New("."),
		sources: []ports.ConfigurationSource{},
		version: "1.0.0",
	}
}

// Load loads configuration from a single source.
func (a *FileAdapter) Load(_ context.Context, source ports.ConfigurationSource) (ports.AppConfiguration, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Reset koanf instance for fresh load
	a.koanf = koanfpkg.New(".")
	a.sources = []ports.ConfigurationSource{source}

	err := a.loadSource(source)
	if err != nil {
		return ports.AppConfiguration{}, fmt.Errorf("failed to load configuration: %w", err)
	}

	config, err := a.parseConfiguration()
	if err != nil {
		return ports.AppConfiguration{}, fmt.Errorf("failed to parse configuration: %w", err)
	}

	a.lastModified = time.Now()

	return config, nil
}

// LoadMultiple loads configuration from multiple sources.
func (a *FileAdapter) LoadMultiple(_ context.Context, sources []ports.ConfigurationSource) (ports.AppConfiguration, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Reset koanf instance for fresh load
	a.koanf = koanfpkg.New(".")
	a.sources = sources

	// Sort sources by priority (higher priority loaded last)
	sortedSources := make([]ports.ConfigurationSource, len(sources))
	copy(sortedSources, sources)

	// Simple bubble sort by priority
	for range sortedSources {
		for j := range len(sortedSources) - 1 {
			if sortedSources[j].Priority > sortedSources[j+1].Priority {
				sortedSources[j], sortedSources[j+1] = sortedSources[j+1], sortedSources[j]
			}
		}
	}

	// Load sources in priority order
	for _, source := range sortedSources {
		err := a.loadSource(source)
		if err != nil && source.Required {
			return ports.AppConfiguration{}, fmt.Errorf("failed to load required source %s: %w", source.Location, err)
		}
	}

	config, err := a.parseConfiguration()
	if err != nil {
		return ports.AppConfiguration{}, fmt.Errorf("failed to parse configuration: %w", err)
	}

	a.lastModified = time.Now()

	return config, nil
}

// Reload reloads configuration from existing sources.
func (a *FileAdapter) Reload(ctx context.Context) (ports.AppConfiguration, error) {
	a.mu.RLock()
	sources := make([]ports.ConfigurationSource, len(a.sources))
	copy(sources, a.sources)
	a.mu.RUnlock()

	return a.LoadMultiple(ctx, sources)
}

// Validate validates the configuration.
func (a *FileAdapter) Validate(config ports.AppConfiguration) ([]ports.ConfigurationError, error) {
	var validationErrors []ports.ConfigurationError

	// Validate environments
	if len(config.Environments) == 0 {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    "environments",
			Err:      ErrAtLeastOneEnvironmentRequired,
			Severity: ports.ErrorSeverityError,
		})
	}

	for envName, env := range config.Environments {
		envErrors := a.validateEnvironment(envName, env)
		validationErrors = append(validationErrors, envErrors...)
	}

	// Validate global settings
	globalErrors := a.validateGlobalSettings(config.GlobalSettings)
	validationErrors = append(validationErrors, globalErrors...)

	return validationErrors, nil
}

// ValidateEnvironment validates a single environment configuration.
func (a *FileAdapter) ValidateEnvironment(env ports.EnvironmentConfiguration) error {
	validationErrors := a.validateEnvironment(env.Name, env)
	if len(validationErrors) > 0 {
		return fmt.Errorf("%w: %d errors found", ErrEnvironmentValidationFailed, len(validationErrors))
	}

	return nil
}

// GetSources returns the configuration sources.
func (a *FileAdapter) GetSources() []ports.ConfigurationSource {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sources := make([]ports.ConfigurationSource, len(a.sources))
	copy(sources, a.sources)

	return sources
}

// GetLastModified returns the last modification time.
func (a *FileAdapter) GetLastModified() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.lastModified
}

// GetVersion returns the configuration version.
func (a *FileAdapter) GetVersion() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.version
}

// Private helper methods

// loadSource loads a single configuration source.
func (a *FileAdapter) loadSource(source ports.ConfigurationSource) error {
	switch source.Type {
	case ports.SourceTypeFile:
		return a.loadFileSource(source)
	case ports.SourceTypeEnvironment:
		return a.loadEnvironmentSource(source)
	case ports.SourceTypeEtcd, ports.SourceTypeConsul, ports.SourceTypeVault, ports.SourceTypeHTTP, ports.SourceTypeDefaults:
		return fmt.Errorf("%w: %s", ErrSourceTypeNotImplemented, source.Type)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedSourceType, source.Type)
	}
}

// loadFileSource loads configuration from a file.
func (a *FileAdapter) loadFileSource(source ports.ConfigurationSource) error {
	if err := a.checkFileExists(source); err != nil {
		// If it's an optional file that doesn't exist, skip it
		if strings.Contains(err.Error(), "optional file not found") {
			return nil
		}

		return err
	}

	format := a.determineFileFormat(source)

	parser, err := a.createParser(format)
	if err != nil {
		return err
	}

	if err := a.loadFileWithParser(source.Location, parser); err != nil {
		return err
	}

	return nil
}

// checkFileExists checks if the configuration file exists and handles optional/required logic.
func (a *FileAdapter) checkFileExists(source ports.ConfigurationSource) error {
	if _, err := os.Stat(source.Location); os.IsNotExist(err) {
		if source.Required {
			return fmt.Errorf("%w: %s", ErrRequiredConfigFileNotFound, source.Location)
		}
		// Return a special error to indicate the file should be skipped
		return fmt.Errorf("%w: %s", ErrOptionalFileNotFound, source.Location)
	}

	return nil
}

// determineFileFormat determines the configuration format based on source format or file extension.
func (a *FileAdapter) determineFileFormat(source ports.ConfigurationSource) ports.ConfigurationFormat {
	if source.Format != "" {
		return source.Format
	}

	ext := strings.ToLower(filepath.Ext(source.Location))
	switch ext {
	case ".yaml", ".yml":
		return ports.ConfigurationFormatYAML
	case ".json":
		return ports.ConfigurationFormatJSON
	case ".env":
		return ports.ConfigurationFormatENV
	default:
		return ports.ConfigurationFormatYAML // Default to YAML
	}
}

// createParser creates the appropriate parser for the given format.
func (a *FileAdapter) createParser(format ports.ConfigurationFormat) (koanfpkg.Parser, error) { //nolint:ireturn
	switch format {
	case ports.ConfigurationFormatYAML:
		return yaml.Parser(), nil
	case ports.ConfigurationFormatENV:
		return dotenv.Parser(), nil
	case ports.ConfigurationFormatJSON, ports.ConfigurationFormatTOML, ports.ConfigurationFormatINI:
		return nil, fmt.Errorf("%w: %s", ErrConfigFormatNotImplemented, format)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedConfigFormat, format)
	}
}

// loadFileWithParser loads the configuration file using the specified parser.
func (a *FileAdapter) loadFileWithParser(location string, parser koanfpkg.Parser) error {
	err := a.koanf.Load(file.Provider(location), parser)
	if err != nil {
		return fmt.Errorf("failed to load configuration file %s: %w", location, err)
	}

	return nil
}

// loadEnvironmentSource loads configuration from environment variables.
func (a *FileAdapter) loadEnvironmentSource(source ports.ConfigurationSource) error {
	prefix := source.Location
	if prefix == "" {
		prefix = "GPS_" // Default prefix for git-provider-sync
	}

	err := a.koanf.Load(env.Provider(prefix, ".", func(envVar string) string {
		// Convert environment variable names to config keys
		// GPS_SOURCE_GITHUB_TOKEN -> source.github.token
		envVar = strings.TrimPrefix(envVar, prefix)
		envVar = strings.ToLower(envVar)
		envVar = strings.ReplaceAll(envVar, "_", ".")

		return envVar
	}), nil)
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	return nil
}

// parseConfiguration parses the loaded configuration into AppConfiguration.
func (a *FileAdapter) parseConfiguration() (ports.AppConfiguration, error) {
	config := ports.NewAppConfiguration()

	// Parse environments
	if a.koanf.Exists("environments") {
		envMap := make(map[string]interface{})

		err := a.koanf.Unmarshal("environments", &envMap)
		if err != nil {
			return config, fmt.Errorf("failed to parse environments: %w", err)
		}

		for envName, envData := range envMap {
			env, err := a.parseEnvironment(envName, envData)
			if err != nil {
				return config, fmt.Errorf("failed to parse environment %s: %w", envName, err)
			}

			config.Environments[envName] = env
		}
	}

	// Parse global settings
	if a.koanf.Exists("global") {
		err := a.koanf.Unmarshal("global", &config.GlobalSettings)
		if err != nil {
			return config, fmt.Errorf("failed to parse global settings: %w", err)
		}
	}

	// Set metadata
	config.Metadata = ports.ConfigurationMetadata{
		Version:     a.version,
		LoadTime:    time.Now(),
		Sources:     a.sources,
		Environment: os.Getenv("ENVIRONMENT"),
		Validated:   false,
		Checksum:    a.calculateChecksum(),
	}

	return config, nil
}

// parseEnvironment parses a single environment configuration.
func (a *FileAdapter) parseEnvironment(name string, data interface{}) (ports.EnvironmentConfiguration, error) {
	env := ports.EnvironmentConfiguration{
		Name:    name,
		Enabled: true,
		Mirrors: make(map[string]ports.MirrorConfiguration),
	}

	envMap, ok := data.(map[string]interface{})
	if !ok {
		return env, ErrEnvironmentMustBeObject
	}

	if err := a.parseEnvironmentEnabled(&env, envMap); err != nil {
		return env, err
	}

	if err := a.parseEnvironmentSource(&env, envMap); err != nil {
		return env, err
	}

	if err := a.parseEnvironmentMirrors(&env, envMap); err != nil {
		return env, err
	}

	if err := a.parseEnvironmentOptions(&env, envMap); err != nil {
		return env, err
	}

	return env, nil
}

// parseSourceConfiguration parses source configuration.
func (a *FileAdapter) parseSourceConfiguration(data interface{}) (ports.SourceConfiguration, error) {
	var source ports.SourceConfiguration

	sourceMap, ok := data.(map[string]interface{})
	if !ok {
		return source, ErrSourceMustBeObject
	}

	// Extract basic fields
	if providerType, exists := sourceMap["provider_type"]; exists {
		if pt, ok := providerType.(string); ok {
			source.ProviderType = pt
		}
	}

	if domain, exists := sourceMap["domain"]; exists {
		if d, ok := domain.(string); ok {
			source.Domain = d
		}
	}

	if owner, exists := sourceMap["owner"]; exists {
		if o, ok := owner.(string); ok {
			source.Owner = o
		}
	}

	// Parse authentication
	if authData, exists := sourceMap["authentication"]; exists {
		auth, err := a.parseAuthConfiguration(authData)
		if err != nil {
			return source, fmt.Errorf("failed to parse authentication: %w", err)
		}

		source.Authentication = auth
	}

	return source, nil
}

// parseMirrorConfiguration parses mirror configuration.
func (a *FileAdapter) parseMirrorConfiguration(name string, data interface{}) (ports.MirrorConfiguration, error) {
	mirror := ports.MirrorConfiguration{
		Name:    name,
		Enabled: true,
	}

	mirrorMap, ok := data.(map[string]interface{})
	if !ok {
		return mirror, ErrMirrorMustBeObject
	}

	if err := a.parseMirrorBasicFields(&mirror, mirrorMap); err != nil {
		return mirror, err
	}

	if err := a.parseMirrorAuthenticationField(&mirror, mirrorMap); err != nil {
		return mirror, err
	}

	return mirror, nil
}

// parseMirrorBasicFields parses the basic fields of a mirror configuration.
func (a *FileAdapter) parseMirrorBasicFields(mirror *ports.MirrorConfiguration, mirrorMap map[string]interface{}) error { //nolint:unparam // Error return reserved for future validation
	// Define field mappings for type-safe parsing
	stringFields := map[string]*string{
		"provider_type": &mirror.ProviderType,
		"domain":        &mirror.Domain,
		"owner":         &mirror.Owner,
		"path":          &mirror.Path,
	}

	// Parse string fields
	for fieldName, fieldPtr := range stringFields {
		if value, exists := mirrorMap[fieldName]; exists {
			if strValue, ok := value.(string); ok {
				*fieldPtr = strValue
			}
		}
	}

	// Parse boolean fields
	if enabled, exists := mirrorMap["enabled"]; exists {
		if boolValue, ok := enabled.(bool); ok {
			mirror.Enabled = boolValue
		}
	}

	return nil
}

// parseMirrorAuthenticationField parses the authentication field of a mirror configuration.
func (a *FileAdapter) parseMirrorAuthenticationField(mirror *ports.MirrorConfiguration, mirrorMap map[string]interface{}) error {
	if authData, exists := mirrorMap["authentication"]; exists {
		auth, err := a.parseAuthConfiguration(authData)
		if err != nil {
			return fmt.Errorf("failed to parse authentication: %w", err)
		}

		mirror.Authentication = auth
	}

	return nil
}

// parseAuthConfiguration parses authentication configuration.
func (a *FileAdapter) parseAuthConfiguration(data interface{}) (ports.AuthenticationConfiguration, error) {
	var auth ports.AuthenticationConfiguration

	authMap, ok := data.(map[string]interface{})
	if !ok {
		return auth, ErrAuthMustBeObject
	}

	a.parseAuthType(&auth, authMap)
	a.parseAuthCredentials(&auth, authMap)

	return auth, nil
}

// validateEnvironment validates a single environment.
func (a *FileAdapter) validateEnvironment(envName string, env ports.EnvironmentConfiguration) []ports.ConfigurationError {
	var validationErrors []ports.ConfigurationError

	// Validate environment name
	if envName == "" {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    "name",
			Err:      ErrEnvironmentNameEmpty,
			Severity: ports.ErrorSeverityError,
		})
	}

	// Validate source
	if env.Source.ProviderType == "" {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    envName + ".source.provider_type",
			Err:      ErrProviderTypeRequired,
			Severity: ports.ErrorSeverityError,
		})
	}

	if env.Source.Owner == "" {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    envName + ".source.owner",
			Err:      ErrOwnerRequired,
			Severity: ports.ErrorSeverityError,
		})
	}

	// Validate mirrors
	if len(env.Mirrors) == 0 {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    envName + ".mirrors",
			Err:      ErrAtLeastOneMirrorRequired,
			Severity: ports.ErrorSeverityWarning,
		})
	}

	for mirrorName, mirror := range env.Mirrors {
		mirrorErrors := a.validateMirror(envName, mirrorName, mirror)
		validationErrors = append(validationErrors, mirrorErrors...)
	}

	return validationErrors
}

// validateMirror validates a single mirror configuration.
func (a *FileAdapter) validateMirror(envName, mirrorName string, mirror ports.MirrorConfiguration) []ports.ConfigurationError {
	var validationErrors []ports.ConfigurationError

	fieldPrefix := fmt.Sprintf("%s.mirrors.%s", envName, mirrorName)

	if mirror.ProviderType == "" {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    fieldPrefix + ".provider_type",
			Err:      ErrProviderTypeRequired,
			Severity: ports.ErrorSeverityError,
		})
	}

	// Validate based on provider type
	switch mirror.ProviderType {
	case "directory", "archive":
		if mirror.Path == "" {
			validationErrors = append(validationErrors, ports.ConfigurationError{
				Field:    fieldPrefix + ".path",
				Err:      fmt.Errorf("%w for %s provider", ErrPathRequired, mirror.ProviderType),
				Severity: ports.ErrorSeverityError,
			})
		}
	case "github", "gitlab", "gitea":
		if mirror.Owner == "" {
			validationErrors = append(validationErrors, ports.ConfigurationError{
				Field:    fieldPrefix + ".owner",
				Err:      fmt.Errorf("%w for %s provider", ErrOwnerRequiredForProvider, mirror.ProviderType),
				Severity: ports.ErrorSeverityError,
			})
		}
	}

	return validationErrors
}

// validateGlobalSettings validates global settings.
func (a *FileAdapter) validateGlobalSettings(settings ports.GlobalSettings) []ports.ConfigurationError {
	var validationErrors []ports.ConfigurationError

	// Validate log level
	validLevels := []string{
		string(ports.LogLevelTrace),
		string(ports.LogLevelDebug),
		string(ports.LogLevelInfo),
		string(ports.LogLevelWarn),
		string(ports.LogLevelError),
		string(ports.LogLevelFatal),
	}

	levelValid := false

	for _, level := range validLevels {
		if string(settings.LogLevel) == level {
			levelValid = true

			break
		}
	}

	if !levelValid {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    "global.log_level",
			Value:    settings.LogLevel,
			Err:      ErrInvalidLogLevel,
			Severity: ports.ErrorSeverityError,
		})
	}

	return validationErrors
}

// calculateChecksum calculates a checksum of the current configuration.
func (a *FileAdapter) calculateChecksum() string {
	data := a.koanf.Sprint()
	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

// parseEnvironmentEnabled parses the enabled flag for an environment.
func (a *FileAdapter) parseEnvironmentEnabled(env *ports.EnvironmentConfiguration, envMap map[string]interface{}) error { //nolint:unparam // Error return reserved for future validation
	if enabled, exists := envMap["enabled"]; exists {
		if enabledBool, ok := enabled.(bool); ok {
			env.Enabled = enabledBool
		}
	}

	return nil
}

// parseEnvironmentSource parses the source configuration for an environment.
func (a *FileAdapter) parseEnvironmentSource(env *ports.EnvironmentConfiguration, envMap map[string]interface{}) error {
	if sourceData, exists := envMap["source"]; exists {
		source, err := a.parseSourceConfiguration(sourceData)
		if err != nil {
			return fmt.Errorf("failed to parse source: %w", err)
		}

		env.Source = source
	}

	return nil
}

// parseEnvironmentMirrors parses the mirrors configuration for an environment.
func (a *FileAdapter) parseEnvironmentMirrors(env *ports.EnvironmentConfiguration, envMap map[string]interface{}) error {
	if mirrorsData, exists := envMap["mirrors"]; exists {
		mirrorsMap, ok := mirrorsData.(map[string]interface{})
		if !ok {
			return ErrMirrorsMustBeObject
		}

		for mirrorName, mirrorData := range mirrorsMap {
			mirror, err := a.parseMirrorConfiguration(mirrorName, mirrorData)
			if err != nil {
				return fmt.Errorf("failed to parse mirror %s: %w", mirrorName, err)
			}

			env.Mirrors[mirrorName] = mirror
		}
	}

	return nil
}

// parseEnvironmentOptions parses the options configuration for an environment.
func (a *FileAdapter) parseEnvironmentOptions(env *ports.EnvironmentConfiguration, envMap map[string]interface{}) error { //nolint:unparam // Error return reserved for future validation
	if optionsData, exists := envMap["options"]; exists {
		var options ports.EnvironmentOptions
		// Simple type assertion for now - a full implementation would use reflection
		_ = optionsData // Use the variable to avoid compiler error
		env.Options = options
	}

	return nil
}

// parseAuthType parses the authentication type field.
func (a *FileAdapter) parseAuthType(auth *ports.AuthenticationConfiguration, authMap map[string]interface{}) {
	if authType, exists := authMap["type"]; exists {
		if at, ok := authType.(string); ok {
			auth.Type = ports.AuthenticationType(at)
		}
	}
}

// parseAuthCredentials parses authentication credential fields.
func (a *FileAdapter) parseAuthCredentials(auth *ports.AuthenticationConfiguration, authMap map[string]interface{}) {
	if token, exists := authMap["token"]; exists {
		if t, ok := token.(string); ok {
			auth.Token = t
		}
	}

	if username, exists := authMap["username"]; exists {
		if u, ok := username.(string); ok {
			auth.Username = u
		}
	}

	if password, exists := authMap["password"]; exists {
		if p, ok := password.(string); ok {
			auth.Password = p
		}
	}

	if sshKeyPath, exists := authMap["ssh_key_path"]; exists {
		if skp, ok := sshKeyPath.(string); ok {
			auth.SSHKeyPath = skp
		}
	}
}
