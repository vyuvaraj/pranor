package import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTopologyGraph_UpdateAndHandler(t *testing.T) {
	tg := NewTopologyGraph()

	tg.UpdateEdge("frontend", "order-service", 25.0, false, "closed")
	tg.UpdateEdge("frontend", "order-service", 30.0, true, "closed")
	tg.UpdateEdge("order-service", "payment-service", 150.0, true, "open")

	edges := tg.GetEdges()
	if len(edges) != 2 {
		t.Fatalf("expected 2 active edges, got %d", len(edges))
	}

	handler := tg.HTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/topology", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp["count"].(float64) != 2 {
		t.Errorf("expected count 2 in topology JSON, got %v", resp["count"])
	}
}
