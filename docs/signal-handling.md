<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Signal Handling

Git Provider Sync handles interruption signals gracefully.

## Supported Signals

| Signal | Action | Cleanup |
|--------|--------|---------|
| `SIGINT` (Ctrl-C) | Cancel operation | Yes |
| `SIGTERM` | Graceful shutdown | Yes |

## User Experience

```text
# During operation
⚡ Cloning repository example-repo (Press Ctrl-C to cancel)

# On interruption
^C
✗ Repository sync cancelled by user (Ctrl-C)
  Cleaning up temporary files...
```

## Features

- **Immediate Response**: Ctrl-C works at all times
- **Clean Shutdown**: No corruption or partial state
- **Automatic Cleanup**: Temporary files removed
- **Exit Code**: Returns 130 on SIGINT

## Notes

- Operations check for cancellation before expensive steps
- Network operations have bounded timeouts
- Cleanup always runs via defer handlers
