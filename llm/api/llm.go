package api

import (
    "context"
    "errors"
    "time"
)

// Role constants for chat messages
const (
    RoleSystem    = "system"
    RoleUser      = "user"
    RoleAssistant = "assistant"
    RoleTool      = "tool"
)

// FinishReason describes why the model stopped generating.
type FinishReason string

const (
    FinishStop      FinishReason = "stop"
    FinishLength    FinishReason = "length"
    FinishToolCall  FinishReason = "tool_call"
    FinishFiltered  FinishReason = "content_filter"
)

// Message is a single turn in a chat conversation.
type Message struct {
    Role    string // RoleSystem, RoleUser, RoleAssistant, RoleTool
    Content string
    Name    string // optional: tool name for RoleTool messages
}

// ChatRequest is the input to a chat completion.
type ChatRequest struct {
    Messages    []Message
    Model       string        // e.g. "gpt-4o", "claude-3-5-sonnet", "gemini-1.5-pro"
    MaxTokens   int           // 0 = provider default
    Temperature float64       // 0.0–2.0; 0 = provider default
    Stream      bool          // true = streaming response (EE)
    BudgetMs    int64         // max latency budget in ms; 0 = no limit
    TopP        float64       // 0 = provider default
    StopWords   []string      // additional stop sequences
}

// ChatResponse is the output from a chat completion.
type ChatResponse struct {
    Content      string
    FinishReason FinishReason
    InputTokens  int
    OutputTokens int
    TotalTokens  int
    CostUSD      float64   // estimated cost in USD; 0 if unknown
    LatencyMs    int64
    Provider     string    // provider name
    Model        string    // model actually used (may differ from request if fallback occurred)
    RequestID    string    // provider-assigned request ID for tracing
    CreatedAt    time.Time
}

// ChatProvider is the interface every LLM provider driver must implement.
type ChatProvider interface {
    // Chat performs a synchronous chat completion.
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
    // Name returns the provider name (e.g. "openai", "anthropic", "echo").
    Name() string
    // Models returns the list of model IDs this provider supports.
    Models() []string
    // HealthCheck verifies the provider is reachable.
    HealthCheck(ctx context.Context) error
}

// Router routes chat requests across registered providers with fallback.
type Router interface {
    // Route selects a provider based on the request and routing strategy, with fallback.
    Route(ctx context.Context, req ChatRequest) (ChatResponse, error)
    // Register adds a ChatProvider to the router.
    Register(p ChatProvider)
    // SetFallbackChain specifies provider priority order by name.
    SetFallbackChain(providerNames []string)
    // HealthCheck checks all registered providers.
    HealthCheck(ctx context.Context) map[string]error
}

// CacheProvider is an optional semantic cache hook for LLM responses.
type CacheProvider interface {
    Get(ctx context.Context, key string) (ChatResponse, bool)
    Set(ctx context.Context, key string, resp ChatResponse, ttl int) error
}

// Sentinel errors
var (
    ErrAllProvidersFailed = errors.New("pranor/llm: all providers in fallback chain failed")
    ErrBudgetExceeded     = errors.New("pranor/llm: token or cost budget exceeded")
    ErrModelUnavailable   = errors.New("pranor/llm: requested model unavailable on any provider")
    ErrNoProviders        = errors.New("pranor/llm: no providers registered")
    ErrEmptyMessages      = errors.New("pranor/llm: messages must not be empty")
    ErrEERequired         = errors.New("pranor/llm: this provider requires Pranor Enterprise Edition")
)
