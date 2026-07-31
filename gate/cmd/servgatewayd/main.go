package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vyuvaraj/pranor/gate/pkg/daemon"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to PranorGateway configuration file")
	port := flag.String("port", ":8080", "Gateway listening port")
	adminPort := flag.String("admin-port", ":8081", "Admin UI listening port")
	version := flag.Bool("version", false, "Show pranor-gated version")
	flag.Parse()

	if *version {
		fmt.Println("pranor-gated v2.0.0 (Standalone Gateway Daemon)")
		os.Exit(0)
	}

	log.Printf("Starting PranorGateway Standalone Daemon (pranor-gated v2.0.0)...")
	d, err := daemon.NewPranorGatewayDaemon(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize pranor-gated: %v", err)
	}

	_ = port
	_ = adminPort

	go daemon.RunDaemonSignalHandler(d)

	if err := d.Start(); err != nil {
		log.Printf("pranor-gated stopped: %v", err)
	}
}
