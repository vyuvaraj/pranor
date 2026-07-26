package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vyuvaraj/serv/packages/ServGate/pkg/daemon"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to ServGateway configuration file")
	port := flag.String("port", ":8080", "Gateway listening port")
	adminPort := flag.String("admin-port", ":8081", "Admin UI listening port")
	version := flag.Bool("version", false, "Show servgatewayd version")
	flag.Parse()

	if *version {
		fmt.Println("servgatewayd v2.0.0 (Standalone Gateway Daemon)")
		os.Exit(0)
	}

	log.Printf("Starting ServGateway Standalone Daemon (servgatewayd v2.0.0)...")
	d, err := daemon.NewServGatewayDaemon(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize servgatewayd: %v", err)
	}

	_ = port
	_ = adminPort

	go daemon.RunDaemonSignalHandler(d)

	if err := d.Start(); err != nil {
		log.Printf("servgatewayd stopped: %v", err)
	}
}
