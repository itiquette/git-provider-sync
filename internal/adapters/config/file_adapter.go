// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

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

	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	koanfpkg "github.com/knadh/koanf/v2"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// FileAdapter implements the Configuration interface using file-based configuration.
type FileAdapter struct {
	koanf         *koanfpkg.Koanf
	sources       []ports.ConfigurationSource
	lastModified  time.Time
	version       string
	watcher       *fsnotify.Watcher
	watchCallback ports.ConfigurationChangeCallback
	stopChan      chan struct{}
	mu            sync.RWMutex
}

// New creates a new file-based configuration adapter.
func New() *FileAdapter {
	return &FileAdapter{
		koanf:    koanfpkg.New("."),
		sources:  []ports.ConfigurationSource{},
		version:  "1.0.0",
		stopChan: make(chan struct{}),
	}
}

// Load loads configuration from a single source.
func (a *FileAdapter) Load(ctx context.Context, source ports.ConfigurationSource) (ports.AppConfiguration, error) {
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
func (a *FileAdapter) LoadMultiple(ctx context.Context, sources []ports.ConfigurationSource) (ports.AppConfiguration, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Reset koanf instance for fresh load
	a.koanf = koanfpkg.New(".")
	a.sources = sources

	// Sort sources by priority (higher priority loaded last)
	sortedSources := make([]ports.ConfigurationSource, len(sources))
	copy(sortedSources, sources)

	// Simple bubble sort by priority
	for i := 0; i < len(sortedSources); i++ {
		for j := 0; j < len(sortedSources)-1; j++ {
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
			Err:      errors.New("at least one environment is required"),
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
		return fmt.Errorf("environment validation failed: %d errors found", len(validationErrors))
	}

	return nil
}

// Watch starts watching configuration files for changes.
func (a *FileAdapter) Watch(ctx context.Context, callback ports.ConfigurationChangeCallback) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.watcher != nil {
		return errors.New("already watching configuration files")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	a.watcher = watcher
	a.watchCallback = callback

	// Add all file sources to watcher
	for _, source := range a.sources {
		if source.Type == ports.SourceTypeFile {
			err = watcher.Add(source.Location)
			if err != nil {
				return fmt.Errorf("failed to watch file %s: %w", source.Location, err)
			}
		}
	}

	// Start watching in goroutine
	go a.watchFiles(ctx)

	return nil
}

// StopWatching stops watching configuration files.
func (a *FileAdapter) StopWatching() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.watcher == nil {
		return nil
	}

	close(a.stopChan)
	err := a.watcher.Close()
	a.watcher = nil
	a.watchCallback = nil
	a.stopChan = make(chan struct{})

	return fmt.Errorf("failed to close file watcher: %w", err)
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
		return fmt.Errorf("source type %s not yet implemented", source.Type)
	default:
		return fmt.Errorf("unsupported source type: %s", source.Type)
	}
}

// loadFileSource loads configuration from a file.
func (a *FileAdapter) loadFileSource(source ports.ConfigurationSource) error {
	// Check if file exists
	if _, err := os.Stat(source.Location); os.IsNotExist(err) {
		if source.Required {
			return fmt.Errorf("required configuration file not found: %s", source.Location)
		}

		return nil // Skip optional files that don't exist
	}

	var parser koanfpkg.Parser

	// Determine parser based on format or file extension
	format := source.Format
	if format == "" {
		ext := strings.ToLower(filepath.Ext(source.Location))
		switch ext {
		case ".yaml", ".yml":
			format = ports.ConfigurationFormatYAML
		case ".json":
			format = ports.ConfigurationFormatJSON
		case ".env":
			format = ports.ConfigurationFormatENV
		default:
			format = ports.ConfigurationFormatYAML // Default to YAML
		}
	}

	switch format {
	case ports.ConfigurationFormatYAML:
		parser = yaml.Parser()
	case ports.ConfigurationFormatENV:
		parser = dotenv.Parser()
	case ports.ConfigurationFormatJSON, ports.ConfigurationFormatTOML, ports.ConfigurationFormatINI:
		return fmt.Errorf("configuration format %s not yet implemented", format)
	default:
		return fmt.Errorf("unsupported configuration format: %s", format)
	}

	// Load the file
	err := a.koanf.Load(file.Provider(source.Location), parser)
	if err != nil {
		return fmt.Errorf("failed to load configuration file %s: %w", source.Location, err)
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
		return env, errors.New("environment must be an object")
	}

	// Parse enabled flag
	if enabled, exists := envMap["enabled"]; exists {
		if enabledBool, ok := enabled.(bool); ok {
			env.Enabled = enabledBool
		}
	}

	// Parse source configuration
	if sourceData, exists := envMap["source"]; exists {
		source, err := a.parseSourceConfiguration(sourceData)
		if err != nil {
			return env, fmt.Errorf("failed to parse source: %w", err)
		}

		env.Source = source
	}

	// Parse mirrors
	if mirrorsData, exists := envMap["mirrors"]; exists {
		mirrorsMap, ok := mirrorsData.(map[string]interface{})
		if !ok {
			return env, errors.New("mirrors must be an object")
		}

		for mirrorName, mirrorData := range mirrorsMap {
			mirror, err := a.parseMirrorConfiguration(mirrorName, mirrorData)
			if err != nil {
				return env, fmt.Errorf("failed to parse mirror %s: %w", mirrorName, err)
			}

			env.Mirrors[mirrorName] = mirror
		}
	}

	// Parse options
	if optionsData, exists := envMap["options"]; exists {
		var options ports.EnvironmentOptions
		// Simple type assertion for now - a full implementation would use reflection
		_ = optionsData // Use the variable to avoid compiler error
		env.Options = options
	}

	return env, nil
}

// parseSourceConfiguration parses source configuration.
func (a *FileAdapter) parseSourceConfiguration(data interface{}) (ports.SourceConfiguration, error) {
	var source ports.SourceConfiguration

	sourceMap, ok := data.(map[string]interface{})
	if !ok {
		return source, errors.New("source must be an object")
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
		return mirror, errors.New("mirror must be an object")
	}

	// Extract basic fields
	if providerType, exists := mirrorMap["provider_type"]; exists {
		if pt, ok := providerType.(string); ok {
			mirror.ProviderType = pt
		}
	}

	if domain, exists := mirrorMap["domain"]; exists {
		if d, ok := domain.(string); ok {
			mirror.Domain = d
		}
	}

	if owner, exists := mirrorMap["owner"]; exists {
		if o, ok := owner.(string); ok {
			mirror.Owner = o
		}
	}

	if path, exists := mirrorMap["path"]; exists {
		if p, ok := path.(string); ok {
			mirror.Path = p
		}
	}

	if enabled, exists := mirrorMap["enabled"]; exists {
		if e, ok := enabled.(bool); ok {
			mirror.Enabled = e
		}
	}

	// Parse authentication
	if authData, exists := mirrorMap["authentication"]; exists {
		auth, err := a.parseAuthConfiguration(authData)
		if err != nil {
			return mirror, fmt.Errorf("failed to parse authentication: %w", err)
		}

		mirror.Authentication = auth
	}

	return mirror, nil
}

// parseAuthConfiguration parses authentication configuration.
func (a *FileAdapter) parseAuthConfiguration(data interface{}) (ports.AuthenticationConfiguration, error) {
	var auth ports.AuthenticationConfiguration

	authMap, ok := data.(map[string]interface{})
	if !ok {
		return auth, errors.New("authentication must be an object")
	}

	if authType, exists := authMap["type"]; exists {
		if at, ok := authType.(string); ok {
			auth.Type = ports.AuthenticationType(at)
		}
	}

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

	return auth, nil
}

// validateEnvironment validates a single environment.
func (a *FileAdapter) validateEnvironment(envName string, env ports.EnvironmentConfiguration) []ports.ConfigurationError {
	var validationErrors []ports.ConfigurationError

	// Validate environment name
	if envName == "" {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    "name",
			Err:      errors.New("environment name cannot be empty"),
			Severity: ports.ErrorSeverityError,
		})
	}

	// Validate source
	if env.Source.ProviderType == "" {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    envName + ".source.provider_type",
			Err:      errors.New("provider type is required"),
			Severity: ports.ErrorSeverityError,
		})
	}

	if env.Source.Owner == "" {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    envName + ".source.owner",
			Err:      errors.New("owner is required"),
			Severity: ports.ErrorSeverityError,
		})
	}

	// Validate mirrors
	if len(env.Mirrors) == 0 {
		validationErrors = append(validationErrors, ports.ConfigurationError{
			Field:    envName + ".mirrors",
			Err:      errors.New("at least one mirror is required"),
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
			Err:      errors.New("provider type is required"),
			Severity: ports.ErrorSeverityError,
		})
	}

	// Validate based on provider type
	switch mirror.ProviderType {
	case "directory", "archive":
		if mirror.Path == "" {
			validationErrors = append(validationErrors, ports.ConfigurationError{
				Field:    fieldPrefix + ".path",
				Err:      fmt.Errorf("path is required for %s provider", mirror.ProviderType),
				Severity: ports.ErrorSeverityError,
			})
		}
	case "github", "gitlab", "gitea":
		if mirror.Owner == "" {
			validationErrors = append(validationErrors, ports.ConfigurationError{
				Field:    fieldPrefix + ".owner",
				Err:      fmt.Errorf("owner is required for %s provider", mirror.ProviderType),
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
			Err:      errors.New("invalid log level"),
			Severity: ports.ErrorSeverityError,
		})
	}

	return validationErrors
}

// watchFiles watches configuration files for changes.
func (a *FileAdapter) watchFiles(ctx context.Context) {
	defer func() {
		if a.watcher != nil {
			if err := a.watcher.Close(); err != nil {
				// Log error but don't return it as this is cleanup
				_ = err
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopChan:
			return
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				// File was modified, reload configuration
				if a.watchCallback != nil {
					oldConfig, _ := a.parseConfiguration()
					newConfig, err := a.Reload(ctx)

					if err == nil {
						a.watchCallback(oldConfig, newConfig)
					}
					_ = oldConfig // Suppress unused variable warning
					_ = newConfig // Suppress unused variable warning
				}
			}
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			// Log error (in a real implementation, use proper logging)
			_ = err
		}
	}
}

// calculateChecksum calculates a checksum of the current configuration.
func (a *FileAdapter) calculateChecksum() string {
	data := a.koanf.Sprint()
	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}
