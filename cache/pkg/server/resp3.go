// Package server provides server transports for Pranor Cache, including RESP3 protocol support.
package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/vyuvaraj/pranor/cache/pkg/cache"
)

// RESP3Type represents RESP3 data types.
type RESP3Type byte

const (
	TypeSimpleString RESP3Type = '+'
	TypeError        RESP3Type = '-'
	TypeInteger      RESP3Type = ':'
	TypeBulkString   RESP3Type = '$'
	TypeArray        RESP3Type = '*'
	TypeNull         RESP3Type = '_'
	TypeMap          RESP3Type = '%'
	TypeBoolean      RESP3Type = '#'
)

// RESP3Value represents a parsed or output RESP3 protocol element.
type RESP3Value struct {
	Type   RESP3Type
	Str    string
	Num    int64
	Bool   bool
	Array  []RESP3Value
	Map    map[string]RESP3Value
	IsNull bool
}

// RESP3Server wraps a Pranor Cache instance and handles Redis RESP3 wire connections.
type RESP3Server struct {
	cache cache.Cache
}

// NewRESP3Server creates a new RESP3 protocol server wrapper over a Pranor Cache.
func NewRESP3Server(c cache.Cache) *RESP3Server {
	return &RESP3Server{cache: c}
}

// HandleConnection processes commands over a net.Conn using RESP3 protocol format.
func (s *RESP3Server) HandleConnection(conn net.Conn) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		val, err := ParseRESP3(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		resp := s.ExecuteCommand(val)
		if err := WriteRESP3(writer, resp); err != nil {
			return err
		}
		if err := writer.Flush(); err != nil {
			return err
		}
	}
}

// ExecuteCommand dispatches parsed RESP3 commands against Pranor Cache.
func (s *RESP3Server) ExecuteCommand(cmd RESP3Value) RESP3Value {
	if cmd.Type != TypeArray || len(cmd.Array) == 0 {
		return RESP3Value{Type: TypeError, Str: "ERR invalid request format"}
	}

	commandName := strings.ToUpper(cmd.Array[0].Str)
	args := cmd.Array[1:]

	switch commandName {
	case "PING":
		if len(args) > 0 {
			return RESP3Value{Type: TypeBulkString, Str: args[0].Str}
		}
		return RESP3Value{Type: TypeSimpleString, Str: "PONG"}

	case "GET":
		if len(args) != 1 {
			return RESP3Value{Type: TypeError, Str: "ERR wrong number of arguments for 'GET' command"}
		}
		val, found, err := s.cache.Get(args[0].Str)
		if err != nil {
			return RESP3Value{Type: TypeError, Str: "ERR " + err.Error()}
		}
		if !found {
			return RESP3Value{Type: TypeNull, IsNull: true}
		}
		return RESP3Value{Type: TypeBulkString, Str: fmt.Sprintf("%v", val)}

	case "SET":
		if len(args) < 2 {
			return RESP3Value{Type: TypeError, Str: "ERR wrong number of arguments for 'SET' command"}
		}
		key := args[0].Str
		val := args[1].Str
		var ttl time.Duration = 0

		// Parse optional EX / PX arguments
		for i := 2; i < len(args); i++ {
			opt := strings.ToUpper(args[i].Str)
			if opt == "EX" && i+1 < len(args) {
				sec, err := strconv.Atoi(args[i+1].Str)
				if err == nil {
					ttl = time.Duration(sec) * time.Second
				}
				i++
			} else if opt == "PX" && i+1 < len(args) {
				ms, err := strconv.Atoi(args[i+1].Str)
				if err == nil {
					ttl = time.Duration(ms) * time.Millisecond
				}
				i++
			}
		}

		if err := s.cache.Set(key, val, ttl); err != nil {
			return RESP3Value{Type: TypeError, Str: "ERR " + err.Error()}
		}
		return RESP3Value{Type: TypeSimpleString, Str: "OK"}

	case "DEL":
		if len(args) == 0 {
			return RESP3Value{Type: TypeError, Str: "ERR wrong number of arguments for 'DEL' command"}
		}
		var count int64
		for _, arg := range args {
			if err := s.cache.Delete(arg.Str); err == nil {
				count++
			}
		}
		return RESP3Value{Type: TypeInteger, Num: count}

	default:
		return RESP3Value{Type: TypeError, Str: fmt.Sprintf("ERR unknown command '%s'", commandName)}
	}
}

// ParseRESP3 reads and parses a single RESP3 value from a bufio.Reader.
func ParseRESP3(r *bufio.Reader) (RESP3Value, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return RESP3Value{}, err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) == 0 {
		return RESP3Value{}, fmt.Errorf("empty RESP3 line")
	}

	respType := RESP3Type(line[0])
	payload := line[1:]

	switch respType {
	case TypeSimpleString:
		return RESP3Value{Type: TypeSimpleString, Str: payload}, nil
	case TypeError:
		return RESP3Value{Type: TypeError, Str: payload}, nil
	case TypeInteger:
		num, err := strconv.ParseInt(payload, 10, 64)
		return RESP3Value{Type: TypeInteger, Num: num}, err
	case TypeBulkString:
		length, err := strconv.Atoi(payload)
		if err != nil {
			return RESP3Value{}, err
		}
		if length == -1 {
			return RESP3Value{Type: TypeNull, IsNull: true}, nil
		}
		buf := make([]byte, length+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return RESP3Value{}, err
		}
		return RESP3Value{Type: TypeBulkString, Str: string(buf[:length])}, nil
	case TypeArray:
		count, err := strconv.Atoi(payload)
		if err != nil {
			return RESP3Value{}, err
		}
		if count == -1 {
			return RESP3Value{Type: TypeNull, IsNull: true}, nil
		}
		elements := make([]RESP3Value, count)
		for i := 0; i < count; i++ {
			elem, err := ParseRESP3(r)
			if err != nil {
				return RESP3Value{}, err
			}
			elements[i] = elem
		}
		return RESP3Value{Type: TypeArray, Array: elements}, nil
	default:
		return RESP3Value{}, fmt.Errorf("unsupported RESP3 type: %c", respType)
	}
}

// WriteRESP3 serializes a RESP3Value into a writer.
func WriteRESP3(w io.Writer, val RESP3Value) error {
	switch val.Type {
	case TypeSimpleString:
		_, err := fmt.Fprintf(w, "+%s\r\n", val.Str)
		return err
	case TypeError:
		_, err := fmt.Fprintf(w, "-%s\r\n", val.Str)
		return err
	case TypeInteger:
		_, err := fmt.Fprintf(w, ":%d\r\n", val.Num)
		return err
	case TypeBulkString:
		_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(val.Str), val.Str)
		return err
	case TypeNull:
		_, err := fmt.Fprintf(w, "_\r\n")
		return err
	case TypeArray:
		if _, err := fmt.Fprintf(w, "*%d\r\n", len(val.Array)); err != nil {
			return err
		}
		for _, elem := range val.Array {
			if err := WriteRESP3(w, elem); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported write RESP3 type: %c", val.Type)
	}
}
