package client

import (
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/mesh/pkg/registry"
)

func TestP2CLoadBalancer_SelectTargetAndLatency(t *testing.T) {
	p2c := NewP2CLoadBalancer("us-east-1", "us-east-1a")

	insts := []registry.Instance{
		{Address: "http://10.0.0.1:8080", Region: "us-east-1"},
		{Address: "http://10.0.0.2:8080", Region: "us-east-1"},
		{Address: "http://10.0.0.3:8080", Region: "us-west-2"},
	}

	// Record high latency for 10.0.0.1, low latency for 10.0.0.2
	p2c.RecordLatency("http://10.0.0.1:8080", 200*time.Millisecond)
	p2c.RecordLatency("http://10.0.0.2:8080", 10*time.Millisecond)

	// Over multiple runs, P2C should favor 10.0.0.2 (same zone & lower latency)
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected := p2c.SelectTarget(insts)
		counts[selected]++
	}

	if counts["http://10.0.0.3:8080"] > 0 {
		t.Errorf("expected zero requests to different region us-west-2, got %d", counts["http://10.0.0.3:8080"])
	}
	if counts["http://10.0.0.2:8080"] < counts["http://10.0.0.1:8080"] {
		t.Errorf("expected 10.0.0.2 (low latency) to receive more traffic than 10.0.0.1, got: %+v", counts)
	}
}
