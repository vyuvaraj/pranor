//go:build !enterprise

package ebpf

import (
	"context"
	"net"
)

type XDPKernelFilter struct {
	InterfaceName string
	Enabled       bool
}

func NewXDPKernelFilter(iface string) *XDPKernelFilter {
	return &XDPKernelFilter{
		InterfaceName: iface,
		Enabled:       false,
	}
}

func (x *XDPKernelFilter) ProcessPacket(ip net.IP, port int) XDPDDoSAction {
	// OSS Fallback: standard user-space packet passing
	return XDPPass
}

func (x *XDPKernelFilter) BlockIP(ctx context.Context, ip string) error {
	return nil
}
