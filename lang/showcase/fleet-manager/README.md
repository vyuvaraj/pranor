# IoT Fleet Management & Telemetry System Showcase

This showcase demonstrates a production-grade **Distributed IoT Fleet Management & Telemetry System** that integrates all core components of the **Pranor** ecosystem:

1. **Pranor Gate (API Gateway)**: Secure entrance proxying authenticated requests, rate-limiting API clients, and forwarding dashboard streams.
2. **Pranor (DSL)**: Compiles clean routing models, async sub/pub events, and scheduled maintenance rollups.
3. **Pranor Pulse (Event Broker)**: Handles the high-throughput `telemetry.ingest` pipeline.
4. **Pranor Vault (Storage)**: Records node states in SQLite and caches active sessions in Redis.
5. **Pranor Mesh (Service Mesh)**: Provides service-to-service routing and mTLS tunnel protection.
6. **Pranor Deploy (Orchestrator)**: Manages and deploys simulators and exporters.
7. **Pranor Hub (Package Registry)**: Hosts and verifies signed firmware packages.

## Architecture

```
                       +-----------------------+
                       |   Pranor Gate (Port 8080)|
                       +-----------+-----------+
                                   | (Reverse Proxy)
                                   v
                       +-----------+-----------+
                       | Fleet API (Port 4500) |
                       +-----+-----------+-----+
                             |           |
               +-------------+           +-------------+
               | (Pub)                                 | (Resolve)
               v                                       v
    +----------+----------+                 +----------+----------+
    |      Pranor Pulse      |                 |       Pranor Mesh      |
    | (telemetry.ingest)  |                 | (Service Discovery) |
    +----------+----------+                 +---------------------+
               |
               v (Sub)
    +----------+----------+                 +---------------------+
    |  Telemetry Workers  |                 |     Pranor Hub    |
    | (Process & Save)    |                 |  (Firmware Packages)|
    +----------+----------+                 +---------------------+
               |
               v (Write)
    +----------+----------+
    |      Pranor Vault      |
    | (SQLite / Redis)    |
    +---------------------+
```

## Running the Showcase

### 1. Start the Fleet API Service
```bash
pranor run main.pnr --watch
```

### 2. Start the Pranor Gate API Gateway
```bash
pranor-gate --config gateway.json
```

Open `http://localhost:8080/dashboard` in your browser to view the interactive real-time visualizer!
