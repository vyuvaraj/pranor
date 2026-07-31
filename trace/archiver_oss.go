//go:build !enterprise

package trace

import (
	"log"
	"github.com/vyuvaraj/pranor/trace/pkg/store"
)

func SetupColdTierArchiver(ts *store.Store) {
	ts.OnEvict = func(traceID string, spans []store.Span) {
		log.Printf("Cold Tier: Evicting trace %s (Archival skipped in Open-Source Edition)", traceID)
	}
}
