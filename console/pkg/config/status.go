package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vyuvaraj/pranor/core"
)

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	statuses := []ComponentStatus{
		CheckStatus("github.com/vyuvaraj/pranor/gate", *GateUrl),
		CheckStatus("github.com/vyuvaraj/pranor/vault", *StoreUrl),
		CheckStatus("github.com/vyuvaraj/pranor/pulse", *QueueUrl),
		CheckStatus("github.com/vyuvaraj/pranor/trace", *TraceUrl),
		CheckStatus("github.com/vyuvaraj/pranor/tunnel", *TunnelUrl),
		CheckStatus("github.com/vyuvaraj/pranor/auth", *AuthUrl),
		CheckStatus("ServDB", *DbUrl),
		CheckStatus("github.com/vyuvaraj/pranor/notify", *MailUrl),
		CheckStatus("github.com/vyuvaraj/pranor/flow", *FlowUrl),
		CheckStatus("github.com/vyuvaraj/pranor/mesh", *MeshUrl),
		CheckStatus("github.com/vyuvaraj/pranor/chrono", *CronUrl),
		CheckStatus("github.com/vyuvaraj/pranor/cache", *CacheUrl),
		CheckStatus("github.com/vyuvaraj/pranor/hub", *RegistryUrl),
		CheckStatus("github.com/vyuvaraj/pranor/deploy", *CloudUrl),
		CheckStatus("ServDocs", *DocsUrl),
	}

	json.NewEncoder(w).Encode(map[string]any{
		"timestamp":  time.Now().Format(time.RFC3339),
		"components": statuses,
	})
}

func CheckStatus(name string, baseUrl string) ComponentStatus {
	if baseUrl == "" {
		return ComponentStatus{Name: name, Online: false, Url: baseUrl}
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	reqUrl := fmt.Sprintf("%s/healthz", strings.TrimSuffix(baseUrl, "/"))
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return ComponentStatus{Name: name, Online: false, Url: baseUrl}
	}

	if jwtSec := os.Getenv("PRANOR_JWT_SECRET"); jwtSec != "" {
		svcToken, _ := Pranor Core.GenerateServiceToken(jwtSec, "github.com/vyuvaraj/pranor/console")
		if svcToken != "" {
			req.Header.Set("Authorization", "Bearer "+svcToken)
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return ComponentStatus{Name: name, Online: false, Url: baseUrl}
	}
	resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		return ComponentStatus{
			Name:      name,
			Online:    false,
			Url:       baseUrl,
			LatencyMs: latency,
		}
	}

	var details any
	var detailsPath string
	switch name {
	case "github.com/vyuvaraj/pranor/vault":
		detailsPath = "/console/metrics"
	case "github.com/vyuvaraj/pranor/pulse":
		detailsPath = "/api/stats"
	case "github.com/vyuvaraj/pranor/gate":
		detailsPath = "/"
	}

	if detailsPath != "" {
		detUrl := fmt.Sprintf("%s%s", strings.TrimSuffix(baseUrl, "/"), detailsPath)
		dreq, derr := http.NewRequest("GET", detUrl, nil)
		if derr == nil {
			if jwtSec := os.Getenv("PRANOR_JWT_SECRET"); jwtSec != "" {
				svcToken, _ := Pranor Core.GenerateServiceToken(jwtSec, "github.com/vyuvaraj/pranor/console")
				if svcToken != "" {
					dreq.Header.Set("Authorization", "Bearer "+svcToken)
				}
			}
			dresp, derr2 := client.Do(dreq)
			if derr2 == nil {
				bodyBytes, _ := io.ReadAll(dresp.Body)
				dresp.Body.Close()
				if len(bodyBytes) > 0 {
					_ = json.Unmarshal(bodyBytes, &details)
				}
			}
		}
	}

	return ComponentStatus{
		Name:      name,
		Online:    true,
		Url:       baseUrl,
		LatencyMs: latency,
		Details:   details,
	}
}
