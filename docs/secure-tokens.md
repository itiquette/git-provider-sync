<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Secure Token Handling

## Token Configuration Methods

### 1. Provider-Specific Environment Variables (Recommended)

```bash
export GPS_GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
export GPS_GITLAB_TOKEN="glpat_xxxxxxxxxxxxxxxxxxxx"
export GPS_GITEA_TOKEN="gitea_xxxxxxxxxxxxxxxxxxxx"
gitprovidersync sync
```

### 2. Environment Variable Expansion

```yaml
auth:
  token: "${GITHUB_TOKEN}"  # Expands from environment
```

### 3. Token Files

```yaml
auth:
  token_file: /run/secrets/github-token  # Read from file
```

## Security Comparison

| Method | Visible in ps | In History | In Environment | Security |
|--------|---------------|------------|----------------|----------|
| Hardcoded | No | No | No | Never use |
| `GPS_*` env | No | No | Yes | Good for CI/CD |
| `${VAR}` | No | No | Yes | Good |
| `token_file` | No | No | No | Best |

## File Permissions

```bash
touch ~/.tokens/github
chmod 600 ~/.tokens/github
echo "ghp_xxx" > ~/.tokens/github
```

## CI/CD Examples

### GitHub Actions

```yaml
env:
  GPS_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  GPS_GITLAB_TOKEN: ${{ secrets.GITLAB_TOKEN }}
```

### GitLab CI

```yaml
variables:
  GPS_GITLAB_TOKEN: $CI_JOB_TOKEN
  GPS_GITHUB_TOKEN: $GITHUB_TOKEN
```

### Kubernetes

```bash
kubectl create secret generic git-tokens \
  --from-literal=GPS_GITHUB_TOKEN=ghp_xxx
```

## Creating Provider Tokens

| Provider | URL | Required Scopes |
|----------|-----|-----------------|
| GitHub | github.com/settings/tokens | `repo` |
| GitLab | gitlab.com/-/profile/personal_access_tokens | `api`, `read_repository`, `write_repository` |
| Gitea | gitea.example.com/user/settings/applications | All repository scopes |

## Precedence Order

1. Provider-specific environment variables (`GPS_GITHUB_TOKEN`)
2. Environment variable expansion (`token: "${MY_TOKEN}"`)
3. Token file (`token_file: "/path/to/file"`)

## Troubleshooting

Common errors:
- "token appears to be an unexpanded variable": Environment variable not set
- "token file is empty": File exists but contains no token
- "permission denied": Check file permissions (should be 600)
