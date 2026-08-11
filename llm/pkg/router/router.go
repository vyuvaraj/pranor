package router

import (
	"context"

	"github.com/vyuvaraj/pranor/llm/api"
)

type ossRouter struct {
	providers     []api.ChatProvider
	fallbackChain []string
	cache         api.CacheProvider
}

func NewOSSRouter() api.Router {
	return &ossRouter{
		providers:     make([]api.ChatProvider, 0),
		fallbackChain: make([]string, 0),
	}
}

func (r *ossRouter) Register(p api.ChatProvider) {
	r.providers = append(r.providers, p)
}

func (r *ossRouter) SetFallbackChain(providerNames []string) {
	r.fallbackChain = providerNames
}

func (r *ossRouter) SetCache(c api.CacheProvider) {
	r.cache = c
}

func (r *ossRouter) Route(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
	if len(req.Messages) == 0 {
		return api.ChatResponse{}, api.ErrEmptyMessages
	}

	if len(r.providers) == 0 {
		return api.ChatResponse{}, api.ErrNoProviders
	}

	providerMap := make(map[string]api.ChatProvider)
	for _, p := range r.providers {
		providerMap[p.Name()] = p
	}

	var lastErr error

	for _, name := range r.fallbackChain {
		p, ok := providerMap[name]
		if !ok {
			continue
		}

		resp, err := p.Chat(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	
	if lastErr != nil {
		return api.ChatResponse{}, api.ErrAllProvidersFailed
	}
	
	return api.ChatResponse{}, api.ErrAllProvidersFailed
}

func (r *ossRouter) RouteStream(ctx context.Context, req api.ChatRequest) (<-chan api.StreamChunk, error) {
	if len(req.Messages) == 0 {
		return nil, api.ErrEmptyMessages
	}

	if len(r.providers) == 0 {
		return nil, api.ErrNoProviders
	}

	providerMap := make(map[string]api.ChatProvider)
	for _, p := range r.providers {
		providerMap[p.Name()] = p
	}

	var lastErr error

	for _, name := range r.fallbackChain {
		p, ok := providerMap[name]
		if !ok {
			continue
		}

		ch, err := p.ChatStream(ctx, req)
		if err == nil {
			return ch, nil
		}
		lastErr = err
	}
	
	if lastErr != nil {
		return nil, api.ErrAllProvidersFailed
	}
	
	return nil, api.ErrAllProvidersFailed
}

func (r *ossRouter) HealthCheck(ctx context.Context) map[string]error {
	res := make(map[string]error)
	for _, p := range r.providers {
		res[p.Name()] = p.HealthCheck(ctx)
	}
	return res
}
