// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides comprehensive functionality for managing, validating,
// and interacting with configurations. It offers:
//
//   - Loading and parsing of application configurations from various sources
//   - Validation of configuration settings to ensure system integrity
//   - Management of configuration data throughout the application lifecycle
//   - Utilities for printing and displaying configuration information
//   - Handling of environment-specific and default configurations
//   - Dynamic configuration updates and hot-reloading capabilities
//   - Interfaces for accessing configuration data in a type-safe manner
//
// Key features:
//
//   - Support for multiple configuration formats (e.g., YAML, JSON, TOML)
//   - Hierarchical configuration structure with inheritance and overrides
//   - Secure handling of sensitive configuration data
//   - Integration with command-line flags and environment variables
//   - Validation rules to catch configuration errors early
//   - Helper functions for common configuration-related tasks
package configuration
