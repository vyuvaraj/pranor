package namespaces

import (
	"testing"
)

func TestVectorNamespaceManager_CreateAndGet(t *testing.T) {
	vnm := NewVectorNamespaceManager()

	cfg := VectorNamespaceConfig{
		BucketName: "ai-knowledge-base",
		Dimension:  1536,
		Metric:     MetricCosine,
	}

	created, err := vnm.CreateNamespace(cfg)
	if err != nil {
		t.Fatalf("CreateNamespace failed: %v", err)
	}

	if created.Dimension != 1536 || created.M != 16 {
		t.Errorf("unexpected namespace config defaults: %+v", created)
	}

	retrieved, found := vnm.GetNamespace("ai-knowledge-base")
	if !found || retrieved.Dimension != 1536 {
		t.Errorf("failed to retrieve namespace: %+v", retrieved)
	}
}
