package import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestChaosInjectorLatencyPartitionCorruption(t *testing.T) {
	chaos := NewChaosInjector()

	// 1. Latency injection
	chaos.AddRule("r-latency", &FaultRule{
		Mode:        ChaosModeLatency,
		Latency:     10 * time.Millisecond,
		Probability: 1.0,
	})

	start := time.Now()
	resPayload, err := chaos.Intercept(context.Background(), "telemetry", "normal_data")
	if err != nil || resPayload != "normal_data" {
		t.Fatalf("Latency intercept failed: %v", err)
	}
	if time.Since(start) < 10*time.Millisecond {
		t.Errorf("Expected at least 10ms latency delay")
	}

	chaos.DisableRule("r-latency")

	// 2. Network Partition injection
	chaos.AddRule("r-partition", &FaultRule{
		Mode:        ChaosModePartition,
		Probability: 1.0,
	})

	_, err = chaos.Intercept(context.Background(), "telemetry", "normal_data")
	if err == nil || !strings.Contains(err.Error(), "simulated network partition") {
		t.Errorf("Expected network partition error, got %v", err)
	}

	chaos.DisableRule("r-partition")

	// 3. Corruption injection
	chaos.AddRule("r-corrupt", &FaultRule{
		Mode:        ChaosModeCorruption,
		Probability: 1.0,
	})

	resCorrupted, err := chaos.Intercept(context.Background(), "telemetry", "normal_data")
	if err != nil || !strings.Contains(resCorrupted, "CORRUPTED_CHAOS") {
		t.Errorf("Expected corrupted payload, got %s", resCorrupted)
	}
}
