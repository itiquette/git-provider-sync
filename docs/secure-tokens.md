<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Secure Token Handling

Git Provider Sync provides multiple methods for securely passing authentication tokens, following security best practices to avoid exposing sensitive information.

## Security Best Practices

### ❌ Never Hardcode Tokens

Never put actual tokens directly in configuration files that might be committed to version control.

### ✅ Recommended: Provider-Specific Environment Variables

Use provider-specific environment variables for the best balance of security and usability, especially in CI/CD environments.

### ✅ Alternative: File-Based Token Storage

For enhanced security in production, use token files with restrictive permissions.

## Token Configuration Methods (In Order of Precedence)

### 1. Provider-Specific Environment Variables

The recommended approach for CI/CD and containerized environments:

```bash
# Set provider-specific tokens
export GPS_GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
export GPS_GITLAB_TOKEN="glpat_xxxxxxxxxxxxxxxxxxxx"
export GPS_GITEA_TOKEN="gitea_xxxxxxxxxxxxxxxxxxxx"

# Tokens are automatically applied to matching providers
gitprovidersync sync
```

### 2. Environment Variable Expansion in Config

Reference environment variables in your configuration file:

```yaml
gitprovidersync:
  production:
    github-source:
      provider_type: github
      owner: myorg
      owner_type: group
      auth:
        token: "${GITHUB_TOKEN}"  # Expands from environment
    gitlab-mirror:
      provider_type: gitlab
      owner: backup
      owner_type: user
      auth:
        token: "${GITLAB_TOKEN}"  # Different token for different provider
```

### 3. Token Files in Configuration

For production deployments with file-based secrets:

```yaml
gitprovidersync:
  production:
    github-source:
      provider_type: github
      owner: myorg
      owner_type: group
      auth:
        token_file: /run/secrets/github-token  # Read from file
```

## Token File Security

### File Permissions

Ensure token files have restrictive permissions:

```bash
# Create token file with secure permissions
touch ~/.tokens/github
chmod 600 ~/.tokens/github
echo "ghp_xxxxxxxxxxxxxxxxxxxx" > ~/.tokens/github
```

### Docker/Kubernetes Secrets

Use platform-native secret management:

```bash
# Docker with environment variables
docker run -e GPS_GITHUB_TOKEN -e GPS_GITLAB_TOKEN \
  gitprovidersync sync

# Kubernetes with secrets as environment variables
kubectl create secret generic git-tokens \
  --from-literal=GPS_GITHUB_TOKEN=ghp_xxx \
  --from-literal=GPS_GITLAB_TOKEN=glpat_xxx
# Reference in pod spec:
# env:
#   - name: GPS_GITHUB_TOKEN
#     valueFrom:
#       secretKeyRef:
#         name: git-tokens
#         key: GPS_GITHUB_TOKEN
```

## CI/CD Examples

### GitHub Actions

```yaml
- name: Sync repositories
  env:
    GPS_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    GPS_GITLAB_TOKEN: ${{ secrets.GITLAB_TOKEN }}
  run: gitprovidersync sync
```

### GitLab CI

```yaml
sync:
  variables:
    GPS_GITLAB_TOKEN: $CI_JOB_TOKEN
    GPS_GITHUB_TOKEN: $GITHUB_TOKEN  # From CI/CD variables
  script:
    - gitprovidersync sync
```

### Jenkins

```groovy
withCredentials([
    string(credentialsId: 'github-token', variable: 'GPS_GITHUB_TOKEN'),
    string(credentialsId: 'gitlab-token', variable: 'GPS_GITLAB_TOKEN')
]) {
    sh 'gitprovidersync sync'
}
```

## Security Comparison

| Method | Visible in ps | In Shell History | In Environment | Security Rating |
|--------|---------------|------------------|----------------|-----------------|
| Hardcoded in config | ✅ No | ✅ No | ✅ No | ⛔ Never use (VCS risk) |
| `GPS_GITHUB_TOKEN` env | ✅ No | ✅ No | ⚠️ Yes | ✅ Recommended for CI/CD |
| `token: ${ENV_VAR}` | ✅ No | ✅ No | ⚠️ Yes | ✅ Good |
| `token_file:` in config | ✅ No | ✅ No | ✅ No | ✅ Best for production |

Note: Environment variables are suitable for CI/CD where the environment is ephemeral and controlled.

## Troubleshooting

### Token Validation

The tool validates tokens and will report issues:
- Empty tokens
- Unexpanded variables (e.g., `${GITHUB_TOKEN}`)
- Tokens with unusual characters

### Common Issues

## Error: "token appears to be an unexpanded variable"

- Your token literally contains `${...}`  
- Solution: Ensure environment variable is set and expanded

## Error: "token file is empty"

- The specified file exists but contains no token
- Solution: Verify file contents

## Token Precedence Order

When multiple token sources are available, Git Provider Sync uses this precedence:

1. **Provider-specific environment variables** (`GPS_GITHUB_TOKEN`, etc.)
2. **Environment variable expansion** in config (`token: "${MY_TOKEN}"`)
3. **Token file** specified in config (`token_file: "/path/to/file"`)

## Summary

For maximum security:
1. Use provider-specific environment variables (`GPS_GITHUB_TOKEN`, etc.) for CI/CD
2. Use `token_file:` in config for persistent configurations with file-based secrets
3. Store token files with restrictive permissions (600)
4. Use platform-native secret management when available (Docker/K8s secrets)
5. Never hardcode tokens directly in configuration files
6. Never pass tokens as command line arguments
