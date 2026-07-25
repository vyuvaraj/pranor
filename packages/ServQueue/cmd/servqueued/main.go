package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/core"
)

func main() {
	port := flag.Int("port", 9092, "Port to listen on for ServQueue broker")
	dataDir := flag.String("data-dir", "./data", "Directory for local WAL storage log segments")
	flag.Parse()

	log.Printf("Starting ServQueue Standalone Daemon (servqueued) on port %d...", *port)
	log.Printf("Local WAL Data Directory: %s", *dataDir)

	driver := core.NewMemoryDriver()
	engine := core.NewEngine(driver)
	defer engine.Close()

	// Embedded HTTP Health / Metrics API
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "UP", "service": "servqueued"}`))
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", *port)}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servqueued failed to start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down servqueued gracefully...")
	_ = server.Close()
}
