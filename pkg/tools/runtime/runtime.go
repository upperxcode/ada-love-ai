package runtimetools

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	toolshared "ada-love-ai/pkg/tools/shared"
)

// --- run_tests ---

type RunTestsTool struct {
	workspace string
}

func NewRunTestsTool(workspace string) *RunTestsTool {
	return &RunTestsTool{workspace: workspace}
}

func (t *RunTestsTool) Name() string { return "run_tests" }
func (t *RunTestsTool) Description() string {
	return "Run the project's test suite and return structured results. Auto-detects the test framework (go test, pytest, jest, npm test, cargo test, etc.) or uses a custom command."
}
func (t *RunTestsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Custom test command to run (e.g. 'go test ./...', 'pytest tests/'). If omitted, auto-detects from project files.",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Subdirectory to run tests in. Defaults to workspace root.",
			},
			"timeout_seconds": map[string]any{
				"type":        "integer",
				"description": "Maximum time to wait for tests. Default: 120.",
				"default":     120,
			},
		},
	}
}

func (t *RunTestsTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	workDir := t.workspace
	if p, ok := args["path"].(string); ok && p != "" {
		workDir = p
		if !filepath.IsAbs(workDir) {
			workDir = filepath.Join(t.workspace, workDir)
		}
	}

	timeoutSec := 120
	if ts, ok := args["timeout_seconds"].(float64); ok && ts > 0 {
		timeoutSec = int(ts)
	}

	var cmd string
	if c, ok := args["command"].(string); ok && c != "" {
		cmd = c
	} else {
		cmd = detectTestCommand(workDir)
	}

	if cmd == "" {
		return toolshared.ErrorResult("Could not detect test framework. Please provide a 'command' parameter.")
	}

	timeout := time.Duration(timeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Split command into args
	parts := strings.Fields(cmd)
	execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	execCmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Command: %s\n", cmd))
	result.WriteString(fmt.Sprintf("Directory: %s\n\n", workDir))

	if ctx.Err() == context.DeadlineExceeded {
		result.WriteString(fmt.Sprintf("[TIMEOUT] Tests exceeded %d second limit.\n", timeoutSec))
		result.WriteString("Partial output:\n")
	}

	if stdout.Len() > 0 {
		out := stdout.String()
		if len(out) > 20000 {
			out = out[:20000] + "\n[TRUNCATED]"
		}
		result.WriteString("STDOUT:\n" + out + "\n")
	}
	if stderr.Len() > 0 {
		errOut := stderr.String()
		if len(errOut) > 10000 {
			errOut = errOut[:10000] + "\n[TRUNCATED]"
		}
		result.WriteString("STDERR:\n" + errOut + "\n")
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.WriteString(fmt.Sprintf("\nExit code: %d", exitErr.ExitCode()))
		} else if ctx.Err() != context.DeadlineExceeded {
			result.WriteString(fmt.Sprintf("\nError: %v", err))
		}
		return toolshared.NewToolResult(result.String())
	}

	result.WriteString("\nExit code: 0 (SUCCESS)")
	return toolshared.NewToolResult(result.String())
}

func detectTestCommand(dir string) string {
	// Check for common test config files
	checks := []struct {
		file    string
		command string
	}{
		{"go.mod", "go test ./..."},
		{"package.json", "npm test"},
		{"Cargo.toml", "cargo test"},
		{"Makefile", "make test"},
		{"pytest.ini", "pytest"},
		{"setup.cfg", "pytest"},
		{"pyproject.toml", "pytest"},
		{"tox.ini", "tox"},
	}

	for _, c := range checks {
		if fileExists(filepath.Join(dir, c.file)) {
			return c.command
		}
	}

	// Check for test directories
	if dirExists(filepath.Join(dir, "tests")) || dirExists(filepath.Join(dir, "test")) {
		// If Python files exist, try pytest
		return "pytest"
	}

	return ""
}

func fileExists(path string) bool {
	return exec.Command("test", "-f", path).Run() == nil
}

func dirExists(path string) bool {
	return exec.Command("test", "-d", path).Run() == nil
}

// --- run_linter_formatter ---

type RunLinterFormatterTool struct {
	workspace string
}

func NewRunLinterFormatterTool(workspace string) *RunLinterFormatterTool {
	return &RunLinterFormatterTool{workspace: workspace}
}

func (t *RunLinterFormatterTool) Name() string { return "run_linter_formatter" }
func (t *RunLinterFormatterTool) Description() string {
	return "Run linters and formatters to check or fix code style. Auto-detects tools (gofmt, ruff, prettier, eslint, etc.) or uses a custom command. Set fix=true to auto-fix issues."
}
func (t *RunLinterFormatterTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Custom linter/formatter command. If omitted, auto-detects.",
			},
			"fix": map[string]any{
				"type":        "boolean",
				"description": "If true, auto-fix issues instead of just reporting. Default: false.",
				"default":     false,
			},
			"path": map[string]any{
				"type":        "string",
				"description": "File or directory to lint. Defaults to workspace root.",
			},
		},
	}
}

func (t *RunLinterFormatterTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	workDir := t.workspace
	if p, ok := args["path"].(string); ok && p != "" {
		workDir = p
		if !filepath.IsAbs(workDir) {
			workDir = filepath.Join(t.workspace, workDir)
		}
	}

	fix, _ := args["fix"].(bool)

	var cmd string
	if c, ok := args["command"].(string); ok && c != "" {
		cmd = c
	} else {
		cmd = detectLinterCommand(workDir, fix)
	}

	if cmd == "" {
		return toolshared.ErrorResult("Could not detect linter/formatter. Please provide a 'command' parameter.")
	}

	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	parts := strings.Fields(cmd)
	execCmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	execCmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Command: %s\n", cmd))
	result.WriteString(fmt.Sprintf("Mode: %s\n\n", map[bool]string{true: "fix", false: "check"}[fix]))

	if stdout.Len() > 0 {
		out := stdout.String()
		if len(out) > 20000 {
			out = out[:20000] + "\n[TRUNCATED]"
		}
		result.WriteString(out + "\n")
	}
	if stderr.Len() > 0 {
		errOut := stderr.String()
		if len(errOut) > 10000 {
			errOut = errOut[:10000] + "\n[TRUNCATED]"
		}
		result.WriteString("STDERR:\n" + errOut + "\n")
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.WriteString(fmt.Sprintf("\nExit code: %d", exitErr.ExitCode()))
		}
	} else {
		result.WriteString("\nExit code: 0 (no issues found)")
	}

	return toolshared.NewToolResult(result.String())
}

func detectLinterCommand(dir string, fix bool) string {
	type detector struct {
		files []string
		check string
		fix   string
	}

	detectors := []detector{
		{[]string{"go.mod"}, "gofmt -l .", "gofmt -w ."},
		{[]string{"package.json"}, "npx prettier --check .", "npx prettier --write ."},
		{[]string{".eslintrc.js", ".eslintrc.json", ".eslintrc.yml", "eslint.config.js"}, "npx eslint .", "npx eslint --fix ."},
		{[]string{"pyproject.toml", "ruff.toml"}, "ruff check .", "ruff check --fix ."},
		{[]string{"setup.cfg", "pyproject.toml"}, "flake8 .", ""},
		{[]string{"Cargo.toml"}, "cargo fmt --check", "cargo fmt"},
	}

	for _, d := range detectors {
		for _, f := range d.files {
			if exec.Command("test", "-f", filepath.Join(dir, f)).Run() == nil {
				if fix && d.fix != "" {
					return d.fix
				}
				return d.check
			}
		}
	}

	return ""
}

// --- check_port_status ---

type CheckPortStatusTool struct {
	workspace string
}

func NewCheckPortStatusTool(workspace string) *CheckPortStatusTool {
	return &CheckPortStatusTool{workspace: workspace}
}

func (t *CheckPortStatusTool) Name() string { return "check_port_status" }
func (t *CheckPortStatusTool) Description() string {
	return "Check if a TCP port is open (listening) or closed on localhost. Useful for verifying if a server started by the agent is running."
}
func (t *CheckPortStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"port": map[string]any{
				"type":        "integer",
				"description": "Port number to check (1-65535).",
			},
			"host": map[string]any{
				"type":        "string",
				"description": "Host to check. Default: localhost.",
				"default":     "localhost",
			},
		},
		"required": []string{"port"},
	}
}

func (t *CheckPortStatusTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	port, ok := args["port"].(float64)
	if !ok || port < 1 || port > 65535 {
		return toolshared.ErrorResult("port must be between 1 and 65535")
	}

	host := "localhost"
	if h, ok := args["host"].(string); ok && h != "" {
		host = h
	}

	addr := fmt.Sprintf("%s:%d", host, int(port))
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return toolshared.NewToolResult(fmt.Sprintf("Port %d on %s: CLOSED (%v)", int(port), host, err))
	}
	conn.Close()

	// Try to get process info via lsof
	var processInfo string
	lsofOut, err := exec.Command("lsof", "-i", fmt.Sprintf(":%d", int(port)), "-P", "-n").CombinedOutput()
	if err == nil && len(lsofOut) > 0 {
		lines := strings.Split(strings.TrimSpace(string(lsofOut)), "\n")
		if len(lines) > 1 {
			processInfo = "\n" + lines[1] // skip header
		}
	}

	return toolshared.NewToolResult(fmt.Sprintf("Port %d on %s: OPEN%s", int(port), host, processInfo))
}
