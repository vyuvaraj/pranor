package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// FrameType specifies the multiplexed tunnel frame type.
type FrameType byte

const (
	FrameData  FrameType = 0x01
	FrameOpen  FrameType = 0x02
	FrameClose FrameType = 0x03
	FramePing  FrameType = 0x04
)

// TunnelFrame represents a multiplexed binary frame over a single tunnel pipe.
type TunnelFrame struct {
	StreamID uint32
	Type     FrameType
	Payload  []byte
}

// WriteFrame serializes a TunnelFrame to binary wire format.
// Wire Format: [4-byte StreamID][1-byte Type][4-byte PayloadLen][Payload Bytes]
func WriteFrame(w io.Writer, frame TunnelFrame) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, frame.StreamID); err != nil {
		return err
	}
	if err := buf.WriteByte(byte(frame.Type)); err != nil {
		return err
	}
	payloadLen := uint32(len(frame.Payload))
	if err := binary.Write(buf, binary.BigEndian, payloadLen); err != nil {
		return err
	}
	if payloadLen > 0 {
		if _, err := buf.Write(frame.Payload); err != nil {
			return err
		}
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// ReadFrame deserializes a TunnelFrame from binary wire format.
func ReadFrame(r io.Reader) (*TunnelFrame, error) {
	var streamID uint32
	if err := binary.Read(r, binary.BigEndian, &streamID); err != nil {
		return nil, err
	}
	typeBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, typeBuf); err != nil {
		return nil, err
	}
	var payloadLen uint32
	if err := binary.Read(r, binary.BigEndian, &payloadLen); err != nil {
		return nil, err
	}
	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
	}
	return &TunnelFrame{
		StreamID: streamID,
		Type:     FrameType(typeBuf[0]),
		Payload:  payload,
	}, nil
}

// TunnelMultiplexer routes multiplexed virtual stream channels over a single physical WebSocket connection pipe.
type TunnelMultiplexer struct {
	mu      sync.RWMutex
	streams map[uint32]chan []byte
	nextID  uint32
}

// NewTunnelMultiplexer creates a TunnelMultiplexer instance.
func NewTunnelMultiplexer() *TunnelMultiplexer {
	return &TunnelMultiplexer{
		streams: make(map[uint32]chan []byte),
		nextID:  1,
	}
}

// OpenStream allocates a new logical multiplexed stream channel.
func (tm *TunnelMultiplexer) OpenStream() (uint32, chan []byte) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	id := tm.nextID
	tm.nextID++

	ch := make(chan []byte, 100)
	tm.streams[id] = ch
	return id, ch
}

// RouteFrame dispatches incoming frame payload to the target stream channel.
func (tm *TunnelMultiplexer) RouteFrame(frame *TunnelFrame) error {
	tm.mu.RLock()
	ch, exists := tm.streams[frame.StreamID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("stream ID %d not found or closed", frame.StreamID)
	}

	if frame.Type == FrameClose {
		tm.mu.Lock()
		delete(tm.streams, frame.StreamID)
		close(ch)
		tm.mu.Unlock()
		return nil
	}

	select {
	case ch <- frame.Payload:
		return nil
	default:
		return fmt.Errorf("stream ID %d buffer full", frame.StreamID)
	}
}
