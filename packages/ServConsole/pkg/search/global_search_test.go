package search

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnifiedGlobalSearchEngine_Search(t *testing.T) {
	engine := NewUnifiedGlobalSearchEngine()

	engine.IndexResource(SearchableResource{
		ID:          "svc-order-api",
		Type:        "service",
		Title:       "Order Processing Service",
		Description: "Handles customer checkout and payment routing",
		URL:         "/services/order-api",
	})
	engine.IndexResource(SearchableResource{
		ID:          "bucket-assets",
		Type:        "bucket",
		Title:       "CDN Media Assets Bucket",
		Description: "ServStore bucket storing static image assets",
		URL:         "/storage/buckets/assets",
	})

	results := engine.Search("checkout", 10)
	if len(results) != 1 || results[0].ID != "svc-order-api" {
		t.Fatalf("expected order-api match, got %+v", results)
	}

	// HTTP Handler test
	req := httptest.NewRequest(http.MethodGet, "/api/v1/console/search?q=CDN", nil)
	w := httptest.NewRecorder()

	engine.HTTPHandler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp["count"].(float64) != 1 {
		t.Errorf("expected 1 result in search JSON response, got %v", resp["count"])
	}
}
