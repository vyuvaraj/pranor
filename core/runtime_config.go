package core

import (
	"os"
	"strconv"
	"time"
)

// RuntimeConfig holds all configuration for a ServRuntime instance.
// Values are loaded from environment variables with sensible defaults.
//
// Environment variables:
//
//	PRANOR_MESH_ADDR      — Pranor Mesh registry address (default: http://localhost:8089)
//	PRANOR_SELF_ADDR      — This service's own HTTP address (e.g. http://localhost:8080)
//	PRANOR_HEALTH_PATH    — Health probe path registered with Pranor Mesh (default: /healthz)
//	PRANOR_HEARTBEAT_TTL  — Heartbeat interval in seconds (default: 5)
//	PRANOR_MAX_RETRIES    — Max retries for outbound calls (default: 3)
//	PRANOR_TIMEOUT_MS     — Outbound call timeout in ms (default: 2000)
//	PRANOR_BACKOFF_MS     — Initial backoff between retries in ms (default: 50)
//	PRANOR_OTEL_ENABLED   — Enable OTel tracing (default: true)
//	PRANOR_REGION         — Optional region tag for geo-aware routing
type RuntimeConfig struct {
	MeshAddr     string
	SelfAddr     string
	HealthPath   string
	HeartbeatTTL time.Duration
	MaxRetries   int
	TimeoutMs    int
	BackoffMs    int
	EnableOtel   bool
	Standalone   bool
	Region       string
}

// IsStandalone returns true if either the environment variable PRANOR_STANDALONE is "true"
// or the CLI argument list contains "--standalone".
func IsStandalone() bool {
	if os.Getenv("PRANOR_STANDALONE") == "true" {
		return true
	}
	for _, arg := range os.Args {
		if arg == "--standalone" {
			return true
		}
	}
	return false
}

// DefaultRuntimeConfig returns a RuntimeConfig populated from environment
// variables, falling back to safe defaults when vars are absent.
func DefaultRuntimeConfig() *RuntimeConfig {
	standalone := IsStandalone()
	enableOtel := getenvBool("PRANOR_OTEL_ENABLED", true)
	if standalone {
		enableOtel = false
	}

	return &RuntimeConfig{
		MeshAddr:     getenv("PRANOR_MESH_ADDR", "http://localhost:8089"),
		SelfAddr:     getenv("PRANOR_SELF_ADDR", ""),
		HealthPath:   getenv("PRANOR_HEALTH_PATH", "/healthz"),
		HeartbeatTTL: time.Duration(getenvInt("PRANOR_HEARTBEAT_TTL", 5)) * time.Second,
		MaxRetries:   getenvInt("PRANOR_MAX_RETRIES", 3),
		TimeoutMs:    getenvInt("PRANOR_TIMEOUT_MS", 2000),
		BackoffMs:    getenvInt("PRANOR_BACKOFF_MS", 50),
		EnableOtel:   enableOtel,
		Standalone:   standalone,
		Region:       getenv("PRANOR_REGION", ""),
	}
}

// Option is a functional option for configuring a RuntimeConfig.
type Option func(*RuntimeConfig)

// WithMeshAddr overrides the Pranor Mesh registry address.
func WithMeshAddr(addr string) Option { return func(c *RuntimeConfig) { c.MeshAddr = addr } }

// WithSelfAddr sets this service's own advertised address.
func WithSelfAddr(addr string) Option { return func(c *RuntimeConfig) { c.SelfAddr = addr } }

// WithHealthPath overrides the health probe path.
func WithHealthPath(path string) Option { return func(c *RuntimeConfig) { c.HealthPath = path } }

// WithHeartbeatTTL overrides the heartbeat interval.
func WithHeartbeatTTL(d time.Duration) Option {
	return func(c *RuntimeConfig) { c.HeartbeatTTL = d }
}

// WithMaxRetries overrides the max retry count for outbound calls.
func WithMaxRetries(n int) Option { return func(c *RuntimeConfig) { c.MaxRetries = n } }

// WithRegion sets the geo-region tag for routing.
func WithRegion(region string) Option { return func(c *RuntimeConfig) { c.Region = region } }

// WithOtel enables or disables OTel tracing.
func WithOtel(enabled bool) Option { return func(c *RuntimeConfig) { c.EnableOtel = enabled } }

// --- helpers ---------------------------------------------------------------

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
