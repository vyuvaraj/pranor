package ebpf

type XDPDDoSAction int

const (
	XDPPass XDPDDoSAction = iota
	XDPDrop
	XDPAbort
)
