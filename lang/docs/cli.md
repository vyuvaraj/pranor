# CLI Reference

## pranor build

Compile a `.pnr` file to a native binary.

```bash
pranor build <file.pnr> [-o <output>]
```

**Examples:**
```bash
pranor build app.pnr                    # → service.exe
pranor build app.pnr -o myapp.exe       # Custom output name
```

## pranor run

Compile and run immediately.

```bash
pranor run <file.pnr> [--watch]
```

**Options:**
- `--watch` — Watch for file changes and hot-reload

## pranor test

Run tests defined in a `.pnr` file.

```bash
pranor test <file.pnr>            # Run tests
pranor test --cover <file.pnr>    # Run tests with coverage report
```

Runs all `test "name" { ... }` blocks and reports results.

**With `--cover`:** Shows statement coverage percentage and saves a coverage profile to `.build/<hash>/coverage.out`.

## pranor lint

Check syntax and perform static analysis without building.

```bash
pranor lint <file.pnr>
```

**Analysis includes:**
- Parse error detection with "did you mean?" suggestions
- Unused variable warnings
- Missing return detection for typed functions
- Type mismatch errors (wrong argument types/count)

**Exit codes:**
- `0` — No errors (may have warnings)
- `1` — Has parse errors or type errors

**Example output:**
```
  warning: variable 'unused' is declared but never used
   7 |     let unused = 42
            ^

  error: argument 1 of 'add' expects type 'int', got 'string'
   6 |     let result = add("hello", true)
                           ^

2 error(s), 1 warning(s)
```

## pranor fmt

Format a `.pnr` file (4-space indent, consistent style).

```bash
pranor fmt <file.pnr>            # Format in place
pranor fmt --check <file.pnr>    # Check only (exit 1 if unformatted)
```

## pranor repl

Interactive Pranor shell.

```bash
pranor repl
```

**Commands inside REPL:**
- Type any expression to evaluate: `1 + 2`, `"hello".toUpper()`
- `let x = 42` — declare variables (persisted across lines)
- `state` — show all declarations
- `clear` — reset state
- `exit` — quit

## pranor add

Generate a `.pnr.d` declaration file for a Go package.

```bash
pranor add <go-package-path>
```

**Examples:**
```bash
pranor add github.com/google/uuid
pranor add encoding/json
pranor add net/url
```

Downloads the package (if needed) and generates type declarations in `declarations/`.

## pranor packages

List installed package declarations.

```bash
pranor packages
```

## pranor remove

Remove a package declaration.

```bash
pranor remove <package-name>
```

## pranor install

Install a community package from Pranor Hub and resolve its transitive dependencies.

```bash
pranor install <package-name>
```

**Examples:**
```bash
pranor install jwt
pranor install retry
pranor install pagination@1.2.0
```

Downloads the package tarball from the configured registry, extracts it to `packages/<name>/`, then reads its `pranor.toml` `[dependencies]` section and recursively installs any missing transitive dependencies.

**Environment variables:**
- `PRANOR_REGISTRY` — Override the registry URL (default: `https://registry.pranor.org`)

**Output example:**
```
Downloading package from https://registry.pranor.org/packages/jwt.tar.gz...
✓ Package 'jwt' installed to packages/jwt/
  Resolving 2 dependencies...
  ↳ Installing dependency: crypto
  ✓ Package 'crypto' installed to packages/crypto/
  ↳ Installing dependency: base64
  • base64 (already installed)
```

## pranor publish

Publish a package directory to Pranor Hub.

```bash
pranor publish <directory>
```

Creates a `.tar.gz` archive of the directory (which should contain a `pranor.toml`) and uploads it to the configured registry. Requires `PRANOR_JWT_SECRET` environment variable for authentication.

## pranor dockerize

Generate a Dockerfile for deployment.

```bash
pranor dockerize <file.pnr>
```

## pranor migrate

Apply declarative `table` schema migrations to the database.

```bash
pranor migrate [file-or-dir] [--db <connection-string>]
```

**Options:**
- `--db` — Database connection string. Falls back to `$DATABASE_URL` then `sqlite://pranor.db`

**Supported connection strings:**

| Format | Example |
|--------|---------|
| SQLite | `sqlite://app.db` |
| PostgreSQL | `postgres://user:pass@localhost/mydb` |
| MySQL | `mysql://user:pass@localhost/mydb` |

**What it does:**

1. Scans `.pnr` files for `table` declarations
2. Connects to the database
3. **Creates** tables that don't exist (`CREATE TABLE IF NOT EXISTS`)
4. **Adds** new columns to existing tables (`ALTER TABLE ADD COLUMN`)
5. Skips anything already up to date

**Example output:**
```
Found 3 table declaration(s):
  • users (5 columns)
  • posts (6 columns)
  • tags (2 columns)

  ✓ users: schema applied
  ✓ posts: schema applied
  - tags: already up to date

Migration complete: 2 table(s) created/updated.
```

## pranor create

AI-scaffold a new `.pnr` file from a natural language description.

```bash
pranor create "<prompt describing your service>"
```

**Examples:**
```bash
pranor create "a REST API for managing blog posts with SQLite"
pranor create "a webhook receiver that processes Stripe payment events"
```

Requires `PRANOR_AI_KEY` environment variable (OpenAI or Gemini API key).

## pranor dev

Start the full development environment with hot-reload and infrastructure services.

```bash
pranor dev [file.pnr] [--services all]
```

Starts ServDB, Pranor Cache, Pranor Pulse, and Pranor Mesh locally, then watches `.pnr`
files for changes and reloads the compiled service automatically.

## pranor deploy

Deploy a `.pnr` service to a target environment (e.g. Pranor Cloud).

```bash
pranor deploy <file.pnr> [--env <production|staging>] [--token <jwt>]
```

**Options:**
- `--env` — Target hosting environment (defaults to `staging`)
- `--token` — Deployment authentication token (falls back to `$PRANOR_DEPLOY_TOKEN`)

## pranor doctor

Execute diagnostic and compatibility checks across all running local services.

```bash
pranor doctor [--integration]
```

**Options:**
- `--integration` — Boot and test the full 15-service pipeline via docker-compose

## pranor audit

Audit third-party package dependency paths for compatibility and vulnerability risks.

```bash
pranor audit [--fix]
```

**Options:**
- `--fix` — Attempt to resolve minor version conflicts automatically

## pranor status

Query and format the health status of active local network registry nodes.

```bash
pranor status [--json]
```

## pranor repl

Start an interactive read-eval-print shell to execute and inspect Pranor statements.

```bash
pranor repl
```

## pranor packages

Inspect or list downloaded dependency structures and standard library overrides.

```bash
pranor packages [--update]
```

## Runtime Flags


Compiled Pranor binaries accept these flags:

```bash
./myservice.exe --port 9090     # Override server port
./myservice.exe --mcp           # Start as MCP tool server
```

**Environment variables:**
- `PORT` — Override server port
- `LOG_FORMAT=json` — JSON log output
- `LOG_LEVEL=debug` — Set log level
- `OTEL_ENDPOINT=http://localhost:4318` — Enable OpenTelemetry
- `OTEL_SERVICE_NAME=my-service` — Service name for traces
- `DATABASE_URL` — Default database connection string
- `PRANOR_MESH_ADDR` — Pranor Mesh registry address (default: `http://localhost:8089`)
- `PRANOR_SELF_ADDR` — This service's advertised address for mesh registration
