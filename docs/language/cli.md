# Pranor CLI Reference

The `pranor` command is the single entry point for all development and deployment tasks.

## Usage

```
pranor <command> [options] [arguments]
```

## Commands

### Development

| Command | Description |
|---------|-------------|
| `pranor init <name>` | Create a new Pranor project |
| `pranor run <file.pnr> [--watch]` | Run a .pnr service (with optional hot-reload) |
| `pranor dev <file.pnr>` | Run with hot-reload enabled (alias for `run --watch`) |
| `pranor build <file.pnr> -o <binary>` | Compile to a native binary |
| `pranor test <file.pnr> [--cover] [--filter]` | Run test blocks |
| `pranor lint <file.pnr>` | Static analysis and syntax checking |
| `pranor fmt <file.pnr>` | Auto-format source code |
| `pranor repl` | Interactive Pranor shell |

### Package Management

| Command | Description |
|---------|-------------|
| `pranor add <package>` | Add a Go package dependency |
| `pranor remove <package>` | Remove a package |
| `pranor packages` | List installed packages |
| `pranor publish` | Publish to Pranor Hub registry |

### Deployment

| Command | Description |
|---------|-------------|
| `pranor deploy [--target docker\|k8s]` | Deploy service to Docker or Kubernetes |
| `pranor dockerize <file.pnr>` | Generate a production Dockerfile |

### Infrastructure

| Command | Description |
|---------|-------------|
| `pranor gate` | Manage Pranor Gate (API gateway) |
| `pranor pulse` | Manage Pranor Pulse (message queue) |
| `pranor cache` | Manage Pranor Cache |
| `pranor mesh` | Manage Pranor Mesh (service discovery) |
| `pranor tunnel` | Manage Pranor Tunnel (dev tunneling) |
| `pranor trace` | Manage Pranor Trace (distributed tracing) |
| `pranor lock` | Acquire/release distributed locks |
| `pranor secret` | Manage secrets (inject, unseal) |

### Tooling

| Command | Description |
|---------|-------------|
| `pranor bench <file.pnr>` | Generate load test scripts from routes |
| `pranor doc <file.pnr>` | Generate HTML API documentation |
| `pranor diff <old.pnr> <new.pnr>` | Detect breaking API changes |
| `pranor migrate` | Run database migrations |
| `pranor doctor` | Diagnose environment issues |
| `pranor upgrade` | Check for and apply Pranor updates |
| `pranor audit` | Scan dependencies for vulnerabilities |

## Global Flags

| Flag | Description |
|------|-------------|
| `--version` | Print Pranor version |
| `--help` | Show help for any command |
| `--env <name>` | Set environment profile (dev, staging, prod) |
| `--verbose` | Enable verbose output |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PRANOR_HOME` | Path to Pranor installation (runtime, stdlib) |
| `PRANOR_OTLP_ENDPOINT` | OpenTelemetry collector URL for tracing |
| `PRANOR_DISCOVERY` | JSON service discovery manifest |

## Examples

```bash
# Create and run a project
pranor init my-api
cd my-api
pranor run main.pnr --watch

# Build for production
pranor build main.pnr -o my-api --target linux/amd64

# Run tests with coverage
pranor test main.pnr --cover

# Deploy to Docker
pranor deploy --target docker

# Add a package
pranor add github.com/google/uuid
```
