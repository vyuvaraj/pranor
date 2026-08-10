// Package mcp implements a native Model Context Protocol (MCP) server for PranorGateway.
// MCP is a JSON-RPC 2.0 protocol that allows AI agents (Claude, GPT-4o, Ollama, etc.)
// to discover and call strongly-typed tools. This server auto-exposes Pranor services
// (Pranor Vault buckets, Pranor Pulse topics) as MCP tools consumable by any MCP-compatible agent.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/gate/pkg/agentgov"
)

// ---- JSON-RPC 2.0 wire types ----

// MCPRequest is an incoming JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// MCPResponse is an outgoing JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// MCPError is the JSON-RPC error object.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standard JSON-RPC error codes.
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternalError  = -32603
)

// ---- MCP Protocol types ----

// MCPContent is a single piece of content returned by a tool call.
type MCPContent struct {
	Type string `json:"type"` // "text" or "json"
	Text string `json:"text"`
}

// MCPToolResult is the result of a tool call.
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPTool describes a callable tool in the MCP registry.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolHandler is the function signature for tool implementations.
type ToolHandler func(ctx context.Context, params map[string]interface{}) (MCPToolResult, error)

// ---- MCP Server ----

type ToolCallAuditEntry struct {
	Timestamp   string                 `json:"timestamp"`
	AgentID     string                 `json:"agent_id"`
	ToolName    string                 `json:"tool_name"`
	Arguments   map[string]interface{} `json:"arguments"`
	IsError     bool                   `json:"is_error"`
	CostUSD     float64                `json:"cost_usd"`
	ExecutionMs int64                  `json:"execution_ms"`
}

// MCPServer is a Model Context Protocol server that exposes Pranor services as AI tools.
// It implements the MCP specification over HTTP, supporting both request/response and
// SSE (Server-Sent Events) transports.
type MCPServer struct {
	mu            sync.RWMutex
	tools         map[string]MCPTool
	handlers      map[string]ToolHandler
	serverName    string
	serverVersion string
	auditLogs     []ToolCallAuditEntry
}

// NewMCPServer creates a new MCPServer with the given name and version.
func NewMCPServer(name, version string) *MCPServer {
	return &MCPServer{
		tools:         make(map[string]MCPTool),
		handlers:      make(map[string]ToolHandler),
		serverName:    name,
		serverVersion: version,
		auditLogs:     make([]ToolCallAuditEntry, 0),
	}
}

// RegisterTool registers a tool and its handler. Thread-safe.
func (s *MCPServer) RegisterTool(tool MCPTool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
}

// toolList returns all registered tools as a slice (must be called with lock held or copy).
func (s *MCPServer) toolList() []MCPTool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]MCPTool, 0, len(s.tools))
	for _, t := range s.tools {
		list = append(list, t)
	}
	return list
}

// Handler returns an http.Handler serving the MCP protocol.
// Supports:
//   - POST /  : JSON-RPC 2.0 request/response
//   - GET  /  : SSE stream with periodic ping events
func (s *MCPServer) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.handleJSONRPC(w, r)
		case http.MethodGet:
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "text/event-stream") {
				s.handleSSE(w, r)
			} else {
				// Return server info for plain GET
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"server": s.serverName,
					"version": s.serverVersion,
					"protocol": "mcp",
					"tools": len(s.tools),
				})
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

// handleJSONRPC processes a single JSON-RPC 2.0 request.
func (s *MCPServer) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(io.LimitReader(r.Body, 4*1024*1024))
	if err != nil {
		s.writeError(w, nil, ErrParseError, "failed to read request body")
		return
	}

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, nil, ErrParseError, "invalid JSON: "+err.Error())
		return
	}

	if req.JSONRPC != "2.0" {
		s.writeError(w, req.ID, ErrInvalidRequest, "jsonrpc must be \"2.0\"")
		return
	}

	ctx := r.Context()
	var result interface{}

	switch req.Method {
	case "initialize":
		result = s.handleInitialize()

	case "tools/list":
		result = s.handleToolsList()

	case "tools/call":
		res, mcpErr := s.handleToolsCall(ctx, req.Params)
		if mcpErr != nil {
			s.writeError(w, req.ID, mcpErr.Code, mcpErr.Message)
			return
		}
		result = res

	default:
		s.writeError(w, req.ID, ErrMethodNotFound, fmt.Sprintf("unknown method: %s", req.Method))
		return
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
	json.NewEncoder(w).Encode(resp)
}

// handleInitialize returns the MCP server capabilities response.
func (s *MCPServer) handleInitialize() interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]interface{}{
			"name":    s.serverName,
			"version": s.serverVersion,
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
	}
}

// handleToolsList returns all registered tools.
func (s *MCPServer) handleToolsList() interface{} {
	return map[string]interface{}{
		"tools": s.toolList(),
	}
}

// handleToolsCall dispatches a tool call to the registered handler with Security Firewall governance.
func (s *MCPServer) handleToolsCall(ctx context.Context, raw json.RawMessage) (*MCPToolResult, *MCPError) {
	var callParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &callParams); err != nil {
		return nil, &MCPError{Code: ErrInvalidParams, Message: "invalid tools/call params: " + err.Error()}
	}

	// 1. Construct Agent Security Chain & Evaluate Firewall Policy
	chain := agentgov.AgentSecurityChain{
		AgentID:      "mcp-agent-session",
		UserID:       "user-mcp",
		TenantID:     "tenant-default",
		CapabilityID: callParams.Name,
	}
	toolCall := agentgov.ToolCallPayload{
		ToolName:  callParams.Name,
		Arguments: callParams.Arguments,
	}

	fw := agentgov.GetFirewall()
	eval := fw.EvaluateToolCall(chain, toolCall)

	if eval.Decision == agentgov.DecisionDeny {
		return nil, &MCPError{Code: -32001, Message: "Tool execution DENIED by Agent Security Firewall: " + eval.Reason}
	} else if eval.Decision == agentgov.DecisionApprove {
		return nil, &MCPError{Code: -32002, Message: fmt.Sprintf("Tool execution paused for Human-In-The-Loop APPROVAL (Approval ID: %s): %s", eval.ApprovalID, eval.Reason)}
	} else if eval.Decision == agentgov.DecisionTransform {
		callParams.Arguments = eval.TransformedArguments
	}

	s.mu.RLock()
	handler, ok := s.handlers[callParams.Name]
	s.mu.RUnlock()

	if !ok {
		return nil, &MCPError{Code: ErrInvalidParams, Message: fmt.Sprintf("unknown tool: %s", callParams.Name)}
	}

	start := time.Now()
	result, err := handler(ctx, callParams.Arguments)
	executionMs := time.Since(start).Milliseconds()

	isError := err != nil || result.IsError
	fw.RecordTrajectoryStep(agentgov.TrajectoryStep{
		Timestamp:   start,
		Chain:       chain,
		ToolCall:    toolCall,
		ResultDecision: eval,
		ExecutionMs: executionMs,
	})
	auditEntry := ToolCallAuditEntry{
		Timestamp:   start.Format(time.RFC3339),
		AgentID:     chain.AgentID,
		ToolName:    callParams.Name,
		Arguments:   callParams.Arguments,
		IsError:     isError,
		CostUSD:     0.0005, // Estimated tool call token attribution
		ExecutionMs: executionMs,
	}

	s.mu.Lock()
	s.auditLogs = append(s.auditLogs, auditEntry)
	s.mu.Unlock()

	if err != nil {
		errResult := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "tool error: " + err.Error()}},
			IsError: true,
		}
		return &errResult, nil
	}

	return &result, nil
}

// GetAuditLogs returns the recorded MCP tool call audit logs (SG.A4)
func (s *MCPServer) GetAuditLogs() []ToolCallAuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	logs := make([]ToolCallAuditEntry, len(s.auditLogs))
	copy(logs, s.auditLogs)
	return logs
}

// handleSSE serves an SSE stream, sending a ping every 30 seconds.
func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"server\":%q}\n\n", s.serverName)
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case t := <-ticker.C:
			fmt.Fprintf(w, "event: ping\ndata: {\"ts\":%d}\n\n", t.UnixMilli())
			flusher.Flush()
		}
	}
}

// writeError writes a JSON-RPC error response.
func (s *MCPServer) writeError(w http.ResponseWriter, id interface{}, code int, msg string) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		Error:   &MCPError{Code: code, Message: msg},
		ID:      id,
	}
	json.NewEncoder(w).Encode(resp)
}

// ---- Built-in Pranor tool registrations ----

// RegisterPranorVaultTools pre-registers built-in Pranor Vault tools on the given MCPServer.
// storeBaseURL is the Pranor Vault HTTP endpoint, e.g. "http://localhost:8085".
func RegisterPranorVaultTools(s *MCPServer, storeBaseURL string) {
	storeBaseURL = strings.TrimRight(storeBaseURL, "/")
	client := &http.Client{Timeout: 10 * time.Second}

	s.RegisterTool(MCPTool{
		Name:        "pranorVault_list_buckets",
		Description: "List all buckets in Pranor Vault",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}, func(ctx context.Context, _ map[string]interface{}) (MCPToolResult, error) {
		resp, err := client.Get(storeBaseURL + "/api/v1/buckets")
		if err != nil {
			return MCPToolResult{}, fmt.Errorf("Pranor Vault unavailable: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return MCPToolResult{Content: []MCPContent{{Type: "json", Text: string(body)}}}, nil
	})

	s.RegisterTool(MCPTool{
		Name:        "pranorVault_get_object",
		Description: "Get an object from a Pranor Vault bucket by key",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"required": []string{"bucket", "key"},
			"properties": map[string]interface{}{
				"bucket": map[string]interface{}{"type": "string", "description": "Bucket name"},
				"key":    map[string]interface{}{"type": "string", "description": "Object key"},
			},
		},
	}, func(ctx context.Context, params map[string]interface{}) (MCPToolResult, error) {
		bucket, _ := params["bucket"].(string)
		key, _ := params["key"].(string)
		if bucket == "" || key == "" {
			return MCPToolResult{}, fmt.Errorf("bucket and key are required")
		}
		resp, err := client.Get(fmt.Sprintf("%s/%s/%s", storeBaseURL, bucket, key))
		if err != nil {
			return MCPToolResult{}, fmt.Errorf("Pranor Vault get failed: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return MCPToolResult{Content: []MCPContent{{Type: "text", Text: string(body)}}}, nil
	})

	s.RegisterTool(MCPTool{
		Name:        "pranorVault_put_object",
		Description: "Upload an object to a Pranor Vault bucket",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"required": []string{"bucket", "key", "content"},
			"properties": map[string]interface{}{
				"bucket":  map[string]interface{}{"type": "string", "description": "Bucket name"},
				"key":     map[string]interface{}{"type": "string", "description": "Object key"},
				"content": map[string]interface{}{"type": "string", "description": "Object content"},
			},
		},
	}, func(ctx context.Context, params map[string]interface{}) (MCPToolResult, error) {
		bucket, _ := params["bucket"].(string)
		key, _ := params["key"].(string)
		content, _ := params["content"].(string)
		if bucket == "" || key == "" {
			return MCPToolResult{}, fmt.Errorf("bucket and key are required")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPut,
			fmt.Sprintf("%s/%s/%s", storeBaseURL, bucket, key),
			strings.NewReader(content),
		)
		if err != nil {
			return MCPToolResult{}, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return MCPToolResult{}, fmt.Errorf("Pranor Vault put failed: %w", err)
		}
		defer resp.Body.Close()
		return MCPToolResult{Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("uploaded %s/%s (status %d)", bucket, key, resp.StatusCode)}}}, nil
	})
}

// RegisterPranorPulseTools pre-registers built-in Pranor Pulse tools on the given MCPServer.
// queueBaseURL is the Pranor Pulse HTTP endpoint, e.g. "http://localhost:8086".
func RegisterPranorPulseTools(s *MCPServer, queueBaseURL string) {
	queueBaseURL = strings.TrimRight(queueBaseURL, "/")
	client := &http.Client{Timeout: 10 * time.Second}

	s.RegisterTool(MCPTool{
		Name:        "pranorPulse_topics",
		Description: "List all topics available in Pranor Pulse",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}, func(ctx context.Context, _ map[string]interface{}) (MCPToolResult, error) {
		resp, err := client.Get(queueBaseURL + "/topics")
		if err != nil {
			return MCPToolResult{}, fmt.Errorf("Pranor Pulse unavailable: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return MCPToolResult{Content: []MCPContent{{Type: "json", Text: string(body)}}}, nil
	})

	s.RegisterTool(MCPTool{
		Name:        "pranorPulse_publish",
		Description: "Publish a message to a Pranor Pulse topic",
		InputSchema: map[string]interface{}{
			"type":     "object",
			"required": []string{"topic", "payload"},
			"properties": map[string]interface{}{
				"topic":   map[string]interface{}{"type": "string", "description": "Topic name"},
				"payload": map[string]interface{}{"type": "string", "description": "Message payload (JSON or plain text)"},
			},
		},
	}, func(ctx context.Context, params map[string]interface{}) (MCPToolResult, error) {
		topic, _ := params["topic"].(string)
		payload, _ := params["payload"].(string)
		if topic == "" {
			return MCPToolResult{}, fmt.Errorf("topic is required")
		}
		body, _ := json.Marshal(map[string]string{"topic": topic, "payload": payload})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, queueBaseURL+"/publish", strings.NewReader(string(body)))
		if err != nil {
			return MCPToolResult{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return MCPToolResult{}, fmt.Errorf("Pranor Pulse publish failed: %w", err)
		}
		defer resp.Body.Close()
		return MCPToolResult{Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("published to %s (status %d)", topic, resp.StatusCode)}}}, nil
	})
}
