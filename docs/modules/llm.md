# LLM Router (`std/llm`)

**Module Path:** `github.com/vyuvaraj/pranor/llm`  
**Introduced:** Phase 91 (Sprint V2.91.3)

---

## Overview

Pranor LLM (`std/llm`) provides a provider-agnostic model routing abstraction with fallback chains, semantic caching hooks, cost tracking, and CGO-free execution.

---

## Key Interfaces

### ChatProvider
Every LLM driver implements `ChatProvider`:

```go
type ChatProvider interface {
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
    Name() string
    Models() []string
    HealthCheck(ctx context.Context) error
}
```

### Router
```go
type Router interface {
    Route(ctx context.Context, req ChatRequest) (ChatResponse, error)
    Register(p ChatProvider)
    SetFallbackChain(providerNames []string)
    HealthCheck(ctx context.Context) map[string]error
}
```

---

## Data Structures

```go
type Message struct {
    Role    string // RoleSystem, RoleUser, RoleAssistant, RoleTool
    Content string
    Name    string
}

type ChatRequest struct {
    Messages    []Message
    Model       string   // e.g. "gpt-4o", "claude-3-5-sonnet"
    MaxTokens   int
    Temperature float64
    Stream      bool
    BudgetMs    int64    // Latency budget in ms
}

type ChatResponse struct {
    Content      string
    FinishReason FinishReason // FinishStop, FinishLength, FinishToolCall, FinishFiltered
    InputTokens  int
    OutputTokens int
    TotalTokens  int
    CostUSD      float64
    LatencyMs    int64
    Provider     string
    Model        string
}
```

---

## Drivers & OSS vs. EE Split

| Provider | Type | Description |
|----------|------|-------------|
| `EchoProvider` | OSS | Test stub echoing the last input message |
| `HTTPProvider` | OSS | Generic OpenAI-compatible REST API driver |
| `OpenAI` | EE | Full OpenAI API driver with streaming & function calling via gRPC sidecar |
| `Anthropic` | EE | Claude 3.5 Sonnet/Haiku driver via gRPC sidecar |
| `Gemini` | EE | Google Gemini 1.5 Pro/Flash driver via gRPC sidecar |
| `Ollama` | EE | Local vLLM/Ollama driver via IPC socket |

---

## Code Example

```go
import (
    "context"
    "github.com/vyuvaraj/pranor/llm"
    "github.com/vyuvaraj/pranor/llm/api"
)

resp, err := llm.Route(ctx, api.ChatRequest{
    Model: "gpt-4o",
    Messages: []api.Message{
        {Role: api.RoleUser, Content: "Hello Pranor!"},
    },
})
```
