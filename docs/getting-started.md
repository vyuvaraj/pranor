# Getting Started with Pranor

Build and deploy a backend service in 5 minutes.

## Prerequisites

- **Go 1.22+** installed ([download](https://golang.org/dl/))
- A terminal (bash, PowerShell, or cmd)

## Install

### macOS / Linux (Homebrew)

```bash
brew tap vyuvaraj/pranor
brew install pranor
```

### Windows (Scoop)

```powershell
scoop bucket add pranor https://github.com/vyuvaraj/scoop-pranor
scoop install pranor
```

### From Source

```bash
git clone https://github.com/vyuvaraj/pranor.git
cd pranor/lang
go build -o pranor .
# Add to PATH or move to /usr/local/bin
```

### Verify

```bash
pranor --version
```

## Create Your First Service

```bash
pranor init myapp
cd myapp
```

This creates a `main.pnr` file:

```pranor
server "8080"
database "sqlite://app.db"

migration "create_users" {
    db.query("CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT UNIQUE
    )")
}

export route "GET" "/api/users" (req) {
    let users = db.query("SELECT * FROM users")
    return { "users": users }
}

export route "POST" "/api/users" (req) {
    let name = req.body.name
    let email = req.body.email
    db.query("INSERT INTO users (name, email) VALUES (?, ?)", name, email)
    return { "status": "created" }
}
```

## Run

```bash
pranor run main.pnr --watch
```

Your API is now running at `http://localhost:8080`. The `--watch` flag auto-reloads on file changes.

## Test It

```bash
# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'

# List users
curl http://localhost:8080/api/users
```

## Build for Production

```bash
pranor build main.pnr -o myapp
./myapp  # Single binary, no runtime needed
```

## Add More Capabilities

Pranor modules extend your service without glue code:

```pranor
// Add a scheduled task
every 5m {
    log.info("Running cleanup...")
    db.query("DELETE FROM sessions WHERE expires_at < datetime('now')")
}

// Add caching
cache "in-memory"

export route "GET" "/api/users/:id" (req) {
    let cached = cache.get("user:" + req.params.id)
    if cached != nil { return cached }
    
    let user = db.query("SELECT * FROM users WHERE id = ?", req.params.id)
    cache.set("user:" + req.params.id, user, 300)
    return user
}
```

## Deploy

```bash
# Docker
pranor deploy --target docker

# Kubernetes
pranor deploy --target k8s --namespace production
```

## What's Next

- [Language Reference](./language/syntax.md) — Full syntax guide
- [Module Docs](./modules/) — Gate, Pulse, Vault, Auth, and more
- [Deployment Guide](./deployment/docker.md) — Production Docker/K8s setup
- [Architecture Overview](./architecture/overview.md) — How modules connect
