//go:build !enterprise

package import (
	"errors"
)

func (s *Server) ResolveNaturalLanguageQuery(query string) (map[string]string, error) {
	return nil, errors.New("Enterprise Edition required for natural language log/trace search")
}
