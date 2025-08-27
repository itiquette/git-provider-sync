<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Configuration Precedence

Git Provider Sync follows the standard configuration precedence order used by most modern CLI tools (docker, kubectl, aws-cli, etc.).

## Precedence Order (Highest to Lowest)

1. **CLI Flags** (highest precedence)

- Command-line arguments always override all other configuration sources
- Example: `--config-file`, `--log-level`, `--dry-run`

2. **Environment Variables**

- `GPS_*` prefixed variables for general configuration
- Provider-specific tokens: `GPS_GITHUB_TOKEN`, `GPS_GITLAB_TOKEN`, `GPS_GITEA_TOKEN`
- Environment variable expansion in config files: `token: "${MY_TOKEN}"`

3. **Project Configuration**

- `./gitprovidersync.yaml` - Project-specific configuration in current directory
- `./.env` - Environment file in current directory (dotenv format)

4. **User Configuration**

- `$XDG_CONFIG_HOME/gitprovidersync/config.yaml` (if XDG_CONFIG_HOME is set)
- `~/.config/gitprovidersync/config.yaml` (fallback when XDG_CONFIG_HOME not set)
- Follows XDG Base Directory Specification

5. **System Configuration** (lowest precedence)

- `/etc/gitprovidersync/config.yaml` (NOT YET IMPLEMENTED)
- For system-wide defaults

## How It Works

Each configuration source overrides values from lower-precedence sources. For example:

```yaml
# ~/.config/gitprovidersync/config.yaml (user config)
gitprovidersync:
  default:
    source:
      owner: default-user
      log_level: info

# ./gitprovidersync.yaml (project config)
gitprovidersync:
  default:
    source:
      owner: project-user  # Overrides user config

# Environment variable
export GPS_LOG_LEVEL=debug  # Overrides both configs

# CLI flag
gitprovidersync sync --log-level trace  # Overrides everything
```

## XDG Base Directory Support

Git Provider Sync follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html):

- **Config**: `$XDG_CONFIG_HOME/gitprovidersync/` or `~/.config/gitprovidersync/`
- **Data**: `$XDG_DATA_HOME/gitprovidersync/` or `~/.local/share/gitprovidersync/` (future)
- **Cache**: `$XDG_CACHE_HOME/gitprovidersync/` or `~/.cache/gitprovidersync/` (future)
- **State**: `$XDG_STATE_HOME/gitprovidersync/` or `~/.local/state/gitprovidersync/` (future)

## Special Behaviors

### --config-file Flag

When you specify a config file explicitly:

```bash
gitprovidersync sync --config-file /path/to/config.yaml
```

This file is loaded as project configuration (level 3), and the default `./gitprovidersync.yaml` is NOT loaded.

### --config-file-only Flag

When you use the config-file-only mode:

```bash
gitprovidersync sync --config-file /path/to/config.yaml --config-file-only
```

- ONLY the specified config file is loaded
- User config is skipped
- Environment variables are skipped
- .env file is skipped
- Only CLI flags can override values

## Token Precedence

Authentication tokens have their own precedence within the configuration system:

1. Provider-specific environment variables: `GPS_GITHUB_TOKEN`, `GPS_GITLAB_TOKEN`, `GPS_GITEA_TOKEN`
2. Environment variable expansion in config: `token: "${GITHUB_TOKEN}"`
3. Token file specified in config: `token_file: "/run/secrets/github-token"`

## Examples

### Typical Development Setup

```bash
# User config for defaults
~/.config/gitprovidersync/config.yaml

# Project config for project-specific settings
./gitprovidersync.yaml

# Environment variables for secrets
export GPS_GITHUB_TOKEN="ghp_xxxxx"
export GPS_GITLAB_TOKEN="glpat_xxxxx"

# Run with everything
gitprovidersync sync
```

### CI/CD Setup

```bash
# Use only environment variables, no config files
export GPS_GITHUB_TOKEN="${SECRET_GITHUB_TOKEN}"
export GPS_GITLAB_TOKEN="${SECRET_GITLAB_TOKEN}"

# Load config from specific location
gitprovidersync sync --config-file /etc/myapp/sync-config.yaml
```

### Testing Specific Config

```bash
# Test with ONLY a specific config file
gitprovidersync sync --config-file test-config.yaml --config-file-only
```

## Best Practices

1. **Never commit secrets** - Use environment variables or token files for sensitive data
2. **User config for personal defaults** - Store your preferences in `~/.config/gitprovidersync/config.yaml`
3. **Project config for team settings** - Commit `gitprovidersync.yaml` to version control
4. **Environment variables for CI/CD** - Use `GPS_*` variables in automated environments
5. **CLI flags for overrides** - Use flags for temporary changes during testing

## Migration from Old Behavior

If you were relying on the old precedence order, you may need to adjust:

- Old: XDG config was loaded first (lowest precedence)
- New: XDG config is user config (middle precedence)

To maintain old behavior, move your XDG config to the project directory as `gitprovidersync.yaml`.
