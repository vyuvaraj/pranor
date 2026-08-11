package capability

import (
    "errors"
    "sync"
    "testing"
)

func TestRegister_OK(t *testing.T) {
    reg := NewInMemoryRegistry()
    err := reg.Register(Capability{ID: "test.cap", Version: "1.0"})
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    c, err := reg.Lookup("test.cap")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if c.ID != "test.cap" {
        t.Errorf("expected test.cap, got %s", c.ID)
    }
}

func TestRegister_DuplicateID(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "test.cap", Version: "1.0"})
    err := reg.Register(Capability{ID: "test.cap", Version: "1.1"})
    if !errors.Is(err, ErrCapabilityAlreadyExists) {
        t.Errorf("expected ErrCapabilityAlreadyExists, got %v", err)
    }
}

func TestRegister_MissingID(t *testing.T) {
    reg := NewInMemoryRegistry()
    err := reg.Register(Capability{ID: "", Version: "1.0"})
    if !errors.Is(err, ErrInvalidCapability) {
        t.Errorf("expected ErrInvalidCapability, got %v", err)
    }
}

func TestLookup_NotFound(t *testing.T) {
    reg := NewInMemoryRegistry()
    _, err := reg.Lookup("nonexistent")
    if !errors.Is(err, ErrCapabilityNotFound) {
        t.Errorf("expected ErrCapabilityNotFound, got %v", err)
    }
}

func TestAuthorize_AllowedAll(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "test.cap", Version: "1.0"}) // empty slices
    err := reg.Authorize("t1", "a1", "test.cap")
    if err != nil {
        t.Errorf("expected allowed, got %v", err)
    }
}

func TestAuthorize_SpecificAgent_Allowed(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "test.cap", Version: "1.0", AllowedAgents: []string{"a1"}})
    err := reg.Authorize("t1", "a1", "test.cap")
    if err != nil {
        t.Errorf("expected allowed, got %v", err)
    }
}

func TestAuthorize_SpecificAgent_Denied(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "test.cap", Version: "1.0", AllowedAgents: []string{"a1"}})
    err := reg.Authorize("t1", "a2", "test.cap")
    if !errors.Is(err, ErrCapabilityUnauthorized) {
        t.Errorf("expected ErrCapabilityUnauthorized, got %v", err)
    }
}

func TestAuthorize_SpecificTenant_Denied(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "test.cap", Version: "1.0", AllowedTenants: []string{"t1"}})
    err := reg.Authorize("t2", "a1", "test.cap")
    if !errors.Is(err, ErrCapabilityUnauthorized) {
        t.Errorf("expected ErrCapabilityUnauthorized, got %v", err)
    }
}

func TestAuthorize_Wildcard(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "test.cap", Version: "1.0", AllowedAgents: []string{"*"}, AllowedTenants: []string{"*"}})
    err := reg.Authorize("any", "any", "test.cap")
    if err != nil {
        t.Errorf("expected allowed, got %v", err)
    }
}

func TestListByAgent(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "c1", Version: "1.0", AllowedAgents: []string{"a1"}})
    reg.Register(Capability{ID: "c2", Version: "1.0", AllowedAgents: []string{"a2"}})
    caps := reg.ListByAgent("a1")
    if len(caps) != 1 || caps[0].ID != "c1" {
        t.Errorf("expected c1, got %v", caps)
    }
}

func TestListByTenant(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "c1", Version: "1.0", AllowedTenants: []string{"t1"}})
    reg.Register(Capability{ID: "c2", Version: "1.0", AllowedTenants: []string{"t2"}})
    caps := reg.ListByTenant("t1")
    if len(caps) != 1 || caps[0].ID != "c1" {
        t.Errorf("expected c1, got %v", caps)
    }
}

func TestUnregister(t *testing.T) {
    reg := NewInMemoryRegistry()
    reg.Register(Capability{ID: "c1", Version: "1.0"})
    err := reg.Unregister("c1")
    if err != nil {
        t.Errorf("expected no error, got %v", err)
    }
    _, err = reg.Lookup("c1")
    if !errors.Is(err, ErrCapabilityNotFound) {
        t.Errorf("expected not found, got %v", err)
    }
}

func TestRiskClass_String(t *testing.T) {
    if RiskLow.String() != "LOW" { t.Errorf("expected LOW") }
    if RiskMedium.String() != "MEDIUM" { t.Errorf("expected MEDIUM") }
    if RiskHigh.String() != "HIGH" { t.Errorf("expected HIGH") }
    if RiskCritical.String() != "CRITICAL" { t.Errorf("expected CRITICAL") }
    if RiskClass(99).String() != "UNKNOWN" { t.Errorf("expected UNKNOWN") }
}

func TestProtocol_String(t *testing.T) {
    if ProtocolMCP.String() != "MCP" { t.Errorf("expected MCP") }
    if ProtocolGRPC.String() != "GRPC" { t.Errorf("expected GRPC") }
    if ProtocolREST.String() != "REST" { t.Errorf("expected REST") }
    if ProtocolWASM.String() != "WASM" { t.Errorf("expected WASM") }
    if ProtocolNative.String() != "NATIVE" { t.Errorf("expected NATIVE") }
    if Protocol(99).String() != "UNKNOWN" { t.Errorf("expected UNKNOWN") }
}

func TestDefaultRegistry(t *testing.T) {
    DefaultRegistry = NewInMemoryRegistry()
    err := Register(Capability{ID: "def", Version: "1"})
    if err != nil {
        t.Errorf("register error: %v", err)
    }
    _, err = Lookup("def")
    if err != nil {
        t.Errorf("lookup error: %v", err)
    }
    err = Authorize("any", "any", "def")
    if err != nil {
        t.Errorf("authorize error: %v", err)
    }
}

func TestConcurrentRegister(t *testing.T) {
    reg := NewInMemoryRegistry()
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id string) {
            defer wg.Done()
            reg.Register(Capability{ID: id, Version: "1.0"})
            reg.ListAll()
            reg.Lookup(id)
        }("id" + string(rune(i)))
    }
    wg.Wait()
}
