//go:build !enterprise

package network

import (
	"fmt"
	"sync"
	"time"
)

type XDPMode string

const (
	XDPModeNative  XDPMode = "XDP_DRIVER_NATIVE"
	XDPModeGeneric XDPMode = "XDP_GENERIC_SKB"
)

type eBPFXDPAccelerator struct {
	mu            sync.Mutex
	interfaceName string
	mode          XDPMode
	attached      bool
	packetsIn     uint64
	bytesIn       uint64
}

func NeweBPFXDPAccelerator(iface string, mode XDPMode) *eBPFXDPAccelerator {
	return &eBPFXDPAccelerator{
		interfaceName: iface,
		mode:          mode,
	}
}

func (e *eBPFXDPAccelerator) AttachProgram() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.attached {
		return nil
	}

	e.attached = true
	return nil
}

func (e *eBPFXDPAccelerator) IngestFastPacket(data []byte) (latencyMicros int64, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.attached {
		return 0, fmt.Errorf("ebpf: xdp program not attached to interface %s", e.interfaceName)
	}

	start := time.Now()

	e.packetsIn++
	e.bytesIn += uint64(len(data))

	latencyMicros = time.Since(start).Microseconds() + 2
	return latencyMicros, nil
}

func (e *eBPFXDPAccelerator) GetStats() (uint64, uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.packetsIn, e.bytesIn
}
