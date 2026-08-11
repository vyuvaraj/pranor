//go:build !enterprise

package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

func (h *HTTPProvider) ChatStream(ctx context.Context, req api.ChatRequest) (<-chan api.StreamChunk, error) {
	var cancel context.CancelFunc
	if req.BudgetMs > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.BudgetMs)*time.Millisecond)
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
		Stream:      true,
		TopP:        req.TopP,
		Stop:        req.StopWords,
	}

	body, err := json.Marshal(oReq)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, h.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if h.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+h.APIKey)
	}

	resp, err := h.client.Do(httpReq)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		if cancel != nil {
			cancel()
		}
		return nil, api.ErrModelUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if cancel != nil {
			cancel()
		}
		return nil, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	ch := make(chan api.StreamChunk)

	go func() {
		defer resp.Body.Close()
		defer close(ch)
		if cancel != nil {
			defer cancel()
		}

		scanner := bufio.NewScanner(resp.Body)
		tokenIndex := 0

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				ch <- api.StreamChunk{Error: err}
				return
			}

			if len(streamResp.Choices) > 0 {
				choice := streamResp.Choices[0]
				ch <- api.StreamChunk{
					Content:      choice.Delta.Content,
					FinishReason: api.FinishReason(choice.FinishReason),
					TokenIndex:   tokenIndex,
				}
				tokenIndex++
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- api.StreamChunk{Error: err}
		}
	}()

	return ch, nil
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
