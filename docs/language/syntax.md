# Pranor: A Programming Language for Background Services

Pranor is a modern, high-level DSL (Domain-Specific Language) designed specifically for building background services, schedulers, event-driven applications, and API microservices. It compiles directly into native binaries via Go code generation, providing high performance, low resource consumption, and rapid development.

---

## Table of Contents
- [Key Features](#key-features)
- [Getting Started](#getting-started)
- [Editor Support](#editor-support)
- [CLI Commands Reference](#cli-commands-reference)
- [Language Syntax Guide](#language-syntax-guide)
- [Multi-File Import System](#multi-file-import-system)
- [Async & Concurrent Primitives](#async--concurrent-primitives)
- [Multi-Target Code Generation](#multi-target-code-generation)
- [Breaking Change Detector](#breaking-change-detector)
- [Web Playground](#web-playground)
- [Standard Library](#standard-library)
- [Package Management](#package-management)
- [Testing Support](#testing-support)
- [Compilation & Deployment](#compilation--deployment)
- [Documentation](#documentation)

---

## Key Features

- **Declarative Infrastructure**: Routes, schedulers, pub/sub, databases, caches, and WebSockets as language keywords — not library calls.
- **Compiles to Native Binaries**: Go code generation → single binary deployment. No runtime dependencies.
- **Optional Static Typing**: Gradual type system with `int`, `float`, `string`, `bool`, optional types (`T?`), union types (`T | error`), and generics with constraints.
- **48 Standard Library Modules**: Auth, JWT, retry, circuit breaker, pagination, CORS, rate limiting, validation, and more — written in Pranor itself.
- **Built-in Test Framework**: `test "name" { assert expr }` blocks with structured assertion messages.
- **Multiple Database Backends**: SQLite, PostgreSQL, Oracle, MongoDB — same `db.query()` API.
- **Multiple Broker Backends**: Kafka, NATS, RabbitMQ, MQTT — same `subscribe`/`publish` syntax.
- **Concurrency Primitives**: `spawn`, `async`/`await`, channels, worker pools.
- **Middleware & Auth**: Declarative middleware with `use [auth, logging]` on routes.
- **Python Interop**: Call Python scripts via `extern fn` bindings.
- **Go Package FFI**: Import any Go package with `pranor add <package>` and auto-generated declarations.
- **VS Code Extension**: Full LSP with diagnostics, autocomplete, hover, go-to-definition, and 30+ snippets.
- **OpenTelemetry & Prometheus**: Built-in tracing and metrics export.
- **Docker Support**: `pranor dockerize` generates production-ready Dockerfiles.
- **Multi-File Import System**: Import types and schemas across `.pnr` files with cross-file type resolution and circular import detection.
- **`async` Task & `concurrent {}` Primitives**: First-class language support for async task execution and parallel concurrent blocks.
- **Multi-Target Code Generation**: Generate Rust (`pranor generate --lang rust`) or Python (`pranor generate --lang python`) client code from `.pnr` service definitions.
- **Breaking Change Detector**: `pranor diff old.pnr new.pnr` detects field removals, type changes, and new required fields — safe to use in CI pipelines.
- **Zero-Install WASM Playground**: Try Pranor in the browser at [playground.pranor.dev](https://playground.pranor.dev) — runs the full compiler in WebAssembly, no install required.

---

## Getting Started

### Prerequisites
- **Go**: Version 1.22 or higher is required to build the compiler and execute Go-transpiled code.
- **Python 3.x**: Optional (needed if using Python external functions).

### Install via Scoop (Windows)
```powershell
scoop bucket add pranor https://github.com/vyuvaraj/scoop-pranor
scoop install pranor
```

### Install via Homebrew (macOS / Linux)
```bash
brew tap vyuvaraj/pranor
brew install pranor
```

### Install via Script (Windows)
```powershell
irm https://raw.githubusercontent.com/vyuvaraj/Pranor/main/release-scripts/install.ps1 | iex
```

### Build from Source
```bash
git clone https://github.com/vyuvaraj/Pranor.git
cd Pranor
go build -o pranor.exe .
```

Add the binary to your system PATH for global access.

---

## Editor Support

### VS Code Extension
Install **Pranor Language Support** from the VS Code Marketplace (or from `.vsix` in the repo):
- Syntax highlighting for `.pnr` files
- Real-time diagnostics (type errors, unused variables, missing returns)
- Autocomplete and hover information
- Go-to-definition across files
- 30+ code snippets (`route`, `fn`, `struct`, `test`, `every`, `subscribe`, etc.)
- Commands: Run (`Ctrl+Shift+R`), Build (`Ctrl+Shift+B`), Test (`Ctrl+Shift+T`)
- Format on save

---

## CLI Commands Reference

| Command | Description |
|---------|-------------|
| `pranor build <file.pnr> [-o output]` | Compile to native binary |
| `pranor run <file.pnr> [--watch]` | Compile and run (with optional hot-reload) |
| `pranor test <file.pnr> [--cover] [--filter name]` | Run test blocks |
| `pranor lint <file.pnr>` | Check syntax and static analysis |
| `pranor fmt <file.pnr> [--check]` | Format code (4-space indent) |
| `pranor repl` | Interactive shell |
| `pranor add <go-package>` | Generate `.pnr.d` declaration for a Go package |
| `pranor packages` | List installed package declarations |
| `pranor remove <package>` | Remove a package declaration |
| `pranor install <name>` | Install a community package |
| `pranor publish <dir>` | Publish a package to the registry |
| `pranor init [name]` | Create a new Pranor project |
| `pranor dockerize <file.pnr>` | Generate a production Dockerfile |
| `pranor debug <file.pnr>` | Debug with Delve |
| `pranor audit` | Audit Go dependencies for vulnerabilities |

---

## Language Syntax Guide

### Core Architecture Statements

Pranor allows you to declare global settings and connections dynamically or using values loaded from environment variables:

```pranor
// Declare port dynamically from environment variables
server env("PORT")

// Setup global message broker (options: "in-memory", or Kafka address)
broker "in-memory"

// Setup databases (SQLite, PostgreSQL, Oracle, MongoDB)
database "sqlite://service_data.db"
database env("DATABASE_URL")

// Setup in-memory cache
cache "in-memory"
```

---

### Static Typing & Type Annotations

Pranor supports optional static typing on variables and function signatures. Providing types compiles them directly into native Go types, skipping the performance overhead of runtime `interface{}` conversions.

Supported types: `int`, `string`, `bool`.

#### Variable Type Annotations
Specify types using `: type` after the identifier:
```pranor
let count: int = 100
let label: string = "Items in queue"
let isActive: bool = true
```

#### Function Signature Type Annotations
Specify parameter and return types to optimize function calls and compiler math:
```pranor
fn calculateTotal(base: int, tax: int) -> int {
    let result: int = base + tax
    return result
}
```

---

### Schedulers (`every` & `cron`)

Easily define background routines that run periodically or at scheduled times.

#### Interval Scheduler
Runs a block of code at a specific time duration (e.g., `s` for seconds, `m` for minutes, `h` for hours).
```pranor
every 5s {
    log.info("System healthcheck running...")
}
```

#### Cron Scheduler
Executes using standard cron patterns. Can load patterns from environment variables.
```pranor
cron "0 */2 * * * *" {
    log.info("This runs every 2 minutes.")
}

// Load from environment variable
cron env("BACKUP_CRON") {
    log.warn("Starting system database backup...")
}
```

---

### Web Servers & HTTP APIs (`route`)

Declare HTTP request endpoints with simple routes. Pranor handles request body parsing natively.

```pranor
route "GET" "/status" (req) {
    log.info("Status check requested")
    return { 
        "status": "Pranor is operating normally", 
        "timestamp": time.now() 
    }
}

route "POST" "/webhook" (req) {
    let body = req.body
    log.info("Received body payload: ", body)
    return { "received": true }
}
```

---

### Pub/Sub Broker (`publish` & `subscribe`)

Publish event messages and register subscriptions.

```pranor
// Publish message onto a topic channel
publish "events.incoming" { "user_id": 101, "action": "login" }

// Subscribe to messages on a topic
subscribe "events.incoming" (msg) {
    log.info("Broker received event: ", msg)
}
```

---

### Concurrency & Worker Pools (`spawn`)

You can execute operations asynchronously without blocking the main workflow thread.

#### Fire-and-Forget Goroutines
```pranor
subscribe "incoming.tasks" (msg) {
    // Spawns a lightweight concurrent thread
    spawn processTask(msg)
}
```

#### Rate-Limited Worker Pools
Specify a worker limit to control resource consumption:
```pranor
// Spawns up to 5 concurrent workers maximum
spawn(5) handleHeavyCalculation(data)
```

---

### Database Operations (`db.query`)

Execute queries directly on the configured databases.

#### SQL Databases (SQLite, PostgreSQL, Oracle)
Supports query parsing and placeholders (`?` translates automatically to appropriate placeholders like `$1` dynamically for PostgreSQL).
```pranor
// Create schema table on startup
db.query("CREATE TABLE IF NOT EXISTS metrics (id INTEGER PRIMARY KEY, ts TEXT)")

// Insert records
db.query("INSERT INTO metrics (ts) VALUES (?)", time.now())

// Read records
let results = db.query("SELECT * FROM metrics LIMIT 5")
log.info("Metrics: ", results)
```

#### MongoDB Operations
Executes collection queries using standardized document queries:
```pranor
let result = db.query("insert", "logs", "{\"service\": \"Pranor\", \"action\": \"db_test\"}")
```

---

### Cache Operations (`cache.set` & `cache.get`)

Leverage native in-memory caching to save and read states quickly:

```pranor
// Set key with cache TTL (Time to Live)
cache.set("session_user_1", { "id": 1, "role": "admin" }, "10m")

// Fetch value from cache
let session = cache.get("session_user_1")
log.info("Active Session: ", session)
```

---

### S3 & Pranor Vault Client Operations (`s3`)

Interact with S3-compatible endpoints or a Pranor Vault gateway using the native `s3` runtime functions. You can also import the helper wrapper from the standard library:

```pranor
import { newClient, put, get, deleteObject, list, at, search } from "stdlib/s3.pnr"

// Initialize client
let client = newClient("http://localhost:8080", "admin", "adminsecret")

// Create and configure a bucket
client.createBucket("my-bucket")
client.setBucketVersioning("my-bucket", true)

// Upload and retrieve objects
client.put("my-bucket", "config.json", "{\"status\": \"active\"}")
let content = client.get("my-bucket", "config.json")
log.info("Content: ", content)

// Time-travel to retrieve previous versions of an object (Pranor Vault only)
let historicalContent = client.at("my-bucket", "config.json", "2026-06-15T09:00:00Z")

// Perform semantic search queries (Pranor Vault only)
let searchResults = client.search("my-bucket", "find active config files", 5)
```

---

### Python Interoperability (`extern fn`)

Map complex algorithms or specialized Python libraries directly to Pranor functions:

```pranor
// Map external Python method
extern fn analyzeText(text) from "python:./scripts/analyzer.py:analyze"

let result = analyzeText("Hello world!")
log.info("Python output: ", result)
```

---

### Built-in Functions & Utilities

#### JSON Support
```pranor
let obj = json.parse("{\"status\": true}")
let rawString = json.stringify(obj)
```

#### String Interpolation (f-strings)
```pranor
let name = "Pranor"
let statusMessage = f"System: {name} is running!"
```

#### Pattern Matching (`match`)
```pranor
match eventType {
    "PAYMENT_COMPLETED" => {
        log.info("Processing checkout success...")
    }
    "USER_LOGOUT" => {
        log.info("Cleaning session...")
    }
    _ => {
        log.warn("Unknown event category received")
    }
}
```

#### Exception Handling (`try-catch`)
```pranor
try {
    let res = http.get("http://invalid-url.com")
} catch (err) {
    log.error("HTTP request failed: ", err)
}
```

---

## Web Playground

Pranor includes an interactive Web Playground for trying the language in-browser.

- **WASM Compiler**: Syntax analysis and formatting run client-side
- **Sandbox Runner**: Compiles and executes code server-side with auto-termination

```bash
go build -o web_playground/server/server.exe web_playground/server/main.go
./web_playground/server/server.exe
# Open http://localhost:8080
```

---

## Standard Library

Pranor ships with 48 importable modules written in Pranor itself:

| Category | Modules |
|----------|---------|
| **Auth & Security** | auth, jwt, crypto, cors, sanitize, ip |
| **Resilience** | retry, circuit_breaker, timeout, semaphore, dlq |
| **HTTP** | http_client, response, middleware, ratelimit, webhook |
| **Data** | validation, pagination, pagination_cursor, csv, diff, sort, collections |
| **Config & Env** | config, env, feature_flags |
| **Observability** | tracing, metrics, health, audit |
| **Utilities** | strings_util, datetime, math, url, base64, mask, idempotency, batch, queue |
| **Infra** | s3, cache_patterns, tenant, scheduler, job, graceful |

Import with:
```pranor
import { hashPassword, verifyPassword } from "stdlib/crypto"
import { ok, notFound, created } from "stdlib/response"
import { retry } from "stdlib/retry"
```

---

## Package Management

### Publishing
```bash
pranor publish <package-dir>
```

### Installing
```bash
pranor install <package-name>
```

### Using
```pranor
import { Helper, helperFunc } from "mypkg"
```
Resolves to `packages/mypkg/index.pnr` or `packages/mypkg/main.pnr`. Only `export`-marked declarations are accessible.

---

## Testing Support

Pranor includes a native test harness built into the language itself. This makes it trivial to write unit tests alongside your code and verify logic without external framework setups.

### Defining Tests
Add `test` blocks and use the `assert` statement to check variables:
```pranor
fn doubleValue(val) {
    return val * 2
}

test "doubling math verification" {
    assert doubleValue(2) == 4
    assert doubleValue(5) == 10
}

test "check string comparison" {
    let val = "Pranor" + "Lang"
    assert val == "ServLang"
}
```

### Running Tests
Execute:
```bash
pranor test test_sample.pnr
```

*Output:*
```
Running tests from test_sample.pnr...
=== RUN   Test_DoublingMathVerification
--- PASS: Test_DoublingMathVerification (0.00s)
=== RUN   Test_CheckStringComparison
--- PASS: Test_CheckStringComparison (0.00s)
PASS
ok  	pranor/.build	1.518s
```

---

## Compilation & Deployment

When `pranor build` or `pranor test` is executed, the compiler compiles the input `.pnr` code into a temporary directory called `.build`.

Inside `.build`:
1. `service.go`: Synthesizes code for all declarations, routes, and background routines.
2. `main.go`: Provides the service runtime engine and entry points.
3. `pranor_test.go`: Aggregates the `test` blocks translated to Go's native testing framework.

The output binary compiles out all debug logs and features a fast, low-overhead native runtime engine.

---

## Documentation

- [Language Reference](../language/syntax.md) — Full syntax and type system
- [Getting Started](../getting-started.md) — First project walkthrough
- [Standard Library](../language/stdlib.md) — All modules documented
- [CLI Reference](../language/cli.md) — All commands and flags
- [Deployment Guide](../deployment/docker.md) — Docker, TLS, observability
- [Examples](../language/examples.md) — Examples & technical articles

---

## Multi-File Import System

Pranor supports splitting service definitions across multiple `.pnr` files. Types and schemas defined in one file can be imported and used in another:

```pranor
// types/user.pnr
type User {
    id:    string
    name:  string
    email: string
}
```

```pranor
// services/auth.pnr
import "./types/user.pnr"

route POST "/api/users" {
    body: User
    handler: createUser
}
```

**Features:**
- Cross-file type resolution with full type checking
- Circular import detection with descriptive error messages
- Wildcard imports: `import "./types/*"`
- Re-export: `export type AdminUser extends User { role: string }`

---

## Async & Concurrent Primitives

Pranor provides first-class `async` and `concurrent` language constructs:

```pranor
// Async task — fire and forget
async fn sendNotification(userID: string) {
    call POST "http://notifications/send" { userID }
}

route POST "/api/orders" {
    handler: fn(req) {
        let order = db.insert("orders", req.body)

        // Fire async — doesn't block the response
        async sendNotification(req.body.userID)

        return { order_id: order.id }
    }
}
```

```pranor
// Concurrent block — run steps in parallel, collect results
route GET "/api/dashboard" {
    handler: fn(req) {
        let results = concurrent {
            orders:      call GET "http://orders/summary"
            inventory:   call GET "http://inventory/levels"
            analytics:   call GET "http://analytics/today"
        }

        return results  // all three completed in parallel
    }
}
```

---

## Multi-Target Code Generation

Generate type-safe client code for other languages from your Pranor service definitions:

```bash
# Generate Go client (default — used within Pranor compilation)
pranor generate --lang go services/orders.pnr

# Generate Rust client
pranor generate --lang rust services/orders.pnr -o ./clients/rust/

# Generate Python client
pranor generate --lang python services/orders.pnr -o ./clients/python/
```

**Generated Rust client:**

```rust
// Auto-generated by `pranor generate --lang rust`
pub struct OrdersClient { base_url: String }
impl OrdersClient {
    pub async fn create_order(&self, body: CreateOrderRequest) -> Result<Order, Error> { ... }
    pub async fn get_order(&self, id: &str) -> Result<Order, Error> { ... }
}
```

**Generated Python client:**

```python
# Auto-generated by `pranor generate --lang python`
class OrdersClient:
    def create_order(self, body: CreateOrderRequest) -> Order: ...
    def get_order(self, id: str) -> Order: ...
```

---

## Breaking Change Detector

`pranor diff` compares two `.pnr` files (or two git revisions) and detects breaking API changes — safe to run in CI before merging:

```bash
pranor diff api/v1/orders.pnr api/v2/orders.pnr
```

**Example output:**

```
⚠️  BREAKING CHANGES DETECTED

[FIELD_REMOVED]   Order.discount_code      (line 12 → removed)
[TYPE_CHANGED]    Order.total: int → float  (line 8)
[REQUIRED_ADDED]  CreateOrderRequest.currency  (line 23 → now required)

✅ NON-BREAKING CHANGES

[FIELD_ADDED]     Order.created_at (optional)
```

**Detected breaking change categories:**
- Field removals from request/response types
- Type changes (e.g., `int` → `string`)
- Making optional fields required
- Route removal or method change
- Removing enum variants

Use in CI:
```yaml
# GitHub Actions
- run: pranor diff main:api/orders.pnr HEAD:api/orders.pnr
  # Exits with code 1 if breaking changes detected
```

---

## License

Apache 2.0 — see [LICENSE](https://github.com/vyuvaraj/pranor/blob/main/LICENSE)

---

## Links

- **GitHub**: [github.com/vyuvaraj/pranor](https://github.com/vyuvaraj/pranor)
- **Playground**: [playground.pranor.dev](https://playground.pranor.dev) — zero-install WASM browser playground
- **VS Code Extension**: Search "Pranor Language Support" in Extensions
- **Issues**: [github.com/vyuvaraj/pranor/issues](https://github.com/vyuvaraj/pranor/issues)
