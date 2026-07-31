package import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	AuthToken  string
	TenantID   string
	HTTPClient *http.Client
}

func NewClient(baseURL, authToken string) *Client {
	return &Client{
		BaseURL:    baseURL,
		AuthToken:  authToken,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) SetTenant(tenant string) {
	c.TenantID = tenant
}

func (c *Client) do(method, endpoint string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}
	if c.TenantID != "" {
		req.Header.Set("X-Tenant-ID", c.TenantID)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Pranor Pulse api error (%d): %s", resp.StatusCode, string(respData))
	}

	return respData, nil
}

func (c *Client) Publish(topic, payload string) error {
	req := map[string]string{
		"topic":   topic,
		"payload": payload,
	}
	_, err := c.do("POST", "/api/v1/publish", req)
	return err
}

func (c *Client) SeekToTime(topic, timeStr string) (int64, error) {
	req := map[string]string{
		"topic": topic,
		"time":  timeStr,
	}
	respBytes, err := c.do("POST", "/api/v1/seekToTime", req)
	if err != nil {
		return 0, err
	}

	var res struct {
		TargetOffset int64 `json:"target_offset"`
	}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return 0, err
	}
	return res.TargetOffset, nil
}
