package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vyuvaraj/pranor/trace/pkg/store"
)

// Exporter converts internal Pranor traces and forwards them to standard OpenTelemetry Collector endpoints
type Exporter struct {
	targetURL  string
	httpClient *http.Client
}

func NewExporter(targetURL string) *Exporter {
	return &Exporter{
		targetURL: targetURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ExportTraces sends trace summaries/spans to an external OTel collector endpoint
func (e *Exporter) ExportTraces(ctx context.Context, spans []store.Span) error {
	if len(spans) == 0 {
		return nil
	}

	otlpPayload := map[string]interface{}{
		"resourceSpans": []map[string]interface{}{
			{
				"resource": map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"key": "service.name",
							"value": map[string]interface{}{
								"stringValue": spans[0].Service,
							},
						},
						{
							"key": "exporter",
							"value": map[string]interface{}{
								"stringValue": "pranor-otel-exporter",
							},
						},
					},
				},
				"scopeSpans": []map[string]interface{}{
					{
						"spans": e.formatSpans(spans),
					},
				},
			},
		},
	}

	data, err := json.Marshal(otlpPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal OTLP payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.targetURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to export OTel payload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("exporter returned unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (e *Exporter) formatSpans(spans []store.Span) []map[string]interface{} {
	var res []map[string]interface{}
	for _, s := range spans {
		attrs := make([]map[string]interface{}, 0)
		for k, v := range s.Attributes {
			attrs = append(attrs, map[string]interface{}{
				"key": k,
				"value": map[string]interface{}{
					"stringValue": fmt.Sprintf("%v", v),
				},
			})
		}

		res = append(res, map[string]interface{}{
			"traceId":           s.TraceID,
			"spanId":            s.SpanID,
			"parentSpanId":      s.ParentSpanID,
			"name":              s.Name,
			"kind":              s.Kind,
			"startTimeUnixNano": fmt.Sprintf("%d", s.StartTime),
			"endTimeUnixNano":   fmt.Sprintf("%d", s.EndTime),
			"status": map[string]interface{}{
				"code": s.Status,
			},
			"attributes": attrs,
		})
	}
	return res
}
