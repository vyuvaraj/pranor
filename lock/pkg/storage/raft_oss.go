//go:build !enterprise

package storage

import "log"

func (s *InMemoryStore) replicateState() {
	// Open-source version does not replicate state to peers
	log.Println("[Raft] Replication disabled in OSS edition")
}
