// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/model"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

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
		if err := validateConfiguration(config); err != nil {
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
