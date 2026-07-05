package integrationtools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	toolshared "ada-love-ai/pkg/tools/shared"
)

const (
	maxHTTPResponseBody = 100000 // 100KB max response body
	httpTimeout         = 30 * time.Second
)

// HTTPRequestTool sends direct HTTP requests to test APIs.
type HTTPRequestTool struct{}

func NewHTTPRequestTool() *HTTPRequestTool {
	return &HTTPRequestTool{}
}

func (t *HTTPRequestTool) Name() string        { return "http_request" }
func (t *HTTPRequestTool) Description() string {
	return "Send an HTTP request (GET, POST, PUT, PATCH, DELETE) to a URL. Useful for testing APIs created by the agent. Returns status code, headers, and response body."
}
func (t *HTTPRequestTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to send the request to.",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "HTTP method. Default: GET.",
				"enum":        []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
				"default":     "GET",
			},
			"headers": map[string]any{
				"type":        "object",
				"description": "Request headers as key-value pairs.",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Request body (for POST/PUT/PATCH).",
			},
		},
		"required": []string{"url"},
	}
}

func (t *HTTPRequestTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	url, ok := args["url"].(string)
	if !ok || url == "" {
		return toolshared.ErrorResult("url is required")
	}

	method := "GET"
	if m, ok := args["method"].(string); ok && m != "" {
		method = strings.ToUpper(m)
	}

	var body io.Reader
	if b, ok := args["body"].(string); ok && b != "" {
		body = strings.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return toolshared.ErrorResult(fmt.Sprintf("failed to create request: %v", err))
	}

	// Set headers
	if headers, ok := args["headers"].(map[string]any); ok {
		for k, v := range headers {
			if s, ok := v.(string); ok {
				req.Header.Set(k, s)
			}
		}
	}

	// Default Content-Type for body requests
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return toolshared.ErrorResult(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	// Read response body with limit
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPResponseBody))
	if err != nil {
		return toolshared.ErrorResult(fmt.Sprintf("failed to read response: %v", err))
	}

	var result bytes.Buffer
	result.WriteString(fmt.Sprintf("%s %s\n", method, url))
	result.WriteString(fmt.Sprintf("Status: %d %s\n\n", resp.StatusCode, resp.Status))

	// Response headers
	result.WriteString("Response Headers:\n")
	for k, v := range resp.Header {
		result.WriteString(fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", ")))
	}

	// Response body
	bodyStr := string(respBody)
	if len(bodyStr) > 0 {
		result.WriteString(fmt.Sprintf("\nBody (%d bytes):\n%s", len(bodyStr), bodyStr))
	}

	if int64(len(respBody)) >= maxHTTPResponseBody {
		result.WriteString("\n\n[TRUNCATED - response body exceeds 100KB]")
	}

	return toolshared.NewToolResult(result.String())
}
