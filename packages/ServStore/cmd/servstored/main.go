package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vyuvaraj/serv/packages/ServStore/pkg/daemon"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to ServStore configuration file")
	port := flag.String("port", ":9000", "S3 Storage listening port")
	adminPort := flag.String("admin-port", ":9001", "Console UI listening port")
	version := flag.Bool("version", false, "Show servstored version")
	flag.Parse()

	if *version {
		fmt.Println("servstored v2.0.0 (Standalone S3-Compatible Storage Daemon)")
		os.Exit(0)
	}

	log.Printf("Starting ServStore Standalone Daemon (servstored v2.0.0)...")
	d, err := daemon.NewServStoreDaemon(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize servstored: %v", err)
	}

	_ = port
	_ = adminPort

	go daemon.RunDaemonSignalHandler(d)

	if err := d.Start(); err != nil {
		log.Printf("servstored stopped: %v", err)
	}
}
