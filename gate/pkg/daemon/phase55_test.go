package daemon

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/gate/pkg/graphql"
	"github.com/vyuvaraj/pranor/gate/pkg/registry"
	"github.com/vyuvaraj/pranor/gate/pkg/waf"
)

func TestPhase55_WASMHotReloadRegistry(t *testing.T) {
	reg := registry.NewWASMHotReloadRegistry()
	err := reg.RegisterPlugin("auth-filter", "https://cdn.example.com/auth.wasm", "v1.2.0", []byte("wasm-binary"))
	if err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	plugin, exists := reg.GetPlugin("auth-filter")
	if !exists || plugin.Version != "v1.2.0" {
		t.Errorf("plugin lookup failed: got %v", plugin)
	}
}

func TestPhase55_GraphQLFederationEngine(t *testing.T) {
	fed := graphql.NewGraphQLFederationEngine()
	_ = fed.RegisterUpstream("users-svc", "http://users:8080/graphql")
	_ = fed.RegisterUpstream("orders-svc", "http://orders:8080/graphql")

	res, err := fed.ExecuteFederatedQuery(context.Background(), "{ users { id } }")
	if err != nil || res == "" {
		t.Errorf("federated query failed: %v", err)
	}
}

func TestPhase55_InlineWAFEngine(t *testing.T) {
	wafEngine := waf.NewInlineWAFEngine()

	// Safe request payload
	safe, _, err := wafEngine.InspectPayload(context.Background(), "GET /api/v1/users?id=10")
	if !safe || err != nil {
		t.Errorf("safe request failed WAF check")
	}

	// SQL injection payload
	sqliSafe, ruleID, err := wafEngine.InspectPayload(context.Background(), "SELECT * FROM users UNION SELECT null, null")
	if sqliSafe || ruleID != "sqli-01" {
		t.Errorf("expected SQLi block by rule sqli-01")
	}
}
