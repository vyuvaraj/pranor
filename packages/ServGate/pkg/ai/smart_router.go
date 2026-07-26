package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type PromptComplexity string

const (
	ComplexityLow  PromptComplexity = "LOW"
	ComplexityHigh PromptComplexity = "HIGH"
)

type SmartAIRouterConfig struct {
	LocalOllamaURL  string `json:"local_ollama_url"`
	OpenAIAPIURL    string `json:"openai_api_url"`
	EnablePrefetch  bool   `json:"enable_prefetch"`
	LowCostModel    string `json:"low_cost_model"`
	HighCostModel   string `json:"high_cost_model"`
}

type SmartAIRouter struct {
	config       SmartAIRouterConfig
	totalSavedUSD float64
	routedLow    uint64
	routedHigh   uint64
	prefetchCache map[string]string
	mu           sync.RWMutex
}

func NewSmartAIRouter(cfg SmartAIRouterConfig) *SmartAIRouter {
	if cfg.LocalOllamaURL == "" {
		cfg.LocalOllamaURL = "http://localhost:11434"
	}
	if cfg.OpenAIAPIURL == "" {
		cfg.OpenAIAPIURL = "https://api.openai.com/v1"
	}
	if cfg.LowCostModel == "" {
		cfg.LowCostModel = "llama3:8b"
	}
	if cfg.HighCostModel == "" {
		cfg.HighCostModel = "gpt-4o"
	}

	return &SmartAIRouter{
		config:        cfg,
		prefetchCache: make(map[string]string),
	}
}

func (s *SmartAIRouter) ClassifyPrompt(prompt string) PromptComplexity {
	wordCount := len(strings.Fields(prompt))
	hasComplexSyntax := strings.Contains(prompt, "refactor") || strings.Contains(prompt, "architecture") || strings.Contains(prompt, "proof") || strings.Contains(prompt, "analyze")

	if wordCount < 50 && !hasComplexSyntax {
		return ComplexityLow
	}
	return ComplexityHigh
}

func (s *SmartAIRouter) RouteAndExecute(ctx context.Context, prompt string) (string, PromptComplexity, float64, error) {
	complexity := s.ClassifyPrompt(prompt)

	s.mu.Lock()
	defer s.mu.Unlock()

	var estimatedSavings float64
	var modelUsed string

	if complexity == ComplexityLow {
		s.routedLow++
		modelUsed = s.config.LowCostModel
		estimatedSavings = 0.015 // Estimated $0.015 saved vs GPT-4o
		s.totalSavedUSD += estimatedSavings
	} else {
		s.routedHigh++
		modelUsed = s.config.HighCostModel
	}

	if s.config.EnablePrefetch {
		s.prefetchCache[prompt] = fmt.Sprintf("speculative-prefetch-response-for-%s", modelUsed)
	}

	res := fmt.Sprintf(`{"model":"%s","complexity":"%s","savings_usd":%.4f,"content":"ServGateway Smart AI Response"}`, modelUsed, complexity, estimatedSavings)
	return res, complexity, estimatedSavings, nil
}

func (s *SmartAIRouter) GetTelemetryStats() (uint64, uint64, float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routedLow, s.routedHigh, s.totalSavedUSD
}
