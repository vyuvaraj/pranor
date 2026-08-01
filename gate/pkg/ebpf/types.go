package ebpf int

const (
	XDPPass XDPDDoSAction = iota
	XDPDrop
	XDPAbort
)
