# Pranor Learn — ML Inference Provider

**Version:** 2.0.0-dev  
**Module Path:** `github.com/vyuvaraj/pranor/learn/api`  
**License:** AGPL-3.0 (OSS) / EE

---

## Overview

Pranor Learn acts as a pluggable ML inference provider for the Decision Engine, powering the Level 5 Learn veto level.

The module is divided into three sub-modules:
- `learn/api`: Shared interfaces and contracts.
- `learn/wasm`: Wazero runner for WASM inference.
- `learn/sidecar`: gRPC IPC sidecar for external models.

---

## Zero-CGO Constraint

All Pranor Learn code enforces `CGO_ENABLED=0`. Complex ML inference is offloaded to the gRPC sidecar.

---

## API Reference

### Predictor Interface

```go
type Predictor interface {
    Predict(ctx context.Context, in PredictInput) (PredictOutput, error)
    HealthCheck(ctx context.Context) error
}
```

### Types

**PredictInput**
Contains the features and context for inference, as well as `BudgetMs` for timeouts.

**PredictOutput**
Contains the prediction result, confidence scores, and advisory actions.

---

## Fault Contracts

- Returns `ErrSidecarTimeout` if the gRPC sidecar exceeds `BudgetMs`.
- Returns `ErrModelBudgetExceeded` for inference compute overruns.

---

## Enterprise Edition

| Feature | OSS | EE |
|---------|:---:|:--:|
| WASM runner | ✓ | ✓ |
| Stubs for sidecar | ✓ | ✓ (Returns `ErrEERequired` in OSS) |
| GPU PyTorch/TabPFN pool | — | ✓ |
