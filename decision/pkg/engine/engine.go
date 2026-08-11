package engine

import (
	graphapi "github.com/vyuvaraj/pranor/graph/api"
)

type VetoLadderEngine struct {
	graphProvider graphapi.GraphProvider
}

func NewVetoLadderEngine(gp graphapi.GraphProvider) *VetoLadderEngine {
	return &VetoLadderEngine{
		graphProvider: gp,
	}
}
