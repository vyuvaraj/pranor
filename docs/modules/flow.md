# Pranor Flow

```bash
docker run -p 8089:8089 ghcr.io/vyuvaraj/pranor-flow:latest
```

`Pranor Flow` is a stateful, DAG-based workflow orchestrator and Saga compensation engine for the **Pranor** ecosystem. It supports durable execution with checkpointing, WASM step functions, sub-workflow composition, per-execution tracing, and a Dead Letter Workflow Queue with manual retry.

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Defining Workflows](#defining-workflows)
- [Saga Compensation](#saga-compensation)
- [WASM Step Functions](#wasm-step-functions)
- [Sub-workflow Composition](#sub-workflow-composition)
- [Getting Started](#getting-started)

---

## Key Features

### 🔀 DAG Orchestration
- **Multi-step DAG execution**: Runs execution graphs sorted topologically by dependency constraints — steps run in parallel when their dependencies are satisfied
- **Step output propagation**: Output from each step is passed as input to dependent steps
- **Fan-out / fan-in**: Parallelize independent branches, synchronize at join steps
- **Conditional branching**: Steps can be skipped based on upstream output conditions

### 💾 Durable Execution
- **Checkpoint persistence**: Workflow state serialized to `.state` files on disk after every step — executions survive engine restarts
- **Resume from checkpoint**: `POST /api/workflows/resume` restarts a workflow from its last successful checkpoint
- **Idempotent step execution**: Steps can be marked idempotent; on replay, Pranor Flow skips already-completed steps

### 🔄 Saga Compensation
- **Automatic rollback on failure**: When a step fails after earlier steps have succeeded, Pranor Flow triggers `CompensateAction` in reverse topological order
- **Per-step compensation actions**: Each step optionally declares a compensate endpoint — called when rolling back
- **Partial compensation**: Compensates only completed steps — not future/skipped steps

### 🧩 WASM Step Functions
- **Sandboxed WASM step execution**: Run any step logic as a WASI-compliant WebAssembly module — language-agnostic step implementations (Rust, C, Go)
- **I/O via stdin/stdout**: Step input passed as JSON on stdin; step output read from stdout
- **Timeout enforcement**: Per-step WASM execution timeout prevents runaway steps

### 🧱 Sub-workflow Composition
- **Nested workflow manager**: Compose complex workflows from smaller reusable sub-workflows
- **Sub-workflow as a step**: Any step can invoke another workflow definition by name — the parent pauses and waits for the child to complete
- **Recursive composition**: Sub-workflows can themselves contain sub-workflows

### 📊 Observability & Cost Tracking
- **Per-execution OTel span attribution**: Each workflow execution and each individual step gets its own OTel span, linked to a root trace
- **AI cost tracking**: Steps that call AI/LLM endpoints have token cost annotations added to their spans
- **Execution timeline**: Full execution log with step start times, durations, status, and outputs

### 📭 Dead Letter Workflow Queue
- **DLQ for failed workflows**: Workflows that exhaust retries are moved to the DLWQ with full failure context
- **Manual retry endpoint**: `POST /api/workflows/dlq/{id}/retry` re-queues a DLWQ workflow from the beginning or from last checkpoint
- **DLWQ browser**: Pranor Console shows failed workflows with their error details

---

## Architecture

```json
{
  "id": "order-checkout-flow",
  "name": "Order Checkout Pipeline",
  "tasks": [
    { "name": "reserve-inventory", "action": "http://inventory-svc/reserve" },
    { "name": "process-payment", "action": "http://payment-svc/charge", "depends_on": ["reserve-inventory"], "compensate_action": "http://payment-svc/refund" },
    { "name": "ship-order", "action": "http://shipping-svc/label", "depends_on": ["process-payment"] }
  ]
}
```

```
Define Workflow (POST /api/workflows/define)
  └── DAG Spec: steps, dependencies, compensations, WASM modules

Execute Workflow (POST /api/workflows/execute)
  │
  ▼
┌────────────────────────────────────────────────────┐
│                    Pranor Flow Engine                  │
│                                                    │
│  Topological Sort → Parallel Ready Steps           │
│       │                                            │
│  ┌────▼─────┐  ┌───────────┐  ┌─────────────────┐ │
│  │ HTTP Step│  │ WASM Step │  │ Sub-workflow    │ │
│  │ Executor │  │ Executor  │  │ Invoker         │ │
│  └────┬─────┘  └─────┬─────┘  └────────┬────────┘ │
│       └──────────────┼─────────────────┘           │
│                      │                             │
│  ┌───────────────────▼───────────────────────────┐ │
│  │  Checkpoint Store (.state files)              │ │
│  └───────────────────────────────────────────────┘ │
│                      │                             │
│  ┌───────────────────▼───────────────────────────┐ │
│  │  On failure: Saga Compensator (reverse order) │ │
│  └───────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────┘
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/workflows/define` | Define a new DAG workflow |
| `GET` | `/api/workflows` | List all workflow definitions |
| `POST` | `/api/workflows/execute` | Execute a workflow instance |
| `GET` | `/api/workflows/instances/{id}` | Get execution status and step logs |
| `POST` | `/api/workflows/resume` | Resume from checkpoint file |
| `GET` | `/api/workflows/dlq` | Browse Dead Letter Workflow Queue |
| `POST` | `/api/workflows/dlq/{id}/retry` | Retry a DLQ workflow |
| `/metrics` | `GET` | Prometheus metrics (workflow success rate, step durations, DLQ depth) |
| `/healthz` | `GET` | Liveness probe |

---

## Defining Workflows

```bash
curl -X POST http://pranor-flow:8089/api/workflows/define \
  -d '{
    "name": "order-fulfillment",
    "steps": [
      {
        "id": "reserve-inventory",
        "type": "http",
        "url": "http://inventory/reserve",
        "depends_on": [],
        "compensate_url": "http://inventory/release"
      },
      {
        "id": "charge-payment",
        "type": "http",
        "url": "http://payments/charge",
        "depends_on": ["reserve-inventory"],
        "compensate_url": "http://payments/refund"
      },
      {
        "id": "notify-customer",
        "type": "http",
        "url": "http://notifications/send",
        "depends_on": ["charge-payment"]
      }
    ]
  }'
```

Execute it:

```bash
curl -X POST http://pranor-flow:8089/api/workflows/execute \
  -d '{"workflow": "order-fulfillment", "input": {"order_id": "ord-123", "amount": 99.99}}'
# → { "instance_id": "wf-abc-001", "status": "running" }
```

---

## Saga Compensation

If `charge-payment` fails after `reserve-inventory` succeeded:

```
1. reserve-inventory → ✅ SUCCESS
2. charge-payment    → ❌ FAILURE
3. Pranor Flow triggers compensations in reverse:
   → POST http://inventory/release  (compensate reserve-inventory)
```

---

## WASM Step Functions

```bash
curl -X POST http://pranor-flow:8089/api/workflows/define \
  -d '{
    "name": "ml-pipeline",
    "steps": [
      {
        "id": "preprocess",
        "type": "wasm",
        "wasm_module": "preprocess.wasm",
        "timeout": "30s",
        "depends_on": []
      },
      {
        "id": "predict",
        "type": "wasm",
        "wasm_module": "model-inference.wasm",
        "depends_on": ["preprocess"]
      }
    ]
  }'
```

---

## Sub-workflow Composition

```bash
curl -X POST http://pranor-flow:8089/api/workflows/define \
  -d '{
    "name": "full-onboarding",
    "steps": [
      { "id": "create-account", "type": "http", "url": "http://accounts/create", "depends_on": [] },
      {
        "id": "setup-billing",
        "type": "sub-workflow",
        "workflow": "billing-setup",
        "depends_on": ["create-account"]
      },
      { "id": "send-welcome", "type": "http", "url": "http://mail/welcome", "depends_on": ["setup-billing"] }
    ]
  }'
```

---

## Getting Started

```bash
docker run -p 8089:8089 \
  -e PRANOR_FLOW_CHECKPOINT_DIR=/data/checkpoints \
  -e PRANOR_FLOW_OTEL_ENDPOINT=http://pranor-trace:4318 \
  -v flow-data:/data \
  ghcr.io/vyuvaraj/pranor-flow:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PRANOR_FLOW_PORT` | `8089` | HTTP listener port |
| `PRANOR_FLOW_CHECKPOINT_DIR` | `./checkpoints` | Directory for workflow state checkpoint files |
| `PRANOR_FLOW_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `PRANOR_FLOW_WASM_MODULES_DIR` | `./wasm` | Directory for WASM step module files |
| `PRANOR_FLOW_DLQ_MAX_SIZE` | `1000` | Max workflows retained in DLQ |
