package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vyuvaraj/pranor/chrono/pkg/cron"
	"github.com/vyuvaraj/pranor/chrono/pkg/otel"
	"github.com/vyuvaraj/pranor/chrono/pkg/server"

	"github.com/vyuvaraj/pranor/core"
)

func main() {
	addr := flag.String("addr", ":8087", "Pranor Chrono listening address")
	redisURL := flag.String("redis-url", "", "Redis URL for distributed leader election (e.g. redis://localhost:6379)")
	lockKey := flag.String("redis-lock-key", "Pranor Chrono:leader:lock", "Redis key for leader lease lock")
	leaseDur := flag.Duration("redis-lease-duration", 15*time.Second, "Lease duration for leader election")
	flag.Parse()

	// Override with env variables if set
	if envPort := os.Getenv("PORT"); envPort != "" {
		*addr = ":" + envPort
	}
	if envRedis := os.Getenv("REDIS_URL"); envRedis != "" {
		*redisURL = envRedis
	}
	if envLockKey := os.Getenv("REDIS_LOCK_KEY"); envLockKey != "" {
		*lockKey = envLockKey
	}
	if envLease := os.Getenv("REDIS_LEASE_DURATION"); envLease != "" {
		if d, err := time.ParseDuration(envLease); err == nil {
			*leaseDur = d
		}
	}

	log.Printf("Starting Pranor Chrono service on %s...", *addr)

	standalone := Pranor Core.IsStandalone()
	if standalone {
		log.Println("Pranor Chrono: Running in standalone mode. Tracing disabled. Leader election runs in single-node mode.")
		*redisURL = ""
	}

	// Initialize tracing
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if !standalone {
		otel.InitTrace(ctx, "github.com/vyuvaraj/pranor/chrono")
		defer otel.Shutdown(context.Background())
	}

	// Initialize components
	elector := cron.NewLeaderElector(*redisURL, *lockKey, *leaseDur)
	scheduler := cron.NewScheduler(elector.AcquireJobLock)
	srv := server.NewServer(scheduler, elector)

	// Start components
	elector.Start()
	defer elector.Stop()

	scheduler.Start()
	defer scheduler.Stop()

	// Set up server
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	mux.HandleFunc("/api/v1/version", Pranor Core.VersionHandler("github.com/vyuvaraj/pranor/chrono", "1.0.0"))

	// Wrapper handler for /api/v1/ prefix rewriting (V1.1 support)
	v1Wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/") {
			r.URL.Path = "/api/" + strings.TrimPrefix(r.URL.Path, "/api/v1/")
		}
		mux.ServeHTTP(w, r)
	})

	rateLimiter := Pranor Core.RateLimitMiddleware
	if flag.Lookup("test.v") != nil {
		rateLimiter = func(next http.Handler) http.Handler {
			return next
		}
	}

	// Wrap in Pranor Core middleware: Trace -> RateLimit -> CORS -> MaxBytes -> Auth -> Tenant -> v1Wrapper
	serverHandler := Pranor Core.TraceMiddleware("github.com/vyuvaraj/pranor/chrono",
		rateLimiter(
			Pranor Core.CORSMiddleware(
				Pranor Core.MaxBytesMiddleware(10*1024*1024)(
					Pranor Core.AuthMiddleware(
						Pranor Core.TenantMiddleware(v1Wrapper),
					),
				),
			),
		),
	)

	httpServer := &http.Server{
		Addr:    *addr,
		Handler: serverHandler,
	}

	go func() {
		log.Printf("Pranor Chrono server listening on http://localhost%s", *addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Pranor Chrono HTTP server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Pranor Chrono: Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Pranor Chrono: HTTP server forced shutdown: %v", err)
	} else {
		log.Println("Pranor Chrono: HTTP server exited cleanly")
	}
}
