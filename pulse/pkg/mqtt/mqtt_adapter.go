package mqtt

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/vyuvaraj/pranor/pulse/pkg/broker"
)

const (
	PacketTypeConnect    byte = 1
	PacketTypeConnAck    byte = 2
	PacketTypePublish    byte = 3
	PacketTypePubAck     byte = 4
	PacketTypeSubscribe  byte = 8
	PacketTypeSubAck     byte = 9
	PacketTypePingReq    byte = 12
	PacketTypePingResp   byte = 13
	PacketTypeDisconnect byte = 14
)

type Server struct {
	addr     string
	engine   *broker.BrokerEngine
	listener net.Listener
	mu       sync.Mutex
	running  bool
}

func NewServer(addr string, engine *broker.BrokerEngine) *Server {
	if addr == "" {
		addr = ":1883"
	}
	return &Server{
		addr:   addr,
		engine: engine,
	}
}

func DecodeHeader(r io.Reader) (byte, int, error) {
	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return 0, 0, err
	}

	pktType := buf[0] >> 4

	// Read remaining length (variable byte integer)
	multiplier := 1
	length := 0
	for {
		if _, err := r.Read(buf); err != nil {
			return 0, 0, err
		}
		digit := buf[0]
		length += int(digit&127) * multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier *= 128
	}

	return pktType, length, nil
}

func EncodeConnAck() []byte {
	return []byte{PacketTypeConnAck << 4, 2, 0, 0}
}

func EncodePubAck() []byte {
	return []byte{PacketTypePubAck << 4, 2, 0, 1}
}

func EncodeSubAck() []byte {
	return []byte{PacketTypeSubAck << 4, 3, 0, 1, 0}
}

func EncodePingResp() []byte {
	return []byte{PacketTypePingResp << 4, 0}
}

func (s *Server) HandlePacket(pktType byte, payload []byte) ([]byte, string, string, error) {
	switch pktType {
	case PacketTypeConnect:
		return EncodeConnAck(), "", "", nil
	case PacketTypePublish:
		// Basic MQTT publish payload parsing
		topic := "iot/telemetry"
		msg := string(payload)
		if len(payload) > 2 {
			topLen := int(payload[0])<<8 | int(payload[1])
			if len(payload) >= 2+topLen {
				topic = string(payload[2 : 2+topLen])
				msg = string(payload[2+topLen:])
			}
		}
		if s.engine != nil {
			_, _ = s.engine.Publish(context.Background(), topic, msg)
		}
		return EncodePubAck(), topic, msg, nil
	case PacketTypeSubscribe:
		return EncodeSubAck(), "", "", nil
	case PacketTypePingReq:
		return EncodePingResp(), "", "", nil
	default:
		return nil, "", "", fmt.Errorf("unsupported mqtt packet type: %d", pktType)
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = l
	s.running = true
	s.mu.Unlock()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				s.mu.Lock()
				if !s.running {
					s.mu.Unlock()
					return
				}
				s.mu.Unlock()
				continue
			}
			go s.handleConnection(conn)
		}
	}()
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		pktType, length, err := DecodeHeader(conn)
		if err != nil {
			return
		}

		payload := make([]byte, length)
		if length > 0 {
			if _, err := io.ReadFull(conn, payload); err != nil {
				return
			}
		}

		resp, _, _, err := s.HandlePacket(pktType, payload)
		if err != nil {
			return
		}
		if resp != nil {
			_, _ = conn.Write(resp)
		}
		if pktType == PacketTypeDisconnect {
			return
		}
	}
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	if s.listener != nil {
		_ = s.listener.Close()
	}
}
