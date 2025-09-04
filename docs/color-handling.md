<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Color Handling

Git Provider Sync respects standard color conventions.

## Color Modes

| Flag | Behavior |
|------|----------|
| `--color=auto` (default) | Colors when output is to TTY |
| `--color=always` | Force colors even when piping |
| `--color=never` | Disable colors |

## Environment Variables

- **`NO_COLOR`**: Disables colors when set (overrides all)
- **`TERM=dumb`**: Disables colors automatically

## Examples

```bash
gitprovidersync status                        # Auto-detect
gitprovidersync status --color=always | less -R  # Force colors for pager
gitprovidersync status --color=never          # No colors
NO_COLOR=1 gitprovidersync status             # No colors via env
```

## Colors Used

- **Green**: Success messages
- **Red**: Error messages  
- **Bold**: Headers and emphasis

Standard ANSI escape codes, minimal usage for clarity.
