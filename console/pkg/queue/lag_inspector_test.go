package import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConsumerLagInspector_RecordAndQuery(t *testing.T) {
	inspector := NewConsumerLagInspector()

	inspector.RecordLag("orders-stream", 0, "payment-workers", 1000, 950)

	lags := inspector.GetGroupLag()
	if len(lags) != 1 || lags[0].Lag != 50 {
		t.Fatalf("expected lag of 50, got %+v", lags)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/console/consumer-groups/lag", nil)
	w := httptest.NewRecorder()
	inspector.HTTPHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp["count"].(float64) != 1 {
		t.Errorf("unexpected JSON response count: %v", resp["count"])
	}
}
