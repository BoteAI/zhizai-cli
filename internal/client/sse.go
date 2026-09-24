package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const sseClientTimeout = 5 * time.Minute

// SSEEventHandler receives one SSE event (event name may be empty).
// Return a non-nil error to abort reading the stream.
type SSEEventHandler func(event, data string) error

func (c *Client) sseHTTPClient() *http.Client {
	base := c.httpClient
	if base == nil {
		base = &http.Client{}
	}
	return &http.Client{
		Timeout:       sseClientTimeout,
		Transport:     base.Transport,
		Jar:           base.Jar,
		CheckRedirect: base.CheckRedirect,
	}
}

// doSSEPOST sends a JSON POST and streams the SSE body through handler.
// Unlike do(), it does not parse resultCode envelopes — SSE is event-based.
func (c *Client) doSSEPOST(path string, payload any, handler SSEEventHandler) error {
	if handler == nil {
		return fmt.Errorf("SSE handler is required")
	}

	var body []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = data
	}

	c.waitRateLimit()

	req, err := c.newRequest(http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.sseHTTPClient().Do(req)
	if err != nil {
		return networkRequestError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 406 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = "无权限或请求无效，请检查登录状态（zhizai auth login / ZHIZAI_REC_API_KEY）"
		}
		return &RequestError{
			APIError: APIError{
				Code:      fmt.Sprintf("%d", resp.StatusCode),
				Message:   msg,
				Reason:    "unauthorized",
				Retryable: false,
			},
			StatusCode: resp.StatusCode,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return &RequestError{
			APIError: APIError{
				Code:      fmt.Sprintf("%d", resp.StatusCode),
				Message:   strings.TrimSpace(string(raw)),
				Reason:    "http_error",
				Retryable: resp.StatusCode >= 500 || resp.StatusCode == 429,
			},
			StatusCode: resp.StatusCode,
		}
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(ct, "application/json") {
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return networkRequestError(err)
		}
		_, err = c.parseSuccessEnvelope(raw, resp.StatusCode)
		if err != nil {
			return err
		}
		return fmt.Errorf("期望 SSE 流，但收到 JSON 成功响应")
	}

	return readSSE(resp.Body, handler)
}

func readSSE(r io.Reader, handler SSEEventHandler) error {
	scanner := bufio.NewScanner(r)
	// Allow larger SSE frames (summaries / long answer chunks).
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var eventName string
	var dataLines []string

	flush := func() error {
		if len(dataLines) == 0 && eventName == "" {
			return nil
		}
		data := strings.Join(dataLines, "\n")
		ev := eventName
		eventName = ""
		dataLines = nil
		return handler(ev, data)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue // comment / keepalive
		}
		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			payload := strings.TrimPrefix(line, "data:")
			if strings.HasPrefix(payload, " ") {
				payload = payload[1:]
			}
			dataLines = append(dataLines, payload)
			continue
		}
		if strings.HasPrefix(line, "id:") {
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return networkRequestError(err)
	}
	return flush()
}
