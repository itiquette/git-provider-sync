// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package ports contains interface definitions for the hexagonal architecture.
// Import aggregation for all git-related ports.
package ports

// Import aggregation point for git operations.
// All git-related interfaces are now split into focused files:
// - git_service.go: Core GitOperations interface
// - git_repository.go: Repository interfaces and supporting types
// - git_options.go: Option types and factory functions
//
// This maintains backward compatibility while providing better organization.

// Re-export interfaces from focused files for backward compatibility
// Consumers can import from this package and get all git-related types.
