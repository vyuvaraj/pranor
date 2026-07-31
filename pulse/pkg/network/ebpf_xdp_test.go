package network

import (
	"testing"
)

func TestEBPFXDPFastPacketIngest(t *testing.T) {
	xdp := NeweBPFXDPAccelerator("eth0", XDPModeNative)

	if err := xdp.AttachProgram(); err != nil {
		t.Fatalf("AttachProgram failed: %v", err)
	}

	packetData := []byte("FAST_STOMP_FRAME_HEADER_PAYLOAD")
	latency, err := xdp.IngestFastPacket(packetData)
	if err != nil {
		t.Fatalf("IngestFastPacket failed: %v", err)
	}

	if latency > 100 {
		t.Errorf("Expected low latency packet ingest (<100µs), got %dµs", latency)
	}

	packets, bytes := xdp.GetStats()
	if packets != 1 || bytes != uint64(len(packetData)) {
		t.Errorf("Unexpected stats: packets %d, bytes %d", packets, bytes)
	}
}
