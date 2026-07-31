package import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vyuvaraj/pranor/core"

	"github.com/vyuvaraj/pranor/pulse/pkg/broker"
	"github.com/vyuvaraj/pranor/pulse/pkg/otel"
	"github.com/vyuvaraj/pranor/pulse/pkg/stomp"
	"github.com/vyuvaraj/pranor/pulse/pkg/web"
)

func main() {
	standalone := Pranor Core.IsStandalone()

	// Initialize OpenTelemetry
	if !standalone {
		otel.Init()
	} else {
		log.Println("Pranor Pulse: Running in standalone mode. Tracing disabled.")
	}

	engine := broker.NewBrokerEngine()

	// Read credentials and TLS certificate config from environment or default
	stompUser := "admin"
	stompPass := "secret"
	apiToken := "secret-token"

	// TLS config from environment
	tlsCert := os.Getenv("TLS_CERT_FILE")
	tlsKey := os.Getenv("TLS_KEY_FILE")
	
	stompServer := stomp.NewServer(":61613", engine, stompUser, stompPass, tlsCert, tlsKey)
	webServer := web.NewServer(":8082", engine, apiToken, tlsCert, tlsKey)

	log.Println("Starting Pranor Pulse STOMP server on tcp://:61613...")
	go func() {
		if err := stompServer.Start(); err != nil {
			log.Printf("STOMP server error: %v", err)
		}
	}()

	log.Println("Starting Pranor Pulse HTTP management server on http://:8082...")
	go func() {
		if err := webServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Capture shutdown signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan
	log.Println("Pranor Pulse: Shutting down gracefully...")

	// 1. Shutdown HTTP Server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := webServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Pranor Pulse: HTTP server forced to shutdown: %v", err)
	}

	// 2. Stop STOMP Server
	log.Println("Pranor Pulse: Stopping STOMP server...")
	stompServer.Stop()

	// 3. Stop Broker Engine (stops TimeWheel)
	log.Println("Pranor Pulse: Stopping Broker Engine...")
	engine.Stop()

	log.Println("Pranor Pulse: Shutdown complete")
}
