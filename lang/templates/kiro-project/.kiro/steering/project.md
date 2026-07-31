---
inclusion: always
---

# Project Conventions

## Build & Run
```bash
pranor build main.pnr -o service.exe
pranor run main.pnr --watch
```

## Testing
```bash
pranor test main.pnr
pranor test --cover main.pnr
```

## Linting & Formatting
```bash
pranor lint main.pnr     # errors + warnings
pranor fmt main.pnr      # auto-format
pranor fmt --check .     # CI check
```

## Project Structure
```
myapp/
├── main.pnr           # Entry point (server, routes)
├── models/            # Struct definitions
│   └── user.pnr
├── handlers/          # Route handler functions
│   └── auth.pnr
├── jobs/              # Scheduled tasks
│   └── cleanup.pnr
├── config.yml         # Runtime configuration
└── tests/             # Test files
    └── user_test.pnr
```

## Configuration
Use `config.yml` for runtime settings, accessed via `config("key")`:
```yaml
db:
  host: "localhost"
  port: "5432"
app:
  secret: "change-me"
```

## Environment Variables
- `PORT` — override server port
- `LOG_FORMAT=json` — JSON structured logging
- `LOG_LEVEL=debug` — log verbosity
- `OTEL_ENDPOINT` — enable OpenTelemetry tracing

## Error Handling Pattern
Prefer the `?` operator for clean error propagation:
```pranor
fn loadUser(id: int) -> User? {
    let row = db.query("SELECT * FROM users WHERE id = ?", id)?
    return User { name: row.name, email: row.email }
}
```

## Code Style
- Use type annotations on public function parameters
- Use `string?` for values that can be nil
- Prefer `stdlib/` imports over reimplementing utilities
- One service concern per file (routes, models, jobs separated)
