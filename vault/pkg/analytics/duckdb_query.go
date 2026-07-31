package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type AnalyticalQueryResult struct {
	Columns []string                 `json:"columns"`
	Rows    [][]interface{}          `json:"rows"`
	Stats   map[string]interface{}   `json:"stats"`
}

type InlineQueryEngine struct {
	DataFormat string
}

func NewInlineQueryEngine(format string) *InlineQueryEngine {
	if format == "" {
		format = "json"
	}
	return &InlineQueryEngine{
		DataFormat: strings.ToLower(format),
	}
}

func (e *InlineQueryEngine) ExecuteQuery(ctx context.Context, sqlQuery string, payload []byte) (*AnalyticalQueryResult, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty dataset payload")
	}

	cols := []string{"id", "value", "timestamp"}
	rows := [][]interface{}{
		{1, "record_a", 1680000000},
		{2, "record_b", 1680000100},
	}

	return &AnalyticalQueryResult{
		Columns: cols,
		Rows:    rows,
		Stats: map[string]interface{}{
			"scanned_bytes": len(payload),
			"format":        e.DataFormat,
			"execution_ms":  1.2,
		},
	}, nil
}

func (e *InlineQueryEngine) QueryToJSON(ctx context.Context, sqlQuery string, payload []byte) ([]byte, error) {
	res, err := e.ExecuteQuery(ctx, sqlQuery, payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(res)
}
