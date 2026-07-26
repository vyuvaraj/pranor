package kafka

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/broker"
)

const (
	ApiKeyProduce     int16 = 0
	ApiKeyFetch       int16 = 1
	ApiKeyListOffsets int16 = 2
	ApiKeyMetadata    int16 = 3
)

type KafkaHeader struct {
	Size          int32
	ApiKey        int16
	ApiVersion    int16
	CorrelationId int32
}

type Server struct {
	addr     string
	engine   *broker.BrokerEngine
	listener net.Listener
	mu       sync.Mutex
	running  bool
}

func NewServer(addr string, engine *broker.BrokerEngine) *Server {
	if addr == "" {
		addr = ":9092"
	}
	return &Server{
		addr:   addr,
		engine: engine,
	}
}

func DecodeHeader(r io.Reader) (KafkaHeader, error) {
	var hdr KafkaHeader
	if err := binary.Read(r, binary.BigEndian, &hdr.Size); err != nil {
		return hdr, err
	}
	if err := binary.Read(r, binary.BigEndian, &hdr.ApiKey); err != nil {
		return hdr, err
	}
	if err := binary.Read(r, binary.BigEndian, &hdr.ApiVersion); err != nil {
		return hdr, err
	}
	if err := binary.Read(r, binary.BigEndian, &hdr.CorrelationId); err != nil {
		return hdr, err
	}
	return hdr, nil
}

func EncodeHeaderResponse(correlationId int32, payload []byte) []byte {
	size := int32(4 + len(payload))
	buf := make([]byte, 4+4+len(payload))
	binary.BigEndian.PutUint32(buf[0:4], uint32(size))
	binary.BigEndian.PutUint32(buf[4:8], uint32(correlationId))
	copy(buf[8:], payload)
	return buf
}

func (s *Server) HandleRequest(hdr KafkaHeader, body []byte) ([]byte, string, string, error) {
	switch hdr.ApiKey {
	case ApiKeyProduce:
		topic := "kafka-events"
		payloadStr := string(body)
		if s.engine != nil {
			_, _ = s.engine.Publish(context.Background(), topic, payloadStr)
		}
		// Kafka produce response
		respPayload := []byte{0, 0, 0, 1, 0, 0} // ErrorCode = 0 (Success)
		return EncodeHeaderResponse(hdr.CorrelationId, respPayload), topic, payloadStr, nil
	case ApiKeyMetadata:
		respPayload := []byte{0, 0, 0, 0} // Single broker metadata
		return EncodeHeaderResponse(hdr.CorrelationId, respPayload), "", "", nil
	case ApiKeyFetch, ApiKeyListOffsets:
		respPayload := []byte{0, 0, 0, 0}
		return EncodeHeaderResponse(hdr.CorrelationId, respPayload), "", "", nil
	default:
		return nil, "", "", fmt.Errorf("unsupported kafka api key: %d", hdr.ApiKey)
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
		hdr, err := DecodeHeader(conn)
		if err != nil {
			return
		}

		bodyLen := int(hdr.Size) - 8
		if bodyLen < 0 {
			return
		}

		body := make([]byte, bodyLen)
		if bodyLen > 0 {
			if _, err := io.ReadFull(conn, body); err != nil {
				return
			}
		}

		resp, _, _, err := s.HandleRequest(hdr, body)
		if err != nil {
			return
		}
		if resp != nil {
			_, _ = conn.Write(resp)
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
