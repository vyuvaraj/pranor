# Deployment Guide

## Building for Production

```bash
pranor build app.pnr -o myservice.exe
```

The output is a single static binary — no runtime dependencies.

## Docker

Generate a Dockerfile:

```bash
pranor dockerize app.pnr
```

Or manually:

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o pranor main.go
RUN ./pranor build app.pnr -o service

FROM alpine:latest
COPY --from=builder /app/service /service
EXPOSE 8080
CMD ["/service"]
```

## Port Configuration

Priority (highest to lowest):

1. `--port` CLI flag: `./service --port 9090`
2. `PORT` env var: `PORT=9090 ./service`
3. Config file: `server.port: "9090"` in `config.yml`
4. Source declaration: `server "8080"`

## TLS / HTTPS

```pranor
server "443" tls "cert.pem" "key.pem"
```

## Configuration File

Create `config.yml` in the working directory:

```yaml
server:
  port: "8080"

db:
  host: "localhost"
  port: "5432"
  name: "myapp"

log:
  level: "info"
  format: "json"

otel:
  endpoint: "http://collector:4318"
  service: "my-service"
```

Access in code: `config("db.host")` → `"localhost"`

## Config Validation

Fail fast on missing required config:

```pranor
validate {
    required "db.host",
    required "db.port",
    required "app.secret"
}
```

If any key is missing at startup, the service exits with an error message showing which keys are missing and how to set them.

## OpenTelemetry (Tracing & Metrics)

Set environment variables to enable:

```bash
OTEL_ENDPOINT=http://localhost:4318 ./service
OTEL_SERVICE_NAME=my-service ./service
```

**Auto-instrumented:**
- HTTP routes (method, path, status, duration)
- Database queries (operation, statement)
- Cache operations (GET/SET, key)
- HTTP client calls (method, URL, status)
- Pub/sub messaging (publish/subscribe, topic)
- Scheduled jobs (every/cron, interval)
- External calls (Python/Go extern functions)

**Protocol:** OTLP/HTTP JSON — compatible with Jaeger, Tempo, Datadog, Honeycomb, etc.

## Health Checks

Auto-generated endpoints (no code needed):

- `GET /health` — Returns `{"status": "healthy"}`
- `GET /ready` — Returns `{"status": "ready"}`
- `GET /metrics` — Prometheus-style metrics

## Structured Logging

```bash
# JSON output (for log aggregators)
LOG_FORMAT=json ./service

# Set level
LOG_LEVEL=debug ./service
```

Output format (JSON mode):
```json
{"level":"info","message":"Request handled","timestamp":"2024-01-01T00:00:00Z","request_id":"abc123"}
```

## Graceful Shutdown

Pranor services handle `SIGINT` and `SIGTERM`:
1. Stop accepting new connections
2. Wait up to 15 seconds for active requests to complete
3. Close database connections
4. Exit cleanly

## Cross-Compilation

Build for different platforms:

```bash
GOOS=linux GOARCH=amd64 go build -o pranor-linux main.go
./pranor-linux build app.pnr -o service-linux
```

## CI/CD

GitHub Actions and GitLab CI templates are included:
- `.github/workflows/ci.yml`
- `.gitlab-ci.yml`

Both compile all examples, run tests, check formatting, and build release binaries on version tags.

## Publishing & Distribution

### Release Build (All Platforms)

Cross-compile for macOS, Linux, and Windows:

```bash
./release-scripts/build-release.sh v1.0.0
```

This produces archives in `release/`:
```
release/
├── pranor-darwin-amd64.tar.gz
├── pranor-darwin-arm64.tar.gz
├── pranor-linux-amd64.tar.gz
├── pranor-linux-arm64.tar.gz
└── pranor-windows-amd64.zip
```

Each archive contains `pranor` (or `pranor.exe`) and `pranor-lsp` (or `pranor-lsp.exe`).

### GitHub Release

1. Tag the release: `git tag v1.0.0 && git push --tags`
2. Create a GitHub Release from the tag
3. Upload all archives from `release/`
4. Note the SHA256 hashes (needed for Homebrew/Scoop):
   ```bash
   shasum -a 256 release/*.tar.gz release/*.zip
   ```

### Homebrew (macOS/Linux)

Formula: `release-scripts/homebrew/pranor.rb`

**Setup (one-time):**
1. Create a Homebrew tap repo: `github.com/user/homebrew-pranor`
2. Copy `pranor.rb` into the tap repo
3. Update SHA256 hashes and download URLs to point to your GitHub release

**User install:**
```bash
brew tap user/pranor
brew install pranor
```

**Updating for new releases:**
1. Update `version` in `pranor.rb`
2. Update SHA256 hashes
3. Push to the tap repo

### Scoop (Windows)

Manifest: `release-scripts/scoop/pranor.json`

**Setup (one-time):**
1. Create a Scoop bucket repo: `github.com/user/scoop-pranor`
2. Copy `pranor.json` into the bucket repo
3. Update the `hash` and `url` fields with actual release URLs

**User install:**
```powershell
scoop bucket add pranor https://github.com/user/scoop-pranor
scoop install pranor
```

**Updating for new releases:**
1. Update `version`, `url`, and `hash` in `pranor.json`
2. Push to the bucket repo (Scoop's `autoupdate` handles future versions automatically)

### VS Code Extension

Publish script: `release-scripts/publish-vscode.sh`

**Prerequisites:**
```bash
npm install -g @vscode/vsce
```

**First-time setup:**
1. Get a Personal Access Token from https://dev.azure.com (Marketplace scope)
2. Run: `vsce login pranor`

**Publishing:**
```bash
cd vscode-support/extension
vsce package          # Creates .vsix file
vsce publish          # Publishes to VS Code Marketplace
```

Or use the script:
```bash
./release-scripts/publish-vscode.sh
```

**User install (after publishing):**
- VS Code: Search "Pranor Language Support" in Extensions
- CLI: `code --install-extension pranor.pnr-vscode`

### Docker Base Image

Dockerfile: `release-scripts/docker/Dockerfile.base`

**Build and push:**
```bash
docker build -t pranor:latest -f release-scripts/docker/Dockerfile.base .
docker tag pranor:latest ghcr.io/user/pranor:latest
docker push ghcr.io/user/pranor:latest
```

**User usage:**
```dockerfile
FROM ghcr.io/user/pranor:latest
WORKDIR /app
COPY myservice.pnr .
RUN pranor build myservice.pnr -o service
CMD ["./service"]
```

### Complete Release Checklist

1. [ ] Run regression tests: `powershell test_regression.ps1`
2. [ ] Update version in:
   - `release-scripts/homebrew/pranor.rb`
   - `release-scripts/scoop/pranor.json`
   - `vscode-support/extension/package.json`
3. [ ] Cross-compile: `./release-scripts/build-release.sh v1.x.x`
4. [ ] Create GitHub Release, upload archives
5. [ ] Compute SHA256 hashes, update Homebrew formula + Scoop manifest
6. [ ] Push Homebrew tap and Scoop bucket repos
7. [ ] Publish VS Code extension: `./release-scripts/publish-vscode.sh`
8. [ ] Build and push Docker image
9. [ ] Announce release
