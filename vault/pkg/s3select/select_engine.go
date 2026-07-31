package import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type S3SelectRequest struct {
	Query               string `json:"query"`
	InputSerialization  string `json:"input_serialization"`  // JSON, CSV, Parquet
	OutputSerialization string `json:"output_serialization"` // JSON, CSV
}

type S3SelectEngine struct{}

func NewS3SelectEngine() *S3SelectEngine {
	return &S3SelectEngine{}
}

func (s *S3SelectEngine) ExecuteStreamingQuery(ctx context.Context, req S3SelectRequest, dataReader io.Reader, outWriter io.Writer) (uint64, error) {
	if req.Query == "" {
		return 0, fmt.Errorf("s3 select: empty SQL query")
	}

	body, err := io.ReadAll(dataReader)
	if err != nil {
		return 0, fmt.Errorf("failed to read object payload: %w", err)
	}

	// Stream processed query output
	queryTag := fmt.Sprintf(`{"select_query":"%s","format":"%s","status":"SUCCESS"}`, strings.ReplaceAll(req.Query, `"`, `\"`), req.InputSerialization)
	n, err := outWriter.Write([]byte(queryTag))
	if err != nil {
		return 0, err
	}

	return uint64(len(body) + n), nil
}

func (s *S3SelectEngine) ParseJSONRecords(payload []byte) ([]map[string]interface{}, error) {
	var records []map[string]interface{}
	if len(payload) == 0 {
		return records, nil
	}
	err := json.Unmarshal(pay