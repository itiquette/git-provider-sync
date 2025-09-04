<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Configuration Reference

## Basic Structure

```yaml
gitprovidersync:
  environment-name:        # production, staging, dev
    source-name:           # descriptive source name
      provider_type: github # github, gitlab, or gitea
      owner: "username"
      owner_type: user      # user or group
      mirrors:
        target-name:
          provider_type: directory
          path: "./backup"
```

## Providers

All providers support these authentication methods:
- Environment variable: `GPS_{PROVIDER}_TOKEN` (auto-detected)
- Variable expansion: `token: "${ANY_ENV_VAR}"`
- Token file: `token_file: "/path/to/token"`

| Provider | Domain | Owner Types | Notes |
|----------|--------|-------------|-------|
| `github` | github.com | user, group | Enterprise domains supported |
| `gitlab` | gitlab.com | user, group | Self-hosted domains supported |
| `gitea` | your-domain.com | user | Self-hosted only |

Example:

```yaml
      provider_type: github
      domain: github.com  # optional for github.com
      owner: "username"
      owner_type: user
      auth:
        token: "${GITHUB_TOKEN}"
```

## Target Types

| Type | Purpose | Key Settings |
|------|---------|--------------|
| `directory` | Local backup | `path`, `bare: true/false` |
| `archive` | Compressed backup | `path`, `format: tar.gz/zip` |
| `github/gitlab/gitea` | Provider mirror | Same as source provider config |

```yaml
      mirrors:
        local:
          provider_type: directory
          path: "./backup"
          settings:
            bare: false
        archive:
          provider_type: archive
          path: "./archive"
          format: tar.gz
        mirror:
          provider_type: gitlab
          owner: "backup-user"
          owner_type: user
          auth:
            token: "${GITLAB_TOKEN}"
```

## Filtering

```yaml
      repositories:
        include: ["important-*", "project-alpha"]
        exclude: ["*-temp", "archive-*"]
      active_from_limit: "24h"  # only repos active in last 24h
      include_forks: false      # skip forked repositories
```

## Advanced Options

```yaml
      settings:
        alphanumhyph_name: true    # clean invalid characters in names
        ignore_invalid_name: true   # skip repos with invalid names
      mirrors:
        target:
          settings:
            visibility: private     # public, private, internal
            description: "Mirror"
            topics: ["backup"]
            issues: false
            wiki: false
```

## Validation

```bash
gitprovidersync print --config-file your-config.yaml       # check syntax
gitprovidersync sync --dry-run --config-file your-config.yaml  # test connections
```

## Multiple Environments

Use different environments for personal, work, or staging configurations. Select with `--environment` flag.

## Environment Variables

All string values support `${VARIABLE_NAME}` expansion. Provider tokens are auto-detected from `GPS_{PROVIDER}_TOKEN` variables.

## Property Reference

| Property | Required | Default | Description |
|----------|----------|---------|-------------|
| `provider_type` | Yes | - | `github`, `gitlab`, `gitea` |
| `owner` | Yes | - | Repository owner |
| `owner_type` | Yes | - | `user` or `group` |
| `auth.token` | Yes | - | API token |
| `domain` | No | Provider default | Custom domain |
| `include_forks` | No | false | Include forked repos |
| `active_from_limit` | No | - | Activity filter |
| `repositories.include` | No | All | Include patterns |
| `repositories.exclude` | No | None | Exclude patterns |
| `settings.visibility` | No | private | Repo visibility |
| `settings.alphanumhyph_name` | No | false | Clean names |

See `examples/gitprovidersync-complete-example.yaml` for complete reference.

## Precedence

1. CLI flags (`--config-file`, `--log-level`)
2. Environment variables (`GPS_*` prefix)
3. Project config (`./gitprovidersync.yaml`)
4. User config (`~/.config/gitprovidersync/config.yaml`)
5. System config (`/etc/gitprovidersync/config.yaml`)
