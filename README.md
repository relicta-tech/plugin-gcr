# plugin-gcr

Relicta plugin for pushing container images to Google Artifact Registry and Google Container Registry.

## Features

- Push container images to Google Artifact Registry (recommended)
- Push container images to legacy Google Container Registry (GCR)
- Authentication via gcloud CLI or service account
- Multiple image tag support with template variables
- Multi-region deployment support
- Dry-run mode for testing

## Installation

```bash
relicta plugin install gcr
```

## Configuration

### Basic Configuration (Artifact Registry)

```yaml
plugins:
  gcr:
    project: my-gcp-project
    region: us-central1
    repository: my-repo
    image: my-app
    source_image: myapp:latest
    tags:
      - "{{.Version}}"
      - latest
```

### Full Configuration

```yaml
plugins:
  gcr:
    # Use Artifact Registry (recommended) vs legacy GCR
    artifact_registry: true

    # GCP Project ID
    project: my-gcp-project

    # Region for Artifact Registry (e.g., us-central1, europe-west1)
    # For legacy GCR: us, eu, asia
    region: us-central1

    # Repository name (Artifact Registry only)
    repository: my-repo

    # Image name
    image: my-app

    # Source image to push
    source_image: myapp:latest

    # Authentication method
    auth:
      method: gcloud  # or "service_account"
      # key_file: /path/to/service-account.json  # for service_account
      # key_json: ${GCP_SERVICE_ACCOUNT_JSON}    # or inline JSON

    # Tags to apply
    tags:
      - "{{.Version}}"
      - "{{.TagName}}"
      - latest

    # Multi-region push
    multi_region:
      enabled: false
      regions:
        - us
        - europe
        - asia

    # Dry run mode
    dry_run: false
```

### Legacy GCR Configuration

```yaml
plugins:
  gcr:
    artifact_registry: false
    project: my-gcp-project
    region: us  # us, eu, or asia
    image: my-app
    source_image: myapp:latest
    tags:
      - "{{.Version}}"
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `artifact_registry` | bool | No | `true` | Use Artifact Registry (vs legacy GCR) |
| `project` | string | Yes | - | GCP project ID |
| `region` | string | No | `us-central1` | Registry region |
| `repository` | string | Conditional | - | Repository name (required for AR) |
| `image` | string | Yes | - | Image name |
| `source_image` | string | Yes | - | Local Docker image to push |
| `tags` | []string | No | `["{{.Version}}"]` | Image tags to apply |
| `auth.method` | string | No | `gcloud` | Auth method: `gcloud` or `service_account` |
| `auth.key_file` | string | No | - | Path to service account key |
| `auth.key_json` | string | No | - | Service account key JSON |
| `multi_region.enabled` | bool | No | `false` | Enable multi-region push |
| `multi_region.regions` | []string | No | - | Regions to push to |
| `dry_run` | bool | No | `false` | Run without making changes |

## Tag Templates

The following template variables are available for tags:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.Version}}` | Release version | `1.2.3` |
| `{{.PreviousVersion}}` | Previous version | `1.2.2` |
| `{{.TagName}}` | Git tag name | `v1.2.3` |
| `{{.ReleaseType}}` | Release type | `patch` |
| `{{.Branch}}` | Git branch (/ replaced with -) | `feature-foo` |

## Authentication

### gcloud CLI (Default)

Uses the gcloud CLI's credential helper:

```bash
gcloud auth configure-docker us-central1-docker.pkg.dev
```

### Service Account

Provide service account credentials via:

1. **Key file path**:
   ```yaml
   auth:
     method: service_account
     key_file: /path/to/service-account.json
   ```

2. **Environment variable** (key file path):
   ```yaml
   auth:
     method: service_account
     key_file: ${GOOGLE_APPLICATION_CREDENTIALS}
   ```

3. **Inline JSON** (from environment):
   ```yaml
   auth:
     method: service_account
     key_json: ${GCP_SERVICE_ACCOUNT_JSON}
   ```

## Required IAM Roles

### Artifact Registry

- `roles/artifactregistry.writer` - Push images
- `roles/artifactregistry.reader` - Pull images

### Legacy GCR

- `roles/storage.objectAdmin` - Push/pull images

## Hooks

This plugin supports the following hooks:

- `post_publish` - Push images after release is published

## Examples

### Multi-Region Deployment

```yaml
plugins:
  gcr:
    project: my-project
    repository: my-repo
    image: my-app
    source_image: my-app:build
    multi_region:
      enabled: true
      regions:
        - us-central1
        - europe-west1
        - asia-east1
    tags:
      - "{{.Version}}"
      - latest
```

### CI/CD with Service Account

```yaml
plugins:
  gcr:
    project: my-project
    region: us-central1
    repository: my-repo
    image: my-app
    source_image: my-app:${CI_COMMIT_SHA}
    auth:
      method: service_account
      key_json: ${GCP_SERVICE_ACCOUNT_JSON}
    tags:
      - "{{.Version}}"
      - "{{.TagName}}"
```

## License

Apache-2.0
