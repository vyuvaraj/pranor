package import (
	"bufio"
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/cache/pkg/cache"
)

func TestRESP3Commands(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	srv := NewRESP3Server(memCache)

	// PING
	pingCmd := RESP3Value{
		Type:  TypeArray,
		Array: []RESP3Value{{Type: TypeBulkString, Str: "PING"}},
	}
	resPing := srv.ExecuteCommand(pingCmd)
	if resPing.Type != TypeSimpleString || resPing.Str != "PONG" {
		t.Errorf("expected +PONG, got %+v", resPing)
	}

	// SET foo bar
	setCmd := RESP3Value{
		Type: TypeArray,
		Array: []RESP3Value{
			{Type: TypeBulkString, Str: "SET"},
			{Type: TypeBulkString, Str: "foo"},
			{Type: TypeBulkString, Str: "bar"},
		},
	}
	resSet := srv.ExecuteCommand(setCmd)
	if resSet.Type != TypeSimpleString || resSet.Str != "OK" {
		t.Errorf("expected +OK, got %+v", resSet)
	}

	// GET foo
	getCmd := RESP3Value{
		Type: TypeArray,
		Array: []RESP3Value{
			{Type: TypeBulkString, Str: "GET"},
			{Type: TypeBulkString, Str: "foo"},
		},
	}
	resGet := srv.ExecuteCommand(getCmd)
	if resGet.Type != TypeBulkString || resGet.Str != "bar" {
		t.Errorf("expected $3\\r\\nbar\\r\\n, got %+v", resGet)
	}

	// GET non-existent key (Null)
	getNullCmd := RESP3Value{
		Type: TypeArray,
		Array: []RESP3Value{
			{Type: TypeBulkString, Str: "GET"},
			{Type: TypeBulkString, Str: "nonexistent"},
		},
	}
	resNull := srv.ExecuteCommand(getNullCmd)
	if resNull.Type != TypeNull || !resNull.IsNull {
		t.Errorf("expected null response, got %+v", resNull)
	}

	// DEL foo
	delCmd := RESP3Value{
		Type: TypeArray,
		Array: []RESP3Value{
			{Type: TypeBulkString, Str: "DEL"},
			{Type: TypeBulkString, Str: "foo"},
		},
	}
	resDel := srv.ExecuteCommand(delCmd)
	if resDel.Type != TypeInteger || resDel.Num != 1 {
		t.Errorf("expected :1, got %+v", resDel)
	}
}

func TestRESP3ParseAndWrite(t *testing.T) {
	rawInput := "*3\r\n$3\r\nSET\r\n$4\r\nuser\r\n$5\r\nalice\r\n"
	reader := bufio.NewReader(bytes.NewBufferString(rawInput))

	val, err := ParseRESP3(reader)
	if err != nil {
		t.Fatalf("ParseRESP3 failed: %v", err)
	}

	if val.Type != TypeArray || len(val.Array) != 3 {
		t.Fatalf("expected array of 3 elements, got %+v", val)
	}
	if val.Array[0].Str != "SET" || val.Array[1].Str != "user" || val.Array[2].Str != "alice" {
		t.Errorf("parsed elements mismatch: %+v", val.Array)
	}

	var buf bytes.Buffer
	outVal := RESP3Value{Type: TypeBulkString, Str: "hello"}
	if err := WriteRESP3(&buf, outVal); err != nil {
		t.Fatalf("WriteRESP3 failed: %v", err)
	}
	if buf.String() != "$5\r\nhello\r\n" {
		t.Errorf("written output mismatch, got: %q", buf.String())
	}
}

func TestRESP3ServerConnection(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	srv := NewRESP3Server(memCache)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen failed: %v", err)
	}
	defer l.Close()

	go func() {
		conn, err := l.Accept()
		if err == nil {
			_ = srv.HandleConnection(conn)
		}
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial failed: %v", err)
	}
	defer conn.Close()

	// Write PING command
	pingReq := "*1\r\n$4\r\nPING\r\n"
	_, err = conn.Write([]byte(pingReq))
	if err != nil {
		t.Fatalf("conn.Write failed: %v", err)
	}

	r := bufio.NewReader(conn)
	res, err := ParseRESP3(r)
	if err != nil {
		t.Fatalf("ParseRESP3 from connection failed: %v", err)
	}
	if res.Type != TypeSimpleString || res.Str != "PONG" {
		t.Errorf("expected +PONG over TCP, got %+v", res)
	}
}
