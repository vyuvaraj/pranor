# Pranor OSS / EE Build-Tag Convention (V2.90.5)

> **Mandatory reading** before contributing any new module or feature to `pranor` or `pranor-ee`.
> Defined in: `requirements_definitive.md §3.3`

---

## Rule Summary

Pranor uses Go build tags to cleanly split open-source and enterprise code. Every module MUST follow this three-file convention:

| File | Build Tag | Location | Purpose |
|------|-----------|----------|---------|
| `feature.go` | *(none)* | `pranor/` | Shared interface, types, constants — compiled in **both** builds |
| `feature_oss.go` | `//go:build !enterprise` | `pranor/` | OSS implementation or stub returning `ErrEERequired` |
| `feature_ee.go` | `//go:build enterprise` | `pranor-ee/src/PranorXxx/` | Full EE implementation |

---

## File Layout Example: `pranor/graph`

```
pranor/graph/
├── api/
│   └── graph.go            ← NO build tag — shared interface
│
├── pkg/context/
│   ├── context.go          ← NO build tag — shared types / interfaces
│   ├── context_oss.go      ← //go:build !enterprise — OSS implementation
│   └── context_test.go     ← NO build tag — tests run on both builds

pranor-ee/src/PranorGraph/
└── pkg/context/
    └── context_ee.go       ← //go:build enterprise — EE implementation
```

---

## Critical Rules

### 1. Shared interface files carry NO build tag
```go
// graph/api/graph.go — NO build tag header
package api

type ContextQuery struct { ... }
type ContextResult struct { ... }

type GraphProvider interface {
    Query(ctx context.Context, q ContextQuery) (ContextResult, error)
}
```

### 2. OSS stubs return `ErrEERequired` — never panic
```go
// graph/pkg/sync/sync_oss.go
//go:build !enterprise

package sync

import "errors"

var ErrEERequired = errors.New("pranor: requires Enterprise Edition")

func SyncCrossDatacenter(ctx context.Context, cfg SyncConfig) error {
    return ErrEERequired   // ✅ Correct
    // panic("EE only")    // ❌ NEVER do this in OSS stubs
}
```

### 3. EE files live ONLY in `pranor-ee` — never in `pranor`
```
pranor/graph/sync_ee.go     ← ❌ FORBIDDEN — EE code in OSS repo
pranor-ee/src/PranorGraph/  ← ✅ CORRECT location for EE code
```

### 4. OSS stubs MUST be API-compatible with EE implementations
The function signature in `feature_oss.go` and `feature_ee.go` MUST be identical. No calling code should ever need to import a build-tag-specific package directly — only the shared interface file.

```go
// Both files MUST have identical signatures:
func SyncCrossDatacenter(ctx context.Context, cfg SyncConfig) error
```

### 5. CGO_ENABLED=0 applies to BOTH OSS and EE builds
The zero-CGO invariant (`requirements_definitive.md §3.1`) applies to both builds. EE implementations MUST also be CGO-free. Heavy ML inference goes through `learn/wasm` or `learn/sidecar`, never via direct CGO.

---

## Scaffold Command

Use the templates in `_templates/` to create new OSS stubs and EE skeletons:

```bash
# Copy and customise the OSS stub
cp _templates/module_oss.go.tmpl pranor/graph/pkg/sync/sync_oss.go
# Edit: replace all {{PLACEHOLDER}} tokens

# Copy and customise the EE skeleton (in pranor-ee repo)
cp _templates/module_ee.go.tmpl pranor-ee/src/PranorGraph/pkg/sync/sync_ee.go
# Edit: replace all {{PLACEHOLDER}} tokens
```

---

## Build Commands

```bash
# OSS build (default — no tags needed)
CGO_ENABLED=0 go build ./...

# EE build (from pranor-ee checkout, with pranor replace directives active)
CGO_ENABLED=0 go build -tags enterprise ./...

# Verify CGO invariant (from pranor root)
bash scripts/check_cgo.sh <module_name>
bash scripts/check_cgo.sh gate
bash scripts/check_cgo.sh graph
```

---

## Error Handling for EE Feature Detection

Callers of EE-gated functions MUST handle `ErrEERequired` gracefully:

```go
result, err := featurepkg.SomeEECapability(ctx, input)
if err != nil {
    if errors.Is(err, featurepkg.ErrEERequired) {
        // Degrade gracefully — log warning, use OSS fallback
        log.Warn("feature requires Enterprise Edition", "feature", "SomeEECapability")
        return ossFallbackResult, nil
    }
    return nil, fmt.Errorf("unexpected error: %w", err)
}
```
