<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Terminal Behavior

## Color Output

Git Provider Sync respects standard color conventions:

| Flag | Behavior |
|------|----------|
| `--color=auto` (default) | Colors when output is to TTY |
| `--color=always` | Force colors even when piping |
| `--color=never` | Disable colors |

Environment variables:
- `NO_COLOR`: Disables colors when set
- `TERM=dumb`: Disables colors automatically

## Signal Handling

Handles interruption signals gracefully:

| Signal | Action | Cleanup |
|--------|--------|---------|
| `SIGINT` (Ctrl-C) | Cancel operation | Yes |
| `SIGTERM` | Graceful shutdown | Yes |

Features:
- Immediate response to Ctrl-C
- Clean shutdown without corruption
- Automatic cleanup of temporary files
- Returns exit code 130 on SIGINT

## Examples

```bash
# Color control
gitprovidersync status --color=always | less -R
NO_COLOR=1 gitprovidersync status

# Interruption
^C cancels current operation cleanly
```
