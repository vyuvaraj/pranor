package sidecar

type GRPCPredictor struct {
	Target string
}

func NewGRPCPredictor(target string) *GRPCPredictor {
	return &GRPCPredictor{Target: target}
}
