package import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vyuvaraj/pranor/core"

	"github.com/vyuvaraj/pranor/cache/pkg/cache"
	"github.com/vyuvaraj/pranor/cache/pkg/otel"
	"github.com/vyuvaraj/pranor/cache/pkg/server"
)

const version = "0.1.0"

// Enterprise hooks (overridden/initialized in ssl_offload files)
var EnterpriseListenAndServeTLS func(srv *http.Server, certFile, keyFile string) error

func main() {
	port := flag.Int("port", 8086, "Pranor Cache listen port")
	backend := flag.String("backend", "memory", "Cache backend: memory or redis")
	redisURL := flag.String("redis-url", "redis://localhost:6379", "Redis connection URL")
	verFlag := flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *verFlag {
		fmt.Printf("Pranor Cache v%s\n", version)
		return
	}

	standalone := Pranor Core.IsStandalone()

	// Initialize OpenTelemetry
	if !standalone {
		otel.Init()
	} else {
		log.Println("Pranor Cache: Running in standalone mode. Tracing disabled.")
		*backend = "memory"
	}

	log.Printf("Starting Pranor Cache v%s...", version)
	log.Printf("Backend engine: %s", *backend)

	var c cache.Cache
	var err error

	if *backend == "redis" {
		log.Printf("Connecting to Redis at: %s", *redisURL)
		c, err = cache.NewRedisCache(*redisURL)
		if err != nil {
			log.Fatalf("Failed to initialize Redis backend: %v", err)
		}
	} else {
		log.Println("Initializing local in-memory backend")
		c = cache.NewInMemoryCache(30 * time.Second)
	}

	srv := server.NewServer(c)
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: srv.Handler(),
	}

	go func() {
		certFile := os.Getenv("PRANOR_CACHE_TLS_CERT")
		keyFile := os.Getenv("PRANOR_CACHE_TLS_KEY")
		if certFile != "" && keyFile != "" {
			log.Printf("Pranor Cache listening with TLS on port %d", *port)
			if err := EnterpriseListenAndServeTLS(httpSrv, certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server TLS listen failed: %v", err)
			}
		} else {
			log.Printf("Pranor Cache listening on http://localhost:%d", *port)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server listen failed: %v", err)
			}
		}
	}()

	// Graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	log.Println("Shutting down Pranor Cache gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Forced shutdown error: %v", err)
	}

	log.Println("Pranor Cache shutdown complete.")
}
