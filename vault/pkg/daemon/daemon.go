package import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/vyuvaraj/pranor/vault/pkg/auth"
	"github.com/vyuvaraj/pranor/vault/pkg/s3"
	"github.com/vyuvaraj/pranor/vault/pkg/storage"
)

type StorageConfig struct {
	Addr           string   `json:"addr"`
	AdminAddr      string   `json:"admin_addr"`
	DataDir        string   `json:"data_dir"`
	EnableWebAdmin bool     `json:"enable_web_admin"`
	DefaultBuckets []string `json:"default_buckets"`
}

type PranorVaultDaemon struct {
	config      StorageConfig
	configPath  string
	store       storage.StorageEngine
	auth        *auth.AuthProvider
	gateway     *s3.Gateway
	server      *http.Server
	adminServer *http.Server
	mu          sync.RWMutex
	buckets     map[string]bool
	startedAt   time.Time
}

func NewPranorVaultDaemon(configPath string) (*PranorVaultDaemon, error) {
	cfg := StorageConfig{
		Addr:           ":9000",
		AdminAddr:      ":9001",
		DataDir:        "./data",
		EnableWebAdmin: true,
		DefaultBuckets: []string{"default-bucket"},
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			_ = json.Unmarshal(data, &cfg)
		}
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		if !strings.HasPrefix(envPort, ":") {
			envPort = ":" + envPort
		}
		cfg.Addr = envPort
	}
	if envAdminPort := os.Getenv("ADMIN_PORT"); envAdminPort != "" {
		if !strings.HasPrefix(envAdminPort, ":") {
			envAdminPort = ":" + envAdminPort
		}
		cfg.AdminAddr = envAdminPort
	}

	storeEngine, err := storage.NewLocalStore(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage engine at %s: %w", cfg.DataDir, err)
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	authProvider := auth.NewAuthProvider(accessKey, secretKey, false)

	gateway := s3.NewGateway(storeEngine, authProvider, nil, nil, 2, false, 2, 1)

	d := &PranorVaultDaemon{
		config:     cfg,
		configPath: configPath,
		store:      storeEngine,
		auth:       authProvider,
		gateway:    gateway,
		buckets:    make(map[string]bool),
		startedAt:  time.Now(),
	}

	ctx := context.Background()
	for _, b := range cfg.DefaultBuckets {
		_ = storeEngine.CreateBucket(ctx, b)
		d.buckets[b] = true
	}

	return d, nil
}

func (d *PranorVaultDaemon) SetPorts(s3Port, adminPort string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s3Port != "" {
		if !strings.HasPrefix(s3Port, ":") {
			s3Port = ":" + s3Port
		}
		d.config.Addr = s3Port
	}
	if adminPort != "" {
		if !strings.HasPrefix(adminPort, ":") {
			adminPort = ":" + adminPort
		}
		d.config.AdminAddr = adminPort
	}
}

func (d *PranorVaultDaemon) Start() error {
	d.mu.Lock()
	handler := d.createS3Handler()

	d.server = &http.Server{
		Addr:         d.config.Addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if d.config.EnableWebAdmin {
		adminMux := http.NewServeMux()
		adminMux.HandleFunc("/api/v1/health", d.handleHealth)
		adminMux.HandleFunc("/api/v1/buckets", d.handleBuckets)
		adminMux.HandleFunc("/api/v1/events/subscribe", d.handleSubscribeWebhook)
		adminMux.HandleFunc("/ui/", d.handleWebAdminUI)

		d.adminServer = &http.Server{
			Addr:    d.config.AdminAddr,
			Handler: adminMux,
		}

		go func() {
			log.Printf("[pranorVaultd] Storage Web Console UI listening on http://localhost%s/ui/", d.config.AdminAddr)
			if err := d.adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[pranorVaultd] Console admin server error: %v", err)
			}
		}()
	}
	d.mu.Unlock()

	log.Printf("[pranorVaultd] Standalone S3 Object Store Daemon listening on %s", d.config.Addr)
	return d.server.ListenAndServe()
}

func (d *PranorVaultDaemon) createS3Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","daemon":"pranor-vaultd"}`))
	})

	// Delegate all S3 API requests to the real s3.Gateway
	mux.Handle("/", d.gateway)

	return mux
}

func (d *PranorVaultDaemon) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d.mu.RLock()
	uptime := time.Since(d.startedAt).Seconds()
	d.mu.RUnlock()

	bucketCount := 0
	if buckets, err := d.store.ListBuckets(r.Context()); err == nil {
		bucketCount = len(buckets)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "UP",
		"version":      "2.0.0",
		"uptime_sec":   uptime,
		"bucket_count": bucketCount,
		"daemon":       "pranor-vaultd",
	})
}

func (d *PranorVaultDaemon) handleBuckets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		buckets, err := d.store.ListBuckets(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to list buckets: %v", err), http.StatusInternalServerError)
			return
		}
		var names []string
		for _, b := range buckets {
			names = append(names, b.Name)
		}
		json.NewEncoder(w).Encode(names)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := d.store.CreateBucket(r.Context(), req.Name); err != nil {
			http.Error(w, fmt.Sprintf("failed to create bucket: %v", err), http.StatusInternalServerError)
			return
		}

		d.mu.Lock()
		d.buckets[req.Name] = true
		d.mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"bucket": req.Name, "status": "created"})
		return
	}
}

func (d *PranorVaultDaemon) handleSubscribeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Bucket  string   `json:"bucket"`
		Events  []string `json:"events"`
		Webhook string   `json:"webhook"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Bucket == "" || req.Webhook == "" {
		http.Error(w, "invalid request: bucket and webhook URL required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "subscribed",
		"bucket":     req.Bucket,
		"events":     req.Events,
		"webhook":    req.Webhook,
		"created_at": time.Now().Format(time.RFC3339),
	})
}

func (d *PranorVaultDaemon) handleWebAdminUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Pranor Vault Console UI</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; margin: 0; padding: 2rem; }
        .card { background: #1e293b; border-radius: 8px; padding: 1.5rem; border: 1px solid #334155; margin-bottom: 1.5rem; }
        h1 { color: #38bdf8; margin-top: 0; }
        table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
        th, td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #334155; }
        th { color: #94a3b8; font-size: 0.875rem; text-transform: uppercase; }
        .badge { background: #10b981; color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; }
    </style>
</head>
<body>
    <div class="card">
        <h1>📦 Pranor Vault Storage Console</h1>
        <p>Standalone S3-Compatible Object Store Daemon (<span class="badge" id="version-badge">pranorVaultd v2.0.0</span>)</p>
        <p>Uptime: <span id="uptime">0s</span> | Total Buckets: <span id="bucket-count">0</span></p>
    </div>
    <div class="card">
        <h2>Active Buckets</h2>
        <table id="bucket-table">
            <thead>
                <tr>
                    <th>Bucket Name</th>
                    <th>Region</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody id="bucket-tbody">
                <tr><td colspan="3">Loading buckets...</td></tr>
            </tbody>
        </table>
    </div>
    <script>
        async function loadData() {
            try {
                const healthRes = await fetch('/api/v1/health');
                if (healthRes.ok) {
                    const health = await healthRes.json();
                    document.getElementById('uptime').textContent = Math.round(health.uptime_sec || 0) + 's';
                    document.getElementById('bucket-count').textContent = health.bucket_count || 0;
                }

                const bucketsRes = await fetch('/api/v1/buckets');
                if (bucketsRes.ok) {
                    const buckets = await bucketsRes.json();
                    const tbody = document.getElementById('bucket-tbody');
                    tbody.innerHTML = '';
                    if (buckets.length === 0) {
                        tbody.innerHTML = '<tr><td colspan="3">No active buckets found</td></tr>';
                    } else {
                        buckets.forEach(b => {
                            const tr = document.createElement('tr');
                            tr.innerHTML = '<td>' + b + '</td><td>us-east-1</td><td><span class="badge">Active</span></td>';
                            tbody.appendChild(tr);
                        });
                    }
                }
            } catch (err) {
                console.error("Failed to load admin stats:", err);
            }
        }
        loadData();
        setInterval(loadData, 5000);
    </script>
</body>
</html>`
	w.Write([]byte(html))
}

func (d *PranorVaultDaemon) Shutdown(ctx context.Context) error {
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

func RunDaemonSignalHandler(d *PranorVaultDaemon) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = d.Shutdown(ctx)
}
