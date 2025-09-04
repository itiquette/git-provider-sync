<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Environment Variables

## Standard Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `NO_COLOR` | Disable colored output | `NO_COLOR=1` |
| `TERM` | Terminal type (colors off when `dumb`) | Auto-detected |
| `HTTP_PROXY` | HTTP proxy URL | `http://proxy:8080` |
| `HTTPS_PROXY` | HTTPS proxy URL | `https://proxy:8443` |
| `NO_PROXY` | Bypass proxy hosts | `localhost,127.0.0.1` |
| `HOME` | User home directory | `/home/user` |
| `XDG_CONFIG_HOME` | XDG config directory | `~/.config` |

## GPS Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GPS_GITHUB_TOKEN` | GitHub auth token | - | `ghp_...` |
| `GPS_GITLAB_TOKEN` | GitLab auth token | - | `glpat-...` |
| `GPS_GITEA_TOKEN` | Gitea auth token | - | Token from Gitea |
| `GPS_CONFIG_FILE` | Config file path | `./gitprovidersync.yaml` | `/etc/sync.yaml` |
| `GPS_LOG_LEVEL` | Log verbosity | `info` | `trace`, `debug`, `warn`, `error` |

## Configuration Expansion

Use `${VAR}` syntax in config files:

```yaml
auth:
  token: "${GITHUB_TOKEN}"
```

## Precedence

1. CLI flags
2. GPS environment variables
3. Standard environment variables
4. Project config (`./gitprovidersync.yaml`)
5. User config (`~/.config/gitprovidersync/config.yaml`)
6. System config (`/etc/gitprovidersync/config.yaml`)

## Examples

```bash
# Proxy and debug
HTTP_PROXY=http://proxy:8080 GPS_LOG_LEVEL=debug gitprovidersync sync

# Provider tokens
GPS_GITHUB_TOKEN=ghp_xxxx GPS_GITLAB_TOKEN=glpat-yyyy gitprovidersync sync

# Disable colors in CI
NO_COLOR=1 gitprovidersync status

# Custom config
GPS_CONFIG_FILE=/etc/sync.yaml gitprovidersync sync
```
