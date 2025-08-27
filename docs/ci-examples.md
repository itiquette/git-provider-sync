<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# CI/CD Examples

← [Back to README](../README.adoc)

Examples for running Git Provider Sync in automated environments.

## Installation Pattern

All examples use this installation pattern:

```bash
# Download and extract latest release
wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
tar -xzf gitprovidersync_linux_amd64.tar.gz
```

## GitHub Actions

### Basic Backup Workflow

```yaml
name: Repository Backup
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Git Provider Sync
        run: |
          # Use installation pattern from above
          wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
          tar -xzf gitprovidersync_linux_amd64.tar.gz
          chmod +x gitprovidersync

      - name: Run Backup
        env:
          GPS_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          ./gitprovidersync sync --config-file .github/backup-config.yaml
```

### Mirror to GitLab

```yaml
name: Mirror to GitLab
on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6 AM

jobs:
  mirror:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Git Provider Sync
        run: |
          wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
          tar -xzf gitprovidersync_linux_amd64.tar.gz
          chmod +x gitprovidersync

      - name: Mirror Repositories
        env:
          GPS_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPS_GITLAB_TOKEN: ${{ secrets.GITLAB_TOKEN }}
        run: |
          ./gitprovidersync sync --config-file configs/gitlab-mirror.yaml
```

## GitLab CI

### Backup Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - backup

repository_backup:
  stage: backup
  image: ubuntu:latest
  before_script:
    - apt-get update -qy
    - apt-get install -y wget tar
    - wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
    - tar -xzf gitprovidersync_linux_amd64.tar.gz
    - chmod +x gitprovidersync
  script:
    - ./gitprovidersync sync --config-file gitlab-backup.yaml
  rules:
    - if: $CI_PIPELINE_SOURCE == "schedule"
  variables:
    GPS_GITLAB_TOKEN: $GITLAB_TOKEN
```

## Docker

### Simple Backup Container

```dockerfile
FROM alpine:latest

RUN apk add --no-cache wget tar

WORKDIR /app

RUN wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz && \
    tar -xzf gitprovidersync_linux_amd64.tar.gz && \
    chmod +x gitprovidersync && \
    rm gitprovidersync_linux_amd64.tar.gz

COPY config.yaml .

CMD ["./gitprovidersync", "sync", "--config-file", "config.yaml"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  git-sync:
    build: .
    environment:
      - GPS_GITHUB_TOKEN=${GITHUB_TOKEN}
      - GPS_GITLAB_TOKEN=${GITLAB_TOKEN}
    volumes:
      - ./backups:/app/backups
      - ./config:/app/config
```

## Kubernetes

### CronJob for Regular Backups

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: git-provider-sync
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: git-sync
            image: your-registry/git-provider-sync:latest
            env:
            - name: GPS_GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: git-tokens
                  key: github-token
            - name: GPS_GITLAB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: git-tokens
                  key: gitlab-token
            volumeMounts:
            - name: config
              mountPath: /app/config
            - name: backup-storage
              mountPath: /app/backups
          volumes:
          - name: config
            configMap:
              name: git-sync-config
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

## Configuration Examples

### GitHub to Directory (CI Config)

```yaml
# .github/backup-config.yaml
gitprovidersync:
  ci:
    automated-backup:
      provider_type: github
      domain: github.com
      owner: "your-username"
      owner_type: user
      auth:
        token: "${GITHUB_TOKEN}"
      repositories:
        exclude:
          - "*-temp"
          - "fork-*"
      mirrors:
        backup:
          provider_type: directory
          path: "./backups"
          settings:
            bare: true
```

### Multi-Provider Mirror

```yaml
# configs/multi-mirror.yaml
gitprovidersync:
  production:
    github-to-gitlab:
      provider_type: github
      domain: github.com
      owner: "your-org"
      owner_type: group
      auth:
        token: "${GITHUB_TOKEN}"
      mirrors:
        gitlab-mirror:
          provider_type: gitlab
          domain: gitlab.com
          owner: "mirrored-repos"
          owner_type: group
          auth:
            token: "${GITLAB_TOKEN}"
          settings:
            visibility: private
```

## Best Practices for CI/CD

1. **Use Secrets Management**: Store tokens in your CI platform's secret manager
2. **Dry Run First**: Test configurations with `--dry-run` before production
3. **Schedule Wisely**: Avoid peak hours and rate limits
4. **Monitor Failures**: Set up alerts for failed synchronization jobs
5. **Version Pin**: Use specific versions rather than `latest` for stability

For more configuration options, see [configuration.md](configuration.md).
