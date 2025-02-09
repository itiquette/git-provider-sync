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
	"strings"
)

const (
	indentSize = 2
)

// PrintConfiguration writes the entire AppConfiguration to the provided writer.
func PrintConfiguration(appCfg model.AppConfiguration, writer io.Writer) {
	fmt.Fprintln(writer, "\nGit Provider Sync Configuration")
	fmt.Fprintln(writer, strings.Repeat("=", 30))

	for envName, env := range appCfg.GitProviderSyncConfs {
		printEnvironment(envName, env, writer, 0)
	}
}

// printEnvironment writes a single environment section with proper indentation.
func printEnvironment(name string, env model.Environment, writer io.Writer, level int) {
	indent := strings.Repeat(" ", level*indentSize)
	fmt.Fprintf(writer, "\n%sEnvironment: %s\n", indent, name)
	fmt.Fprintf(writer, "%s%s\n", indent, strings.Repeat("-", 20))

	for sourceName, syncConfig := range env {
		printSyncConfig(sourceName, syncConfig, writer, level+1)
	}
}

// printSyncConfig writes the details of a single SyncConfig with proper indentation.
func printSyncConfig(name string, syncCfg model.SyncConfig, writer io.Writer, level int) {
	indent := strings.Repeat(" ", level*indentSize)
	fmt.Fprintf(writer, "\n%sSync Configuration: %s\n", indent, name)

	// Print mandatory fields
	fmt.Fprintf(writer, "%sProvider Type: %s\n", indent, syncCfg.ProviderType)
	fmt.Fprintf(writer, "%sDomain: %s\n", indent, syncCfg.GetDomain())
	fmt.Fprintf(writer, "%sOwner: %s\n", indent, syncCfg.Owner)
	fmt.Fprintf(writer, "%sOwner Type: %s\n", indent, syncCfg.OwnerType)

	// Print optional fields only if they have non-default values
	if syncCfg.IncludeForks {
		fmt.Fprintf(writer, "%sInclude Forks: %t\n", indent, syncCfg.IncludeForks)
	}

	if syncCfg.UseGitBinary {
		fmt.Fprintf(writer, "%sUse Git Binary: %t\n", indent, syncCfg.UseGitBinary)
	}

	if syncCfg.ActiveFromLimit != "" {
		fmt.Fprintf(writer, "%sActive From Limit: %s\n", indent, syncCfg.ActiveFromLimit)
	}

	// Print Auth Configuration
	if !isEmptyAuthConfig(syncCfg.Auth) {
		printAuthConfig(syncCfg.Auth, writer, level+1)
	}

	// Print Repositories Configuration
	if !isEmptyRepositoriesOption(syncCfg.Repositories) {
		printRepositoriesOption(syncCfg.Repositories, writer, level+1)
	}

	// Print Mirror Configurations
	if len(syncCfg.Mirrors) > 0 {
		indentSub := strings.Repeat(" ", level*indentSize)
		fmt.Fprintf(writer, "\n%sMirror Configurations:\n", indentSub)
		fmt.Fprintf(writer, "%s"+strings.Repeat("-", 20)+"\n", "  ")

		for name, mirror := range syncCfg.Mirrors {
			printMirrorConfig(name, mirror, writer, level+1)
		}
	}
}

// printAuthConfig writes authentication configuration details with proper indentation.
func printAuthConfig(authCfg model.AuthConfig, writer io.Writer, level int) {
	indent := strings.Repeat(" ", level*indentSize)
	fmt.Fprintf(writer, "\n%sAuthentication:\n", indent)

	// Print mandatory fields
	fmt.Fprintf(writer, "%sProtocol: %s\n", indent, authCfg.Protocol)

	// Print optional fields only if they have values
	if authCfg.HTTPScheme != "" {
		fmt.Fprintf(writer, "%sHTTP Scheme: %s\n", indent, authCfg.HTTPScheme)
	}

	if authCfg.Token != "" {
		fmt.Fprintf(writer, "%sToken: <*****>\n", indent)
	}

	if authCfg.ProxyURL != "" {
		fmt.Fprintf(writer, "%sProxy URL: %s\n", indent, authCfg.ProxyURL)
	}

	if authCfg.CertDirPath != "" {
		fmt.Fprintf(writer, "%sCertificate Directory: %s\n", indent, authCfg.CertDirPath)
	}

	// Print SSH configuration if any SSH-related fields are set
	if authCfg.SSHCommand != "" || authCfg.SSHURLRewriteFrom != "" || authCfg.SSHURLRewriteTo != "" {
		fmt.Fprintf(writer, "\n%sSSH Configuration:\n", indent)

		if authCfg.SSHCommand != "" {
			fmt.Fprintf(writer, "%sCommand: %s\n", indent, authCfg.SSHCommand)
		}

		if authCfg.SSHURLRewriteFrom != "" {
			fmt.Fprintf(writer, "%sURL Rewrite From: %s\n", indent, authCfg.SSHURLRewriteFrom)
		}

		if authCfg.SSHURLRewriteTo != "" {
			fmt.Fprintf(writer, "%sURL Rewrite To: %s\n", indent, authCfg.SSHURLRewriteTo)
		}
	}
}

// printMirrorConfig writes the details of a mirror configuration with proper indentation.
func printMirrorConfig(name string, mirrorCfg model.MirrorConfig, writer io.Writer, level int) {
	indent := strings.Repeat(" ", level*indentSize)
	fmt.Fprintf(writer, "\n%sMirror: %s\n", indent, name)

	// Print mandatory fields
	fmt.Fprintf(writer, "%sType: %s\n", indent, mirrorCfg.ProviderType)

	if mirrorCfg.Domain != "" {
		fmt.Fprintf(writer, "%sDomain: %s\n", indent, mirrorCfg.GetDomain())
	}

	if mirrorCfg.Owner != "" {
		fmt.Fprintf(writer, "%sOwner: %s\n", indent, mirrorCfg.Owner)
	}

	fmt.Fprintf(writer, "%sOwner Type: %s\n", indent, mirrorCfg.OwnerType)

	// Print optional fields only if they have non-default values
	if mirrorCfg.UseGitBinary {
		fmt.Fprintf(writer, "%sUse Git Binary: %t\n", indent, mirrorCfg.UseGitBinary)
	}

	if mirrorCfg.Path != "" {
		fmt.Fprintf(writer, "%sPath: %s\n", indent, mirrorCfg.Path)
	}

	// Print Mirror Settings if they're not empty
	if !isEmptyMirrorSettings(mirrorCfg.Settings) {
		printMirrorSettings(mirrorCfg.Settings, writer, level+1)
	}

	// Print Mirror Auth Configuration if it's not empty
	if !isEmptyAuthConfig(mirrorCfg.Auth) {
		printAuthConfig(mirrorCfg.Auth, writer, level+1)
	}
}

// printMirrorSettings writes mirror-specific settings with proper indentation.
func printMirrorSettings(settings model.MirrorSettings, writer io.Writer, level int) {
	indent := strings.Repeat(" ", level*indentSize)
	fmt.Fprintf(writer, "\n%sSettings:\n", indent)

	// Print only non-default values
	if settings.ASCIIName {
		fmt.Fprintf(writer, "%sASCII Name: %t\n", indent, settings.ASCIIName)
	}

	if settings.DescriptionPrefix != "" {
		fmt.Fprintf(writer, "%sDescription Prefix: %s\n", indent, settings.DescriptionPrefix)
	}

	if settings.Disabled {
		fmt.Fprintf(writer, "%sDisabled: %t\n", indent, settings.Disabled)
	}

	if settings.ForcePush {
		fmt.Fprintf(writer, "%sForce Push: %t\n", indent, settings.ForcePush)
	}

	if settings.GitHubUploadURL != "" {
		fmt.Fprintf(writer, "%sGitHub Upload URL: %s\n", indent, settings.GitHubUploadURL)
	}

	if settings.IgnoreInvalidName {
		fmt.Fprintf(writer, "%sIgnore Invalid Name: %t\n", indent, settings.IgnoreInvalidName)
	}

	if settings.Visibility != "" {
		fmt.Fprintf(writer, "%sVisibility: %s\n", indent, settings.Visibility)
	}
}

// printRepositoriesOption writes repository configuration options with proper indentation.
func printRepositoriesOption(opt model.RepositoriesOption, writer io.Writer, level int) {
	indent := strings.Repeat(" ", level*indentSize)
	fmt.Fprintf(writer, "\n%sRepositories:\n", indent)

	if len(opt.Include) > 0 {
		fmt.Fprintf(writer, "%sInclude:\n", indent)

		for _, pattern := range opt.Include {
			fmt.Fprintf(writer, "%s  %s\n", indent, pattern)
		}
	}

	if len(opt.Exclude) > 0 {
		fmt.Fprintf(writer, "%sExclude:\n", indent)

		for _, pattern := range opt.Exclude {
			fmt.Fprintf(writer, "%s  %s\n", indent, pattern)
		}
	}
}

// Helper functions to check if configurations are empty.
func isEmptyAuthConfig(authCfg model.AuthConfig) bool {
	return authCfg.Protocol == "" &&
		authCfg.HTTPScheme == "" &&
		authCfg.Token == "" &&
		authCfg.ProxyURL == "" &&
		authCfg.CertDirPath == "" &&
		authCfg.SSHCommand == "" &&
		authCfg.SSHURLRewriteFrom == "" &&
		authCfg.SSHURLRewriteTo == ""
}

func isEmptyRepositoriesOption(opt model.RepositoriesOption) bool {
	return len(opt.Include) == 0 && len(opt.Exclude) == 0
}

func isEmptyMirrorSettings(settings model.MirrorSettings) bool {
	return !settings.ASCIIName &&
		settings.DescriptionPrefix == "" &&
		!settings.Disabled &&
		!settings.ForcePush &&
		settings.GitHubUploadURL == "" &&
		!settings.IgnoreInvalidName &&
		settings.Visibility == ""
}
