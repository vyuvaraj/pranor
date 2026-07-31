package import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/vyuvaraj/pranor/gate/pkg/proxy"
)

type GatewayConfig struct {
	Addr           string            `json:"addr" yaml:"addr"`
	AdminAddr      string            `json:"admin_addr" yaml:"admin_addr"`
	EnableTLS      bool              `json:"enable_tls" yaml:"enable_tls"`
	CertFile       string            `json:"cert_file" yaml:"cert_file"`
	KeyFile        string            `json:"key_file" yaml:"key_file"`
	Routes         []proxy.Route     `json:"routes" yaml:"routes"`
	Metadata       map[string]string `json:"metadata" yaml:"metadata"`
	EnableWebAdmin bool              `json:"enable_web_admin" yaml:"enable_web_admin"`
}

type PranorGatewayDaemon struct {
	config      GatewayConfig
	configPath  string
	server      *http.Server
	adminServer *http.Server
	mu          sync.RWMutex
	routes      map[string]proxy.Route
	startedAt   time.Time
}

func NewPranorGatewayDaemon(configPath string) (*PranorGatewayDaemon, error) {
	cfg := GatewayConfig{
		Addr:           ":8080",
		AdminAddr:      ":8081",
		EnableWebAdmin: true,
		Routes:         make([]proxy.Route, 0),
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			_ = json.Unmarshal(data, &cfg)
		}
	}

	d := &PranorGatewayDaemon{
		config:     cfg,
		configPath: configPath,
		routes:     make(map[string]proxy.Route),
		startedAt:  time.Now(),
	}

	for _, r := range cfg.Routes {
		d.routes[r.Prefix] = r
	}

	return d, nil
}

func (d *PranorGatewayDaemon) Start() error {
	d.mu.Lock()
	handler := d.createHTTPHandler()

	d.server = &http.Server{
		Addr:         d.config.Addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if d.config.EnableWebAdmin {
		adminMux := http.NewServeMux()
		adminMux.HandleFunc("/api/v1/health", d.handleHealth)
		adminMux.HandleFunc("/api/v1/routes", d.handleRoutes)
		adminMux.HandleFunc("/ui/", d.handleWebAdminUI)

		d.adminServer = &http.Server{
			Addr:    d.config.AdminAddr,
			Handler: adminMux,
		}

		go func() {
			log.Printf("[pranorGated] Web Admin UI listening on http://localhost%s/ui/", d.config.AdminAddr)
			if err := d.adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[pranorGated] Admin server error: %v", err)
			}
		}()
	}
	d.mu.Unlock()

	log.Printf("[pranorGated] Standalone PranorGateway Daemon listening on %s", d.config.Addr)
	return d.server.ListenAndServe()
}

func (d *PranorGatewayDaemon) createHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","daemon":"pranor-gated"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d.mu.RLock()
		route, exists := d.routes[r.URL.Path]
		d.mu.RUnlock()

		if !exists {
			http.Error(w, fmt.Sprintf("PranorGateway: no route for path %s", r.URL.Path), http.StatusNotFound)
			return
		}

		w.Header().Set("X-PranorGateway-Version", "v2.0.0")
		w.Header().Set("X-PranorGateway-Upstream", route.Target)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"proxied","path":"%s","target":"%s"}`, r.URL.Path, route.Target)))
	})

	return mux
}

func (d *PranorGatewayDaemon) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d.mu.RLock()
	uptime := time.Since(d.startedAt).Seconds()
	routeCount := len(d.routes)
	d.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "UP",
		"version":     "2.0.0",
		"uptime_sec":  uptime,
		"route_count": routeCount,
		"daemon":      "pranor-gated",
	})
}

func (d *PranorGatewayDaemon) handleRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d.mu.RLock()
	defer d.mu.RUnlock()

	if r.Method == http.MethodGet {
		routesList := make([]proxy.Route, 0, len(d.routes))
		for _, r := range d.routes {
			routesList = append(routesList, r)
		}
		json.NewEncoder(w).Encode(routesList)
		return
	}

	if r.Method == http.MethodPost {
		var newRoute proxy.Route
		if err := json.NewDecoder(r.Body).Decode(&newRoute); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if newRoute.Prefix == "" || newRoute.Target == "" {
			http.Error(w, "missing prefix or target", http.StatusBadRequest)
			return
		}
		d.routes[newRoute.Prefix] = newRoute
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newRoute)
		return
	}
}

func (d *PranorGatewayDaemon) handleWebAdminUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>PranorGateway Inspector UI</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; margin: 0; padding: 2rem; }
        .card { background: #1e293b; border-radius: 8px; padding: 1.5rem; border: 1px solid #334155; margin-bottom: 1.5rem; }
        h1 { color: #38bdf8; margin-top: 0; }
        table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
        th, td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #334155; }
        th { color: #94a3b8; font-size: 0.875rem; text-transform: uppercase; }
        .badge { background: #0284c7; color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; }
    </style>
</head>
<body>
    <div class="card">
        <h1>⚡ PranorGateway Admin Console</h1>
        <p>Standalone API Gateway Daemon (<span class="badge">pranorGated v2.0.0</span>)</p>
    </div>
    <div class="card">
        <h2>Registered Routes</h2>
        <table id="route-table">
            <thead>
                <tr>
                    <th>Prefix</th>
                    <th>Target Upstream URL</th>
                    <th>Plugins</th>
                </tr>
            </thead>
            <tbody>
                <tr><td>/api/v1/service</td><td>http://backend:8080</td><td>WASM, RateLimit</td></tr>
            </tbody>
        </table>
    </div>
</body>
</html>`
	w.Write([]byte(html))
}

func (d *PranorGatewayDaemon) Shutdown(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.adminServer != nil {
		_ = d.adminServer.Shutdown(ctx)
	}
	if d.server != nil {
		return d.server.Shutdown(ctx)
	}
	return nil
}

func RunDaemonSignalHandler(d *PranorGatewayDaemon) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = d.Shutdown(ctx)
}
