package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/vyuvaraj/pranor/llm/api"
	"github.com/vyuvaraj/pranor/llm/pkg/providers"
	"github.com/vyuvaraj/pranor/llm/pkg/router"
)

type errProvider struct{}
func (e *errProvider) Name() string { return "err" }
func (e *errProvider) Models() []string { return []string{"err-1"} }
func (e *errProvider) Chat(_ context.Context, _ api.ChatRequest) (api.ChatResponse, error) { return api.ChatResponse{}, errors.New("provider error") }
func (e *errProvider) HealthCheck(_ context.Context) error { return errors.New("unhealthy") }

func TestEchoProvider_Chat(t *testing.T) {
	p := providers.NewEchoProvider()
	resp, err := p.Chat(context.Background(), api.ChatRequest{
		Messages: []api.Message{{Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "hello" {
		t.Errorf("expected hello, got %s", resp.Content)
	}
}

func TestEchoProvider_HealthCheck(t *testing.T) {
	p := providers.NewEchoProvider()
	if err := p.HealthCheck(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestEchoProvider_Models(t *testing.T) {
	p := providers.NewEchoProvider()
	models := p.Models()
	if len(models) != 1 || models[0] != "echo-1" {
		t.Errorf("expected [echo-1], got %v", models)
	}
}

func TestRouter_NoProviders(t *testing.T) {
	r := router.NewOSSRouter()
	_, err := r.Route(context.Background(), api.ChatRequest{Messages: []api.Message{{Content: "hi"}}})
	if !errors.Is(err, api.ErrNoProviders) {
		t.Errorf("expected ErrNoProviders, got %v", err)
	}
}

func TestRouter_EmptyMessages(t *testing.T) {
	r := router.NewOSSRouter()
	_, err := r.Route(context.Background(), api.ChatRequest{})
	if !errors.Is(err, api.ErrEmptyMessages) {
		t.Errorf("expected ErrEmptyMessages, got %v", err)
	}
}

func TestRouter_EchoSuccess(t *testing.T) {
	r := router.NewOSSRouter()
	r.Register(providers.NewEchoProvider())
	r.SetFallbackChain([]string{"echo"})
	resp, err := r.Route(context.Background(), api.ChatRequest{Messages: []api.Message{{Content: "test"}}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "test" {
		t.Errorf("expected test, got %s", resp.Content)
	}
}

func TestRouter_FallbackChain_AllFail(t *testing.T) {
	r := router.NewOSSRouter()
	r.Register(&errProvider{})
	r.SetFallbackChain([]string{"err"})
	_, err := r.Route(context.Background(), api.ChatRequest{Messages: []api.Message{{Content: "hi"}}})
	if !errors.Is(err, api.ErrAllProvidersFailed) {
		t.Errorf("expected ErrAllProvidersFailed, got %v", err)
	}
}

func TestRouter_HealthCheck(t *testing.T) {
	r := router.NewOSSRouter()
	r.Register(providers.NewEchoProvider())
	res := r.HealthCheck(context.Background())
	if err, ok := res["echo"]; !ok || err != nil {
		t.Errorf("expected nil error for echo, got %v", err)
	}
}

func TestSentinelErrors(t *testing.T) {
	if api.ErrAllProvidersFailed == nil || api.ErrAllProvidersFailed.Error() != "pranor/llm: all providers in fallback chain failed" {
		t.Error("ErrAllProvidersFailed mismatch")
	}
	if api.ErrBudgetExceeded == nil || api.ErrBudgetExceeded.Error() != "pranor/llm: token or cost budget exceeded" {
		t.Error("ErrBudgetExceeded mismatch")
	}
	if api.ErrModelUnavailable == nil || api.ErrModelUnavailable.Error() != "pranor/llm: requested model unavailable on any provider" {
		t.Error("ErrModelUnavailable mismatch")
	}
	if api.ErrNoProviders == nil || api.ErrNoProviders.Error() != "pranor/llm: no providers registered" {
		t.Error("ErrNoProviders mismatch")
	}
	if api.ErrEmptyMessages == nil || api.ErrEmptyMessages.Error() != "pranor/llm: messages must not be empty" {
		t.Error("ErrEmptyMessages mismatch")
	}
	if api.ErrEERequired == nil || api.ErrEERequired.Error() != "pranor/llm: this provider requires Pranor Enterprise Edition" {
		t.Error("ErrEERequired mismatch")
	}
}

func TestHTTPProvider_Name(t *testing.T) {
	p := providers.NewHTTPProvider("", "", "")
	if p.Name() != "http" {
		t.Errorf("expected http, got %s", p.Name())
	}
}

func TestFinishReason_Constants(t *testing.T) {
	if api.FinishStop != "stop" {
		t.Error("FinishStop mismatch")
	}
	if api.FinishLength != "length" {
		t.Error("FinishLength mismatch")
	}
	if api.FinishToolCall != "tool_call" {
		t.Error("FinishToolCall mismatch")
	}
	if api.FinishFiltered != "content_filter" {
		t.Error("FinishFiltered mismatch")
	}
}
