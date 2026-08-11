//go:build enterprise

package handlers

import "net/http"

// IsVisualDesignerSupported indicates if the real-time visual DAG designer is supported.
const IsVisualDesignerSupported = true

// HandleDesignerSave handles saving visual DAG designer templates in Enterprise Edition.
func (ctx *HandlerContext) HandleDesignerSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"saved","edition":"enterprise"}`))
}
