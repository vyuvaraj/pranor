//go:build !enterprise

package storage (s *InMemoryStore) replicateState() {
	// Open-source version does not replicate state to peers
}
