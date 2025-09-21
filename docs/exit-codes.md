<!-- SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors -->
<!-- SPDX-License-Identifier: CC0-1.0 -->

# Exit Codes

Git Provider Sync follows standard Unix/Linux exit code conventions to indicate the result of operations.

## Exit Code Reference

| Exit Code | Description | When Used |
|-----------|-------------|-----------|
| 0 | Success | Normal successful completion |
| 1 | General Error | Unspecified errors and failures |
| 2 | Misuse | Invalid command usage, bad arguments, or incorrect flags |
| 77 | Permission Denied | API authentication failures (401/403), insufficient token scopes, or file permission errors |
| 78 | Configuration Error | Invalid configuration file, YAML syntax errors, or missing required configuration |
| 126 | Cannot Execute | Command exists but cannot be executed (e.g., git binary lacks execute permission) |
| 127 | Command Not Found | Required command not found in PATH (e.g., git binary missing) |
| 129 | SIGHUP | Terminated by SIGHUP signal (terminal hangup) |
| 130 | SIGINT | Terminated by SIGINT signal (Ctrl+C) |
| 131 | SIGQUIT | Terminated by SIGQUIT signal (Ctrl+\, quit with core dump) |
| 143 | SIGTERM | Terminated by SIGTERM signal (graceful termination request) |

## Signal Handling

Git Provider Sync handles the following signals for graceful shutdown:

- **SIGHUP** (1): Terminal hangup, often from SSH disconnection. Performs graceful shutdown, exits with 129.
- **SIGINT** (2): User interrupt via Ctrl+C. Performs graceful shutdown, exits with 130.
- **SIGQUIT** (3): User quit via Ctrl+\. Exits immediately without cleanup (traditional core dump behavior), exits with 131.
- **SIGTERM** (15): System termination request. Performs graceful shutdown, exits with 143.

### Signal Behavior

- **First signal**: Initiates graceful shutdown with cleanup (except SIGQUIT which exits immediately)
- **Second signal**: Forces immediate termination without cleanup
- **Cleanup timeout**: 10 seconds for graceful shutdown before forced termination

## Error Precedence

When multiple error conditions occur:

1. **Application errors take precedence** over signal exits
2. **Specific error codes** (77, 78, 126, 127) override general error (1)
3. **Signal exit codes** only used when no application error occurred

### Examples

- Configuration error + SIGINT = Exit 78 (config error takes precedence)
- Permission denied + SIGTERM = Exit 77 (permission error takes precedence)
- Successful run + SIGINT = Exit 130 (signal code used)

## Common Scenarios

| Scenario | Exit Code | Meaning |
|----------|-----------|---------|
| `git` command not installed | 127 | Command not found |
| GitHub token lacks required scopes | 77 | Permission denied |
| Invalid YAML in config file | 78 | Configuration error |
| User presses Ctrl+C during sync | 130 | SIGINT received |
| SSH connection drops during operation | 129 | SIGHUP received |
| Invalid command-line flag | 2 | Misuse of command |
