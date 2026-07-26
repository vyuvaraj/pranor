package ai

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type LLMProvider string

const (
	ProviderOpenAI    LLMProvider = "openai"
	ProviderAnthropic LLMProvider = "anthropic"
	ProviderOllama    LLMProvider = "ollama"
)

type LLMProxyConfig struct {
	Provider    LLMProvider `json:"provider"`
	TargetURL   string      `json:"target_url"`
	APIKey      string      `json:"api_key"`
	MaxTPM      int         `json:"max_tpm"` // Tokens per minute
	MaxRPM      int         `json:"max_rpm"` // Requests per minute
	EnableCache bool        `json:"enable_cache"`
}

type LLMEdgeProxy struct {
	config     LLMProxyConfig
	cache      map[string][]byte
	cacheMu    sync.RWMutex
	tpmCounter int
	rpmCounter int
	lastReset  time.Time
	rateMu     sync.Mutex
}

func NewLLMEdgeProxy(cfg LLMProxyConfig) *LLMEdgeProxy {
	if cfg.TargetURL == "" {
		switch cfg.Provider {
		case ProviderAnthropic:
			cfg.TargetURL = "https://api.anthropic.com"
		case ProviderOllama:
			cfg.TargetURL = "http://localhost:11434"
		default:
			cfg.TargetURL = "https://api.openai.com"
		}
	}

	return &LLMEdgeProxy{
		config:    cfg,
		cache:     make(map[string][]byte),
		lastReset: time.Now(),
	}
}

func (p *LLMEdgeProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !p.checkRateLimit() {
		http.Error(w, `{"error":{"message":"ServGateway AI Rate Limit Exceeded (TPM/RPM threshold hit)","type":"rate_limit_error"}}`, http.StatusTooManyRequests)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if p.config.EnableCache && r.Method == http.MethodPost {
		cacheKey := p.computeCacheKey(body)
		p.cacheMu.RLock()
		cachedResp, found := p.cache[cacheKey]
		p.cacheMu.RUnlock()

		if found {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-ServGateway-AI-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			w.Write(cachedResp)
			return
		}
	}

	// Forward request to upstream LLM API
	targetReq, err := http.NewRequestWithContext(r.Context(), r.Method, p.config.TargetURL+r.URL.Path, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	targetReq.Header = r.Header.Clone()
	if p.config.APIKey != "" {
		targetReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(targetReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("LLM Upstream Error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read upstream response", http.StatusInternalServerError)
		return
	}

	if p.config.EnableCache && resp.StatusCode == http.StatusOK {
		cacheKey := p.computeCacheKey(body)
		p.cacheMu.Lock()
		p.cache[cacheKey] = respBody
		p.cacheMu.Unlock()
	}

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.Header().Set("X-ServGateway-AI-Cache", "MISS")
	w.Header().Set("X-ServGateway-AI-Provider", string(p.config.Provider))
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func (p *LLMEdgeProxy) checkRateLimit() bool {
	p.rateMu.Lock()
	defer p.rateMu.Unlock()

	if time.Since(p.lastReset) > time.Minute {
		p.tpmCounter = 0
		p.rpmCounter = 0
		p.lastReset = time.Now()
	}

	if p.config.MaxRPM > 0 && p.rpmCounter >= p.config.MaxRPM {
		return false
	}

	p.rpmCounter++
	return true
}

func (p *LLMEdgeProxy) computeCacheKey(body []byte) string {
	h := sha256.New()
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

type OpenAICompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
