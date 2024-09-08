// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/model"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

// ProviderConfig represents the configuration for a single provider.
type ProviderConfig struct {
	Provider         string            `koanf:"provider"`
	Domain           string            `koanf:"domain"`
	Token            string            `koanf:"token"`
	User             string            `koanf:"user"`
	Group            string            `koanf:"group"`
	Exclude          map[string]string `koanf:"exclude"`
	Include          map[string]string `koanf:"include"`
	Providerspecific map[string]string `koanf:"providerspecific"`
}

// String returns a string representation of ProviderConfig, masking the token.
func (p ProviderConfig) String() string {
	return fmt.Sprintf("ProviderConfig: Provider: %s, Domain: %s, Token: <****>, User: %s, Group: %s, Exclude: %v, Include: %v, ProviderSpecific: %v",
		p.Provider, p.Domain, p.User, p.Group, p.Exclude, p.Include, p.Providerspecific)
}

// DebugLog logs the ProviderConfig details at debug level.
func (p ProviderConfig) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	event := logger.Debug(). //nolint:zerologlint
					Str("provider", p.Provider).
					Fields(p.Exclude).
					Fields(p.Include)

	switch strings.ToLower(p.Provider) {
	case DIRECTORY:
		event.Str("target directory", p.DirectoryTargetDir())
	case ARCHIVE:
		event.Str("target directory", p.ArchiveTargetDir())
	default:
		event.Str("domain", p.Domain).
			Strs("user,group", []string{p.User, p.Group})
	}

	return event
}

// ProvidersConfig represents the configuration for source and target providers.
type ProvidersConfig struct {
	SourceProvider  ProviderConfig            `koanf:"source"`
	ProviderTargets map[string]ProviderConfig `koanf:"targets"`
}

// AppConfiguration represents the entire application configuration.
type AppConfiguration struct {
	Configurations map[string]ProvidersConfig `koanf:"configurations"`
}

// ArchiveTargetDir returns the archive target directory.
func (p ProviderConfig) ArchiveTargetDir() string {
	return p.Providerspecific["archivetargetdir"]
}

// DirectoryTargetDir returns the directory target directory.
func (p ProviderConfig) DirectoryTargetDir() string {
	return p.Providerspecific["directorytargetdir"]
}

// IsGroup returns true if the configuration is for a group.
func (p ProviderConfig) IsGroup() bool {
	return p.Group != ""
}

// IncludedRepositories returns a slice of included repository names.
func (p ProviderConfig) IncludedRepositories() []string {
	return splitAndTrim(p.Include["repositories"])
}

// ExcludedRepositories returns a slice of excluded repository names.
func (p ProviderConfig) ExcludedRepositories() []string {
	return splitAndTrim(p.Exclude["repositories"])
}

// DebugLog logs the AppConfiguration details at debug level.
func (a AppConfiguration) DebugLog(logger *zerolog.Logger) {
	for name, config := range a.Configurations {
		config.SourceProvider.DebugLog(logger).Msg(name + ":SourceProvider")

		for key, target := range config.ProviderTargets {
			target.DebugLog(logger).Msg(key + ":TargetProvider")
		}
	}
}

// configuration/configuration.go

type ConfigLoader interface {
	LoadConfiguration(ctx context.Context) (*AppConfiguration, error)
}

type DefaultConfigLoader struct{}

// LoadConfiguration loads the application configuration from various sources.
func (DefaultConfigLoader) LoadConfiguration(ctx context.Context) (*AppConfiguration, error) {
	cliOption := model.CLIOptions(ctx)
	appConfig := &AppConfiguration{}

	if err := ReadConfigurationFile(appConfig, cliOption.ConfigFilePath, cliOption.ConfigFileOnly); err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	for _, config := range appConfig.Configurations {
		if err := ValidateConfiguration(config); err != nil {
			return nil, fmt.Errorf("failed to validate configuration: %w", err)
		}
	}

	return appConfig, nil
}

func ReadConfigurationFile(appConfiguration *AppConfiguration, configfile string, configfileOnly bool) error {
	const (
		xdgConfigHomeEnv        = "XDG_CONFIG_HOME"
		xdgConfigHomeConfigPath = "/gitprovidersync/" + "gitprovidersync.yaml"
		dotEnvFilename          = ".env"
	)

	koanfConf := koanf.New(".")
	xdgConfigfileExists, xdgConfigFilePath := hasXDGConfigFile(xdgConfigHomeEnv, xdgConfigHomeConfigPath)
	localConfigfileExists := hasLocalConfigFile(configfile)
	dotEnvFileExists, dotEnvFilePath := hasDotEnvFile(dotEnvFilename)

	// xdg config file
	if xdgConfigfileExists && !configfileOnly {
		if err := koanfConf.Load(file.Provider(xdgConfigFilePath), yaml.Parser()); err != nil {
			return fmt.Errorf("error loading xdg_config_home configuration. %w", err)
		}
	}

	// local config file
	if localConfigfileExists {
		if err := koanfConf.Load(file.Provider(configfile), yaml.Parser()); err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}
	}

	// .env file
	if dotEnvFileExists && !configfileOnly {
		err := koanfConf.Load(file.Provider(dotEnvFilePath), dotenv.ParserEnv("", "_", strings.ToLower))
		if err != nil {
			return fmt.Errorf("error loading dotenvfile config: %w", err)
		}
	}
	// env variable (prefix: GPS_)
	if !configfileOnly {
		if err := koanfConf.Load(env.Provider("GPS_", ".", func(s string) string {
			return strings.ReplaceAll(strings.ToLower(
				strings.TrimPrefix(s, "GPS_")), "_", ".")
		}), nil); err != nil {
			return fmt.Errorf("failed to read environment conf: %w", err)
		}
	}

	// Unmarshal the YAML data into the config struct
	if err := koanfConf.Unmarshal("", appConfiguration); err != nil {
		panic(fmt.Errorf("error unmarshalling yaml config: %w", err))
	}

	if len(appConfiguration.Configurations) == 0 {
		panic("No configuration could be found!")
	}

	return nil
}

func hasXDGConfigFile(xdgconfighome string, xdgconfighomeconfigpath string) (bool, string) {
	xdgConfigfileExists := false

	var xdgConfigFilePath string

	envValue, xdgHomeIsSet := os.LookupEnv(xdgconfighome)
	if xdgHomeIsSet {
		xdgConfigFilePath = filepath.Join(envValue, xdgconfighomeconfigpath)
		if _, err := os.Stat(xdgConfigFilePath); err == nil {
			xdgConfigfileExists = true
		}
	}

	return xdgConfigfileExists, xdgConfigFilePath
}

func hasLocalConfigFile(configFile string) bool {
	localConfigfileExists := false
	if _, err := os.Stat(configFile); err == nil {
		localConfigfileExists = true
	}

	return localConfigfileExists
}

func hasDotEnvFile(dotEnvFilePath string) (bool, string) {
	dotEnvFileExists := false

	envValue, testConfigHomeIsSet := os.LookupEnv("GPS_TESTCONFIG_HOME")
	if testConfigHomeIsSet {
		dotEnvFilePath = filepath.Join(envValue, dotEnvFilePath)
	}

	if _, err := os.Stat(dotEnvFilePath); err == nil {
		dotEnvFileExists = true
	}

	return dotEnvFileExists, dotEnvFilePath
}

func splitAndTrim(s string) []string {
	return strings.Split(strings.ReplaceAll(s, " ", ""), ",")
}
