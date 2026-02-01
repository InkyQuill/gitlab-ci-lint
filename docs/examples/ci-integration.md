# CI/CD Integration Guide

This guide covers integrating GitLab CI Lint into your CI/CD pipelines.

## GitLab CI Integration

### Basic Integration

Validate your `.gitlab-ci.yml` before running jobs:

```yaml
# .gitlab-ci.yml
stages:
  - lint
  - build
  - test
  - deploy

# Lint stage first
lint:ci:
  stage: lint
  image: alpine:latest
  script:
    - apk add --no-cache curl
    - curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o gitlab-ci-lint
    - chmod +x gitlab-ci-lint
    - ./gitlab-ci-lint --skip-api .gitlab-ci.yml
  only:
    changes:
      - .gitlab-ci.yml

# Your actual jobs
build:
  stage: build
  script:
    - echo "Building..."
  needs:
    - lint:ci
```

### Using CI Job Token

GitLab provides `$CI_JOB_TOKEN` automatically for authentication:

```yaml
lint:ci:
  stage: lint
  image: alpine:latest
  script:
    - apk add --no-cache curl
    - curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o gitlab-ci-lint
    - chmod +x gitlab-ci-lint
    - ./gitlab-ci-lint --token $CI_JOB_TOKEN .gitlab-ci.yml
  only:
    changes:
      - .gitlab-ci.yml
```

### Docker Image

If you have a Docker image with gitlab-ci-lint:

```yaml
lint:ci:
  stage: lint
  image: yourregistry/gitlab-ci-lint:latest
  script:
    - gitlab-ci-lint --token $CI_JOB_TOKEN .gitlab-ci.yml
  only:
    changes:
      - .gitlab-ci.yml
```

### Project-Specific Validation

```yaml
lint:ci:
  stage: lint
  script:
    - gitlab-ci-lint --token $CI_JOB_TOKEN --project $CI_PROJECT_ID .gitlab-ci.yml
  only:
    changes:
      - .gitlab-ci.yml
```

### Multiple CI Files

```yaml
lint:ci:
  stage: lint
  script:
    # Validate main CI file
    - gitlab-ci-lint --token $CI_JOB_TOKEN .gitlab-ci.yml
    # Validate other CI files
    - gitlab-ci-lint --token $CI_JOB_TOKEN ci/docker.gitlab-ci.yml
    - gitlab-ci-lint --token $CI_JOB_TOKEN ci/deploy.gitlab-ci.yml
  only:
    changes:
      - .gitlab-ci.yml
      - ci/**/*.gitlab-ci.yml
```

### Exit Code Handling

```yaml
lint:ci:
  stage: lint
  script:
    - gitlab-ci-lint --token $CI_JOB_TOKEN .gitlab-ci.yml
  allow_failure: false  # Pipeline fails if lint fails
  only:
    changes:
      - .gitlab-ci.yml
```

### Conditional Linting

```yaml
lint:ci:
  stage: lint
  script:
    - if [ -f ".gitlab-ci.yml" ]; then gitlab-ci-lint --token $CI_JOB_TOKEN .gitlab-ci.yml; fi
  allow_failure: true
```

## GitHub Actions Integration

### Basic Setup

```yaml
# .github/workflows/ci-lint.yml
name: CI Lint

on:
  push:
    branches: [main, develop]
    paths:
      - '.gitlab-ci.yml'
  pull_request:
    paths:
      - '.gitlab-ci.yml'

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Download gitlab-ci-lint
        run: |
          curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o gitlab-ci-lint
          chmod +x gitlab-ci-lint

      - name: Lint CI configuration
        env:
          GCL_TOKEN: ${{ secrets.GITLAB_TOKEN }}
        run: ./gitlab-ci-lint .gitlab-ci.yml
```

### With Multiple Files

```yaml
name: CI Lint

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        file:
          - .gitlab-ci.yml
          - ci/docker.gitlab-ci.yml
          - ci/deploy.gitlab-ci.yml
    steps:
      - uses: actions/checkout@v4

      - name: Download gitlab-ci-lint
        run: |
          curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o gitlab-ci-lint
          chmod +x gitlab-ci-lint

      - name: Lint ${{ matrix.file }}
        env:
          GCL_TOKEN: ${{ secrets.GITLAB_TOKEN }}
        run: ./gitlab-ci-lint "${{ matrix.file }}"
```

### JSON Output for Processing

```yaml
name: CI Lint

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download gitlab-ci-lint
        run: |
          curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o gitlab-ci-lint
          chmod +x gitlab-ci-lint

      - name: Lint and parse output
        id: lint
        env:
          GCL_TOKEN: ${{ secrets.GITLAB_TOKEN }}
        run: |
          result=$(./gitlab-ci-lint --output json .gitlab-ci.yml)
          echo "result=$result" >> $GITHUB_OUTPUT

          valid=$(echo "$result" | jq -r '.valid')
          echo "valid=$valid" >> $GITHUB_OUTPUT

      - name: Check result
        if: steps.lint.outputs.valid != 'true'
        run: exit 1
```

## Jenkins Integration

### Pipeline Example

```groovy
pipeline {
    agent any

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Lint CI') {
            steps {
                sh '''
                    if [ ! -f "gitlab-ci-lint" ]; then
                        curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o gitlab-ci-lint
                        chmod +x gitlab-ci-lint
                    fi
                    ./gitlab-ci-lint --token ${GITLAB_TOKEN} .gitlab-ci.yml
                '''
            }
        }
    }

    environment {
        GITLAB_TOKEN = credentials('gitlab-token')
    }
}
```

## CircleCI Integration

```yaml
# .circleci/config.yml
version: 2.1

orbs:
  gitlab-ci-lint: your-organization/gitlab-ci-lint@1.0.0

workflows:
  lint-workflow:
    jobs:
      - gitlab-ci-lint/lint:
          file: .gitlab-ci.yml
```

## Pre-commit Hooks

### Using Husky (Node.js)

```json
// package.json
{
  "name": "my-project",
  "scripts": {
    "prepare": "husky install"
  },
  "devDependencies": {
    "husky": "^8.0.0"
  }
}
```

```bash
# .husky/pre-commit
#!/bin/sh
. "$(dirname "$0")/_/husky.sh"

gitlab-ci-lint --skip-api .gitlab-ci.yml
```

### Using pre-commit (Python)

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: gitlab-ci-lint
        name: GitLab CI Lint
        entry: gitlab-ci-lint --skip-api
        language: system
        files: '\.gitlab-ci\.ya?ml$'
```

## Docker Integration

### Dockerfile with Validation

```dockerfile
FROM alpine:latest AS lint

RUN apk add --no-cache curl

RUN curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o /usr/local/bin/gitlab-ci-lint && \
    chmod +x /usr/local/bin/gitlab-ci-lint

# Your actual image
FROM python:3.11

COPY --from=lint /usr/local/bin/gitlab-ci-lint /usr/local/bin/gitlab-ci-lint

WORKDIR /app
COPY . .

RUN gitlab-ci-lint --skip-api .gitlab-ci.yml

# Continue with your build...
```

### Docker Compose

```yaml
version: '3.8'

services:
  lint:
    image: alpine:latest
    command:
      - sh
      - -c
      - |
        apk add --no-cache curl &&
        curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o /tmp/gitlab-ci-lint &&
        chmod +x /tmp/gitlab-ci-lint &&
        /tmp/gitlab-ci-lint --skip-api /app/.gitlab-ci.yml
    volumes:
      - .:/app
```

## Terraform Integration

### Validate CI files in Terraform-managed projects

```hcl
# terraform/main.tf
resource "gitlab_project" "example" {
  # ... project configuration
}

resource "null_resource" "lint_ci" {
  triggers = {
    ci_hash = filesha256(".gitlab-ci.yml")
  }

  provisioner "local-exec" {
    command = "gitlab-ci-lint .gitlab-ci.yml"
  }

  depends_on = [gitlab_project.example]
}
```

## Kubernetes Integration

### Admission Controller (Conceptual)

Validate GitLab CI files before committing to GitOps repository:

```yaml
# webhook-validation.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: gitlab-ci-lint
webhooks:
  - name: validate-gitlab-ci
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["configmaps"]
    clientConfig:
      service:
        name: gitlab-ci-lint-webhook
        namespace: default
        path: /validate
    admissionReviewVersions: ["v1"]
    sideEffects: None
```

## Monitoring and Alerting

### Prometheus Metrics (Conceptual)

```bash
# Script that runs periodically and exports metrics
#!/bin/bash

gitlab-ci-lint --output json .gitlab-ci.yml > /tmp/result.json

valid=$(jq -r '.valid' /tmp/result.json)
errors=$(jq '[.results[].errors] | add | length' /tmp/result.json)
warnings=$(jq '[.results[].warnings] | add | length' /tmp/result.json)

# Export to Prometheus node exporter
echo "# HELP gitlab_ci_lint_valid CI configuration validity"
echo "# TYPE gitlab_ci_lint_valid gauge"
echo "gitlab_ci_lint_valid ${valid}"

echo "# HELP gitlab_ci_lint_errors Number of CI errors"
echo "# TYPE gitlab_ci_lint_errors gauge"
echo "gitlab_ci_lint_errors ${errors}"

echo "# HELP gitlab_ci_lint_warnings Number of CI warnings"
echo "# TYPE gitlab_ci_lint_warnings gauge"
echo "gitlab_ci_lint_warnings ${warnings}"
```

## Best Practices

### 1. Run Lint Early

```yaml
# ✅ Good: Lint stage is first
stages:
  - lint
  - build
  - test
  - deploy

lint:ci:
  stage: lint
  script:
    - gitlab-ci-lint .gitlab-ci.yml
```

### 2. Fail Fast

```yaml
# ✅ Good: Fail immediately on invalid config
lint:ci:
  stage: lint
  script:
    - gitlab-ci-lint .gitlab-ci.yml
  allow_failure: false
```

### 3. Cache the Binary

```yaml
# ✅ Good: Cache gitlab-ci-lint binary
.cache:
  paths:
    - gitlab-ci-lint

lint:ci:
  stage: lint
  cache:
    policy: pull
  script:
    - |
      if [ ! -f "gitlab-ci-lint" ]; then
        curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v1.0.0/gitlab-ci-lint-linux-amd64 -o gitlab-ci-lint
        chmod +x gitlab-ci-lint
      fi
      ./gitlab-ci-lint .gitlab-ci.yml
```

### 4. Use Project-Specific Validation

```yaml
# ✅ Good: Validate with project context
lint:ci:
  stage: lint
  script:
    - gitlab-ci-lint --project $CI_PROJECT_ID .gitlab-ci.yml
```

### 5. Validate on Changes Only

```yaml
# ✅ Good: Only run when CI files change
lint:ci:
  stage: lint
  script:
    - gitlab-ci-lint .gitlab-ci.yml
  only:
    changes:
      - .gitlab-ci.yml
      - ci/**/*.yml
```

### 6. Use Local-Only for Speed

```yaml
# ✅ Good: Fast feedback during development
lint:ci-syntax:
  stage: lint
  script:
    - gitlab-ci-lint --skip-api .gitlab-ci.yml

lint:ci-full:
  stage: lint
  script:
    - gitlab-ci-lint .gitlab-ci.yml
  only:
    - main
    - develop
```

## Troubleshooting CI Integration

### Issue: CI Job Token Not Working

**Solution**: Ensure `read_api` scope is enabled (default for CI job tokens in GitLab 14+)

### Issue: Timeouts in CI

**Solution**: Increase timeout
```bash
gitlab-ci-lint --timeout 120s .gitlab-ci.yml
```

### Issue: Different Behavior in CI vs Local

**Possible causes**:
1. Different GitLab instance
2. Different token permissions
3. Missing environment variables

**Solution**: Use `--verbose` to debug
```bash
gitlab-ci-lint --verbose .gitlab-ci.yml
```
