package daemon

import (
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

	"github.com/vyuvaraj/serv/packages/ServStore/pkg/auth"
	"github.com/vyuvaraj/serv/packages/ServStore/pkg/s3"
	"github.com/vyuvaraj/serv/packages/ServStore/pkg/storage"
)

type StorageConfig struct {
	Addr           string   `json:"addr"`
	AdminAddr      string   `json:"admin_addr"`
	DataDir        string   `json:"data_dir"`
	EnableWebAdmin bool     `json:"enable_web_admin"`
	DefaultBuckets []string `json:"default_buckets"`
}

type ServStoreDaemon struct {
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

func NewServStoreDaemon(configPath string) (*ServStoreDaemon, error) {
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

	d := &ServStoreDaemon{
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

func (d *ServStoreDaemon) Start() error {
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
		adminMux.HandleFunc("/ui/", d.handleWebAdminUI)

		d.adminServer = &http.Server{
			Addr:    d.config.AdminAddr,
			Handler: adminMux,
		}

		go func() {
			log.Printf("[servstored] Storage Web Console UI listening on http://localhost%s/ui/", d.config.AdminAddr)
			if err := d.adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[servstored] Console admin server error: %v", err)
			}
		}()
	}
	d.mu.Unlock()

	log.Printf("[servstored] Standalone S3 Object Store Daemon listening on %s", d.config.Addr)
	return d.server.ListenAndServe()
}

func (d *ServStoreDaemon) createS3Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","daemon":"servstored"}`))
	})

	// Delegate all S3 API requests to the real s3.Gateway
	mux.Handle("/", d.gateway)

	return mux
}

func (d *ServStoreDaemon) handleHealth(w http.ResponseWriter, r *http.Request) {
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
		"daemon":       "servstored",
	})
}

func (d *ServStoreDaemon) handleBuckets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		buckets, err := d.store.ListBuckets(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to list buckets: %v", err), http.StatusInternalServerError)
			return
		}
		bucketNames := make([]string, 0, len(buckets))
		for _, b := range buckets {
			bucketNames = append(bucketNames, b.Name)
		}
		json.NewEncoder(w).Encode(bucketNames)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "missing bucket name", http.StatusBadRequest)
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

func (d *ServStoreDaemon) handleWebAdminUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>ServStore Console UI</title>
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
        <h1>📦 ServStore Storage Console</h1>
        <p>Standalone S3-Compatible Object Store Daemon (<span class="badge">servstored v2.0.0</span>)</p>
    </div>
    <div class="card">
        <h2>Active Buckets</h2>
        <table id="bucket-table">
            <thead>
                <tr>
                    <th>Bucket Name</th>
                    <th>Region</th>
                    <th>Storage Usage</th>
                    <th>WORM Status</th>
                </tr>
            </thead>
            <tbody>
                <tr><td>default-bucket</td><td>us-east-1</td><td>12.4 GB</td><td>Disabled</td></tr>
            </tbody>
        </table>
    </div>
</body>
</html>`
	w.Write([]byte(html))
}

func (d *ServStoreDaemon) Shutdown(ctx context.Context) error {
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

func RunDaemonSignalHandler(d *ServStoreDaemon) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = d.Shutdown(ctx)
}
