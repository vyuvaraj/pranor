package import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestANNQueryEngine_QueryANNWithFilter(t *testing.T) {
	aqe := NewANNQueryEngine()

	// Insert vectors
	vec1 := []float32{1.0, 0.0, 0.0}
	vec2 := []float32{0.9, 0.1, 0.0}
	vec3 := []float32{0.0, 1.0, 0.0}

	aqe.InsertVector("kb-bucket", "doc1", vec1, map[string]interface{}{"category": "tech"})
	aqe.InsertVector("kb-bucket", "doc2", vec2, map[string]interface{}{"category": "tech"})
	aqe.InsertVector("kb-bucket", "doc3", vec3, map[string]interface{}{"category": "finance"})

	req := ANNQueryRequest{
		BucketName: "kb-bucket",
		Vector:     []float32{1.0, 0.0, 0.0},
		TopK:       5,
		MinScore:   0.5,
		Filter:     map[string]interface{}{"category": "tech"},
	}

	matches, err := aqe.QueryANN(req)
	if err != nil {
		t.Fatalf("QueryANN failed: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matching tech documents, got %d", len(matches))
	}

	if matches[0].ObjectID != "doc1" || matches[0].Score < 0.99 {
		t.Errorf("unexpected top match: %+v", matches[0])
	}

	// HTTP Handler test
	body, _ := json.Marshal(req)
	httpreq := httptest.NewRequest(http.MethodPost, "/api/v1/store/vector/query", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	aqe.HTTPHandler().ServeHTTP(w, httpreq)
	if w.Code != http.StatusOK {
		t.Fatalf(