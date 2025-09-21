// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package ports contains interface definitions for the hexagonal architecture
// Import aggregation for all git-related ports
package ports

// Import aggregation point for git operations
// All git-related interfaces are split into focused files:
// - git_service.go: Core GitOperations interface
// - git_repository.go: Repository interfaces and supporting types
// - git_options.go: Option types and factory functions
