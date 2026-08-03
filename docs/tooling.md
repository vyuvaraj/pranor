# Tooling & IDE Support

Pranor provides dedicated developer tooling and IDE integration to support language editing, code refactoring, diagnostics, and visual cloud control.

---

## 1. VS Code Extension (`pranor-vscode`)

The official **Pranor Platform & Language Tools** extension converts VS Code into an integrated control plane for your entire microservice infrastructure.

### Installation
- Search for `Pranor Platform & Language Tools` in the VS Code Marketplace, or install the `.vsix` package:
  ```bash
  code --install-extension pranor-vscode-1.0.0.vsix
  ```

### Key Features & Control Panels
- **Language Intelligence**: Syntax highlighting, formatting, diagnostics, and code lens shortcuts for `.pnr` files.
- **Interactive API Client Panel (`pranor.apiClient`)**: In-editor HTTP test runner targeting Pranor Gate API routers.
- **Live Event Stream & DLQ Tailer Panel (`pranor.tailPulseEvents`)**: Tail Pranor Pulse event topics in real-time with one-click Dead Letter Queue (DLQ) replay.
- **S3 & Vector Search Explorer (`pranor.vectorSearch`)**: Browse Pranor Vault object buckets and run natural language HNSW cosine vector searches directly inside VS Code.
- **Live Distributed Flamegraph Viewer (`pranor.flamegraphLogs`)**: Trace execution bottlenecks with CPU/latency flamegraphs correlated side-by-side with log entries by `trace_id`.
- **Visual Secret Console (`pranor.secretConsole`)**: Manage cluster master keys, unseal vault stores, and inspect environment secrets.
- **Multi-Cluster Infrastructure Dashboard (`pranor.clusterDeployments`)**: Monitor multi-region cluster health and trigger one-click blue/green canary deployments.

---

## 2. Language Server Protocol (`pranor-lsp`)

`pranor-lsp` is an enterprise-grade Language Server implementing the Language Server Protocol (LSP) specification for standard editor integration (VS Code, Neovim, Emacs, Sublime Text, JetBrains).

### Capabilities
- **Workspace-Wide Rename (`textDocument/rename`)**: Safe multi-file refactoring emitting `WorkspaceEdit` diffs across all workspace `.pnr` files.
- **Auto-Imports & Code Actions (`textDocument/codeAction`)**: Quick-fixes including missing `use std/...` imports and error handler stubs.
- **Fuzzy Workspace Symbol Search (`workspace/symbol`)**: High-performance background symbol indexing (`Ctrl+T` / `Cmd+T`).
- **Call Hierarchy Navigation (`textDocument/prepareCallHierarchy`)**: Visual incoming and outgoing call tree inspection for functions and HTTP routes.
- **Chained Type Inference (`textDocument/completion`)**: Context-aware member completions for chained calls (e.g. `db.query().first()`, `encoding.base64.`).
- **Document Highlighting & Incremental Sync (`textDocument/documentHighlight`)**: Zero-latency symbol occurrence highlighting on cursor focus.

### Standalone Server Usage
Start `pranor-lsp` via standard stdin/stdout JSON-RPC:
```bash
pranor lsp
# or directly:
pranor-lsp
```
