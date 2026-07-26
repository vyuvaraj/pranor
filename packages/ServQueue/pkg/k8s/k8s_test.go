package k8s

import (
	"context"
	"testing"
)

func TestK8sOperatorAndKEDAScaler(t *testing.T) {
	op := NewOperator()
	cluster := &ServQueueCluster{
		Name:      "prod-queue",
		Namespace: "default",
		Spec: ServQueueClusterSpec{
			Replicas: 3,
		},
	}

	status, err := op.Reconcile(context.Background(), cluster)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
	if status.Phase != "Running" || status.ReadyReplicas != 3 {
		t.Errorf("Unexpected status: %+v", status)
	}

	scaler := NewKEDAScaler("orders", 50)
	active, _ := scaler.IsActive(context.Background(), 10)
	if !active {
		t.Errorf("Expected scaler to be active for lag 10")
	}

	metrics, err := scaler.GetMetrics(context.Background(), 10)
	if err != nil || metrics.MetricValue != 10 {
		t.Errorf("Unexpected metric value: %v", metrics)
	}
}
