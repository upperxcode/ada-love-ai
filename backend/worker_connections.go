package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// ConnectionDefinition defines a predefined connection type.
type ConnectionDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`    // "ada", "cli", "rest", "mcp"
	Command     string `json:"command"` // CLI command (for cli/mcp type)
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// PredefinedConnections returns the list of available connections.
func PredefinedConnections() []ConnectionDefinition {
	return []ConnectionDefinition{
		{
			Name:        "Ada",
			Type:        "ada",
			Description: "Built-in Ada engine (uses backend agent loop)",
			Icon:        "🤖",
		},
		{
			Name:        "Crush",
			Type:        "cli",
			Command:     "crush",
			Description: "Crush AI coding assistant",
			Icon:        "💎",
		},
		{
			Name:        "OpenCode",
			Type:        "cli",
			Command:     "opencode",
			Description: "OpenCode AI coding assistant",
			Icon:        "🔓",
		},
		{
			Name:        "Aider",
			Type:        "cli",
			Command:     "aider",
			Description: "Aider AI pair programming",
			Icon:        "🤝",
		},
		{
			Name:        "Claude Code",
			Type:        "cli",
			Command:     "claude",
			Description: "Claude Code CLI",
			Icon:        "🧠",
		},
		{
			Name:        "Custom CLI",
			Type:        "cli",
			Description: "Custom CLI tool",
			Icon:        "⌨️",
		},
		{
			Name:        "REST API",
			Type:        "rest",
			Description: "REST-compatible API endpoint",
			Icon:        "🌐",
		},
		{
			Name:        "MCP Server",
			Type:        "mcp",
			Description: "Model Context Protocol server",
			Icon:        "🔌",
		},
	}
}

// ConnectionTestResult holds the result of a connection test.
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency int64  `json:"latency_ms"`
}

// TestConnection tests a worker connection.
func (e *Engine) TestConnection(connectionType, connectionName, connectionConfig string) ConnectionTestResult {
	start := time.Now()

	switch connectionType {
	case "ada":
		// Ada is always available
		return ConnectionTestResult{
			Success: true,
			Message: "Ada engine is running",
			Latency: 0,
		}

	case "cli":
		return testCLIConnection(connectionName, connectionConfig, start)

	case "rest":
		return testRESTConnection(connectionConfig, start)

	case "mcp":
		return testMCPConnection(connectionName, connectionConfig, start)

	default:
		return ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Unknown connection type: %s", connectionType),
		}
	}
}

func testCLIConnection(name, config string, start time.Time) ConnectionTestResult {
	// Parse config to get command
	command := name // default: use connection name as command
	if config != "" {
		var cfg struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}
		if err := json.Unmarshal([]byte(config), &cfg); err == nil && cfg.Command != "" {
			command = cfg.Command
		}
	}

	// Try to run command with --version or --help
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First try --version
	cmd := exec.CommandContext(ctx, command, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try --help
		cmd = exec.CommandContext(ctx, command, "--help")
		output, err = cmd.CombinedOutput()
	}
	if err != nil {
		return ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Command '%s' not found or not responding: %v", command, err),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	version := strings.TrimSpace(string(output))
	if len(version) > 100 {
		version = version[:100] + "..."
	}

	return ConnectionTestResult{
		Success: true,
		Message: fmt.Sprintf("Connected to %s: %s", name, version),
		Latency: time.Since(start).Milliseconds(),
	}
}

func testRESTConnection(config string, start time.Time) ConnectionTestResult {
	var cfg struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
	}
	if config != "" {
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return ConnectionTestResult{
				Success: false,
				Message: fmt.Sprintf("Invalid config: %v", err),
			}
		}
	}

	if cfg.URL == "" {
		return ConnectionTestResult{
			Success: false,
			Message: "URL is required",
		}
	}

	method := cfg.Method
	if method == "" {
		method = "GET"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, cfg.URL, nil)
	if err != nil {
		return ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Invalid URL: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ConnectionTestResult{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	status := resp.Status
	if len(status) > 80 {
		status = status[:80]
	}

	return ConnectionTestResult{
		Success: resp.StatusCode >= 200 && resp.StatusCode < 500,
		Message: fmt.Sprintf("HTTP %s", status),
		Latency: time.Since(start).Milliseconds(),
	}
}

func testMCPConnection(name, config string, start time.Time) ConnectionTestResult {
	// MCP servers can be CLI-based or WebSocket-based
	var cfg struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
		URL     string   `json:"url"`
	}
	if config != "" {
		json.Unmarshal([]byte(config), &cfg)
	}

	// If URL is provided, try WebSocket connection
	if cfg.URL != "" {
		return testRESTConnection(config, start)
	}

	// If command is provided, try CLI
	if cfg.Command != "" {
		return testCLIConnection(name, config, start)
	}

	return ConnectionTestResult{
		Success: false,
		Message: "MCP server requires either 'command' or 'url' in config",
	}
}

// LanguageInstructions returns the language instruction to inject at the end
// of the persona/system prompt for a given worker. The AI should use English
// internally for reasoning but respond to the user in the specified language.
func LanguageInstructions(w WorkerConfig) string {
	lang := strings.TrimSpace(w.Language)
	if lang == "" {
		return ""
	}

	langName := lang
	switch lang {
	case "pt-BR", "pt":
		langName = "Portuguese (Brazilian)"
	case "en":
		langName = "English"
	case "es":
		langName = "Spanish"
	case "fr":
		langName = "French"
	case "de":
		langName = "German"
	case "it":
		langName = "Italian"
	case "ja":
		langName = "Japanese"
	case "zh":
		langName = "Chinese"
	case "ko":
		langName = "Korean"
	case "ru":
		langName = "Russian"
	case "ar":
		langName = "Arabic"
	}

	return fmt.Sprintf(
		"IMPORTANT: Always respond to the user in %s (%s). You may think and reason internally in English, but the final response must be in %s.",
		langName, lang, langName,
	)
}

// FullPersona returns the complete persona for a worker: the base persona
// concatenated with the language instruction (if any).
func FullPersona(w WorkerConfig) string {
	parts := []string{}
	if w.Persona != "" {
		parts = append(parts, w.Persona)
	}
	if instr := LanguageInstructions(w); instr != "" {
		parts = append(parts, instr)
	}
	return strings.Join(parts, "\n\n")
}
