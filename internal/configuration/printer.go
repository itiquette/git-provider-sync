// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package model provides functionality for handling and printing
// Git Provider Sync configurations.
package configuration

import (
	"fmt"
	"io"
	model "itiquette/git-provider-sync/internal/model/configuration"
)

// PrintConfiguration writes the entire AppConfiguration to the provided writer.
func PrintConfiguration(appCfg model.AppConfiguration, writer io.Writer) {
	fmt.Fprintln(writer, "\n----------------------------")
	fmt.Fprintln(writer, "Git Provider Sync")
	fmt.Fprintln(writer, "----------------------------")

	for envName, env := range appCfg.GitProviderSyncConfs {
		printEnvironment(envName, env, writer)
	}
}

// printEnvironment writes a single environment section to the provided writer.
func printEnvironment(name string, env model.Environment, writer io.Writer) {
	fmt.Fprintf(writer, "\nEnvironment: %s\n", name)
	fmt.Fprintln(writer, "============================")

	for sourceName, syncConfig := range env {
		printSyncConfig(sourceName, syncConfig, writer)
	}
}

// printSyncConfig writes the details of a single SyncConfig to the provided writer.
func printSyncConfig(name string, config model.SyncConfig, writer io.Writer) {
	fmt.Fprintf(writer, "\nSync Configuration: %s\n", name)
	fmt.Fprintf(writer, "Provider Type: %s\n", config.ProviderType)
	fmt.Fprintf(writer, "Domain: %s\n", config.GetDomain())
	fmt.Fprintf(writer, "Owner: %s\n", config.Owner)
	fmt.Fprintf(writer, "Owner Type: %s\n", config.OwnerType)
	fmt.Fprintf(writer, "Include Forks: %t\n", config.IncludeForks)
	fmt.Fprintf(writer, "Use Git Binary: %t\n", config.UseGitBinary)

	if config.ActiveFromLimit != "" {
		fmt.Fprintf(writer, "Active From Limit: %s\n", config.ActiveFromLimit)
	}

	// Print Auth Configuration
	printAuthConfig(config.Auth, writer)

	// Print Repositories Configuration
	if config.Repositories != (model.RepositoriesOption{}) {
		printRepositoriesOption(config.Repositories, writer)
	}

	// Print Mirror Configurations
	if len(config.Mirrors) > 0 {
		fmt.Fprintln(writer, "\nMirror Configurations:")
		fmt.Fprintln(writer, "------------------")

		for _, mirror := range config.Mirrors {
			printMirrorConfig(mirror, writer)
		}
	}
}

// printAuthConfig writes authentication configuration details.
func printAuthConfig(auth model.AuthConfig, writer io.Writer) {
	fmt.Fprintln(writer, "\nAuthentication Configuration:")
	fmt.Fprintln(writer, "-------------------------")
	fmt.Fprintf(writer, "Protocol: %s\n", auth.Protocol)

	if auth.HTTPScheme != "" {
		fmt.Fprintf(writer, "HTTP Scheme: %s\n", auth.HTTPScheme)
	}

	if auth.Token != "" {
		fmt.Fprintln(writer, "Token: <*****>")
	}

	if auth.ProxyURL != "" {
		fmt.Fprintf(writer, "Proxy URL: %s\n", auth.ProxyURL)
	}

	if auth.CertDirPath != "" {
		fmt.Fprintf(writer, "Certificate Directory: %s\n", auth.CertDirPath)
	}

	// Print SSH-specific configuration if present
	if auth.SSHCommand != "" || auth.SSHURLRewriteFrom != "" || auth.SSHURLRewriteTo != "" {
		fmt.Fprintln(writer, "\nSSH Configuration:")
		fmt.Fprintf(writer, "SSH Command: %s\n", auth.SSHCommand)
		fmt.Fprintf(writer, "SSH URL Rewrite From: %s\n", auth.SSHURLRewriteFrom)
		fmt.Fprintf(writer, "SSH URL Rewrite To: %s\n", auth.SSHURLRewriteTo)
	}
}

// printMirrorConfig writes the details of a mirror configuration.
func printMirrorConfig(mirror model.MirrorConfig, writer io.Writer) {
	//fmt.Fprintf(writer, "\nMirror: %s\n", mirror)
	fmt.Fprintf(writer, "Type: %s\n", mirror.ProviderType)
	fmt.Fprintf(writer, "Domain: %s\n", mirror.GetDomain())
	fmt.Fprintf(writer, "Owner: %s\n", mirror.Owner)
	fmt.Fprintf(writer, "Owner Type: %s\n", mirror.OwnerType)

	if mirror.Path != "" {
		fmt.Fprintf(writer, "Path: %s\n", mirror.Path)
	}

	// Print Mirror Settings
	printMirrorSettings(mirror.Settings, writer)

	// Print Mirror Auth Configuration
	fmt.Fprintln(writer, "\nMirror Authentication:")
	printAuthConfig(mirror.Auth, writer)
}

// printMirrorSettings writes mirror-specific settings.
func printMirrorSettings(settings model.MirrorSettings, writer io.Writer) {
	fmt.Fprintln(writer, "\nMirror Settings:")
	fmt.Fprintf(writer, "ASCII Name: %t\n", settings.ASCIIName)

	if settings.DescriptionPrefix != "" {
		fmt.Fprintf(writer, "Description Prefix: %s\n", settings.DescriptionPrefix)
	}

	fmt.Fprintf(writer, "Disabled: %t\n", settings.Disabled)
	fmt.Fprintf(writer, "Force Push: %t\n", settings.ForcePush)

	if settings.GitHubUploadURL != "" {
		fmt.Fprintf(writer, "GitHub Upload URL: %s\n", settings.GitHubUploadURL)
	}

	fmt.Fprintf(writer, "Ignore Invalid Name: %t\n", settings.IgnoreInvalidName)

	if settings.Visibility != "" {
		fmt.Fprintf(writer, "Visibility: %s\n", settings.Visibility)
	}
}

// printRepositoriesOption writes repository configuration options.
func printRepositoriesOption(repos model.RepositoriesOption, writer io.Writer) {
	fmt.Fprintln(writer, "\nRepository Configuration:")
	fmt.Fprintln(writer, "------------------------")

	if len(repos.Include) > 0 {
		fmt.Fprintf(writer, "Include: %v\n", repos.Include)
	}

	if len(repos.Exclude) > 0 {
		fmt.Fprintf(writer, "Exclude: %v\n", repos.Exclude)
	}
}
