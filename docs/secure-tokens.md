<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Secure Token Handling

## Token Configuration Methods

See [Environment Variables](environment-variables.md) for provider-specific tokens.

### Token Files (Most Secure)

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

## CI/CD Integration

See [CI/CD Examples](ci-examples.md) for platform-specific token configuration.

## Creating Provider Tokens

| Provider | URL | Required Scopes |
|----------|-----|-----------------|
| GitHub | github.com/settings/tokens | `repo` |
| GitLab | gitlab.com/-/profile/personal_access_tokens | `api`, `read_repository`, `write_repository` |
| Gitea | gitea.example.com/user/settings/applications | All repository scopes |

## Token Precedence

See [Configuration](configuration.md#authentication) for token precedence details.

## Troubleshooting

Common errors:
- "token appears to be an unexpanded variable": Environment variable not set
- "token file is empty": File exists but contains no token
- "permission denied": Check file permissions (should be 600)
