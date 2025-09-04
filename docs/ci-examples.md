<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# CI/CD Examples

## Common Setup

```bash
# Download latest release (all examples use this)
wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
tar -xzf gitprovidersync_linux_amd64.tar.gz
chmod +x gitprovidersync
```

## GitHub Actions

```yaml
name: Repository Sync
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Git Provider Sync
        run: |
          wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
          tar -xzf gitprovidersync_linux_amd64.tar.gz
          chmod +x gitprovidersync

      - name: Run Sync
        env:
          GPS_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPS_GITLAB_TOKEN: ${{ secrets.GITLAB_TOKEN }}  # If mirroring to GitLab
        run: ./gitprovidersync sync --config-file sync-config.yaml
```

## GitLab CI

```yaml
sync:
  image: alpine:latest
  before_script:
    - apk add --no-cache wget tar git
    - wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
    - tar -xzf gitprovidersync_linux_amd64.tar.gz
    - chmod +x gitprovidersync
  script:
    - ./gitprovidersync sync --config-file sync-config.yaml
  variables:
    GPS_GITLAB_TOKEN: $CI_JOB_TOKEN
    GPS_GITHUB_TOKEN: $GITHUB_TOKEN  # From CI/CD variables
  only:
    - schedules
```

## Jenkins

```groovy
pipeline {
    agent any
    triggers {
        cron('H 2 * * *')  // Daily at 2 AM
    }
    stages {
        stage('Setup') {
            steps {
                sh '''
                    wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
                    tar -xzf gitprovidersync_linux_amd64.tar.gz
                    chmod +x gitprovidersync
                '''
            }
        }
        stage('Sync') {
            steps {
                withCredentials([
                    string(credentialsId: 'github-token', variable: 'GPS_GITHUB_TOKEN'),
                    string(credentialsId: 'gitlab-token', variable: 'GPS_GITLAB_TOKEN')
                ]) {
                    sh './gitprovidersync sync --config-file sync-config.yaml'
                }
            }
        }
    }
}
```

## Docker

```dockerfile
FROM alpine:latest
RUN apk add --no-cache wget tar git && \
    wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz && \
    tar -xzf gitprovidersync_linux_amd64.tar.gz && \
    chmod +x gitprovidersync && \
    mv gitprovidersync /usr/local/bin/
COPY sync-config.yaml /config/sync-config.yaml
CMD ["gitprovidersync", "sync", "--config-file", "/config/sync-config.yaml"]
```

```bash
# Run container
docker run -e GPS_GITHUB_TOKEN -e GPS_GITLAB_TOKEN your-sync-image
```

## Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: git-sync
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: sync
            image: alpine:latest
            command:
            - /bin/sh
            - -c
            - |
              wget https://github.com/itiquette/git-provider-sync/releases/latest/download/gitprovidersync_linux_amd64.tar.gz
              tar -xzf gitprovidersync_linux_amd64.tar.gz
              chmod +x gitprovidersync
              ./gitprovidersync sync --config-file /config/sync-config.yaml
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
              mountPath: /config
          volumes:
          - name: config
            configMap:
              name: sync-config
          restartPolicy: OnFailure
```

## Configuration File Example

All CI/CD examples reference a config file:

```yaml
gitprovidersync:
  production:
    github-source:
      provider_type: github
      owner: "your-org"
      owner_type: group
      auth:
        token: "${GPS_GITHUB_TOKEN}"  # From environment
      mirrors:
        gitlab-backup:
          provider_type: gitlab
          owner: "backup-org"
          owner_type: group
          auth:
            token: "${GPS_GITLAB_TOKEN}"  # From environment
```

## Tips

- Use scheduled runs for regular backups
- Store tokens as secrets in your CI/CD platform
- Use `--dry-run` first to test configuration
- Set `GPS_LOG_LEVEL=debug` for troubleshooting
- Consider using `--environment` flag for multiple configurations
