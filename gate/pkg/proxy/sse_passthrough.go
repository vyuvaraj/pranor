package import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SSEPassthroughProxy proxies Server-Sent Events (SSE) token streams with zero-buffering chunk flushing.
type SSEPassthroughProxy struct {
	client *http.Client
}

// NewSSEPassthroughProxy creates an SSEPassthroughProxy instance.
func NewSSEPassthroughProxy() *SSEPassthroughProxy {
	return &SSEPassthroughProxy{
		client: &http.Client{},
	}
}

// ProxySSEStream streams incoming SSE response events to the client ResponseWriter in real-time.
func (ssp *SSEPassthroughProxy) ProxySSEStream(w http.ResponseWriter, r *http.Request, upstreamURL string) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported by client ResponseWriter")
	}

	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, r.Body)
	if err != nil {
		return fmt.Errorf("failed to create upstream request: %w", err)
	}

	for k, v := range r.Header {
		upstreamReq.Header[k] = v
	}
	upstreamReq.Header.Set("Accept", "text/event-stream")

	resp, err := ssp.client.Do(upstreamReq)
	if err != nil {
		return fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	// Copy headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(resp.StatusCode)

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			_, _ = io.WriteString(w, line)
			if strings.HasPrefix(line, "data: ") || line == "\n" {
				flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				flusher.Flush()
				return nil
			}
			return err
		}
	}
}
