<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Color Handling in Git Provider Sync

This document describes how color output is handled in git-provider-sync, following best practices for CLI applications.

## Implementation Overview

The color handling implementation follows these principles:
- **Hexagonal**: Color mode is injected, not read from global state
- **Functional**: Pure functions with explicit inputs for color decisions
- **Idiomatic**: Uses standard naming from tools like `ls` and `grep`
- **Simple**: Only 3 colors (red, green, bold) for essential highlighting

## Color Modes

The `--color` flag supports three modes:

- `auto` (default): Enable colors only when output is to a TTY
- `always`: Force colors even when piping/redirecting
- `never`: Disable colors completely

## Environment Variables

The following environment variables are respected:

1. **`NO_COLOR`**: When set to any non-empty value, disables colors completely

- Takes precedence over all other settings including `--color=always`
- Standard from <https://no-color.org>

2. **`TERM=dumb`**: Indicates a terminal with no color support

- Disables colors when detected

## Usage Examples

```bash
# Default behavior (auto-detect based on TTY)
gitprovidersync status

# Force colors even when piping
gitprovidersync status --color=always | less -R

# Disable colors
gitprovidersync status --color=never

# Disable colors via environment variable
NO_COLOR=1 gitprovidersync status

# Override with --color=always still respects NO_COLOR
NO_COLOR=1 gitprovidersync status --color=always  # No colors
```

## Architecture

### Color Module (`internal/adapters/terminal/color.go`)

- `ColorMode` type: Represents when to use colors (auto/always/never)
- `Color` struct: Provides minimal ANSI color codes
- `NewColor()`: Creates color codes based on mode and TTY detection
- `shouldUseColor()`: Pure function determining if colors should be used

### Integration Points

1. **CLI Flag** (`cmd/root.go`):  

- `--color` flag with default value "auto"

2. **Base Options** (`cmd/baseoption/baseoption.go`):

- Extracts color mode from command flags
- Passes to CLI configuration

3. **Output Formatters** (`internal/adapters/cli/outputformatter.go`):

- Uses color mode to determine output formatting
- Respects TTY detection and environment variables

## Testing

Comprehensive tests ensure correct behavior:

- Color mode handling (auto/always/never)
- Environment variable precedence (NO_COLOR, TERM=dumb)
- TTY detection integration
- Color method outputs (Success, Error, Header)

## Best Practices Followed

1. ✅ **NO_COLOR environment variable support**
2. ✅ **TERM=dumb detection**
3. ✅ **--color flag with auto/always/never modes**
4. ✅ **Minimal color usage** (only 3 colors for semantic meaning)
5. ✅ **TTY detection for auto mode**
6. ✅ **Proper precedence** (NO_COLOR > TERM=dumb > --color flag)
