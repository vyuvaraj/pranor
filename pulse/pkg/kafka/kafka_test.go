//go:build !enterprise

package kafka

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestKafkaAdapterHeaderDecodingAndRouting(t *testing.T) {
	srv := NewServer(":9092", nil)

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, int32(16))          // Size
	binary.Write(&buf, binary.BigEndian, ApiKeyProduce)      // ApiKey
	binary.Write(&buf, binary.BigEndian, int16(1))           // ApiVersion
	binary.Write(&buf, binary.BigEndian, int32(1001))        // CorrelationId

	hdr, err := DecodeHeader(&buf)
	if err != nil {
		t.Fatalf("DecodeHeader failed: %v", err)
	}

	if hdr.ApiKey != ApiKeyProduce || hdr.CorrelationId != 1001 {
		t.Errorf("Unexpected header values: %+v", hdr)
	}

	respBytes, topic, msg, err := srv.HandleRequest(hdr, []byte("kafka-message-body"))
	if err != nil || len(respBytes) == 0 {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	if topic != "kafka-events" || msg != "kafka-message-body" {
		t.Errorf("Unexpected topic/msg: topic=%s, msg=%s", topic, msg)
	}

	respCorrId := int32(binary.BigEndian.Uint32(respBytes[4:8]))
	if respCorrId != 1001 {
		t.Errorf("Expected response correlation ID 1001, got %d", respCorrId)
	}
}
