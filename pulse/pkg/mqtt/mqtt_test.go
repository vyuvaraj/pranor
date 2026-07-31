package import (
	"bytes"
	"testing"
)

func TestMQTTAdapterPacketDecoding(t *testing.T) {
	srv := NewServer(":1883", nil)

	// Test Connect
	resp, _, _, err := srv.HandlePacket(PacketTypeConnect, nil)
	if err != nil || len(resp) == 0 {
		t.Fatalf("HandlePacket Connect failed: %v", err)
	}
	if resp[0]>>4 != PacketTypeConnAck {
		t.Errorf("Expected ConnAck packet type, got %d", resp[0]>>4)
	}

	// Test Publish
	pubPayload := append([]byte{0, 5, 't', 'o', 'p', 'i', 'c'}, []byte("hello-mqtt")...)
	respPub, topic, msg, err := srv.HandlePacket(PacketTypePublish, pubPayload)
	if err != nil || topic != "topic" || msg != "hello-mqtt" {
		t.Errorf("Publish decode error: topic=%s, msg=%s, err=%v", topic, msg, err)
	}
	if respPub[0]>>4 != PacketTypePubAck {
		t.Errorf("Expected PubAck packet type, got %d", respPub[0]>>4)
	}

	// Header decoding test
	hdrBytes := []byte{PacketTypePublish << 4, 12}
	pktType, lenVal, err := DecodeHeader(bytes.NewBuffer(hdrBytes))
	if err != nil || pktType != PacketTypePublish || lenVal != 12 {
		t.Errorf("DecodeHeader failed: pktType=%d, len=%d, err=%v", pktType, lenVal, err)
	}
}
