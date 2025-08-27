<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Environment Variables

Git Provider Sync respects standard environment variables and provides GPS-specific variables for configuration.

## Standard Environment Variables

The following standard environment variables are respected per CLI best practices:

### Display Control

- **`NO_COLOR`** - Disables colored output when set to any non-empty value (<https://no-color.org>)
  - Takes precedence over all other color settings
  - Example: `NO_COLOR=1 gitprovidersync status`

- **`TERM`** - Terminal type
  - When set to `dumb`, colors are disabled automatically
  - Automatically detected by the system

### Network Configuration  

- **`HTTP_PROXY`**/**`http_proxy`** - HTTP proxy server URL
  - Used by Git providers for HTTP connections
  - Format: `http://proxy.example.com:8080`
  - Example: `HTTP_PROXY=http://proxy:8080 gitprovidersync sync`

- **`HTTPS_PROXY`**/**`https_proxy`** - HTTPS proxy server URL
  - Used by Git providers for HTTPS connections
  - Format: `https://proxy.example.com:8443`
  - Example: `HTTPS_PROXY=https://proxy:8443 gitprovidersync sync`

- **`NO_PROXY`**/**`no_proxy`** - Comma-separated list of hosts to bypass proxy
  - Example: `NO_PROXY=localhost,127.0.0.1,.internal.company.com`

### System Paths

- **`HOME`** - User home directory
  - Used to locate user configuration: `~/.config/gitprovidersync/config.yaml`
  - Fallback when `XDG_CONFIG_HOME` is not set

- **`XDG_CONFIG_HOME`** - XDG base directory for user configuration
  - When set, config is at: `$XDG_CONFIG_HOME/gitprovidersync/config.yaml`
  - Follows XDG Base Directory Specification

## GPS-Specific Environment Variables

### Provider Authentication

Provider-specific tokens have the highest precedence for authentication:

- **`GPS_GITHUB_TOKEN`** - GitHub personal access token
  - Overrides any GitHub token in configuration files
  - Example: `GPS_GITHUB_TOKEN=ghp_xxxx gitprovidersync sync`

- **`GPS_GITLAB_TOKEN`** - GitLab personal access token
  - Overrides any GitLab token in configuration files
  - Example: `GPS_GITLAB_TOKEN=glpat-xxxx gitprovidersync sync`

- **`GPS_GITEA_TOKEN`** - Gitea access token
  - Overrides any Gitea token in configuration files
  - Example: `GPS_GITEA_TOKEN=xxxx gitprovidersync sync`

### General Configuration

- **`GPS_CONFIG_FILE`** - Path to configuration file
  - Overrides default search paths
  - Example: `GPS_CONFIG_FILE=/custom/path/config.yaml gitprovidersync status`

- **`GPS_LOG_LEVEL`** - Logging verbosity
  - Values: `trace`, `debug`, `info`, `warn`, `error`
  - Default: `info`
  - Example: `GPS_LOG_LEVEL=debug gitprovidersync sync`

### Configuration Expansion

Environment variables can be referenced in configuration files using `${VAR}` syntax:

```yaml
gitprovidersync:
  production:
    github-source:
      provider_type: github
      auth:
        token: "${GITHUB_TOKEN}"  # Expands GITHUB_TOKEN environment variable
```

## Precedence Order

Configuration values are resolved in the following order (highest to lowest):

1. **CLI flags** - Command-line arguments
2. **Provider-specific environment variables** - `GPS_GITHUB_TOKEN`, etc.
3. **General environment variables** - `GPS_*` variables
4. **Standard environment variables** - `NO_COLOR`, proxy settings
5. **Project configuration** - `./gitprovidersync.yaml` and `./.env`
6. **User configuration** - `~/.config/gitprovidersync/config.yaml`
7. **System configuration** - `/etc/gitprovidersync/config.yaml` (not yet implemented)

## Security Considerations

- **Never commit tokens** - Always use environment variables or secure secret management
- **Use specific provider tokens** - `GPS_GITHUB_TOKEN` is more secure than a general token
- **Minimal permissions** - Create tokens with only necessary scopes
- **Token rotation** - Regularly rotate access tokens

## Examples

### Running with proxy and debug logging

```bash
HTTP_PROXY=http://proxy:8080 \
HTTPS_PROXY=https://proxy:8443 \
GPS_LOG_LEVEL=debug \
gitprovidersync sync production
```

### Using provider tokens without config file

```bash
GPS_GITHUB_TOKEN=ghp_xxxx \
GPS_GITLAB_TOKEN=glpat-yyyy \
gitprovidersync sync --config-file production.yaml
```

### Disabling colors in CI/CD

```bash
NO_COLOR=1 gitprovidersync status
```

### Custom config location

```bash
GPS_CONFIG_FILE=/etc/myapp/sync.yaml gitprovidersync sync
```

## Debugging

To see which environment variables are being used:

```bash
# Show current environment variables
GPS_LOG_LEVEL=trace gitprovidersync status

# Check proxy settings
curl -I https://api.github.com  # Should use same proxy as gitprovidersync
```

## Migration Guide

### From GPS_COLOR_MODE (Removed)

The `GPS_COLOR_MODE` environment variable has been removed. Use standard variables instead:

- To disable colors: Set `NO_COLOR=1`
- Colors are automatically detected based on terminal capabilities
- Use `--color=never` flag to disable for a single command
- Use `--color=always` flag to force colors (e.g., for piping to `less -R`)

### From GPS_SECURE_TOKEN

The global `GPS_SECURE_TOKEN` has been replaced with provider-specific tokens:

- Use `GPS_GITHUB_TOKEN` for GitHub authentication
- Use `GPS_GITLAB_TOKEN` for GitLab authentication  
- Use `GPS_GITEA_TOKEN` for Gitea authentication

This provides better security by limiting token scope to specific providers.
