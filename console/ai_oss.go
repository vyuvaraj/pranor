//go:build !enterprise

package console

import (
	"net/http"
)

func registerAIHandlers(mux *http.ServeMux) {
	// Return 403 Forbidden for AI endpoints in Open-Source edition
	forbiddenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSONError(w, r, "AI features require Pranor Console Enterprise Edition", "ERR_EE_REQUIRED", http.StatusForbidden)
	})

	mux.Handle("/api/incidents/analyze", authorizeConsole(forbiddenHandler))
	mux.Handle("/api/ai/metrics", authorizeConsole(forbiddenHandler))
}

// handleIncidentAnalyze returns 403 in the OSS build.
func handleIncidentAnalyze(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, r, "AI Incident Analysis requires Pranor Console Enterprise Edition", "ERR_EE_REQUIRED", http.StatusForbidden)
}

// handleAIMetrics returns 403 in the OSS build.
func handleAIMetrics(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, r, "AI Metrics require Pranor Console Enterprise Edition", "ERR_EE_REQUIRED", http.StatusForbidden)
}
