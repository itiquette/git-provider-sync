<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Configuration Reference

← [Back to README](../README.adoc)

Complete reference for configuring Git Provider Sync.

> **Note**: For configuration precedence and where to place config files, see [Configuration Precedence](configuration-precedence.md).  
> **See Also**: [Environment Variables](environment-variables.md) for detailed environment variable reference.

## Basic Structure

```yaml
gitprovidersync:
  environment-name:        # Environment (e.g., production, staging)
    github-backup:         # Descriptive name for this sync
      # Source provider configuration
      provider_type: github # Source: github, gitlab, or gitea
      owner: "username"     # Repository owner/organization name
      owner_type: user      # 'user' for individuals, 'group' for organizations
      mirrors:
        # Target configurations (where to sync TO)
        local-backup:        # Descriptive target name
          provider_type: directory  # Target: directory, archive, or git provider
          path: "./backup"   # Local path for backup
```

## Source Providers

### GitHub

```yaml
      provider_type: github
      domain: github.com  # or your enterprise domain
      owner: "your-username"
      owner_type: user  # or use owner_type: group for organizations
      auth:
        # Option 1: Auto-detected from GPS_GITHUB_TOKEN env var (recommended)
        # Option 2: Reference any env var using expansion
        token: "${GITHUB_TOKEN}"
        # Option 3: Read from file (most secure for production)
        token_file: "/run/secrets/github-token"
```

### GitLab

```yaml
      provider_type: gitlab
      domain: gitlab.com  # or your self-hosted domain
      owner: "your-username"
      owner_type: user  # or use owner_type: group for groups
      auth:
        # Option 1: Auto-detected from GPS_GITLAB_TOKEN env var (recommended)
        # Option 2: Reference any env var using expansion
        token: "${GITLAB_TOKEN}"
        # Option 3: Read from file (most secure for production)
        token_file: "/run/secrets/gitlab-token"
```

### Gitea

```yaml
      provider_type: gitea
      domain: gitea.example.com  # Your Gitea instance domain
      owner: "your-username"
      owner_type: user           # Gitea typically uses 'user' (not group)
      auth:
        # Option 1: Auto-detected from GPS_GITEA_TOKEN env var (recommended)
        # Option 2: Reference any env var using expansion
        token: "${GITEA_TOKEN}"
        # Option 3: Read from file (most secure for production)
        token_file: "/run/secrets/gitea-token"
```

## Target Types

### Directory (Local Backup)

```yaml
      mirrors:
        local-backup:
          provider_type: directory
          path: "./backup"
          settings:
            bare: false  # Creates working copies (default)
            # bare: true  # Creates bare repositories
```

### Archive (Compressed Backup)

```yaml
      mirrors:
        archive-backup:
          provider_type: archive
          path: "./archive"
          format: tar.gz  # or zip
```

### Git Provider (Mirror)

```yaml
      mirrors:
        target-mirror:
          provider_type: github  # or gitlab, gitea
          domain: target-domain.com
          owner: "target-username"
          owner_type: user
          auth:
            token: "${TARGET_TOKEN}"
          settings:
            visibility: private  # public, private, internal
            # Additional provider-specific options
```

## Repository Filtering

```yaml
      repositories:
        include:                # Only sync repositories matching these patterns
          - "important-*"       # All repos starting with "important-"
          - "project-alpha"     # Exact repo name match
        exclude:                # Skip repositories matching these patterns
          - "*-temp"            # All repos ending with "-temp"
          - "archive-*"         # All repos starting with "archive-"
      # Optional: limit by activity (prevents syncing stale repos)
      active_from_limit: "24h"  # Only sync repos with commits in last 24h
```

## Advanced Options

### Repository Name Transformation

```yaml
      settings:
        alphanumhyph_name: true  # Clean names to alphanumeric + hyphens
        ignore_invalid_name: true  # Skip repos with invalid names
```

### Target-Specific Options

```yaml
      mirrors:
        github-target:
          provider_type: github
          # ... provider config
          settings:
            visibility: private
            description: "Mirrored from source"
            topics: ["mirror", "backup"]
            issues: false
            wiki: false
```

## Configuration Validation

```bash
# Check configuration syntax
gitprovidersync print --config-file your-config.yaml

# Test connection to providers
gitprovidersync sync --dry-run --config-file your-config.yaml
```

## Multiple Environments

Organize different sync configurations:

```yaml
gitprovidersync:
  personal:              # Personal repos backup
    github-backup:
      provider_type: github
      owner: "username"
      owner_type: user
      auth:
        token: "${GITHUB_TOKEN}"
      mirrors:
        local-backup:
          provider_type: directory
          path: "./backup"

  work:                  # Organization mirror
    org-mirror:
      provider_type: github
      owner: "company"
      owner_type: group
      auth:
        token: "${GITHUB_ORG_TOKEN}"
      mirrors:
        gitlab-mirror:
          provider_type: gitlab
          owner: "company-backup"
          owner_type: group
          auth:
            token: "${GITLAB_TOKEN}"
```

## Environment Variable Support

All string values support environment variable substitution using `${VARIABLE_NAME}` syntax:

```yaml
gitprovidersync:
  production:
    github-backup:
      auth:
        token: "${GITHUB_TOKEN}"
      mirrors:
        target:
          auth:
            token: "${GITLAB_TOKEN}"
```

For detailed configuration property reference with all available options, see [Configuration Properties Reference](#configuration-properties-reference) below.

---

## Configuration Properties Reference

### Essential Properties

Most commonly used configuration properties:

| Property | Description | Required | Example |
|----------|-------------|----------|---------|
| `provider_type` | Source provider | ✓ Yes | `github`, `gitlab`, `gitea` |
| `owner` | Repository owner | ✓ Yes | `"my-username"` or `"my-org"` |
| `owner_type` | Owner type | ✓ Yes | `user` or `group` |
| `auth.token` | API token | ✓ Yes | `"${GITHUB_TOKEN}"` |
| `mirrors` | Target configurations | ✓ Yes | See examples above |

### Advanced Properties

Complete reference of all configuration properties:

| Property Path | Description | Required | Default |
|---------------|-------------|----------|---------|
| gitprovidersync.{env}.{source}.domain | Custom domain | Optional | Provider default |
| gitprovidersync.{env}.{source}.include_forks | Include forked repos | Optional | false |
| gitprovidersync.{env}.{source}.active_from_limit | Activity filter | Optional | N/A |
| gitprovidersync.{env}.{source}.use_git_binary | Use system git | Optional | false |

| `auth.http_scheme` | Protocol scheme | Optional | https |
| `auth.cert_dir_path` | Custom certificates | Optional | N/A |
| `auth.proxy_url` | Proxy URL | Optional | N/A |
| `auth.ssh_command` | SSH proxy command | Optional | N/A |
| `repositories.include` | Include patterns | Optional | All |
| `repositories.exclude` | Exclude patterns | Optional | None |
| `mirrors.{target}.provider_type` | Target type | ✓ Yes | N/A |
| `mirrors.{target}.path` | Target path | Optional | N/A |
| `settings.visibility` | Repository visibility | Optional | private |
| `settings.alphanumhyph_name` | Clean names | Optional | false |

For complete configuration details, see the examples above and inline comments in `examples/gitprovidersync-complete-example.yaml`.
