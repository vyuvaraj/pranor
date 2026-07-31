package tunnel

import (
	"bytes"
	"testing"
)

func TestTunnelMultiplexer_FrameSerializationAndRouting(t *testing.T) {
	buf := new(bytes.Buffer)

	frameOut := TunnelFrame{
		StreamID: 101,
		Type:     FrameData,
		Payload:  []byte("hello multiplexed tunnel"),
	}

	err := WriteFrame(buf, frameOut)
	if err != nil {
		t.Fatalf("WriteFrame failed: %v", err)
	}

	frameIn, err := ReadFrame(buf)
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}

	if frameIn.StreamID != 101 || frameIn.Type != FrameData || string(frameIn.Payload) != "hello multiplexed tunnel" {
		t.Errorf("frame deserialization mismatch: %+v", frameIn)
	}

	// Route through TunnelMultiplexer
	mux := NewTunnelMultiplexer()
	id, ch := mux.OpenStream()
	if id != 1 {
		t.Errorf("expected stream ID 1, got %d", id)
	}

	frameOut.StreamID = 1
	err = mux.RouteFrame(&frameOut)
	if err != nil {
		t.Fatalf("RouteFrame failed: %v", err)
	}

	received := <-ch
	if string(received) != "hello multiplexed tunnel" {
		t.Errorf("unexpected payload received on stream channel: %s", string(received))
	}
}
