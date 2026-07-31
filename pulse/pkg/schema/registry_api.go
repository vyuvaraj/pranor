package import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Handler returns an http.Handler exposing the Schema Registry REST API.
//
// Routes:
//
//	GET    /api/v1/schemas/subjects                              - list subjects
//	POST   /api/v1/schemas/subjects/{subject}                   - register schema
//	GET    /api/v1/schemas/subjects/{subject}/versions          - list versions
//	GET    /api/v1/schemas/subjects/{subject}/versions/latest   - get latest schema
//	GET    /api/v1/schemas/subjects/{subject}/versions/{v}      - get specific version
//	POST   /api/v1/schemas/subjects/{subject}/validate          - validate a message
//	DELETE /api/v1/schemas/subjects/{subject}                   - delete subject
func Handler(r *Registry) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/schemas/subjects", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		jsonResp(w, r.ListSubjects())
	})

	// /api/v1/schemas/subjects/{subject}[/versions[/...] | /validate]
	mux.HandleFunc("/api/v1/schemas/subjects/", func(w http.ResponseWriter, req *http.Request) {
		// Strip prefix to get: {subject}[/versions[/latest|/{v}] | /validate]
		trimmed := strings.TrimPrefix(req.URL.Path, "/api/v1/schemas/subjects/")
		parts := strings.SplitN(trimmed, "/", 3)

		subject := parts[0]
		if subject == "" {
			http.Error(w, "subject required", http.StatusBadRequest)
			return
		}

		// No suffix: register (POST) or delete (DELETE)
		if len(parts) == 1 {
			switch req.Method {
			case http.MethodPost:
				handleRegister(w, req, r, subject)
			case http.MethodDelete:
				if err := r.DeleteSubject(subject); err != nil {
					http.Error(w, jsonErr(err.Error()), http.StatusNotFound)
					return
				}
				jsonResp(w, map[string]bool{"deleted": true})
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		suffix := parts[1] // "versions" or "validate"

		if suffix == "validate" {
			handleValidate(w, req, r, subject)
			return
		}

		if suffix != "versions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// /versions[/latest | /{v}]
		if len(parts) == 2 {
			// GET /versions — list version numbers
			if req.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			versions, err := r.ListVersions(subject)
			if err != nil {
				http.Error(w, jsonErr(err.Error()), http.StatusNotFound)
				return
			}
			jsonResp(w, versions)
			return
		}

		versionStr := parts[2]
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if versionStr == "latest" {
			s, err := r.GetLatest(subject)
			if err != nil {
				http.Error(w, jsonErr(err.Error()), http.StatusNotFound)
				return
			}
			jsonResp(w, s)
			return
		}

		v, err := strconv.Atoi(versionStr)
		if err != nil {
			http.Error(w, jsonErr("version must be an integer"), http.StatusBadRequest)
			return
		}
		s, err := r.GetSchema(subject, v)
		if err != nil {
			http.Error(w, jsonErr(err.Error()), http.StatusNotFound)
			return
		}
		jsonResp(w, s)
	})

	return mux
}

// handleRegister processes POST /api/v1/schemas/subjects/{subject}
func handleRegister(w http.ResponseWriter, req *http.Request, r *Registry, subject string) {
	body, err := io.ReadAll(io.LimitReader(req.Body, 1*1024*1024))
	if err != nil {
		http.Error(w, jsonErr("failed to read body"), http.StatusBadRequest)
		return
	}

	var payload struct {
		Format SchemaFormat `json:"format"`
		Schema string       `json:"schema"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, jsonErr("invalid JSON: "+err.Error()), http.StatusBadRequest)
		return
	}
	if payload.Format == "" {
		payload.Format = FormatJSONSchema
	}

	s, err := r.RegisterSchema(subject, payload.Format, payload.Schema)
	if err != nil {
		http.Error(w, jsonErr(err.Error()), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonResp(w, map[string]interface{}{"id": s.ID, "version": s.Version, "subject": s.Subject})
}

// handleValidate processes POST /api/v1/schemas/subjects/{subject}/validate
func handleValidate(w http.ResponseWriter, req *http.Request, r *Registry, subject string) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(req.Body, 4*1024*1024))
	if err != nil {
		http.Error(w, jsonErr("failed to read body"), http.StatusBadRequest)
		return
	}

	valErrs, err := r.ValidateMessage(subject, body)
	if err != nil {
		http.Error(w, jsonErr(err.Error()), http.StatusNotFound)
		return
	}

	if len(valErrs) == 0 {
		jsonResp(w, map[string]interface{}{"valid": true})
		return
	}
	w.WriteHeader(http.StatusUnprocessableEntity)
	jsonResp(w, map[string]interface{}{"valid": false, "errors": valErrs})
}

// ---- helpers ----

func jsonResp(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonErr(msg string) string {
	return fmt.Sprintf(`{"error":%q}`, msg)
}
