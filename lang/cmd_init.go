package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func initProject() {
	name := "my-service"
	if len(os.Args) >= 3 {
		name = os.Args[2]
	}

	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// main.pnr
	mainSrv := `server "8080"

// Path parameter: curl http://localhost:8080/api/hello/Alice
route "GET" "/api/hello/:name" (req) {
    let name = req.params.name
    return { "message": f"Hello, {name}!" }
}

// Query parameter: curl http://localhost:8080/api/greet?name=Bob
route "GET" "/api/greet" (req) {
    let name = req.params.name
    if name == nil {
        return { "message": "Hello, world!" }
    }
    return { "message": f"Hello, {name}!" }
}
`
	if err := os.WriteFile(filepath.Join(name, "main.pnr"), []byte(mainSrv), 0644); err != nil {
		fmt.Printf("Failed to write main.pnr: %v\n", err)
		os.Exit(1)
	}

	// config.yml
	configYml := `server:
  port: "8080"

log:
  level: "info"
  format: "text"
`
	if err := os.WriteFile(filepath.Join(name, "config.yml"), []byte(configYml), 0644); err != nil {
		fmt.Printf("Failed to write config.yml: %v\n", err)
		os.Exit(1)
	}

	// test file
	testSrv := `test "health check returns ok" {
    // TODO: add your tests here
    assert true
}
`
	if err := os.WriteFile(filepath.Join(name, "main_test.pnr"), []byte(testSrv), 0644); err != nil {
		fmt.Printf("Failed to write main_test.pnr: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created project: %s/\n", name)
	fmt.Println("")
	fmt.Println("  Files:")
	fmt.Println("    main.pnr       — Your service (routes, logic)")
	fmt.Println("    main_test.pnr  — Tests")
	fmt.Println("    config.yml     — Runtime configuration")
	fmt.Println("")
	fmt.Println("  Get started:")
	fmt.Printf("    cd %s\n", name)
	fmt.Println("    pranor run main.pnr --watch")
	fmt.Println("")
	fmt.Println("  Then visit: http://localhost:8080/health")
}

// runInitFullStack generates a docker-compose.yml wiring up all Pranor services.
// Usage: pranor init --full-stack
func runInitFullStack() {
	const compose = `version: "3.9"

services:
  Pranor Gate:
    image: ghcr.io/vyuvaraj/Pranor Gate:latest
    ports: ["8080:8080"]
    environment:
      - PRANOR_MESH_URL=http://Pranor Mesh:8083
      - PRANOR_QUEUE_URL=http://Pranor Pulse:8085
      - PRANOR_STORE_URL=http://Pranor Vault:9000
    depends_on: [Pranor Mesh, Pranor Pulse, Pranor Vault]

  Pranor Mesh:
    image: ghcr.io/vyuvaraj/Pranor Mesh:latest
    ports: ["8083:8083"]

  Pranor Pulse:
    image: ghcr.io/vyuvaraj/Pranor Pulse:latest
    ports: ["8085:8085"]
    volumes: ["queue_data:/data"]

  Pranor Vault:
    image: ghcr.io/vyuvaraj/Pranor Vault:latest
    ports: ["9000:9000"]
    volumes: ["store_data:/data"]

  servdb:
    image: ghcr.io/vyuvaraj/pranor-db:latest
    ports: ["5432:5432"]
    environment:
      - POSTGRES_PASSWORD=Pranor
    volumes: ["db_data:/var/lib/postgresql/data"]

  Pranor Cache:
    image: ghcr.io/vyuvaraj/Pranor Cache:latest
    ports: ["8086:8086"]

  Pranor Chrono:
    image: ghcr.io/vyuvaraj/Pranor Chrono:latest
    ports: ["8087:8087"]
    environment:
      - PRANOR_QUEUE_URL=http://Pranor Pulse:8085

  Pranor Trace:
    image: ghcr.io/vyuvaraj/Pranor Trace:latest
    ports: ["4317:4317", "16686:16686"]

  Pranor Console:
    image: ghcr.io/vyuvaraj/Pranor Console:latest
    ports: ["8888:8888"]
    environment:
      - PRANOR_GATE_URL=http://Pranor Gate:8080
      - PRANOR_MESH_URL=http://Pranor Mesh:8083
      - PRANOR_QUEUE_URL=http://Pranor Pulse:8085
      - PRANOR_STORE_URL=http://Pranor Vault:9000
      - PRANOR_DB_URL=http://servdb:5432
      - PRANOR_CACHE_URL=http://Pranor Cache:8086
      - PRANOR_CRON_URL=http://Pranor Chrono:8087
      - PRANOR_TRACE_URL=http://Pranor Trace:4317
    depends_on: [Pranor Gate, Pranor Mesh, Pranor Pulse, Pranor Vault, servdb, Pranor Cache, Pranor Chrono, Pranor Trace]

volumes:
  queue_data:
  store_data:
  db_data:
`
	outPath := "docker-compose.yml"
	if err := os.WriteFile(outPath, []byte(compose), 0644); err != nil {
		fmt.Printf("Failed to write docker-compose.yml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generated docker-compose.yml with all Pranor services.")
	fmt.Println("")
	fmt.Println("  Start the full stack:")
	fmt.Println("    docker compose up -d")
	fmt.Println("")
	fmt.Println("  Pranor Console dashboard → http://localhost:8888")
	fmt.Println("  Pranor Gate API gateway  → http://localhost:8080")
	fmt.Println("  Pranor Trace UI          → http://localhost:16686")
}

