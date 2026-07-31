package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vyuvaraj/pranor/gate/pkg/ai"
	"github.com/vyuvaraj/pranor/gate/pkg/autotls"
	"github.com/vyuvaraj/pranor/gate/pkg/edge"
	"github.com/vyuvaraj/pranor/gate/pkg/transcode"
	"github.com/vyuvaraj/pranor/gate/pkg/wasm"
)

func TestPhase46_StandaloneDaemonAndCLI(t *testing.T) {
	d, err := NewPranorGatewayDaemon("")
	if err != nil {
		t.Fatalf("failed to create daemon: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	d.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestPhase46_InlineWASMEdgeMiddleware(t *testing.T) {
	engine := wasm.NewWASMEdgeEngine()
	engine.RegisterFilter("test-auth", []byte("wasm-bytecode"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	action, err := engine.ProcessRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected process error: %v", err)
	}
	if action != wasm.ActionAllow {
		t.Errorf("expected action allow, got %v", action)
	}
	if req.Header.Get("X-WASM-Edge-Processed") != "true" {
		t.Errorf("expected WASM header tag")
	}
}

func TestPhase46_EdgeAILLMProxy(t *testing.T) {
	proxy := ai.NewLLMEdgeProxy(ai.LLMProxyConfig{
		Provider:    ai.ProviderOllama,
		TargetURL:   "http://localhost:11434",
		MaxRPM:      100,
		EnableCache: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()

	proxy.ServeHTTP(w, req)
	if w.Code == 0 {
		t.Errorf("expected valid HTTP response code")
	}
}

func TestPhase46_AutoTLSManager(t *testing.T) {
	mgr := autotls.NewAutoTLSManager(autotls.AutoTLSConfig{
		IsDev: true,
	})

	tlsCfg, err := mgr.GetTLSConfig()
	if err != nil {
		t.Fatalf("failed to get self-signed TLS config: %v", err)
	}
	if len(tlsCfg.Certificates) == 0 {
		t.Errorf("expected self-signed certificate")
	}
}

func TestPhase46_GRPCTranscoderAndGraphQL(t *testing.T) {
	transcoder := transcode.NewGRPCTranscoder("localhost:50051")
	req := httptest.NewRequest(http.MethodPost, "/api/grpc", nil)
	w := httptest.NewRecorder()

	transcoder.TranscodeHTTPToGRPC(w, req)
	if w.Header().Get("Content-Type") != "application/grpc" {
		t.Errorf("expected application/grpc Content-Type, got %s", w.Header().Get("Content-Type"))
	}

	agg := transcode.NewGraphQLAggregator([]string{"http://service-a/graphql", "http://service-b/graphql"})
	res, err := agg.AggregateSchema(context.Background(), "{ user { id } }")
	if err != nil || res == "" {
		t.Errorf("graphql aggregation failed: %v", err)
	}
}

func TestPhase46_BrowserEdgeSDK(t *testing.T) {
	sdk := edge.NewBrowserGatewaySDK()
	sdk.RegisterClientRoute("/app", "/offline.html", true)

	data, status, err := sdk.InterceptFetch(context.Background(), "/app")
	if err != nil || status != http.StatusOK {
		t.Fatalf("browser edge intercept failed: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("empty browser intercept payload")
	}
}
