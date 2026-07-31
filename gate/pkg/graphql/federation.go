package import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type GraphQLUpstream struct {
	Name string
	URL  string
}

type GraphQLFederationEngine struct {
	upstreams map[string]GraphQLUpstream
	mu        sync.RWMutex
}

func NewGraphQLFederationEngine() *GraphQLFederationEngine {
	return &GraphQLFederationEngine{
		upstreams: make(map[string]GraphQLUpstream),
	}
}

func (f *GraphQLFederationEngine) RegisterUpstream(name, url string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if name == "" || url == "" {
		return fmt.Errorf("graphql federation: missing upstream name or URL")
	}

	f.upstreams[name] = GraphQLUpstream{Name: name, URL: url}
	return nil
}

func (f *GraphQLFederationEngine) ExecuteFederatedQuery(ctx context.Context, query string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if query == "" {
		return "", fmt.Errorf("graphql federation: empty query")
	}

	upstreamList := make([]string, 0, len(f.upstreams))
	for name := range f.upstreams {
		upstreamList = append(upstreamList, name)
	}

	return fmt.Sprintf(`{"data":{"federated_nodes":[%s],"status":"SUCCESS"}}`, strings.Join(upstreamList, ",")), nil
}
