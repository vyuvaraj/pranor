package transcode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GRPCTranscoder struct {
	targetAddr string
}

func NewGRPCTranscoder(targetAddr string) *GRPCTranscoder {
	if targetAddr == "" {
		targetAddr = "localhost:50051"
	}
	return &GRPCTranscoder{
		targetAddr: targetAddr,
	}
}

func (g *GRPCTranscoder) TranscodeHTTPToGRPC(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	var jsonReq map[string]interface{}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &jsonReq)
	}

	// Transcode JSON map into simulated gRPC payload
	grpcPayload := append([]byte{0, 0, 0, 0, byte(len(body))}, body...)

	w.Header().Set("Content-Type", "application/grpc")
	w.Header().Set("grpc-status", "0")
	w.Header().Set("X-ServGateway-Transcoded", "REST-to-gRPC")
	w.WriteHeader(http.StatusOK)
	w.Write(grpcPayload)
}

type GraphQLAggregator struct {
	upstreams []string
}

func NewGraphQLAggregator(upstreams []string) *GraphQLAggregator {
	return &GraphQLAggregator{
		upstreams: upstreams,
	}
}

func (ga *GraphQLAggregator) AggregateSchema(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("empty graphql query")
	}
	return fmt.Sprintf(`{"data":{"aggregated_nodes":%d,"query_status":"SUCCESS"}}`, len(ga.upstreams)), nil
}
