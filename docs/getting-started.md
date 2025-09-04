<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Quick Start Guide

Prerequisites: Git Provider Sync installed ([Installation Guide](../INSTALL.adoc))

## 1. Basic Configuration

Create a `gitprovidersync.yaml` file:

```yaml
gitprovidersync:
  production:
    github-backup:
      provider_type: github
      domain: github.com
      owner: "your-username"
      owner_type: user
      auth:
        token: "${GITHUB_TOKEN}"
      mirrors:
        local-backup:
          provider_type: directory
          path: "./backup"
```

## 2. Set Authentication Token

```bash
export GPS_GITHUB_TOKEN="ghp_your_github_token"   # GitHub
export GPS_GITLAB_TOKEN="glpat_your_gitlab_token" # GitLab
export GPS_GITEA_TOKEN="your_gitea_token"         # Gitea
```

Or use any environment variable in config: `token: "${GITHUB_TOKEN}"`

See [Configuration Reference](configuration.md) for more options.

## 3. Verify Installation & Configuration

```bash
gitprovidersync --version
gitprovidersync print --config-file gitprovidersync.yaml
gitprovidersync sync --dry-run --config-file gitprovidersync.yaml
```

## 4. Run Synchronization

```bash
gitprovidersync sync --config-file gitprovidersync.yaml
gitprovidersync sync --environment production --config-file gitprovidersync.yaml
```

## Next Steps

- See [configuration.md](configuration.md) for detailed configuration options
- Check [ci-examples.md](ci-examples.md) for automated deployment examples
- View the complete [usage documentation](usage.adoc) for advanced features
