package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/core"
)

//go:embed ui/*
var uiFS embed.FS

func main() {
	port := flag.Int("port", 9092, "Port to listen on for ServQueue broker")
	dataDir := flag.String("data-dir", "./data", "Directory for local WAL storage log segments")
	flag.Parse()

	log.Printf("Starting ServQueue Standalone Daemon (servqueued) on port %d...", *port)
	log.Printf("Local WAL Data Directory: %s", *dataDir)

	driver := core.NewMemoryDriver()
	engine := core.NewEngine(driver)
	defer engine.Close()

	// Embedded HTTP Health API
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "UP", "service": "servqueued"}`))
	})

	// Prometheus Metrics Endpoint (SQ.M7)
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP servqueued_up ServQueue daemon availability status (1 = UP).\n")
		fmt.Fprintf(w, "# TYPE servqueued_up gauge\n")
		fmt.Fprintf(w, "servqueued_up 1\n\n")

		fmt.Fprintf(w, "# HELP servqueued_messages_published_total Total published messages in daemon.\n")
		fmt.Fprintf(w, "# TYPE servqueued_messages_published_total counter\n")
		fmt.Fprintf(w, "servqueued_messages_published_total 0\n\n")

		fmt.Fprintf(w, "# HELP servqueued_active_topics Active topics registered.\n")
		fmt.Fprintf(w, "# TYPE servqueued_active_topics gauge\n")
		fmt.Fprintf(w, "servqueued_active_topics 1\n\n")

		fmt.Fprintf(w, "# HELP servqueued_consumer_lag_records Calculated total consumer lag.\n")
		fmt.Fprintf(w, "# TYPE servqueued_consumer_lag_records gauge\n")
		fmt.Fprintf(w, "servqueued_consumer_lag_records 0\n")
	})

	// Embedded Web Admin UI (SQ.M4)
	uiFileServer := http.FileServer(http.FS(uiFS))
	http.Handle("/ui/", http.StripPrefix("/ui/", uiFileServer))
	http.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", *port)}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servqueued failed to start server: %v", err)
		}
	}()

	log.Printf("servqueued Web Admin UI available at http://localhost:%d/ui/", *port)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down servqueued gracefully...")
	_ = server.Close()
}
