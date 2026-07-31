# Kubernetes Deployment

Deploy Pranor modules to Kubernetes using Helm charts or raw manifests.

## Helm Chart (Recommended)

```bash
helm repo add pranor https://vyuvaraj.github.io/pranor/charts
helm install pranor-vault pranor/pranor-vault --namespace pranor --create-namespace
helm install pranor-gate pranor/pranor-gate --namespace pranor
helm install pranor-pulse pranor/pranor-pulse --namespace pranor
```

## Minimal Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pranor-gate
  namespace: pranor
spec:
  replicas: 2
  selector:
    matchLabels:
      app: pranor-gate
  template:
    metadata:
      labels:
        app: pranor-gate
    spec:
      containers:
      - name: pranor-gate
        image: ghcr.io/vyuvaraj/pranor-gate:latest
        ports:
        - containerPort: 8080
        env:
        - name: PRANOR_OTLP_ENDPOINT
          value: "http://pranor-trace:8090"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: pranor-gate
  namespace: pranor
spec:
  selector:
    app: pranor-gate
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

## KEDA Auto-Scaling (Pranor Pulse)

Scale consumers based on message queue lag:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: pranor-pulse-consumer
spec:
  scaleTargetRef:
    name: order-processor
  minReplicaCount: 1
  maxReplicaCount: 10
  triggers:
  - type: external
    metadata:
      scalerAddress: pranor-pulse:8082
      topic: orders
      consumerGroup: processors
      lagThreshold: "100"
```

## Service Discovery

Set `PRANOR_DISCOVERY` as a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pranor-discovery
data:
  PRANOR_DISCOVERY: |
    {
      "gate": "http://pranor-gate:8080",
      "vault": "http://pranor-vault:8081",
      "pulse": "http://pranor-pulse:8082",
      "cache": "http://pranor-cache:8086",
      "trace": "http://pranor-trace:8090",
      "auth": "http://pranor-auth:8098"
    }
```

## Next Steps

- [Docker Deployment](./docker.md) — Local/staging setup
- [Standalone Binaries](./standalone.md) — No containers needed
- [Security Model](../architecture/security.md) — mTLS between modules
