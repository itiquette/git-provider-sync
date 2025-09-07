// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"itiquette/git-provider-sync/internal/adapters/auth"
	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/adapters/configuration/dto"
	"itiquette/git-provider-sync/internal/adapters/log"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
)

// ConfigLoader defines the interface for loading application configuration.
type ConfigLoader interface {
	LoadConfiguration(ctx context.Context) (*dto.AppConfiguration, error)
}

// DefaultConfigLoader provides default configuration loading implementation.
type DefaultConfigLoader struct{}

// LoadConfiguration loads the application configuration from various sources.
func (DefaultConfigLoader) LoadConfiguration(ctx context.Context) (*dto.AppConfiguration, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering LoadConfiguration")

	// Get CLI configuration from CLI adapter
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		logger.Warn().Msg("CLI configuration not found in context, using defaults")

		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	appCfg := &dto.AppConfiguration{}

	if err := ReadConfigurationFile(ctx, cliConfig.ConfigFilePath(), cliConfig.ConfigFileOnly(), appCfg); err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	if err := validateConfiguration(ctx, appCfg); err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	return appCfg, nil
}

func processEnvKey(str string, prefix string) string {
	var fieldKeywords = []string{
		"provider_type",
		"owner_type",
		"active_from_limit",
		"include_forks",
		"use_git_binary",
		"cert_dir_path",
		"http_scheme",
		"proxy_url",
		"ssh_command",
		"ssh_url_rewrite_from",
		"ssh_url_rewrite_to",
		"alphanumhyph_name",
		"description_prefix",
		"force_push",
		"github_uploadurl",
		"ignore_invalid_name",
	}

	lowered := strings.ToLower(strings.TrimPrefix(str, prefix))

	for _, keyword := range fieldKeywords {
		if strings.HasSuffix(lowered, "_"+keyword) {
			prefix := strings.TrimSuffix(lowered, "_"+keyword)

			return strings.ReplaceAll(prefix, "_", ".") + "." + keyword
		}
	}

	return strings.ReplaceAll(lowered, "_", ".")
}

// LoadConfigurationSources loads configuration from various file sources
// Follows standard precedence order (lowest to highest):
// 1. System config (/etc/gitprovidersync/dto.yaml) - NOT IMPLEMENTED YET
// 2. User config (~/.config/gitprovidersync/dto.yaml or XDG_CONFIG_HOME)
// 3. Project config (./gitprovidersync.yaml and ./.env)
// Note: Environment variables and CLI flags are handled separately with higher precedence.
func loadConfigurationSources(ctx context.Context, koanfConf *koanf.Koanf, configfile string, configfileOnly bool) error {
	// If --config-file-only is set, ONLY load the specified file
	if configfileOnly {
		return loadConfigFileOnly(koanfConf, configfile)
	}

	// Load in order of precedence (lowest to highest)
	// 1. User config (optional)
	if err := loadUserConfig(koanfConf); err != nil {
		// User config is optional, log but don't fail
		log.Logger(ctx).Debug().Err(err).Msg("Failed to load user config (continuing)")
	}

	// 2. Project config
	if err := loadProjectConfig(koanfConf, configfile); err != nil {
		return err
	}

	// 3. .env file (highest file precedence)
	if err := loadDotEnv(koanfConf); err != nil {
		return err
	}

	return nil
}

func loadConfigFileOnly(koanfConf *koanf.Koanf, configfile string) error {
	if configfile == "" {
		return errors.New("config file path cannot be empty when using --config-file-only")
	}

	if err := koanfConf.Load(file.Provider(configfile), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading specified config file %s: %w", configfile, err)
	}

	return nil
}

func loadUserConfig(koanfConf *koanf.Koanf) error {
	userConfigPath := getUserConfigPath()
	if userConfigPath == "" {
		return nil
	}

	if _, err := os.Stat(userConfigPath); errors.Is(err, fs.ErrNotExist) {
		return nil // File doesn't exist, skip
	}

	if err := koanfConf.Load(file.Provider(userConfigPath), yaml.Parser()); err != nil {
		return fmt.Errorf("failed to load user config from %s: %w", userConfigPath, err)
	}

	return nil
}

func loadProjectConfig(koanfConf *koanf.Koanf, configfile string) error {
	if configfile == "" {
		return nil
	}

	if _, err := os.Stat(configfile); errors.Is(err, fs.ErrNotExist) {
		return nil // File doesn't exist, skip
	}

	if err := koanfConf.Load(file.Provider(configfile), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading project config %s: %w", configfile, err)
	}

	return nil
}

func loadDotEnv(koanfConf *koanf.Koanf) error {
	dotEnvPath := getDotEnvPath()
	if _, err := os.Stat(dotEnvPath); errors.Is(err, fs.ErrNotExist) {
		return nil // File doesn't exist, skip
	}

	err := koanfConf.Load(file.Provider(dotEnvPath), dotenv.ParserEnv("", ".", func(s string) string {
		return processEnvKey(s, "")
	}))
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	return nil
}

// LoadEnvironmentConfiguration loads environment variables into configuration.
func loadEnvironmentConfiguration(koanfConf *koanf.Koanf) error {
	if err := koanfConf.Load(env.Provider("GPS_", ".", func(s string) string {
		return processEnvKey(s, "GPS_")
	}), nil); err != nil {
		return fmt.Errorf("failed to read environment conf: %w", err)
	}

	return nil
}

// ProcessRepositoryLists converts comma-separated repository strings to slices.
func processRepositoryLists(koanfConf *koanf.Koanf) {
	keys := koanfConf.Keys()

	for _, key := range keys {
		if strings.HasSuffix(key, "repositories.include") || strings.HasSuffix(key, "repositories.exclude") {
			convertCommaSeparatedToSlice(koanfConf, key)
		}
	}
}

// ConvertCommaSeparatedToSlice converts a comma-separated string value to a string slice.
func convertCommaSeparatedToSlice(koanfConf *koanf.Koanf, key string) {
	value := koanfConf.Get(key)
	if value == nil {
		return
	}

	strValue, ok := value.(string)
	if !ok || !strings.Contains(strValue, ",") {
		return
	}

	rawRepos := strings.Split(strValue, ",")
	repos := make([]string, 0, len(rawRepos))

	for _, repo := range rawRepos {
		trimmedRepo := strings.TrimSpace(repo)
		if trimmedRepo != "" { // Skip empty entries
			repos = append(repos, trimmedRepo)
		}
	}

	_ = koanfConf.Set(key, repos) // Error ignored - configuration setting failure is not critical
}

// ProcessTokenFiles reads tokens from files when token_file is specified
// And applies provider-specific environment variables
// More secure than embedding tokens in dto.
func processTokenFiles(appConfig *dto.AppConfiguration) error {
	// Process each environment and source
	for _, env := range appConfig.GitProviderSyncConfs {
		for sourceName, source := range env {
			// Apply provider-specific environment variables first
			applyProviderTokenFromEnv(&source.Auth, source.ProviderType)

			// Expand environment variables in token field
			if source.Auth.Token != "" {
				source.Auth.Token = os.ExpandEnv(source.Auth.Token)
			}

			// Process source auth token file
			if err := processAuthTokenFile(&source.Auth); err != nil {
				return fmt.Errorf("failed to read token file for source %s: %w", sourceName, err)
			}

			env[sourceName] = source

			// Process mirror auth
			for mirrorName, mirror := range source.Mirrors {
				// Apply provider-specific environment variables first
				applyProviderTokenFromEnv(&mirror.Auth, mirror.ProviderType)

				// Expand environment variables in token field
				if mirror.Auth.Token != "" {
					mirror.Auth.Token = os.ExpandEnv(mirror.Auth.Token)
				}

				// Process mirror auth token file
				if err := processAuthTokenFile(&mirror.Auth); err != nil {
					return fmt.Errorf("failed to read token file for mirror %s: %w", mirrorName, err)
				}

				source.Mirrors[mirrorName] = mirror
			}
		}
	}

	return nil
}

// ApplyProviderTokenFromEnv checks for provider-specific environment variables
// And applies them with highest precedence.
func applyProviderTokenFromEnv(authConfig *dto.AuthConfig, providerType string) {
	// Check for provider-specific token environment variable
	// GPS_GITHUB_TOKEN, GPS_GITLAB_TOKEN, GPS_GITEA_TOKEN
	envVarName := fmt.Sprintf("GPS_%s_TOKEN", strings.ToUpper(providerType))
	if token := os.Getenv(envVarName); token != "" {
		// Provider-specific env var has highest precedence, always override
		authConfig.Token = token
		authConfig.TokenFile = "" // Clear token_file if provider env var is set
	}
}

// ProcessAuthTokenFile reads token from file if token_file is specified.
func processAuthTokenFile(authConfig *dto.AuthConfig) error {
	// Skip if no token_file specified or token already set
	if authConfig.TokenFile == "" || authConfig.Token != "" {
		return nil
	}

	// Read token from file
	token, err := auth.ReadTokenFromFile(authConfig.TokenFile)
	if err != nil {
		return fmt.Errorf("failed to read token from file: %w", err)
	}

	// Validate token
	if err := auth.ValidateToken(token); err != nil {
		return fmt.Errorf("invalid token in file: %w", err)
	}

	// Set token and clear token_file for security
	authConfig.Token = token
	authConfig.TokenFile = "" // Clear to avoid leaking file path

	return nil
}

// ReadConfigurationFile reads configuration following standard precedence order:
// 1. CLI flags (handled by caller via configfile and configfileOnly params)
// 2. Environment variables (GPS_* prefix)
// 3. Project config (./gitprovidersync.yaml and ./.env)
// 4. User config (~/.config/gitprovidersync/dto.yaml or XDG_CONFIG_HOME)
// 5. System config (/etc/gitprovidersync/dto.yaml) - NOT IMPLEMENTED
// Each higher precedence source overrides values from lower precedence sources.
func ReadConfigurationFile(ctx context.Context, configfile string, configfileOnly bool, appConfiguration *dto.AppConfiguration) error {
	koanfConf := koanf.New(".")

	// Load file sources (system -> user -> project precedence)
	if err := loadConfigurationSources(ctx, koanfConf, configfile, configfileOnly); err != nil {
		return err
	}

	// Load environment variables (higher precedence than files)
	// Skip env vars if --config-file-only flag is set
	if !configfileOnly {
		if err := loadEnvironmentConfiguration(koanfConf); err != nil {
			return err
		}

		processRepositoryLists(koanfConf)
	}

	// Unmarshal consolidated configuration
	if err := koanfConf.Unmarshal("", appConfiguration); err != nil {
		return fmt.Errorf("error unmarshalling yaml config: %w", err)
	}

	if len(appConfiguration.GitProviderSyncConfs) == 0 {
		return domain.ErrConfigurationNotFound
	}

	appConfiguration.FillDefaults()

	// Process token files and provider-specific env vars (highest precedence)
	if err := processTokenFiles(appConfiguration); err != nil {
		return fmt.Errorf("failed to process token files: %w", err)
	}

	return nil
}

// GetUserConfigPath returns the user configuration path following XDG spec
// Checks XDG_CONFIG_HOME first, then falls back to ~/.config/gitprovidersync/dto.yaml
// For backward compatibility, also checks for gitprovidersync.yaml if dto.yaml doesn't exist.
func getUserConfigPath() string {
	const appName = "gitprovidersync"

	const newConfigFile = "dto.yaml"

	const oldConfigFile = "gitprovidersync.yaml" // Backward compatibility

	// Helper to check both new and old config files
	checkPaths := func(dir string) string {
		// Prefer new dto.yaml
		newPath := filepath.Join(dir, appName, newConfigFile)
		if _, err := os.Stat(newPath); err == nil {
			return newPath
		}

		// Fall back to old gitprovidersync.yaml for backward compatibility
		oldPath := filepath.Join(dir, appName, oldConfigFile)
		if _, err := os.Stat(oldPath); err == nil {
			return oldPath
		}

		// Return new path even if neither exists (for future creation)
		return newPath
	}

	// Check XDG_CONFIG_HOME first (follows XDG spec)
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return checkPaths(xdgConfig)
	}

	// Fall back to ~/.config (XDG default when XDG_CONFIG_HOME not set)
	if homeDir, err := os.UserHomeDir(); err == nil {
		return checkPaths(filepath.Join(homeDir, ".config"))
	}

	return ""
}

// GetDotEnvPath returns the path to .env file, with test override support.
func getDotEnvPath() string {
	// Allow test override
	if testHome := os.Getenv("GPS_TESTCONFIG_HOME"); testHome != "" {
		return filepath.Join(testHome, ".env")
	}

	return ".env"
}
