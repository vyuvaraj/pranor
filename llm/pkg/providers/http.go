//go:build !enterprise

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vyuvaraj/pranor/llm/api"
)

type HTTPProvider struct {
	BaseURL      string
	APIKey       string
	DefaultModel string
	client       *http.Client
}

func NewHTTPProvider(baseURL, apiKey, defaultModel string) *HTTPProvider {
	return &HTTPProvider{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		DefaultModel: defaultModel,
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *HTTPProvider) Name() string {
	return "http"
}

func (h *HTTPProvider) Models() []string {
	return []string{h.DefaultModel}
}

type openaiReq struct {
	Model       string        `json:"model"`
	Messages    []api.Message `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
}

func (h *HTTPProvider) Chat(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
	start := time.Now()

	if req.BudgetMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.BudgetMs)*time.Millisecond)
		defer cancel()
	}

	model := req.Model
	if model == "" {
		model = h.DefaultModel
	}

	oReq := openaiReq{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      req.Stream,
		TopP:        req.TopP,
		Stop:        req.StopWords,
	}

	body, err := json.Marshal(oReq)
	if err != nil {
		return api.ChatResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, h.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return api.ChatResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if h.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+h.APIKey)
	}

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return api.ChatResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return api.ChatResponse{}, api.ErrModelUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return api.ChatResponse{}, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	var oResp struct {
		ID      string `json:"id"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return api.ChatResponse{}, err
	}

	var content string
	var finish string
	if len(oResp.Choices) > 0 {
		content = oResp.Choices[0].Message.Content
		finish = oResp.Choices[0].FinishReason
	}

	return api.ChatResponse{
		Content:      content,
		FinishReason: api.FinishReason(finish),
		InputTokens:  oResp.Usage.PromptTokens,
		OutputTokens: oResp.Usage.CompletionTokens,
		TotalTokens:  oResp.Usage.TotalTokens,
		LatencyMs:    time.Since(start).Milliseconds(),
		Provider:     "http",
		Model:        model,
		RequestID:    oResp.ID,
		CreatedAt:    time.Now(),
	}, nil
}

func (h *HTTPProvider) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, h.BaseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	if h.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+h.APIKey)
	}
	resp, err := h.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
	}
	return nil
}
