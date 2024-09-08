// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides functionality for handling and printing
// Git Provider Sync configurations.
package configuration

import (
	"fmt"
	"io"
	"strings"
)

// PrintConfiguration writes the entire AppConfiguration to the provided writer.
func PrintConfiguration(config AppConfiguration, writer io.Writer) {
	fmt.Fprintln(writer, "\n----------------------------")
	fmt.Fprintln(writer, "Git Provider Sync")
	fmt.Fprintln(writer, "----------------------------")

	for name, configuration := range config.Configurations {
		printConfigurationSection(name, configuration.SourceProvider, configuration.ProviderTargets, writer)
	}
}

// printConfigurationSection writes a single configuration section to the provided writer.
func printConfigurationSection(name string, sourceConfig ProviderConfig, targetConfigs map[string]ProviderConfig, writer io.Writer) {
	fmt.Fprintf(writer, "\nConfiguration Name: %s\n\n", name)
	printProviderConfig("Source Provider", sourceConfig, writer)

	for provider, target := range targetConfigs {
		fmt.Fprintln(writer, "Target Provider:")
		fmt.Fprintf(writer, " Configuration Name: %s\n", provider)
		printProviderConfig("", target, writer)
	}
}

// printProviderConfig writes the details of a single ProviderConfig to the provided writer.
func printProviderConfig(header string, config ProviderConfig, writer io.Writer) {
	if header != "" {
		fmt.Fprintf(writer, "%s:\n", header)
	}

	fmt.Fprintf(writer, " Provider: %s\n", config.Provider)

	if !isLocalProvider(config.Provider) {
		printRemoteProviderDetails(config, writer)
	}

	printStringMap(" Providerspecific", config.Providerspecific, writer)
	fmt.Fprintln(writer)
}

// isLocalProvider checks if the provider is a local type (ARCHIVE or DIRECTORY).
func isLocalProvider(provider string) bool {
	return strings.EqualFold(provider, ARCHIVE) || strings.EqualFold(provider, DIRECTORY)
}

// printRemoteProviderDetails writes the details specific to remote providers.
func printRemoteProviderDetails(config ProviderConfig, writer io.Writer) {
	fmt.Fprintf(writer, " Domain: %s\n", config.Domain)

	if len(config.Token) == 0 {
		fmt.Fprintln(writer, " Token not specified")
	} else {
		fmt.Fprintln(writer, " Token: <*****>")
	}

	if len(config.User) == 0 {
		fmt.Fprintf(writer, " Group: %s\n", config.Group)
	} else {
		fmt.Fprintf(writer, " User: %s\n", config.User)
	}

	printStringMap(" Include", config.Include, writer)
	printStringMap(" Exclude", config.Exclude, writer)
}

// printStringMap writes a map of strings to the provided writer if the map is not empty.
func printStringMap(header string, aMap map[string]string, writer io.Writer) {
	if len(aMap) == 0 {
		return
	}

	fmt.Fprintf(writer, "%s:\n", header)

	for key, value := range aMap {
		fmt.Fprintf(writer, "  %s: %s\n", key, value)
	}
}
