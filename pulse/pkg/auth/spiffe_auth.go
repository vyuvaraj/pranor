package import (
	"fmt"
	"strings"
	"sync"
)

type SPIFFEIdentity struct {
	ID        string `json:"spiffe_id"`
	Domain    string `json:"domain"`
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
}

type SPIFFEAuthManager struct {
	mu            sync.RWMutex
	allowedSpiffes map[string]bool
}

func NewSPIFFEAuthManager() *SPIFFEAuthManager {
	return &SPIFFEAuthManager{
		allowedSpiffes: make(map[string]bool),
	}
}

func (s *SPIFFEAuthManager) AllowIdentity(spiffeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allowedSpiffes[spiffeID] = true
}

// ParseSPIFFEID parses standard SPIFFE URI format spiffe://domain/ns/namespace/sa/service
func ParseSPIFFEID(uri string) (*SPIFFEIdentity, error) {
	if !strings.HasPrefix(uri, "spiffe://") {
		return nil, fmt.Errorf("spiffe: invalid URI scheme")
	}

	trimmed := strings.TrimPrefix(uri, "spiffe://")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 1 {
		return nil, fmt.Errorf("spiffe: malformed identity")
	}

	identity := &SPIFFEIdentity{
		ID:     uri,
		Domain: parts[0],
	}

	for i := 1; i < len(parts)-1; i += 2 {
		if parts[i] == "ns" {
			identity.Namespace = parts[i+1]
		} else if parts[i] == "sa" {
			identity.Service = parts[i+1]
		}
	}

	return identity, nil
}

func (s *SPIFFEAuthManager) VerifySPIFFEID(spiffeID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.allowedSpiffes[spiffeID]
}
