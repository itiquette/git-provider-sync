// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package cmd provides the command-line interface for Git Provider Sync,
// a utility for mirroring Git repositories across multiple providers and storage formats.
//
// Available Commands:
//   - sync: Mirror repositories from source to target providers
//   - print: Display configuration in various formats
//   - status: Show synchronization status and health checks
//   - man: Generate manual pages
//
// Supported providers: GitHub, GitLab, Gitea, local directory, tar.gz archive
//
// The package structure includes:
//   - Root command setup and global CLI configuration
//   - Individual command implementations in dedicated subdirectories
//   - Shared CLI utilities and option handling
package cmd
