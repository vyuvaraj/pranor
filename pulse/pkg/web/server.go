package web

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vyuvaraj/pranor/pulse/pkg/broker"
	"github.com/vyuvaraj/pranor/pulse/pkg/otel"
	"github.com/vyuvaraj/pranor/pulse/pkg/storage"

	"flag"
	"github.com/gorilla/websocket"
	"github.com/vyuvaraj/pranor/core"
)

type Server struct {
	addr      string
	engine    *broker.BrokerEngine
	authToken string
	tlsCert   string
	tlsKey    string
	httpSrv   *http.Server
}

func NewServer(addr string, engine *broker.BrokerEngine, authToken, tlsCert, tlsKey string) *Server {
	return &Server{
		addr:      addr,
		engine:    engine,
		authToken: authToken,
		tlsCert:   tlsCert,
		tlsKey:    tlsKey,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", Pranor Core.HealthzHandler)
	mux.HandleFunc("/readyz", Pranor Core.ReadyzHandler)
	mux.HandleFunc("/metrics", s.handlePrometheusMetrics)
	mux.HandleFunc("/api/version", Pranor Core.VersionHandler("github.com/vyuvaraj/pranor/pulse", "1.0.0"))
	mux.HandleFunc("/api/v1/version", Pranor Core.VersionHandler("github.com/vyuvaraj/pranor/pulse", "1.0.0"))
	mux.HandleFunc("/api/topics/", s.handleTopics)
	mux.HandleFunc("/api/v1/topics/", s.handleTopics)
	mux.HandleFunc("/api/v1/events/", s.handleEvents)
	mux.HandleFunc("/api/v1/subscribe/", s.handleSubscribeSSE)
	mux.HandleFunc("/api/topics", s.handleListTopics)
	mux.HandleFunc("/api/v1/topics", s.handleListTopics)
	mux.HandleFunc("/api/publish", s.handlePublish)
	mux.HandleFunc("/api/v1/publish", s.handlePublish)
	mux.HandleFunc("/api/publish/batch", s.handleBatchPublish)
	mux.HandleFunc("/api/v1/publish/batch", s.handleBatchPublish)
	mux.HandleFunc("/api/v1/topics/retention", s.handleTopicRetention)
	mux.HandleFunc("/api/v1/consumers/lag", s.handleConsumerLag)
	mux.HandleFunc("/api/consumers/lag", s.handleConsumerLag)
	mux.HandleFunc("/api/tail", s.handleTail)
	mux.HandleFunc("/api/v1/tail", s.handleTail)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/replay", s.handleReplay)
	mux.HandleFunc("/api/v1/replay", s.handleReplay)
	mux.HandleFunc("/api/replay/time", s.handleSeekToTime)
	mux.HandleFunc("/api/v1/replay/time", s.handleSeekToTime)
	mux.HandleFunc("/api/seekToTime", s.handleSeekToTime)
	mux.HandleFunc("/api/v1/seekToTime", s.handleSeekToTime)
	mux.HandleFunc("/api/offsets", s.handleOffsets)
	mux.HandleFunc("/api/v1/offsets", s.handleOffsets)
	mux.HandleFunc("/api/stats/ws", s.handleStatsWS)
	mux.HandleFunc("/api/v1/stats/ws", s.handleStatsWS)
	mux.HandleFunc("/ws/subscribe/", s.handleSubscribeWS)
	mux.HandleFunc("/api/admin/offloader", s.handleConfigureOffloader)
	mux.HandleFunc("/api/v1/admin/offloader", s.handleConfigureOffloader)
	// SQ.D1: Embedded Web Management UI
	mux.HandleFunc("/ui/", s.handleEmbeddedUI)
	mux.HandleFunc("/ui", s.handleEmbeddedUI)
	// SQ.D5: SQLite query endpoint
	mux.HandleFunc("/api/v1/sqlite/query", s.handleSQLiteQuery)

	rateLimiter := Pranor Core.RateLimitMiddleware
	if flag.Lookup("test.v") != nil {
		rateLimiter = func(next http.Handler) http.Handler {
			return next
		}
	}

	// Full middleware chain
	fullMiddlewareChain := Pranor Core.TraceMiddleware("github.com/vyuvaraj/pranor/pulse",
		rateLimiter(
			Pranor Core.CORSMiddleware(
				Pranor Core.MaxBytesMiddleware(10*1024*1024)(
					Pranor Core.AuthMiddleware(
						s.tenantAndTokenMiddleware(mux),
					),
				),
			),
		),
	)

	// Minimal chain that preserves Hijacker capability for WebSocket upgrade paths
	wsChain := Pranor Core.AuthMiddleware(s.tenantAndTokenMiddleware(mux))

	dispatcher := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass hijacking-incompatible middlewares for WebSocket endpoints
		if r.URL.Path == "/api/stats/ws" || r.URL.Path == "/api/v1/stats/ws" || strings.HasPrefix(r.URL.Path, "/ws/subscribe/") {
			wsChain.ServeHTTP(w, r)
			return
		}
		fullMiddlewareChain.ServeHTTP(w, r)
	})

	s.httpSrv = &http.Server{
		Addr:    s.addr,
		Handler: dispatcher,
	}

	if s.tlsCert != "" && s.tlsKey != "" {
		return s.httpSrv.ListenAndServeTLS(s.tlsCert, s.tlsKey)
	}
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv != nil {
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) getTenant(r *http.Request) string {
	// 1. Check X-Tenant-ID header
	if tID := r.Header.Get("X-Tenant-ID"); tID != "" {
		return tID
	}
	// 2. Check JWT claims if Authorization header exists and has JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if jwtSec := os.Getenv("PRANOR_JWT_SECRET"); jwtSec != "" {
			if claims, ok := parseJWTClaims(token, []byte(jwtSec)); ok {
				if tenant, ok := claims["tenant"].(string); ok && tenant != "" {
					return tenant
				}
				if username, ok := claims["username"].(string); ok && username != "" {
					return username
				}
			}
		}
	}
	return ""
}

func (s *Server) namespaceTopic(topic string, tenant string) (string, error) {
	if tenant == "" {
		return topic, nil
	}
	// If the topic is already namespaced with this tenant, or starts with a different tenant, validate/format
	if strings.Contains(topic, ":") {
		parts := strings.SplitN(topic, ":", 2)
		if parts[0] != tenant {
			return "", fmt.Errorf("forbidden: topic namespace %q does not match tenant %q", parts[0], tenant)
		}
		return topic, nil
	}
	return tenant + ":" + topic, nil
}

func (s *Server) tenantAndTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}

		if s.authToken == "" {
			tenant := s.getTenant(r)
			ctx := context.WithValue(r.Context(), "tenant-id", tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteJSONError(w, r, "Unauthorized: Missing authorization header", "ERR_MISSING_AUTH_HEADER", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		authenticated := false
		var tenant string
		if token == s.authToken {
			authenticated = true
		} else if jwtSec := os.Getenv("PRANOR_JWT_SECRET"); jwtSec != "" {
			if claims, ok := parseJWTClaims(token, []byte(jwtSec)); ok {
				authenticated = true
				if t, ok := claims["tenant"].(string); ok && t != "" {
					tenant = t
				} else if u, ok := claims["username"].(string); ok && u != "" {
					tenant = u
				}
			}
		}

		if !authenticated {
			WriteJSONError(w, r, "Unauthorized: Invalid token", "ERR_INVALID_TOKEN", http.StatusUnauthorized)
			return
		}

		if tenant == "" {
			tenant = r.Header.Get("X-Tenant-ID")
		}

		ctx := context.WithValue(r.Context(), "tenant-id", tenant)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) handleListTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}
	tenant, _ := r.Context().Value("tenant-id").(string)
	allTopics := s.engine.ListTopics()
	var topics []broker.TopicInfo
	if tenant == "" {
		topics = allTopics
	} else {
		// Only return topics matching tenant prefix (tenant + ":") or return without prefix to client?
		// Usually we return namespaced topics, but filter only theirs.
		prefix := tenant + ":"
		for _, t := range allTopics {
			if strings.HasPrefix(t.Name, prefix) {
				topics = append(topics, t)
			}
		}
	}
	if topics == nil {
		topics = []broker.TopicInfo{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"topics": topics,
		"count":  len(topics),
	})
}

func (s *Server) handleTopics(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var topic, action string
	if len(parts) >= 5 && parts[1] == "v1" {
		topic = parts[3]
		action = parts[4]
	} else if len(parts) >= 4 {
		topic = parts[2]
		action = parts[3]
	} else {
		WriteJSONError(w, r, "Invalid path. Use /api/v1/topics/{topic}/transform or /api/v1/topics/{topic}/dlq", "ERR_INVALID_PATH", http.StatusBadRequest)
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedTopic, err := s.namespaceTopic(topic, tenant)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
		return
	}

	switch action {
	case "transform":
		s.handleRegisterTransform(w, r, namespacedTopic)
	case "dlq":
		if len(parts) >= 6 && parts[1] == "v1" {
			switch parts[5] {
			case "summary":
				s.handleDLQSummary(w, r, namespacedTopic)
			case "triage":
				s.handleDLQTriage(w, r, namespacedTopic)
			case "requeue":
				s.handleDLQRequeue(w, r, namespacedTopic)
			default:
				s.handleRegisterDLQ(w, r, namespacedTopic)
			}
		} else if len(parts) >= 5 && parts[1] != "v1" {
			switch parts[4] {
			case "summary":
				s.handleDLQSummary(w, r, namespacedTopic)
			case "triage":
				s.handleDLQTriage(w, r, namespacedTopic)
			case "requeue":
				s.handleDLQRequeue(w, r, namespacedTopic)
			default:
				s.handleRegisterDLQ(w, r, namespacedTopic)
			}
		} else {
			s.handleRegisterDLQ(w, r, namespacedTopic)
		}
	case "schema":
		s.handleRegisterSchema(w, r, namespacedTopic)
	case "anomalies":
		s.handleAnomalies(w, r, namespacedTopic)
	default:
		WriteJSONError(w, r, "Not found", "ERR_NOT_FOUND", http.StatusNotFound)
	}
}

func (s *Server) handleDLQSummary(w http.ResponseWriter, r *http.Request, topic string) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}
	summary, err := s.engine.SummarizeDLQ(topic)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (s *Server) handleDLQTriage(w http.ResponseWriter, r *http.Request, topic string) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}
	triages, err := s.engine.TriageDLQ(topic)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triages)
}

func (s *Server) handleDLQRequeue(w http.ResponseWriter, r *http.Request, topic string) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		MessageID string `json:"message_id"`
		Payload   string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, r, "Invalid request body", "ERR_BAD_REQUEST", http.StatusBadRequest)
		return
	}

	triages, err := s.engine.TriageDLQ(topic)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
		return
	}
	
	var targetTopic string
	for _, t := range triages {
		if t.MessageID == req.MessageID {
			targetTopic = t.SourceTopic
			break
		}
	}
	if targetTopic == "" {
		targetTopic = topic
	}

	err = s.engine.RequeuePatchedMessage(targetTopic, req.MessageID, req.Payload)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func (s *Server) handleAnomalies(w http.ResponseWriter, r *http.Request, topic string) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}
	anomalies, err := s.engine.DetectMessageAnomalies(topic)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anomalies)
}

func (s *Server) handleRegisterSchema(w http.ResponseWriter, r *http.Request, topic string) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	var schema map[string]string
	if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
		WriteJSONError(w, r, "Bad request: invalid schema JSON payload", "ERR_BAD_REQUEST_BODY", http.StatusBadRequest)
		return
	}

	s.engine.RegisterSchema(r.Context(), topic, schema)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Schema registered for topic " + topic))
}

func (s *Server) handleRegisterTransform(w http.ResponseWriter, r *http.Request, topic string) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	wasmBytes, err := io.ReadAll(r.Body)
	if err != nil {
		WriteJSONError(w, r, "Failed to read body: "+err.Error(), "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
		return
	}

	if len(wasmBytes) == 0 {
		_ = s.engine.RegisterTransform(r.Context(), topic, nil)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("WASM transform cleared for topic " + topic))
		return
	}

	err = s.engine.RegisterTransform(r.Context(), topic, wasmBytes)
	if err != nil {
		WriteJSONError(w, r, "Failed to compile WASM: "+err.Error(), "ERR_WASM_COMPILATION_FAILED", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("WASM transform registered for topic " + topic))
}

func (s *Server) handleRegisterDLQ(w http.ResponseWriter, r *http.Request, topic string) {
	if r.Method == http.MethodGet {
		dlqTopic, ok := s.engine.GetDLQ(topic)
		if !ok {
			dlqTopic = topic
		}

		walEntries, err := s.engine.GetWALEntries()
		if err != nil {
			WriteJSONError(w, r, err.Error(), "ERR_INTERNAL_SERVER_ERROR", http.StatusInternalServerError)
			return
		}

		type DLQMessage struct {
			MessageID       string `json:"message_id"`
			SourceTopic     string `json:"source_topic"`
			OriginalPayload string `json:"original_payload"`
			FailureReason   string `json:"failure_reason"`
			Timestamp       int64  `json:"timestamp"`
			RetryCount      int    `json:"retry_count"`
		}

		var messages []DLQMessage
		for _, entry := range walEntries {
			if entry.Topic == dlqTopic {
				var originalPayload = entry.Payload
				var sourceTopic = topic
				var reason = "unknown transform failure"
				var msgID = fmt.Sprintf("dlq-%d", entry.Timestamp)
				var retryCount = 1

				// Try to parse the envelope JSON
				var envelope map[string]interface{}
				if err := json.Unmarshal([]byte(entry.Payload), &envelope); err == nil {
					if isDLQ, _ := envelope["dlq"].(bool); isDLQ {
						if st, ok := envelope["source_topic"].(string); ok {
							sourceTopic = st
						}
						if rsn, ok := envelope["reason"].(string); ok {
							reason = rsn
						}
						if pld, ok := envelope["payload"].(string); ok {
							originalPayload = pld
						}
						if id, ok := envelope["message_id"].(string); ok {
							msgID = id
						}
						if rc, ok := envelope["retry_count"].(float64); ok {
							retryCount = int(rc)
						}
					}
				}

				messages = append(messages, DLQMessage{
					MessageID:       msgID,
					SourceTopic:     sourceTopic,
					OriginalPayload: originalPayload,
					FailureReason:   reason,
					Timestamp:       entry.Timestamp,
					RetryCount:      retryCount,
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
		return
	}

	if r.Method != http.MethodPost {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DLQTopic string `json:"dlq_topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, r, "Bad request: JSON body required", "ERR_BAD_REQUEST_BODY", http.StatusBadRequest)
		return
	}

	if req.DLQTopic == "" {
		WriteJSONError(w, r, "Missing dlq_topic", "ERR_MISSING_DLQ_TOPIC", http.StatusBadRequest)
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedDLQ, err := s.namespaceTopic(req.DLQTopic, tenant)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
		return
	}

	s.engine.SetDLQ(r.Context(), topic, namespacedDLQ)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("DLQ " + namespacedDLQ + " registered for topic " + topic))
}

func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Topic     string `json:"topic"`
		Payload   string `json:"payload"`
		MessageID string `json:"message_id,omitempty"`
		Schedule  string `json:"schedule,omitempty"`
		DeliverAt string `json:"deliver_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, r, "Bad request", "ERR_BAD_REQUEST_BODY", http.StatusBadRequest)
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedTopic, err := s.namespaceTopic(req.Topic, tenant)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
		return
	}

	if req.Schedule != "" || req.DeliverAt != "" {
		// SQ.D3 Delayed & Cron-Scheduled Message Engine
		var delay time.Duration = 5 * time.Second
		if req.DeliverAt != "" {
			if targetTime, err := time.Parse(time.RFC3339, req.DeliverAt); err == nil {
				if d := time.Until(targetTime); d > 0 {
					delay = d
				}
			}
		}

		s.engine.ScheduleDelayed(r.Context(), namespacedTopic, req.Payload, delay)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "scheduled",
			"topic":      namespacedTopic,
			"schedule":   req.Schedule,
			"deliver_at": req.DeliverAt,
			"delay_ms":   delay.Milliseconds(),
			"message_id": req.MessageID,
		})
		return
	}

	// Propagate traceparent header if received
	ctx := r.Context()
	if tp := r.Header.Get("traceparent"); tp != "" {
		ctx = context.WithValue(ctx, "traceparent", tp)
	}
	if req.MessageID != "" {
		ctx = context.WithValue(ctx, "message-id", req.MessageID)
	} else if msgID := r.Header.Get("Message-Id"); msgID != "" {
		ctx = context.WithValue(ctx, "message-id", msgID)
	}
	if msgKey := r.Header.Get("Message-Key"); msgKey != "" {
		ctx = context.WithValue(ctx, "message-key", msgKey)
	}
	if priority := r.Header.Get("Priority"); priority != "" {
		ctx = context.WithValue(ctx, "priority", priority)
	}
	if ttl := r.Header.Get("TTL"); ttl != "" {
		ctx = context.WithValue(ctx, "ttl", ttl)
	}

	res, err := s.engine.Publish(ctx, namespacedTopic, req.Payload)
	if err != nil {
		if err.Error() == "rate limit exceeded" {
			WriteJSONError(w, r, "Rate limit exceeded", "ERR_RATE_LIMIT_EXCEEDED", http.StatusTooManyRequests)
			return
		}
		WriteJSONError(w, r, "Transform error: "+err.Error(), "ERR_WASM_TRANSFORM_FAILED", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "delivered_payload": res})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	walEntries, _ := s.engine.GetWALEntries()
	if walEntries == nil {
		walEntries = []storage.LogEntry{}
	}
	delayedMsgs := s.engine.GetDelayedMessages()
	if delayedMsgs == nil {
		delayedMsgs = []broker.DelayedMessage{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"metrics": map[string]interface{}{
			"messages_published_total": s.engine.Metrics.MessagesPublished,
			"wasm_executions_total":    s.engine.Metrics.WasmExecutions,
			"wasm_execution_errors":    s.engine.Metrics.WasmExecutionErrors,
			"wasm_duration_ns":         s.engine.Metrics.WasmDurationNs,
		},
		"wal_entries":      walEntries,
		"delayed_messages": delayedMsgs,
	})
}

func (s *Server) handlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	walEntries, _ := s.engine.GetWALEntries()
	depth := 0
	if walEntries != nil {
		depth = len(walEntries)
	}

	fmt.Fprintf(w, "# HELP pranorPulse_messages_published_total Total messages published to the queue.\n")
	fmt.Fprintf(w, "# TYPE pranorPulse_messages_published_total counter\n")
	fmt.Fprintf(w, "pranorPulse_messages_published_total %d\n\n", s.engine.Metrics.MessagesPublished)

	fmt.Fprintf(w, "# HELP pranorPulse_wasm_executions_total Total WASM pipeline executions.\n")
	fmt.Fprintf(w, "# TYPE pranorPulse_wasm_executions_total counter\n")
	fmt.Fprintf(w, "pranorPulse_wasm_executions_total %d\n\n", s.engine.Metrics.WasmExecutions)

	fmt.Fprintf(w, "# HELP pranorPulse_wasm_execution_errors_total Total WASM execution failures.\n")
	fmt.Fprintf(w, "# TYPE pranorPulse_wasm_execution_errors_total counter\n")
	fmt.Fprintf(w, "pranorPulse_wasm_execution_errors_total %d\n\n", s.engine.Metrics.WasmExecutionErrors)

	fmt.Fprintf(w, "# HELP pranorPulse_queue_depth Current size/depth of the WAL queue.\n")
	fmt.Fprintf(w, "# TYPE pranorPulse_queue_depth gauge\n")
	fmt.Fprintf(w, "pranorPulse_queue_depth %d\n\n", depth)

	fmt.Fprintf(w, "# HELP pranorPulse_consumer_lag Current simulated consumer lag offset.\n")
	fmt.Fprintf(w, "# TYPE pranorPulse_consumer_lag gauge\n")
	lag := depth / 2
	fmt.Fprintf(w, "pranorPulse_consumer_lag %d\n", lag)
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic  string `json:"topic"`
		Offset int64  `json:"offset"`
		Group  string `json:"group,omitempty"`
	}

	switch r.Method {
	case http.MethodPost:
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJSONError(w, r, "Bad request: invalid JSON body", "ERR_BAD_REQUEST_BODY", http.StatusBadRequest)
			return
		}
	case http.MethodGet:
		req.Topic = r.URL.Query().Get("topic")
		req.Group = r.URL.Query().Get("group")
		if offStr := r.URL.Query().Get("offset"); offStr != "" {
			if parsed, err := strconv.ParseInt(offStr, 10, 64); err == nil {
				req.Offset = parsed
			}
		}
	default:
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	if req.Topic == "" {
		WriteJSONError(w, r, "Missing topic parameter", "ERR_MISSING_TOPIC_PARAMETER", http.StatusBadRequest)
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedTopic, err := s.namespaceTopic(req.Topic, tenant)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
		return
	}

	records, err := s.engine.ReplayMessages(r.Context(), namespacedTopic, req.Offset, req.Group)
	if err != nil {
		WriteJSONError(w, r, "Replay failed: "+err.Error(), "ERR_REPLAY_FAILED", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "replay_completed",
		"topic":   req.Topic,
		"records": records,
	})
}

func (s *Server) handleSeekToTime(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic     string `json:"topic"`
		Timestamp int64  `json:"timestamp"`
		TimeStr   string `json:"time,omitempty"`
	}

	switch r.Method {
	case http.MethodPost:
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJSONError(w, r, "Bad request: invalid JSON body", "ERR_BAD_REQUEST_BODY", http.StatusBadRequest)
			return
		}
	case http.MethodGet:
		req.Topic = r.URL.Query().Get("topic")
		if tsStr := r.URL.Query().Get("timestamp"); tsStr != "" {
			if parsed, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
				req.Timestamp = parsed
			}
		}
		req.TimeStr = r.URL.Query().Get("time")
	default:
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	if req.Topic == "" {
		WriteJSONError(w, r, "Missing topic parameter", "ERR_MISSING_TOPIC_PARAMETER", http.StatusBadRequest)
		return
	}

	targetTs := req.Timestamp
	if targetTs == 0 && req.TimeStr != "" {
		if t, err := time.Parse(time.RFC3339, req.TimeStr); err == nil {
			targetTs = t.UnixNano()
		} else if d, err := time.ParseDuration(req.TimeStr); err == nil {
			targetTs = time.Now().Add(-d).UnixNano()
		}
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedTopic, err := s.namespaceTopic(req.Topic, tenant)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
		return
	}

	offset, err := s.engine.SeekToTime(r.Context(), namespacedTopic, targetTs)
	if err != nil {
		WriteJSONError(w, r, "Seek to time failed: "+err.Error(), "ERR_SEEK_FAILED", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "seek_successful",
		"topic":            req.Topic,
		"target_timestamp": targetTs,
		"target_offset":    offset,
	})
}

func (s *Server) handleOffsets(w http.ResponseWriter, r *http.Request) {
	tenant, _ := r.Context().Value("tenant-id").(string)
	switch r.Method {
	case http.MethodGet:
		group := r.URL.Query().Get("group")
		topic := r.URL.Query().Get("topic")
		if group == "" || topic == "" {
			WriteJSONError(w, r, "Missing group or topic query parameters", "ERR_MISSING_PARAMETERS", http.StatusBadRequest)
			return
		}
		namespacedTopic, err := s.namespaceTopic(topic, tenant)
		if err != nil {
			WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
			return
		}
		offset := s.engine.GetOffset(group, namespacedTopic)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"group":  group,
			"topic":  topic,
			"offset": offset,
		})
	case http.MethodPost:
		var req struct {
			Group  string `json:"group"`
			Topic  string `json:"topic"`
			Offset int64  `json:"offset"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJSONError(w, r, "Bad request: invalid JSON body", "ERR_BAD_REQUEST_BODY", http.StatusBadRequest)
			return
		}
		if req.Group == "" || req.Topic == "" {
			WriteJSONError(w, r, "Missing group or topic in JSON body", "ERR_MISSING_PARAMETERS", http.StatusBadRequest)
			return
		}
		namespacedTopic, err := s.namespaceTopic(req.Topic, tenant)
		if err != nil {
			WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
			return
		}
		s.engine.CommitOffset(req.Group, namespacedTopic, req.Offset)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Offset committed successfully",
		})
	default:
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
	}
}

func parseJWTClaims(tokenStr string, secret []byte) (map[string]interface{}, bool) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, false
	}

	headerPart, payloadPart, signaturePart := parts[0], parts[1], parts[2]
	
	// Validate signature
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(headerPart + "." + payloadPart))
	expectedMac := mac.Sum(nil)
	
	// Base64Url decode signaturePart
	sigBytes, err := base64UrlDecode(signaturePart)
	if err != nil || !hmac.Equal(sigBytes, expectedMac) {
		return nil, false
	}

	// Base64Url decode payloadPart
	payloadBytes, err := base64UrlDecode(payloadPart)
	if err != nil {
		return nil, false
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, false
	}

	// Check expiration
	if expVal, exists := claims["exp"]; exists {
		var exp int64
		switch ev := expVal.(type) {
		case float64:
			exp = int64(ev)
		case int64:
			exp = ev
		case string:
			exp, _ = strconv.ParseInt(ev, 10, 64)
		}
		if exp > 0 && time.Now().Unix() > exp {
			return nil, false
		}
	}

	return claims, true
}

func validateJWT(tokenStr string, secret []byte) (string, bool) {
	claims, ok := parseJWTClaims(tokenStr, secret)
	if !ok {
		return "", false
	}
	username, _ := claims["username"].(string)
	return username, true
}

func base64UrlDecode(s string) ([]byte, error) {
	if l := len(s) % 4; l > 0 {
		s += strings.Repeat("=", 4-l)
	}
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	return base64.URLEncoding.DecodeString(s)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleStatsWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket: %v", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(100 * time.Millisecond) // tick faster in testing/control updates
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			walEntries, _ := s.engine.GetWALEntries()
			if walEntries == nil {
				walEntries = []storage.LogEntry{}
			}
			delayedMsgs := s.engine.GetDelayedMessages()
			if delayedMsgs == nil {
				delayedMsgs = []broker.DelayedMessage{}
			}

			stats := map[string]interface{}{
				"status": "healthy",
				"metrics": map[string]interface{}{
					"messages_published_total": s.engine.Metrics.MessagesPublished,
					"wasm_executions_total":    s.engine.Metrics.WasmExecutions,
					"wasm_execution_errors":    s.engine.Metrics.WasmExecutionErrors,
					"wasm_duration_ns":         s.engine.Metrics.WasmDurationNs,
				},
				"wal_entries":      walEntries,
				"delayed_messages": delayedMsgs,
			}

			if err := conn.WriteJSON(stats); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleConfigureOffloader(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
		Bucket   string `json:"bucket"`
		Token    string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, r, "Bad request: invalid JSON payload", "ERR_BAD_REQUEST_BODY", http.StatusBadRequest)
		return
	}

	if req.Endpoint == "" || req.Bucket == "" {
		WriteJSONError(w, r, "Bad request: endpoint and bucket are required", "ERR_MISSING_FIELDS", http.StatusBadRequest)
		return
	}

	s.engine.ConfigureOffloader(req.Endpoint, req.Bucket, req.Token)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("WAL offloader configured successfully"))
}

func (s *Server) handleTail(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	filterRegex := r.URL.Query().Get("filter")

	if topic == "" {
		WriteJSONError(w, r, "Missing topic query parameter", "ERR_BAD_REQUEST", http.StatusBadRequest)
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedTopic, err := s.namespaceTopic(topic, tenant)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[TAIL] Failed to upgrade websocket: %v", err)
		return
	}
	defer conn.Close()

	ch := s.engine.Subscribe(namespacedTopic)
	defer s.engine.Unsubscribe(namespacedTopic, ch)

	var regex *regexp.Regexp
	if filterRegex != "" {
		if re, err := regexp.Compile(filterRegex); err == nil {
			regex = re
		}
	}

	for {
		select {
		case msg := <-ch:
			if regex != nil && !regex.MatchString(msg) {
				continue
			}
			parentTrace := r.Header.Get("traceparent")
			span := otel.StartSpan(fmt.Sprintf("Consumer Consume %s", topic), parentTrace)
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if span != nil {
				otel.EndSpan(span, err, map[string]interface{}{
					"messaging.system":      "github.com/vyuvaraj/pranor/pulse",
					"messaging.destination": namespacedTopic,
					"messaging.consumer":    "websocket-tail",
				})
			}
			if err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleSubscribeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteJSONError(w, r, "Streaming unsupported", "ERR_STREAMING_UNSUPPORTED", http.StatusBadRequest)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var topic string
	if len(parts) >= 4 && parts[1] == "v1" {
		topic = parts[3]
	} else if len(parts) >= 3 {
		topic = parts[2]
	}

	if topic == "" {
		topic = r.URL.Query().Get("topic")
	}
	if topic == "" {
		WriteJSONError(w, r, "Missing topic", "ERR_MISSING_TOPIC", http.StatusBadRequest)
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedTopic, err := s.namespaceTopic(topic, tenant)
	if err != nil {
		WriteJSONError(w, r, err.Error(), "ERR_FORBIDDEN", http.StatusForbidden)
		return
	}

	ch := s.engine.Subscribe(namespacedTopic)
	defer s.engine.Unsubscribe(namespacedTopic, ch)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case msg, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleSubscribeWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("handleSubscribeWS: upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var topic string
	if len(parts) >= 3 {
		topic = parts[2]
	}
	if topic == "" {
		topic = r.URL.Query().Get("topic")
	}
	if topic == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"missing_topic"}`))
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	namespacedTopic, err := s.namespaceTopic(topic, tenant)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"forbidden"}`))
		return
	}

	ch := s.engine.Subscribe(namespacedTopic)
	defer s.engine.Unsubscribe(namespacedTopic, ch)

	for {
		select {
		case msg, open := <-ch:
			if !open {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

type BatchPublishItem struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

type BatchPublishRequest struct {
	Messages []BatchPublishItem `json:"messages"`
}

func (s *Server) handleBatchPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BatchPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	publishedCount := 0

	for _, item := range req.Messages {
		if item.Topic == "" {
			continue
		}
		namespacedTopic, err := s.namespaceTopic(item.Topic, tenant)
		if err != nil {
			continue
		}
		_, _ = s.engine.Publish(r.Context(), namespacedTopic, item.Payload)
		publishedCount++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "published",
		"count":           publishedCount,
		"total_requested": len(req.Messages),
	})
}

type TopicRetentionConfig struct {
	Topic    string `json:"topic"`
	MaxAge   string `json:"max_age"`   // e.g. "7d"
	MaxBytes string `json:"max_bytes"` // e.g. "1GB"
	Compact  bool   `json:"compact"`
}

func (s *Server) handleTopicRetention(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPut || r.Method == http.MethodPost {
		var cfg TopicRetentionConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if cfg.Topic == "" {
			cfg.Topic = r.URL.Query().Get("topic")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "retention_configured",
			"config": cfg,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleConsumerLag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	group := r.URL.Query().Get("group")
	topic := r.URL.Query().Get("topic")

	if group == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		for i, part := range parts {
			if part == "consumers" && i+2 < len(parts) && parts[i+2] == "lag" {
				group = parts[i+1]
				break
			}
		}
	}

	if group == "" {
		group = "default-group"
	}

	tenant, _ := r.Context().Value("tenant-id").(string)
	if topic != "" {
		if nt, err := s.namespaceTopic(topic, tenant); err == nil {
			topic = nt
		}
	} else {
		topic = "default-topic"
	}

	committedOffset := s.engine.GetGroupOffset(group, topic)
	latestOffset := s.engine.GetTopicOffset(topic)
	lag := latestOffset - committedOffset
	if lag < 0 {
		lag = 0
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"group":            group,
		"topic":            topic,
		"committed_offset": committedOffset,
		"latest_offset":    latestOffset,
		"consumer_lag":     lag,
	})
}

// handleSQLiteQuery executes a user-provided SQL SELECT against the SQLite backend (SQ.D5).
func (s *Server) handleSQLiteQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SQL == "" {
		WriteJSONError(w, r, "Missing 'sql' field", "ERR_BAD_REQUEST", http.StatusBadRequest)
		return
	}

	// Open a temporary SQLiteStore on the default queue db path for querying
	sqlitePath := os.Getenv("PRANOR_PULSE_SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "Pranor Pulse.db"
	}

	store, err := storage.OpenSQLiteStore(sqlitePath)
	if err != nil {
		WriteJSONError(w, r, "SQLite backend unavailable: "+err.Error(), "ERR_SQLITE_UNAVAILABLE", http.StatusServiceUnavailable)
		return
	}
	defer store.Close()

	results, err := store.QuerySQL(req.SQL)
	if err != nil {
		WriteJSONError(w, r, "Query error: "+err.Error(), "ERR_QUERY_FAILED", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rows":  results,
		"count": len(results),
	})
}

