# Standalone Deployment

Run individual Pranor modules as native binaries without containers.

## Build from Source

Each module can be built independently:

```bash
cd pranor/gate && go build -o pranor-gate .
cd pranor/vault && go build -o pranor-vault .
cd pranor/pulse && go build -o pranor-pulse .
cd pranor/trace && go build -o pranor-trace .
cd pranor/auth && go build -o pranor-auth .
cd pranor/cache && go build -o pranor-cache .
```

## Run

```bash
# Start tracing first (other modules send traces here)
./pranor-trace --port 8090 &

# Start object storage
./pranor-vault --port 8081 --data-dir ./data &

# Start API gateway
./pranor-gate --port 8080 --config config.json &

# Start message broker
./pranor-pulse --port 8082 &
```

## Unified Binary (pranord)

Run all modules in a single process:

```bash
cd pranor/platform
go build -o pranord .
./pranord --modules gate,vault,pulse,trace,auth,cache
```

## systemd Service

```ini
[Unit]
Description=Pranor Gate API Gateway
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/pranor-gate --port 8080
Restart=always
RestartSec=5
Environment=PRANOR_OTLP_ENDPOINT=http://localhost:8090

[Install]
WantedBy=multi-user.target
```

## Environment Setup

```bash
export PRANOR_HOME=/opt/pranor
export PRANOR_OTLP_ENDPOINT=http://localhost:8090
export PRANOR_DISCOVERY='{"gate":"http://localhost:8080","vault":"http://localhost:8081"}'
```

## Next Steps

- [Docker Deployment](./docker.md) — Container-based setup
- [Kubernetes](./kubernetes.md) — Production cluster deployment
