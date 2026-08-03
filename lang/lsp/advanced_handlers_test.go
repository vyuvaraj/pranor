package lsp

import (
	"encoding/json"
	"testing"
)

func TestWorkspaceSymbol(t *testing.T) {
	server := NewServer()
	server.documents["file:///app/main.pnr"] = "fn calculateTotal() { return 100 }"
	server.analyzeAndPublishDiagnostics("file:///app/main.pnr", server.documents["file:///app/main.pnr"])

	msg := JSONRPCMessage{
		ID:     1,
		Params: json.RawMessage(`{"query":"calculate"}`),
	}

	server.handleWorkspaceSymbol(msg)
}

func TestCallHierarchy(t *testing.T) {
	server := NewServer()
	code := "fn processOrder() { print(\"ok\") }\nfn main() { processOrder() }"
	server.documents["file:///app/main.pnr"] = code
	server.analyzeAndPublishDiagnostics("file:///app/main.pnr", code)

	prepMsg := JSONRPCMessage{
		ID:     2,
		Params: json.RawMessage(`{"textDocument":{"uri":"file:///app/main.pnr"},"position":{"line":0,"character":5}}`),
	}
	server.handlePrepareCallHierarchy(prepMsg)

	incMsg := JSONRPCMessage{
		ID:     3,
		Params: json.RawMessage(`{"item":{"name":"processOrder","kind":12,"uri":"file:///app/main.pnr","range":{"start":{"line":0,"character":0},"end":{"line":0,"character":12}},"selectionRange":{"start":{"line":0,"character":0},"end":{"line":0,"character":12}}}}`),
	}
	server.handleCallHierarchyIncomingCalls(incMsg)
}

func TestDocumentHighlight(t *testing.T) {
	server := NewServer()
	code := "fn foo() { let x = 10; print(x); }"
	server.documents["file:///app/main.pnr"] = code

	msg := JSONRPCMessage{
		ID:     4,
		Params: json.RawMessage(`{"textDocument":{"uri":"file:///app/main.pnr"},"position":{"line":0,"character":15}}`),
	}
	server.handleDocumentHighlight(msg)
}
