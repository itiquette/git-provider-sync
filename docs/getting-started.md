<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Quick Start Guide

← [Back to README](../README.adoc)

Get up and running with Git Provider Sync in minutes.

**Prerequisites**: Git Provider Sync installed ([Installation Guide](../INSTALL.adoc))

## 1. Basic Configuration

Create a `gitprovidersync.yaml` file:

```yaml
gitprovidersync:
  production:              # Environment name (production/staging/dev)
    github-backup:         # Descriptive name for this sync
      provider_type: github # Source: github, gitlab, or gitea
      domain: github.com   # Use github.com or your GitHub Enterprise domain
      owner: "your-username" # Replace with your GitHub username/org
      owner_type: user     # Use 'user' for personal, 'group' for organizations
      auth:
        token: "${GITHUB_TOKEN}" # Environment variable for security
      mirrors:
        local-backup:      # Descriptive target name
          provider_type: directory # Creates local directory backup
          path: "./backup" # Local path where repos will be backed up
```

## 2. Set Authentication Token

Git Provider Sync uses provider-specific environment variables:

```bash
# For GitHub sources/mirrors
export GPS_GITHUB_TOKEN="ghp_your_github_token"

# For GitLab sources/mirrors  
export GPS_GITLAB_TOKEN="glpat_your_gitlab_token"

# For Gitea sources/mirrors
export GPS_GITEA_TOKEN="your_gitea_token"
```

Alternative: You can also use environment variable expansion in the config:

```yaml
auth:
  token: "${GITHUB_TOKEN}"  # Uses your existing GITHUB_TOKEN env var
```

For detailed authentication setup (token files, SSH, etc.), see the [Secure Tokens Guide](secure-tokens.md) and [Configuration Reference](configuration.md).

## 3. Verify Installation & Configuration

```bash
# Verify installation worked
gitprovidersync --version
# Expected: gitprovidersync version vX.X.X

# Verify your configuration
gitprovidersync print --config-file gitprovidersync.yaml
# Expected: Shows your parsed configuration with resolved environment variables

# See what would be synced (dry run)
gitprovidersync sync --dry-run --config-file gitprovidersync.yaml
# Expected: Lists repositories that would be synced without actually doing it
```

## 4. Run Synchronization

```bash
# Perform actual synchronization
gitprovidersync sync --config-file gitprovidersync.yaml
# Expected: Progress messages showing repositories being cloned/synced to ./backup/
```

## Common Use Cases

### Mirror to Another Provider

```yaml
      mirrors:
        gitlab-mirror:
          provider_type: gitlab
          domain: gitlab.com
          owner: "your-gitlab-username"
          owner_type: user
          auth:
            token: "${GITLAB_TOKEN}"
```

### Archive Repositories

```yaml
      mirrors:
        archive:
          provider_type: archive
          path: "./archive"
          format: tar.gz
```

### Filter Repositories

```yaml
      repositories:
        include:
          - "important-*"
        exclude:
          - "*-temp"
```

## Next Steps

- See [configuration.md](configuration.md) for detailed configuration options
- Check [ci-examples.md](ci-examples.md) for automated deployment examples
- View the complete [usage documentation](usage.adoc) for advanced features
