# Integrations & Developer Tooling

Pranor seamlessly integrates with existing cloud-native infrastructure tools, CI/CD pipelines, observability platforms, and IDE containers.

---

## 1. Terraform Provider (`terraform-provider-pranor`)

Declaratively manage your Pranor cloud infrastructure using standard HCL configurations.

### Configuration
```hcl
terraform {
  required_providers {
    pranor = {
      source  = "vyuvaraj/pranor"
      version = "~> 1.0.0"
    }
  }
}

provider "pranor" {
  address = "http://localhost:8096"
  token   = var.pranor_admin_token
}

resource "pranor_bucket" "user_uploads" {
  name       = "user-uploads"
  versioning = true
}

resource "pranor_topic" "order_events" {
  name       = "orders.created"
  partitions = 4
}

resource "pranor_cron_job" "nightly_cleanup" {
  name     = "nightly-cleanup"
  schedule = "0 0 * * *"
  endpoint = "http://api-service:8080/internal/cleanup"
}
```

---

## 2. GitHub Action (`pranor/deploy-action@v1`)

Automate `.pnr` application compilation, artifact packaging, and zero-downtime blue/green deployment directly inside GitHub Actions workflows.

### Workflow Example (`.github/workflows/deploy.yml`)
```yaml
name: Pranor CI/CD Deployment

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to Pranor Cluster
        uses: pranor/deploy-action@v1
        with:
          entrypoint: 'main.pnr'
          output-binary: 'app.pnr'
          deploy-target: 'docker'
          cluster-url: 'https://deploy.pranor.dev'
          api-token: ${{ secrets.PRANOR_DEPLOY_TOKEN }}
          environment: 'production'
```

---

## 3. Observability Integrations

### Prometheus Remote Write Receiver (`Pranor Trace`)
Pranor Trace accepts Prometheus `remote_write` payloads natively. Point existing Prometheus server or agent scrapers directly to Pranor Trace:
```yaml
# prometheus.yml
remote_write:
  - url: "http://pranor-trace:8087/api/v1/prom/remote_write"
```

### OpenTelemetry Collector Exporter (`Pranor Trace`)
Export traces and metrics from the standard OpenTelemetry Collector to Pranor Trace via OTLP/HTTP:
```yaml
# otel-collector-config.yaml
exporters:
  otlphttp/pranor:
    endpoint: "http://pranor-trace:8087"

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [otlphttp/pranor]
```

### Grafana Data Source Plugin
Pranor Trace and Pranor Pulse provide a native Grafana datasource plugin for visualizing distributed traces, span latency distributions, and topic event metrics on Grafana dashboards.
- Connection URL: `http://localhost:8087` (Trace) or `http://localhost:8083` (Pulse).

---

## 4. Onboarding & DX Automation

### Interactive Quickstart Wizard (`pranor quickstart`)
Interactively scaffold new projects with optional module presets (REST API, Auth, Vault, Pulse events, Chrono jobs):
```bash
pranor quickstart
```

### Infrastructure Health Diagnostics (`pranor doctor`)
Run comprehensive system health checks across ports, binary dependencies, Docker environment, and configuration validity:
```bash
pranor doctor
```

### VS Code Dev Container & GitHub Codespaces Template
One-click cloud development container pre-configured with Go 1.22+, `pranor` compiler, `pranor-lsp`, and forwarded ports for `pranord` console (`8096`):
- Open `.devcontainer/devcontainer.json` in VS Code or launch via GitHub Codespaces.
