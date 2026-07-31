# Changelog

## 3.4.0

### Added
- **`async` & `concurrent` Syntax & Snippets (VS.G1)**: Full TextMate grammar highlighting and code snippets for `async fn`, `async` task calls, and `concurrent {}` parallel blocks.
- **`pranorctl` Cluster Administration Integration (VS.G2)**: Command palette command `Pranor: pranorctl Cluster Administration` to run `pranorctl get services`, `pranorctl get nodes`, `pranorctl restart service`, and `pranorctl apply config`.
- **`pranor diff` Breaking Change Detector (VS.G3)**: Command `Pranor: Check Breaking API Changes (pranor diff)` to run schema diffing against git base branch (`main`) with dedicated output logging.
- **Multi-Target Client Code Generation (VS.G4)**: Commands `Pranor: Generate Rust Client Code` (`--lang rust`) and `Pranor: Generate Python Client Code` (`--lang python`).
- **Platform Chaos Control Panel (VS.G5)**: Dedicated Webview panel to trigger/abort network delay, CPU stress, memory pressure, disk throttle, and clock skew faults across cluster nodes (`PL.G3`).
- **WASM Playground & Pranor Console Export (VS.G6)**: Deep-linking command `Pranor: Export Current File to WASM Playground` (`playground.pranor.dev`).
- **`pranord` Single-Binary Unified Console Webview (VS.G7)**: Unified multi-tab webview console (`pranor.openServdConsole`) with auto-detection for `pranord` single-binary status and health rollups.

## 3.3.0

### Added
- **Phase 35 Built-in Namespace Support**: Integrated autocomplete suggestion lists, signature helper tooltips, and hover documentation for all 23 new Phase 35 built-in utility namespaces (including `exec`, `csv`, `yaml`, `diff`, `proto`, etc.) and their sub-namespaces (e.g. `encoding.base64`, `encoding.hex`).

## 3.2.0

### Added
- **Symbol Renaming (CD.114)**: Added workspace-wide rename symbol refactoring support, allowing renaming variables, functions, and structs across all `.pnr` files.

### Fixed
- **Light Theme Sidebar Contrast**: Fixed sidebar action button contrast issues in light themes by using standard VS Code secondary state color variables.

## 3.1.0

### Added
- **`pranor.openPlayground` Command (CD.121)**: Embedded Monaco Web Playground directly inside a VS Code Webview panel, launching a local background compiler sandbox server.
- **Extended `pranor doctor` (17.1)**: Enhanced diagnostics to automatically verify installed local WASM runtimes (node, wasmtime, wasmer) and local plugin/extension versions.
- **WinGet Installer Manifest (PKG.7)**: Created the `Yuvaraj.Pranor.yaml` package manifest under `release-scripts/` to support automated Winget platform setups.

### Fixed
- **LSP Windows URI Normalization**: Fixed a bug where differences in Windows path/URI casing and URL-encoding caused autocomplete lookups to return empty results.
- **Robust JSON-RPC Parser**: Fixed a stream desynchronization hang by parsing multiple incoming headers (e.g. `Content-Type`) correctly and using `io.ReadFull`.
- **Trace Options Fix**: Fixed a `LanguageClient` start hang by correcting the trace configuration type to string `'verbose'` for compatibility with `vscode-languageclient` v9.

## 3.0.7

### Added
- **Project Scaffolding** (CD.117) — `Pranor: New Project from Template` opens a 3-step flow: (1) Quick Pick from 5 templates (API Service, Worker, Scheduled, Full Stack, Minimal); (2) Input project name with validation; (3) Folder picker. Generates `main.pnr`, `tests/`, `pranor.toml`, `.gitignore`, and `README.md` ready to run. Opens the new project immediately.
- **One-Click Deploy** (CD.118) — `Pranor: Deploy to Pranor Deploy` opens an environment picker (Production / Staging / Preview), then shows a dark Webview panel with live build log: compile → test → package → upload → provision → health check → deployed URL. Calls Pranor Deploy API at `:8084`; animates a mock flow when offline.
- **Coverage Line Highlights** (CD.122) — `Pranor: Run Tests with Coverage Highlights` runs `pranor test --coverage`, then paints green-tinted lines for covered code and red-highlighted lines with `✗ uncovered` annotations for uncovered code. Results appear in both the editor and the overview ruler. Falls back to realistic mock coverage when the binary isn’t available. `Pranor: Clear Coverage Highlights` resets all decorations.

## 3.0.6

### Added
- **Pranor Activity Bar Panel** (CD.119) — Dedicated sidebar icon in VS Code's Activity Bar showing all 17 services with live 🟢/🔴 health icons, port numbers, and uptime. Polls Pranor Hub every 6s. Shows mock data with `offline` badge when registry is unreachable. Refresh button in panel title bar.
- **Pranor Tunnel Session Viewer** (CD.120) — `pranor.viewTunnels` Webview dashboard showing active tunnel sessions with client IP, target host:port, protocol, duration, bytes in/out totals. Completes 17/17 service dashboard coverage.
- **Import Auto-Organization** (CD.116) — Three-part feature: (1) Completion provider on `use <Tab>` shows all 18 stdlib modules with description and API signature docs; (2) CodeActions quick-fix lightbulb adds missing `use <module>` when `db.`, `cache.`, `http.` etc. are used without import; (3) `Pranor: Add Missing Imports` command adds all missing imports at once.

## 3.0.5

### Added
- **Inlay Type Hints** (CD.113) — Always-on inline type hints in the editor for `fn` return types (`→ string`) and `let` bindings (`: int`). Infers from return expression patterns: `db.query()` → `Result`, `http.get()` → `Response`, literals → `string`/`int`/`bool`/`float`. Togglable via `pranor.enableInlayHints` setting.
- **Test Gutter Decorations** (CD.115) — Run `pranor test` via the new `Pranor: Run Tests (with Gutter Decorations)` command to paint 🟡 yellow dots on all test blocks before running, then 🟢 green or 🔴 red based on results. Results persist when switching tabs. Parses PASS/FAIL output lines; falls back to exit-code if unstructured. Includes `Pranor: Clear Test Gutter Markers` to reset all decorations.

## 3.0.4

### Added
- **Pranor Test Explorer** — Sidebar panel in Explorer listing all `test "..."` blocks from every `.pnr` file, grouped by file with collapse/expand. Refreshes on save.
- **pranor bench panel** (`pranor.runBench`) — Runs `pranor bench <file>` in terminal and opens a live p50/p99/throughput results panel per route.
- **Pranor Deploy Deployments** (`pranor.viewDeployments`) — Live table of branch preview deployments with URLs, build status, and auto-refresh.
- **Pranor Pool Inspector** (`pranor.inspectPool`) — DB connection pool dashboard showing active/idle/max connections per named pool, with wait-queue alerts.
- **Pranor Notify Queue** (`pranor.inspectMail`) — Email queue dashboard showing queued/sent/bounced counts and per-item status with template names.

## 3.0.3

### Added
- **Pranor Auth Progressive Risk Scoring Dashboard** (`pranor.inspectAuth`) tracing user devices, countries, and MFA step-ups.
- **Interactive REPL Launcher** (`pranor.openREPL`) — Spawns a `pranor repl` terminal inside VS Code for live expression evaluation without a full project build.
- **Pranor Mesh Topology Viewer** (`pranor.viewMesh`) — Renders a live Mermaid.js graph of all mesh service connections, with fallback static topology offline.
- **Pranor Trace Request Tracer** (`pranor.traceRequests`) — Shows distributed trace spans with filterable trace ID, service, operation, latency, and OK/ERROR status. Auto-refreshes every 5s.
- **Pranor Hub Health Monitor** (`pranor.viewRegistry`) — Full table of all registered microservices with live health checks, ports, and uptime. Auto-refreshes every 4s.
- **Status Bar Health Indicator** — Persistent `$(circuit-board) Pranor` item in the editor footer, clicking opens the Registry Monitor. Turns amber with service count when any service is down.

## 3.0.2

### Added
- **Visual DAG Flowchart Designer** (`pranor.visualizeWorkflow`) rendering step sequences using Mermaid.js.
- **Pranor Pulse Broker Explorer** (`pranor.exploreQueue`) listing active partitions and consumer groups.
- **Pranor Vault Bucket Manager** (`pranor.exploreStore`) showing S3 directories.
- **Pranor Lock Contention Dashboard** (`pranor.exploreLocks`) tracing active lock waiters.
- **Pranor Gate Route Simulator** (`pranor.simulateRoute`) validating paths against config routes.
- **Pranor Chrono Scheduler Explorer** (`pranor.exploreCron`) monitoring schedules and smart analysis warnings.
- **Pranor Cache Stats Dashboard** (`pranor.inspectCache`) displaying cache hit ratios.

## 3.0.1

### Fixed
- Colocated LSP path autodetection fixes and cross-platform terminal escaping corrections.

## 3.0.0

### Added
- Full LSP integration (diagnostics, autocomplete, hover, go-to-definition)
- Commands: Run, Build, Test, Watch with keybindings
- Format on Save support via `pranor fmt`
- 30+ code snippets for common patterns
- Real-time diagnostics (type errors, unused variables, missing returns)
- Hover information for all symbols and built-in objects
- Editor title run button for `.pnr` files

### Improved
- TextMate grammar extended for generics, optional types, union types
- Snippet coverage for all language features (MCP tools, migrations, WebSocket, etc.)

## 2.0.0

### Added
- Extended snippet library (structs, methods, error handling, middleware)
- Support for new language features (enums, generics, optional chaining)
- Configuration options for LSP and compiler paths

## 1.0.0

### Added
- Initial release
- TextMate syntax highlighting for `.pnr` files
- Basic code snippets for routes, functions, and schedulers
- Extension icon and branding
