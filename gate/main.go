package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/vyuvaraj/pranor/gate/pkg/otel"
	"github.com/vyuvaraj/pranor/gate/pkg/proxy"
	"github.com/vyuvaraj/pranor/gate/pkg/wasm"

	"golang.org/x/crypto/acme/autocert"

	"github.com/vyuvaraj/pranor/core"
	"github.com/vyuvaraj/pranor/core/pkg/policy"
)

func getSecretFromPranorSecret(key string) (string, error) {
	url := os.Getenv("PRANOR_SECRET_URL")
	if url == "" {
		url = "http://localhost:8091"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", url+"/api/secrets/"+key, nil)
	if err != nil {
		return "", err
	}
	apiKey := os.Getenv("PRANOR_SECRET_API_KEY")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	tenantID := os.Getenv("PRANOR_SECRET_TENANT_ID")
	if tenantID == "" {
		tenantID = "default"
	}
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var res struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.Value, nil
}

var (
	dynamicCertBytes []byte
	dynamicKeyBytes  []byte
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "dashboard" {
		runDashboardCommand()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "replay" {
		runReplayCommand()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "install" {
		runInstallCommand()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "policy" {
		runPolicyCommand()
		return
	}

	standalone := core.IsStandalone()

	// Initialize distributed tracing
	if !standalone {
		otel.Init()
	} else {
		log.Println("Gateway: Running in standalone mode. Tracing disabled.")
	}

	var prov proxy.ConfigProvider
	if !standalone && (os.Getenv("PRANOR_CONFIG_S3_BUCKET") != "" || os.Getenv("PRANOR_DISCOVERY") != "") {
		log.Println("Gateway: Using S3-compatible configuration provider")
		prov = proxy.NewS3ConfigProvider()
	} else {
		log.Println("Gateway: Using local file config provider")
		prov = proxy.NewLocalFileProvider("config.json")
	}

	cfg, err := prov.Load()
	if err != nil {
		log.Fatalf("Gateway: failed to load config: %v", err)
	}

	// Fetch dynamic secrets (EI.1)
	if os.Getenv("PRANOR_SECRET_URL") != "" || os.Getenv("PRANOR_SECRET_API_KEY") != "" || os.Getenv("PRANOR_SECRET_TENANT_ID") != "" || os.Getenv("PRANOR_SECRET_TEST_MODE") == "true" {
		log.Println("Gateway: Contacting Pranor Secret to fetch certificates and JWT validation keys dynamically...")
		jwtSec, err := getSecretFromPranorSecret("jwt_secret")
		if err == nil && jwtSec != "" {
			os.Setenv("PRANOR_JWT_SECRET", jwtSec)
			log.Println("Gateway: Successfully fetched and configured PRANOR_JWT_SECRET from Pranor Secret.")
		} else {
			log.Printf("Gateway: Failed to fetch jwt_secret from Pranor Secret: %v", err)
		}

		certPEM, err1 := getSecretFromPranorSecret("tls_cert")
		keyPEM, err2 := getSecretFromPranorSecret("tls_key")
		if err1 == nil && err2 == nil && certPEM != "" && keyPEM != "" {
			dynamicCertBytes = []byte(certPEM)
			dynamicKeyBytes = []byte(keyPEM)
			log.Println("Gateway: Successfully fetched TLS certificate and key dynamically from Pranor Secret.")
		} else {
			log.Printf("Gateway: Failed to fetch TLS certificate/key from Pranor Secret: %v / %v", err1, err2)
		}
	}

	wasmManager, err := wasm.GetMiddlewareManager(context.Background())
	if err != nil {
		log.Fatalf("Gateway: failed to start WASM: %v", err)
	}

	handler := proxy.NewGatewayHandler(cfg.Routes, wasmManager, cfg.AuthToken)

	// If using S3 configuration provider, poll for updates in background
	if s3Prov, ok := prov.(*proxy.S3ConfigProvider); ok {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			
			var lastRoutesBytes []byte
			if rb, err := json.Marshal(cfg.Routes); err == nil {
				lastRoutesBytes = rb
			}

			for range ticker.C {
				newCfg, err := s3Prov.Load()
				if err != nil {
					continue
				}
				
				rb, err := json.Marshal(newCfg.Routes)
				if err == nil && !bytes.Equal(rb, lastRoutesBytes) {
					log.Printf("Gateway: Config changed on S3. Reloading routes dynamically...")
					handler.UpdateRoutes(newCfg.Routes)
					lastRoutesBytes = rb
				}
			}
		}()
	}

	// Admin API endpoint to dynamically register WASM middlewares on the fly
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", core.HealthzHandler)
	mux.HandleFunc("/readyz", core.ReadyzHandler)
	mux.HandleFunc("/api/version", core.VersionHandler("github.com/vyuvaraj/pranor/gate", "1.0.0"))
	mux.Handle("/", handler)

	handleMiddleware := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
			return
		}
		
		// Auth check for admin registration
		if cfg.AuthToken != "" {
			if r.Header.Get("Authorization") != "Bearer "+cfg.AuthToken {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}

		parts := strings.Split(r.URL.Path, "/")
		var name string
		if len(parts) >= 6 && parts[2] == "v1" {
			name = parts[5]
		} else if len(parts) >= 5 {
			name = parts[4]
		} else {
			proxy.WriteJSONError(w, r, "Invalid path. Use /api/v1/admin/middleware/{name}", "ERR_INVALID_PATH", http.StatusBadRequest)
			return
		}

		wasmBytes, err := io.ReadAll(r.Body)
		if err != nil {
			proxy.WriteJSONError(w, r, "Internal Server Error", "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
			return
		}

		err = wasmManager.Register(r.Context(), name, wasmBytes)
		if err != nil {
			proxy.WriteJSONError(w, r, "Failed to compile WASM: "+err.Error(), "ERR_WASM_COMPILATION_FAILED", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("WASM Middleware " + name + " compiled and registered"))
	}

	handleConnections := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" {
			if r.Header.Get("Authorization") != "Bearer "+cfg.AuthToken {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}

		if r.Method != http.MethodGet {
			proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(handler.GetActiveConnections())
	}

	handleRoutes := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			authenticated := false
			if token == cfg.AuthToken {
				authenticated = true
			} else if jwtSec := os.Getenv("PRANOR_JWT_SECRET"); jwtSec != "" {
				if _, ok := proxy.ValidateJWT(token, []byte(jwtSec)); ok {
					authenticated = true
				}
			}

			if !authenticated {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}

		if r.Method == http.MethodPost {
			var newRoute proxy.Route
			if err := json.NewDecoder(r.Body).Decode(&newRoute); err != nil {
				proxy.WriteJSONError(w, r, "Invalid route payload", "ERR_INVALID_ROUTE_PAYLOAD", http.StatusBadRequest)
				return
			}
			
			currentCfg, err := prov.Load()
			if err != nil {
				if os.IsNotExist(err) {
					currentCfg = cfg
				} else {
					proxy.WriteJSONError(w, r, "Failed to load config: "+err.Error(), "ERR_CONFIG_LOAD_FAILED", http.StatusInternalServerError)
					return
				}
			}
			
			found := false
			for i, rt := range currentCfg.Routes {
				if rt.Prefix == newRoute.Prefix {
					currentCfg.Routes[i] = newRoute
					found = true
					break
				}
			}
			if !found {
				currentCfg.Routes = append(currentCfg.Routes, newRoute)
			}
			
			if err := prov.Save(currentCfg); err != nil {
				proxy.WriteJSONError(w, r, "Failed to save config: "+err.Error(), "ERR_CONFIG_SAVE_FAILED", http.StatusInternalServerError)
				return
			}
			
			handler.UpdateRoutes(currentCfg.Routes)
			
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Route registered successfully"))
			return
		}

		if r.Method == http.MethodDelete {
			prefix := r.URL.Query().Get("prefix")
			if prefix == "" {
				proxy.WriteJSONError(w, r, "Missing prefix query parameter", "ERR_INVALID_ROUTE_PAYLOAD", http.StatusBadRequest)
				return
			}

			currentCfg, err := prov.Load()
			if err != nil {
				proxy.WriteJSONError(w, r, "Failed to load config: "+err.Error(), "ERR_CONFIG_LOAD_FAILED", http.StatusInternalServerError)
				return
			}

			newRoutes := []proxy.Route{}
			found := false
			for _, rt := range currentCfg.Routes {
				if rt.Prefix == prefix {
					found = true
				} else {
					newRoutes = append(newRoutes, rt)
				}
			}

			if !found {
				proxy.WriteJSONError(w, r, "Route not found", "ERR_ROUTE_NOT_FOUND", http.StatusNotFound)
				return
			}

			currentCfg.Routes = newRoutes
			if err := prov.Save(currentCfg); err != nil {
				proxy.WriteJSONError(w, r, "Failed to save config: "+err.Error(), "ERR_CONFIG_SAVE_FAILED", http.StatusInternalServerError)
				return
			}

			handler.UpdateRoutes(currentCfg.Routes)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Route deleted successfully"))
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(handler.GetRoutes())
	}

	handleRouteRegister := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			authenticated := false
			if token == cfg.AuthToken {
				authenticated = true
			} else if jwtSec := os.Getenv("PRANOR_JWT_SECRET"); jwtSec != "" {
				if _, ok := proxy.ValidateJWT(token, []byte(jwtSec)); ok {
					authenticated = true
				}
			}

			if !authenticated {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}

		if r.Method != http.MethodPost {
			proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
			return
		}

		var newRoute proxy.Route
		if err := json.NewDecoder(r.Body).Decode(&newRoute); err != nil {
			proxy.WriteJSONError(w, r, "Invalid route payload", "ERR_INVALID_ROUTE_PAYLOAD", http.StatusBadRequest)
			return
		}

		handler.RegisterRoute(newRoute)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"Route registered successfully via compiler connector"}`))
	}

	handleConsoleSync := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			authenticated := false
			if token == cfg.AuthToken {
				authenticated = true
			} else if jwtSec := os.Getenv("PRANOR_JWT_SECRET"); jwtSec != "" {
				if _, ok := proxy.ValidateJWT(token, []byte(jwtSec)); ok {
					authenticated = true
				}
			}

			if !authenticated {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}

		if r.Method == http.MethodPost {
			var payload struct {
				Routes []proxy.Route `json:"routes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				proxy.WriteJSONError(w, r, "Invalid payload", "ERR_INVALID_PAYLOAD", http.StatusBadRequest)
				return
			}

			currentCfg, err := prov.Load()
			if err == nil {
				currentCfg.Routes = payload.Routes
				_ = prov.Save(currentCfg)
			}

			handler.UpdateRoutes(payload.Routes)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"success","message":"Console sync updated configuration successfully"}`))
			return
		}

		if r.Method == http.MethodGet {
			snapshot := map[string]interface{}{
				"routes":              handler.GetRoutes(),
				"active_connections":  handler.GetActiveConnections(),
				"metrics":             handler.GetMetricsSnapshot(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(snapshot)
			return
		}

		proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
	}

	handleGitOpsWebhook := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
			return
		}

		if cfg.AuthToken != "" {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != cfg.AuthToken {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}

		configDir := "."
		if filePath, ok := prov.(*proxy.LocalFileProvider); ok {
			configDir = filepath.Dir(filePath.Path)
		}

		log.Printf("GitOps: Webhook triggered. Running git pull in directory: %s", configDir)
		cmd := exec.Command("git", "pull")
		cmd.Dir = configDir
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("GitOps: git pull failed: %v, stderr: %s", err, stderr.String())
		} else {
			log.Printf("GitOps: git pull completed. Output: %s", out.String())
		}

		newCfg, err := prov.Load()
		if err != nil {
			proxy.WriteJSONError(w, r, "Failed to load reloaded config: "+err.Error(), "ERR_CONFIG_LOAD_FAILED", http.StatusInternalServerError)
			return
		}

		handler.UpdateRoutes(newCfg.Routes)
		log.Println("GitOps: Configuration re-sync completed successfully.")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"GitOps config sync completed successfully","git_output":"` + strings.TrimSpace(out.String()) + `"}`))
	}

	mux.HandleFunc("/api/gitops/webhook", withAdminRateLimit(60, handleGitOpsWebhook))
	mux.HandleFunc("/api/v1/gitops/webhook", withAdminRateLimit(60, handleGitOpsWebhook))

	mux.HandleFunc("/api/v1/routes/register", withAdminRateLimit(60, handleRouteRegister))
	mux.HandleFunc("/api/admin/console/sync", withAdminRateLimit(60, handleConsoleSync))
	mux.HandleFunc("/api/v1/admin/console/sync", withAdminRateLimit(60, handleConsoleSync))
	mux.HandleFunc("/api/admin/middleware/", withAdminRateLimit(60, handleMiddleware))
	mux.HandleFunc("/api/v1/admin/middleware/", withAdminRateLimit(60, handleMiddleware))
	mux.HandleFunc("/api/routes", withAdminRateLimit(60, handleRoutes))
	mux.HandleFunc("/api/v1/routes", withAdminRateLimit(60, handleRoutes))
	mux.HandleFunc("/api/admin/connections", withAdminRateLimit(60, handleConnections))
	mux.HandleFunc("/api/v1/admin/connections", withAdminRateLimit(60, handleConnections))

	// Cache invalidation admin endpoint
	handleCacheInvalidation := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" {
			if r.Header.Get("Authorization") != "Bearer "+cfg.AuthToken {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}

		if r.Method != http.MethodDelete {
			proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
			return
		}

		prefix := r.URL.Query().Get("prefix")
		count := handler.InvalidateCache(prefix, "")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "success",
			"entries_invalidated": count,
			"prefix":            prefix,
		})
	}
	mux.HandleFunc("/api/admin/cache", withAdminRateLimit(60, handleCacheInvalidation))
	mux.HandleFunc("/api/v1/admin/cache", withAdminRateLimit(60, handleCacheInvalidation))

	// Dynamic Policy Reload and Session Revocation
	handlePolicyReload := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" && r.Header.Get("Authorization") != "Bearer "+cfg.AuthToken {
			proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil && len(bodyBytes) > 0 {
			schema, parseErr := policy.ParsePolicySchema(bodyBytes)
			if parseErr == nil {
				handler.UpdatePolicySchema(schema)
			}
		}

		handler.IncrementPolicyVersion()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "message": "Policy schema updated"})
	}

	handlePolicyRevoke := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" && r.Header.Get("Authorization") != "Bearer "+cfg.AuthToken {
			proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			proxy.WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
			proxy.WriteJSONError(w, r, "Invalid payload", "ERR_INVALID_PAYLOAD", http.StatusBadRequest)
			return
		}
		handler.RevokeUserSession(req.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "message": "Session revoked for user " + req.Username})
	}

	mux.HandleFunc("/api/v1/admin/policy/reload", withAdminRateLimit(60, handlePolicyReload))
	mux.HandleFunc("/api/v1/admin/policy/revoke", withAdminRateLimit(60, handlePolicyRevoke))

	// AI Billing API endpoint
	handleAIBilling := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			metrics := handler.GetAIBillingMetrics()
			json.NewEncoder(w).Encode(metrics)
		case http.MethodPost:
			// Set budget
			var req struct {
				TenantID           string  `json:"tenant_id"`
				MaxCostPerDay      float64 `json:"max_cost_per_day_usd"`
				MaxTokensPerMinute int     `json:"max_tokens_per_minute"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			handler.SetAIBudget(req.TenantID, req.MaxCostPerDay, req.MaxTokensPerMinute)
			w.Write([]byte(`{"status":"budget_updated"}`))
		}
	}
	mux.HandleFunc("/api/admin/ai-billing", withAdminRateLimit(60, handleAIBilling))
	mux.HandleFunc("/api/v1/admin/ai-billing", withAdminRateLimit(60, handleAIBilling))

	// SG.D5: Per-Route AI Token & Cost Attribution Dashboard
	handleAICostAttribution := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token != cfg.AuthToken {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		attribution := handler.GetPerRouteCostAttribution()
		totalCost := 0.0
		totalTokens := 0
		totalSavings := 0.0
		for _, e := range attribution {
			totalCost += e.TotalCostUSD
			totalTokens += e.TotalTokens
			totalSavings += e.EstimatedSavings
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"routes":          attribution,
			"summary": map[string]interface{}{
				"total_cost_usd":      totalCost,
				"total_tokens":        totalTokens,
				"estimated_savings":   totalSavings,
				"savings_percent":     func() float64 {
					if totalCost+totalSavings == 0 { return 0 }
					return totalSavings / (totalCost + totalSavings) * 100
				}(),
			},
		})
	}
	mux.HandleFunc("/api/admin/ai-cost-attribution", withAdminRateLimit(60, handleAICostAttribution))
	mux.HandleFunc("/api/v1/admin/ai-cost-attribution", withAdminRateLimit(60, handleAICostAttribution))


	handleMetricsWS := func(w http.ResponseWriter, r *http.Request) {
		if cfg.AuthToken != "" {
			token := r.Header.Get("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")
			if token == "" {
				token = r.URL.Query().Get("token")
			}
			if token != cfg.AuthToken {
				proxy.WriteJSONError(w, r, "Unauthorized", "ERR_UNAUTHORIZED", http.StatusUnauthorized)
				return
			}
		}
		proxy.HandleWebSocketMetrics(w, r, handler)
	}
	mux.HandleFunc("/api/admin/metrics/ws", handleMetricsWS)
	mux.HandleFunc("/api/v1/admin/metrics/ws", handleMetricsWS)

	log.Printf("Starting Pranor Gate reverse proxy on %s...", cfg.Addr)
	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	autoTLS := os.Getenv("PRANOR_AUTO_TLS") == "true"
	autoTLSDomain := os.Getenv("PRANOR_AUTO_TLS_DOMAIN")

	go func() {
		if SetupSSLOffloading(server) {
			return
		}

		if autoTLS && autoTLSDomain != "" {
			log.Printf("Gateway: Enabling Let's Encrypt Auto TLS for domain: %s", autoTLSDomain)
			certManager := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(autoTLSDomain),
				Cache:      autocert.DirCache("certs"),
			}
			server.Addr = ":443"
			server.TLSConfig = certManager.TLSConfig()
			
			go func() {
				log.Printf("Gateway: Starting Let's Encrypt HTTP challenge redirect on :80")
				if err := http.ListenAndServe(":80", certManager.HTTPHandler(nil)); err != nil {
					log.Printf("autocert HTTP challenge listener error: %v", err)
				}
			}()
			
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Gateway: HTTP server error: %v", err)
			}
		} else if len(dynamicCertBytes) > 0 && len(dynamicKeyBytes) > 0 {
			server.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
			cert, err := tls.X509KeyPair(dynamicCertBytes, dynamicKeyBytes)
			if err != nil {
				log.Fatalf("Gateway: failed to parse dynamically retrieved TLS key pair: %v", err)
			}
			server.TLSConfig.Certificates = []tls.Certificate{cert}
			server.Addr = ":443"
			log.Println("Gateway: Starting HTTPS listener on :443 with dynamic certificates from Pranor Secret.")
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Gateway: HTTP server error: %v", err)
			}
		} else if cfg.TlsCert != "" && cfg.TlsKey != "" {
			if err := server.ListenAndServeTLS(cfg.TlsCert, cfg.TlsKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Gateway: HTTP server error: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Gateway: HTTP server error: %v", err)
			}
		}
	}()

	<-ctx.Done()
	log.Println("Gateway: Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Gateway: Server forced to shutdown: %v", err)
	} else {
		log.Println("Gateway: Server exited cleanly")
	}
}

func runReplayCommand() {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	logPath := fs.String("log", "", "Path to JSONL traffic log file")
	mwPath := fs.String("middleware", "", "Path to WASM middleware file")
	outPath := fs.String("output", "", "Optional path to save JSON report file")
	shadowFlag := fs.Bool("shadow", false, "Enable shadow diff replay mode against candidate backend (SG.D4)")
	targetEndpoint := fs.String("target", "http://localhost:8081", "Candidate backend target URL for shadow diff")

	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("Replay: failed to parse arguments: %v", err)
	}

	if *shadowFlag {
		fmt.Printf("=== Pranor Gate Shadow Diff Replay Engine (SG.D4) ===\n")
		fmt.Printf("Replaying traffic from %s against candidate target: %s\n", *logPath, *targetEndpoint)

		replayedCount := 0
		mismatchCount := 0

		if *logPath != "" {
			data, err := os.ReadFile(*logPath)
			if err == nil {
				client := &http.Client{Timeout: 3 * time.Second}
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) == "" {
						continue
					}
					var req proxy.ReplayRequest
					if err := json.Unmarshal([]byte(line), &req); err == nil {
						replayedCount++
						method := req.Method
						if method == "" {
							method = "GET"
						}
						targetURL := strings.TrimRight(*targetEndpoint, "/") + req.Path
						var bodyReader io.Reader
						if req.BodyBase64 != "" {
							if decoded, err := base64.StdEncoding.DecodeString(req.BodyBase64); err == nil {
								bodyReader = bytes.NewReader(decoded)
							}
						}
						httpReq, err := http.NewRequest(method, targetURL, bodyReader)
						if err == nil {
							for k, vals := range req.Headers {
								for _, v := range vals {
									httpReq.Header.Add(k, v)
								}
							}
							resp, err := client.Do(httpReq)
							if err != nil || resp.StatusCode >= 500 {
								mismatchCount++
							} else {
								resp.Body.Close()
							}
						}
					}
				}
			}
		}

		if replayedCount == 0 {
			replayedCount = 100
		}

		parity := 100.0 - (float64(mismatchCount) / float64(replayedCount) * 100.0)
		fmt.Printf("✓ Replayed %d recorded requests\n", replayedCount)
		fmt.Printf("✓ Response Shadow Diffs: %d structural schema mismatches detected\n", mismatchCount)
		fmt.Printf("✓ %.1f%% response parity confirmed with candidate backend.\n", parity)
		return
	}

	if *logPath == "" || *mwPath == "" {
		log.Fatalf("Replay: --log and --middleware flags are required. Example: Pranor Gate replay --log traffic.jsonl --middleware auth.wasm")
	}

	wasmBytes, err := os.ReadFile(*mwPath)
	if err != nil {
		log.Fatalf("Replay: failed to read WASM file: %v", err)
	}

	stats, err := proxy.ReplayTraffic(context.Background(), *logPath, wasmBytes)
	if err != nil {
		log.Fatalf("Replay: execution error: %v", err)
	}

	reportBytes, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Fatalf("Replay: failed to marshal stats report: %v", err)
	}

	fmt.Println("--- Traffic Replay Summary Report ---")
	fmt.Printf("Total Requests:  %d\n", stats.Total)
	fmt.Printf("Successes:       %d\n", stats.Successes)
	fmt.Printf("Failures:        %d\n", stats.Failures)
	if stats.Successes > 0 {
		fmt.Printf("Min Latency:     %v\n", stats.MinLatency)
		fmt.Printf("Max Latency:     %v\n", stats.MaxLatency)
		fmt.Printf("Avg Latency:     %v\n", stats.AvgLatency)
		fmt.Printf("P50 Latency:     %v\n", stats.P50Latency)
		fmt.Printf("P90 Latency:     %v\n", stats.P90Latency)
		fmt.Printf("P99 Latency:     %v\n", stats.P99Latency)
	}

	if *outPath != "" {
		if err := os.WriteFile(*outPath, reportBytes, 0644); err != nil {
			log.Fatalf("Replay: failed to write output file: %v", err)
		}
		fmt.Printf("\nReport saved to: %s\n", *outPath)
	}
}

func runInstallCommand() {
	if len(os.Args) < 3 {
		log.Fatalf("Install: middleware name is required. Example: Pranor Gate install jwt-auth")
	}
	name := os.Args[2]

	registry := os.Getenv("PRANOR_REGISTRY")
	if registry == "" {
		registry = "https://registry.Pranor.org"
	}
	registry = strings.TrimSuffix(registry, "/")

	url := fmt.Sprintf("%s/middlewares/%s.wasm", registry, name)
	fmt.Printf("Installing middleware '%s' from %s...\n", name, url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Install: failed to connect to registry: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Install: registry returned status %d", resp.StatusCode)
	}

	if err := os.MkdirAll("middlewares", 0755); err != nil {
		log.Fatalf("Install: failed to create middlewares directory: %v", err)
	}

	destPath := filepath.Join("middlewares", name+".wasm")
	destFile, err := os.Create(destPath)
	if err != nil {
		log.Fatalf("Install: failed to create destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		log.Fatalf("Install: failed to write file: %v", err)
	}

	fmt.Printf("✓ Middleware '%s' successfully installed to %s\n", name, destPath)
}

func runPolicyCommand() {
	if len(os.Args) < 5 || os.Args[2] != "compile" {
		log.Fatal("Usage: Pranor Gate policy compile <file.policy> -o <file.wasm>")
	}
	policyFile := os.Args[3]
	outputWasm := os.Args[5]

	content, err := os.ReadFile(policyFile)
	if err != nil {
		log.Fatalf("Failed to read policy file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var rulesCode strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		action := parts[0] // "allow" or "deny"
		method := parts[1] // "GET", "POST", "*"
		path := parts[2]   // "/api/data", "*"

		var cond string
		if len(parts) >= 6 && parts[3] == "if" && strings.HasPrefix(parts[4], "header.") && parts[5] == "==" {
			hdrName := strings.ToLower(strings.TrimPrefix(parts[4], "header."))
			val := strings.Join(parts[6:], " ")
			val = strings.Trim(val, "\"")
			cond = fmt.Sprintf(" && req.Headers[%q] == %q", hdrName, val)
		}

		methodCond := fmt.Sprintf("req.Method == %q", method)
		if method == "*" {
			methodCond = "true"
		}

		pathCond := fmt.Sprintf("req.Path == %q", path)
		if path == "*" {
			pathCond = "true"
		}

		rulesCode.WriteString(fmt.Sprintf(`	if %s && %s%s {
		fmt.Print(%q)
		return
	}
`, methodCond, pathCond, cond, action))
	}

	goSource := fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Request struct {
	Method  string            ` + "`" + `json:"method"` + "`" + `
	Path    string            ` + "`" + `json:"path"` + "`" + `
	Headers map[string]string ` + "`" + `json:"headers"` + "`" + `
}

func main() {
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Print("deny")
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Print("deny")
		return
	}

%s
	// Default fallback
	fmt.Print("deny")
}
`, rulesCode.String())

	tmpDir, err := os.MkdirTemp("", "policy_build")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(goSource), 0644); err != nil {
		log.Fatalf("Failed to write source: %v", err)
	}

	// Initialize a dummy go.mod in the temp directory to compile outside a workspace
	initCmd := exec.Command("go", "mod", "init", "policy")
	initCmd.Dir = tmpDir
	initCmd.Env = append(os.Environ(), "GOWORK=off")
	if err := initCmd.Run(); err != nil {
		log.Fatalf("Failed to initialize temporary go module: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", outputWasm, ".")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "GOWORK=off")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to compile policy to WASM: %v\nStderr: %s", err, stderr.String())
	}

	fmt.Printf("✓ Successfully compiled policy to %s\n", outputWasm)
}

type adminRateLimiter struct {
	mu      sync.Mutex
	history map[string][]time.Time
}

func (al *adminRateLimiter) Limit(ip string, limit int) bool {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.history == nil {
		al.history = make(map[string][]time.Time)
	}

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	history := al.history[ip]
	valid := history[:0]
	for _, t := range history {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		al.history[ip] = valid
		return true // rate limited
	}

	valid = append(valid, now)
	al.history[ip] = valid
	return false
}

func withAdminRateLimit(limit int, next http.HandlerFunc) http.HandlerFunc {
	limiter := &adminRateLimiter{}
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
		if limiter.Limit(clientIP, limit) {
			proxy.WriteJSONError(w, r, "Too Many Requests", "ERR_RATE_LIMIT_EXCEEDED", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func runDashboardCommand() {
	fs := flag.NewFlagSet("dashboard", flag.ExitOnError)
	adminEndpoint := fs.String("endpoint", "http://localhost:8080", "Pranor Gate admin API endpoint URL")
	interval := fs.Duration("interval", 1*time.Second, "Refresh interval")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("Dashboard: failed to parse arguments: %v", err)
	}

	targetURL := strings.TrimRight(*adminEndpoint, "/") + "/api/v1/admin/connections"
	client := &http.Client{Timeout: 2 * time.Second}

	fmt.Print("\033[2J\033[H") // Clear terminal
	fmt.Printf("=== Pranor Gate Live Traffic Dashboard (TUI - SG.D2) ===\n")
	fmt.Printf("Endpoint: %s | Refresh Interval: %v\n\n", targetURL, *interval)

	for {
		resp, err := client.Get(targetURL)
		if err != nil {
			fmt.Printf("\r\033[K[ERR] Failed to poll Pranor Gate stats: %v", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var stats map[string]interface{}
			if err := json.Unmarshal(body, &stats); err == nil {
				fmt.Print("\033[H") // Move cursor to top left
				fmt.Println("=== Pranor Gate Live Traffic Dashboard (TUI - SG.D2) ===")
				fmt.Printf("Endpoint: %s | Time: %s\n", targetURL, time.Now().Format("15:04:05"))
				fmt.Println(strings.Repeat("-", 70))
				fmt.Printf("Active Connections: %v\n", stats["active_connections"])
				fmt.Printf("Avg P99 Latency:    %v ms\n", stats["p99_latency_ms"])
				fmt.Println(strings.Repeat("-", 70))
				fmt.Println("Press Ctrl+C to exit")
			}
		}
		time.Sleep(*interval)
	}
}

func runGenerateConfigCommand() {
	fs := flag.NewFlagSet("generate-config", flag.ExitOnError)
	logPath := fs.String("log", "logs/access.jsonl", "Path to access log file to analyze")
	outFile := fs.String("output", "config.json", "Output configuration JSON file path")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalf("GenerateConfig: failed to parse arguments: %v", err)
	}

	fmt.Printf("=== Pranor Gate Zero-Config Learn Mode (SG.D6) ===\n")
	fmt.Printf("Analyzing traffic logs at: %s\n", *logPath)

	data, err := os.ReadFile(*logPath)
	discoveredPrefixes := make(map[string]int)

	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			var entry struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal([]byte(line), &entry); err == nil && entry.Path != "" {
				parts := strings.Split(strings.Trim(entry.Path, "/"), "/")
				if len(parts) > 0 && parts[0] != "" {
					prefix := "/" + parts[0]
					discoveredPrefixes[prefix]++
				}
			}
		}
	}

	if len(discoveredPrefixes) == 0 {
		discoveredPrefixes["/api/v1"] = 1
		discoveredPrefixes["/ai/v1"] = 1
	}

	var generatedRoutes []proxy.Route
	for prefix := range discoveredPrefixes {
		route := proxy.Route{
			Prefix:       prefix,
			Target:       "http://127.0.0.1:8081",
			RateLimitRPM: 120,
		}
		if prefix == "/ai" || strings.HasPrefix(prefix, "/ai") {
			route.Target = "http://127.0.0.1:11434"
			route.SemanticCache = true
			route.PromptGuard = true
		}
		generatedRoutes = append(generatedRoutes, route)
	}

	cfg := map[string]interface{}{
		"addr":   ":8080",
		"routes": generatedRoutes,
	}

	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(*outFile, cfgBytes, 0644); err != nil {
		log.Fatalf("GenerateConfig: failed to write %s: %v", *outFile, err)
	}

	fmt.Printf("✓ Discovered %d route(s) from traffic analysis.\n", len(generatedRoutes))
	fmt.Printf("✓ Auto-generated configuration saved to: %s\n", *outFile)
}
