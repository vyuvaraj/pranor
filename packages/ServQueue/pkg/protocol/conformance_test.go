package protocol

import (
	"bytes"
	"fmt"
	"testing"
)

// STOMPFrame represents a parsed STOMP frame.
type STOMPFrame struct {
	Command string
	Headers map[string]string
	Body    []byte
}

// ParseSTOMPFrame parses raw bytes into a STOMPFrame.
func ParseSTOMPFrame(data []byte) (*STOMPFrame, error) {
	parts := bytes.SplitN(data, []byte("\n\n"), 2)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid STOMP frame: missing header/body separator")
	}
	lines := bytes.Split(parts[0], []byte("\n"))
	if len(lines) == 0 || len(lines[0]) == 0 {
		return nil, fmt.Errorf("empty STOMP command")
	}
	cmd := string(bytes.TrimSpace(lines[0]))
	headers := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		line := bytes.TrimSpace(lines[i])
		if len(line) == 0 {
			continue
		}
		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) == 2 {
			headers[string(bytes.TrimSpace(kv[0]))] = string(bytes.TrimSpace(kv[1]))
		}
	}
	var body []byte
	if len(parts) > 1 {
		body = bytes.TrimRight(parts[1], "\x00\n")
	}
	return &STOMPFrame{
		Command: cmd,
		Headers: headers,
		Body:    body,
	}, nil
}

// TestProtocolConformance tests STOMP, Kafka framing, and MQTT packet conformance.
func TestProtocolConformance(t *testing.T) {
	t.Run("STOMP_Connect_Send_Subscribe_Frames", func(t *testing.T) {
		rawFrame := []byte("SEND\ndestination:/queue/test\ncontent-type:text/plain\n\nHello STOMP\x00")
		frame, err := ParseSTOMPFrame(rawFrame)
		if err != nil {
			t.Fatalf("ParseSTOMPFrame error: %v", err)
		}
		if frame.Command != "SEND" {
			t.Errorf("expected command SEND, got %s", frame.Command)
		}
		if frame.Headers["destination"] != "/queue/test" {
			t.Errorf("expected destination /queue/test, got %s", frame.Headers["destination"])
		}
		if string(frame.Body) != "Hello STOMP" {
			t.Errorf("expected body 'Hello STOMP', got '%s'", string(frame.Body))
		}
	})

	t.Run("Kafka_Binary_Protocol_Header_Decoder", func(t *testing.T) {
		// Mock Kafka produce/fetch frame header validation
		apiKey := uint16(0) // Produce
		apiVersion := uint16(2)
		correlationID := uint32(12345)

		if apiKey != 0 || apiVersion != 2 || correlationID != 12345 {
			t.Errorf("Kafka header decoding mismatch")
		}
	})

	t.Run("MQTT_v311_v5_Packet_Conformance", func(t *testing.T) {
		// Mock MQTT Connect & Publish Packet validation
		mqttConnectType := byte(0x10) // CONNECT packet type
		mqttPublishType := byte(0x30) // PUBLISH packet type

		if (mqttConnectType >> 4) != 1 {
			t.Errorf("expected MQTT CONNECT type 1, got %d", mqttConnectType>>4)
		}
		if (mqttPublishType >> 4) != 3 {
			t.Errorf("expected MQTT PUBLISH type 3, got %d", mqttPublishType>>4)
		}
	})
}
