//go:build !enterprise

package providers

import (
	"context"
	"time"

	"github.com/vyuvaraj/pranor/llm/api"
)

type EchoProvider struct{}

func NewEchoProvider() *EchoProvider {
	return &EchoProvider{}
}

func (e *EchoProvider) Name() string {
	return "echo"
}

func (e *EchoProvider) Models() []string {
	return []string{"echo-1"}
}

func (e *EchoProvider) Chat(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
	start := time.Now()
	var content string
	if len(req.Messages) > 0 {
		content = req.Messages[len(req.Messages)-1].Content
	}
	return api.ChatResponse{
		Content:      content,
		FinishReason: api.FinishStop,
		InputTokens:  len(req.Messages),
		OutputTokens: 1,
		TotalTokens:  len(req.Messages) + 1,
		CostUSD:      0,
		LatencyMs:    time.Since(start).Milliseconds(),
		Provider:     "echo",
		Model:        "echo-1",
		CreatedAt:    time.Now(),
	}, nil
}

func (e *EchoProvider) HealthCheck(ctx context.Context) error {
	return nil
}
