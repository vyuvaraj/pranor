package import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mcpRPC(t *testing.T, srv http.Handler, method string, params interface{}) map[string]interface{} {
	t.Helper()
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      1,
	}
	if params != nil {
		body["params"] = params
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

func TestMCPInitialize(t *testing.T) {
	s := NewMCPServer("test-Pranor", "1.0.0")
	resp := mcpRPC(t, s.Handler(), "initialize", nil)

	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	si, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected serverInfo in result, got: %v", result)
	}
	if si["name"] != "test-Pranor" {
		t.Errorf("expected server name 'test-Pranor', got %v", si["name"])
	}
	if _, ok := result["capabilities"]; !ok {
		t.Error("expected capabilities in initialize response")
	}
}

func TestMCPToolsList(t *testing.T) {
	s := NewMCPServer("test-Pranor", "1.0.0")
	s.RegisterTool(MCPTool{
		Name:        "tool_alpha",
		Description: "Alpha tool",
		InputSchema: map[string]interface{}{"type": "object"},
	}, func(_ context.Context, _ map[string]interface{}) (MCPToolResult, error) {
		return MCPToolResult{}, nil
	})
	s.RegisterTool(MCPTool{
		Name:        "tool_beta",
		Description: "Beta tool",
		InputSchema: map[string]interface{}{"type": "object"},
	}, func(_ context.Context, _ map[string]interface{}) (MCPToolResult, error) {
		return MCPToolResult{}, nil
	})

	resp := mcpRPC(t, s.Handler(), "tools/list", nil)
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object: %v", resp)
	}
	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatalf("expected tools array: %v", result)
	}
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

func TestMCPToolsCall(t *testing.T) {
	s := NewMCPServer("test-Pranor", "1.0.0")
	s.RegisterTool(MCPTool{
		Name:        "greet",
		Description: "Returns a greeting",
		InputSchema: map[string]interface{}{"type": "object"},
	}, func(_ context.Context, params map[string]interface{}) (MCPToolResult, error) {
		name, _ := params["name"].(string)
		if name == "" {
			name = "world"
		}
		return MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "hello, " + name}},
		}, nil
	})

	resp := mcpRPC(t, s.Handler(), "tools/call", map[string]interface{}{
		"name":      "greet",
		"arguments": map[string]interface{}{"name": "Pranor"},
	})
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result: %v", resp)
	}
	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Fatalf("expected content array: %v", result)
	}
	item := content[0].(map[string]interface{})
	if !strings.Contains(item["text"].(string), "Pranor") {
		t.Errorf("expected greeting to contain 'Pranor', got: %v", item["text"])
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	s := NewMCPServer("test-Pranor", "1.0.0")
	resp := mcpRPC(t, s.Handler(), "nonexistent/method", nil)
	if resp["error"] == nil {
		t.Errorf("expected error for unknown method, got: %v", resp)
	}
}

func TestMCPUnknownTool(t *testing.T) {
	s := NewMCPServer("test-Pranor", "1.0.0")
	resp := mcpRPC(t, s.Handler(), "tools/call", map[string]interface{}{
		"name":      "does_not_exist",
		"arguments": map[string]interface{}{},
	})
	if resp["error"] == nil {
		t.Errorf("expected error for unknown tool, got: %v", resp)
	}
}
