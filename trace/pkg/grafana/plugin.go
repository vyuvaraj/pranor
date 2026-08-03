package grafana

import (
	"encoding/json"
	"net/http"

	"github.com/vyuvaraj/pranor/trace/pkg/store"
)

type GrafanaQueryRequest struct {
	Targets []struct {
		Target string `json:"target"`
		RefID  string `json:"refId"`
		Type   string `json:"type"`
	} `json:"targets"`
}

type GrafanaQueryResult struct {
	Target string      `json:"target"`
	Data   interface{} `json:"datapoints"`
}

type GrafanaPlugin struct {
	traceStore *store.Store
}

func NewGrafanaPlugin(ts *store.Store) *GrafanaPlugin {
	return &GrafanaPlugin{traceStore: ts}
}

// HandleTestConnection implements Grafana data source test endpoint
func (gp *GrafanaPlugin) HandleTestConnection(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Pranor Telemetry Data Source Connected"}`))
}

// HandleQuery processes Grafana dashboard query requests
func (gp *GrafanaPlugin) HandleQuery(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var qReq GrafanaQueryRequest
	if err := json.NewDecoder(req.Body).Decode(&qReq); err != nil {
		http.Error(w, "Invalid query payload", http.StatusBadRequest)
		return
	}

	traces := gp.traceStore.ListTraces()
	var results []GrafanaQueryResult

	for _, target := range qReq.Targets {
		datapoints := [][]interface{}{}
		for _, t := range traces {
			datapoints = append(datapoints, []interface{}{t.DurationMs, t.TimestampNano / 1000000})
		}

		results = append(results, GrafanaQueryResult{
			Target: target.Target,
			Data:   datapoints,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
