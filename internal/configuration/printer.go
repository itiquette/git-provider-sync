// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides functionality for handling and printing
// Git Provider Sync configurations.
package configuration

import (
	"fmt"
	"io"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"strings"
)

// PrintConfiguration writes the entire AppConfiguration to the provided writer.
func PrintConfiguration(config config.AppConfiguration, writer io.Writer) {
	fmt.Fprintln(writer, "\n----------------------------")
	fmt.Fprintln(writer, "Git Provider Sync")
	fmt.Fprintln(writer, "----------------------------")

	for name, configuration := range config.Configurations {
		printConfigurationSection(name, configuration.SourceProvider, configuration.ProviderTargets, writer)
	}
}

// printConfigurationSection writes a single configuration section to the provided writer.
func printConfigurationSection(name string, sourceConfig config.ProviderConfig, targetConfigs map[string]config.ProviderConfig, writer io.Writer) {
	fmt.Fprintf(writer, "\nConfiguration Name: %s\n\n", name)
	printProviderConfig("Source Provider", sourceConfig, writer)

	for provider, target := range targetConfigs {
		fmt.Fprintln(writer, "Target Provider:")
		fmt.Fprintf(writer, " Configuration Name: %s\n", provider)
		printProviderConfig("", target, writer)
	}
}

// printProviderConfig writes the details of a single ProviderConfig to the provided writer.
func printProviderConfig(header string, config config.ProviderConfig, writer io.Writer) {
	if header != "" {
		fmt.Fprintf(writer, "%s:\n", header)
	}

	fmt.Fprintf(writer, " ProviderType: %s\n", config.ProviderType)

	if !isLocalProvider(config.ProviderType) {
		printRemoteProviderDetails(config, writer)
	}

	printStringMap(" Additional", config.Additional, writer)
	fmt.Fprintln(writer)
}

// isLocalProvider checks if the provider is a local type (ARCHIVE or DIRECTORY).
func isLocalProvider(provider string) bool {
	return strings.EqualFold(provider, config.ARCHIVE) || strings.EqualFold(provider, config.DIRECTORY)
}

// printRemoteProviderDetails writes the details specific to remote providers.
func printRemoteProviderDetails(config config.ProviderConfig, writer io.Writer) {
	fmt.Fprintf(writer, " Domain: %s\n", config.Domain)

	if len(config.HTTPClient.Token) == 0 {
		fmt.Fprintln(writer, " Token not specified")
	} else {
		fmt.Fprintln(writer, " HttpClient.Token: <*****>")
	}

	if len(config.HTTPClient.ProxyURL) > 0 {
		fmt.Fprintf(writer, "  HTTPClient.ProxyURL: %s\n", config.HTTPClient.ProxyURL)
	}

	if len(config.User) == 0 {
		fmt.Fprintf(writer, " Group: %s\n", config.Group)
	} else {
		fmt.Fprintf(writer, " User: %s\n", config.User)
	}

	// Protocol
	printGitProtocol(writer, config)

	// SSHClientOptions
	printSSHClientOption(writer, config)

	fmt.Fprintf(writer, " Include: %s\n", config.Repositories.Include)
	fmt.Fprintf(writer, " Exclude: %s\n", config.Repositories.Exclude)
}

func printGitProtocol(writer io.Writer, providerConfig config.ProviderConfig) {
	fmt.Fprint(writer, " Git:\n")

	if len(providerConfig.Git.Type) == 0 {
		fmt.Fprintf(writer, "  Type: %s\n", config.HTTPS)
		fmt.Fprintf(writer, "  UseGitBinary: %t\n", providerConfig.Git.UseGitBinary)
		fmt.Fprintf(writer, "  IncludeForks: %t\n", providerConfig.Git.IncludeForks)
	} else {
		fmt.Fprintf(writer, "  Type: %s\n", providerConfig.Git.Type)
		fmt.Fprintf(writer, "  UseGitBinary: %t\n", providerConfig.Git.UseGitBinary)
		fmt.Fprintf(writer, "  IncludeForks: %t\n", providerConfig.Git.IncludeForks)
	}
}

func printSSHClientOption(writer io.Writer, providerConfig config.ProviderConfig) {
	if len(providerConfig.SSHClient.ProxyCommand) >= 0 {
		fmt.Fprint(writer, " SSHClientOptions:\n")
		fmt.Fprintf(writer, "  ProxyCommand: %s\n", providerConfig.SSHClient.ProxyCommand)
		fmt.Fprintf(writer, "  RewriteSSHURLFrom: %s\n", providerConfig.SSHClient.RewriteSSHURLFrom)
		fmt.Fprintf(writer, "  RewriteSSHURLTo: %s\n", providerConfig.SSHClient.RewriteSSHURLTo)
	}
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
