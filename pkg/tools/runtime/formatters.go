package runtimetools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	toolshared "ada-love-ai/pkg/tools/shared"
)

// --- prettier_format ---

// PrettierFormatTool formats code using Prettier with the project's config.
type PrettierFormatTool struct {
	workspace string
}

func NewPrettierFormatTool(workspace string) *PrettierFormatTool {
	return &PrettierFormatTool{workspace: workspace}
}

func (t *PrettierFormatTool) Name() string { return "prettier_format" }
func (t *PrettierFormatTool) Description() string {
	return "Format code using Prettier. Can format a specific file or code snippet. Uses the project's .prettierrc config automatically."
}
func (t *PrettierFormatTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "File path to format. Relative to workspace or absolute.",
			},
			"code": map[string]any{
				"type":        "string",
				"description": "Raw code string to format (stdin mode). Use this OR 'path', not both.",
			},
			"parser": map[string]any{
				"type":        "string",
				"description": "Parser to use (e.g. 'typescript', 'tsx', 'json', 'css', 'markdown'). Auto-detected from file extension when using 'path'.",
			},
		},
	}
}

func (t *PrettierFormatTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	path, _ := args["path"].(string)
	code, _ := args["code"].(string)
	parser, _ := args["parser"].(string)

	if path == "" && code == "" {
		return toolshared.ErrorResult("provide either 'path' or 'code' to format")
	}

	// Resolve Prettier binary
	prettierBin, err := findBinary(t.workspace, "prettier", "node_modules/.bin/prettier")
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if code != "" {
		// stdin mode
		args := []string{"--stdin"}
		if parser != "" {
			args = append(args, "--stdin-filepath", "file."+parserToExt(parser))
		}
		cmd = exec.CommandContext(ctx, prettierBin, args...)
		cmd.Dir = t.workspace
		cmd.Stdin = strings.NewReader(code)
	} else {
		// file mode
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(t.workspace, path)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return toolshared.ErrorResult(fmt.Sprintf("file not found: %s", path))
		}
		cmd = exec.CommandContext(ctx, prettierBin, "--write", absPath)
		cmd.Dir = t.workspace
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if len(errMsg) > 5000 {
			errMsg = errMsg[:5000]
		}
		return toolshared.ErrorResult(fmt.Sprintf("prettier failed: %s", errMsg))
	}

	if code != "" {
		formatted := stdout.String()
		if formatted == code {
			return toolshared.NewToolResult("Code is already formatted.")
		}
		return toolshared.NewToolResult(fmt.Sprintf("Formatted code:\n\n%s", formatted))
	}

	return toolshared.NewToolResult(fmt.Sprintf("File formatted: %s", path))
}

// --- eslint_check ---

// ESLintCheckTool runs ESLint on files and returns structured results.
type ESLintCheckTool struct {
	workspace string
}

func NewESLintCheckTool(workspace string) *ESLintCheckTool {
	return &ESLintCheckTool{workspace: workspace}
}

func (t *ESLintCheckTool) Name() string { return "eslint_check" }
func (t *ESLintCheckTool) Description() string {
	return "Run ESLint to check code for errors and warnings. Uses the project's eslint.config.mjs automatically. Returns structured results with file, line, severity, and message."
}
func (t *ESLintCheckTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "File or directory to lint. Defaults to 'frontend/src/'.",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results to return. Default: 50.",
				"default":     50,
			},
		},
	}
}

func (t *ESLintCheckTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	lintPath := "frontend/src/"
	if p, ok := args["path"].(string); ok && p != "" {
		lintPath = p
	}

	maxResults := 50
	if mr, ok := args["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
	}

	eslintBin, err := findBinary(t.workspace, "eslint", "node_modules/.bin/eslint")
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}

	absPath := lintPath
	if !filepath.IsAbs(lintPath) {
		absPath = filepath.Join(t.workspace, lintPath)
	}

	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, eslintBin,
		"--format", "json",
		absPath,
	)
	cmd.Dir = t.workspace

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// ESLint exits with non-zero when there are errors, so we ignore the error
	cmd.Run()

	if stdout.Len() == 0 {
		errMsg := stderr.String()
		if len(errMsg) > 3000 {
			errMsg = errMsg[:3000]
		}
		if errMsg != "" {
			return toolshared.ErrorResult(fmt.Sprintf("eslint error: %s", errMsg))
		}
		return toolshared.NewToolResult("No ESLint issues found.")
	}

	// Parse JSON output
	type eslintMessage struct {
		Line     int    `json:"line"`
		Column   int    `json:"column"`
		Severity int    `json:"severity"`
		Message  string `json:"message"`
		RuleID   string `json:"ruleId"`
	}
	type eslintResult struct {
		FilePath string          `json:"filePath"`
		Messages []eslintMessage `json:"messages"`
	}

	// Simple JSON parsing without importing encoding/json for the struct
	// We'll format the output directly
	output := stdout.String()
	totalErrors := 0
	totalWarnings := 0

	// Count issues from the JSON
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"severity": 2`) {
			totalErrors++
		} else if strings.Contains(line, `"severity": 1`) {
			totalWarnings++
		}
	}

	// Format a summary
	if totalErrors == 0 && totalWarnings == 0 {
		return toolshared.NewToolResult("No ESLint issues found.")
	}

	// Run again with compact format for readable output
	cmd2 := exec.CommandContext(ctx, eslintBin, absPath)
	cmd2.Dir = t.workspace
	var stdout2 bytes.Buffer
	cmd2.Stdout = &stdout2
	cmd2.Stderr = &stderr
	cmd2.Run()

	compactOutput := stdout2.String()
	if len(compactOutput) > 20000 {
		compactOutput = compactOutput[:20000] + "\n[TRUNCATED]"
	}

	result := fmt.Sprintf("ESLint Results: %d error(s), %d warning(s)\n\n%s",
		totalErrors, totalWarnings, compactOutput)

	if totalErrors+totalWarnings > maxResults {
		result += fmt.Sprintf("\n[Showing first %d issues. Refine your path to see more.]", maxResults)
	}

	return toolshared.NewToolResult(result)
}

// --- eslint_fix ---

// ESLintFixTool runs ESLint with --fix to auto-correct issues.
type ESLintFixTool struct {
	workspace string
}

func NewESLintFixTool(workspace string) *ESLintFixTool {
	return &ESLintFixTool{workspace: workspace}
}

func (t *ESLintFixTool) Name() string { return "eslint_fix" }
func (t *ESLintFixTool) Description() string {
	return "Run ESLint with --fix to auto-correct fixable issues. Uses the project's eslint.config.mjs automatically."
}
func (t *ESLintFixTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "File or directory to fix. Defaults to 'frontend/src/'.",
			},
		},
	}
}

func (t *ESLintFixTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	fixPath := "frontend/src/"
	if p, ok := args["path"].(string); ok && p != "" {
		fixPath = p
	}

	eslintBin, err := findBinary(t.workspace, "eslint", "node_modules/.bin/eslint")
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}

	absPath := fixPath
	if !filepath.IsAbs(fixPath) {
		absPath = filepath.Join(t.workspace, fixPath)
	}

	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// First run without fix to count before
	cmdBefore := exec.CommandContext(ctx, eslintBin, absPath)
	cmdBefore.Dir = t.workspace
	var beforeOut bytes.Buffer
	cmdBefore.Stdout = &beforeOut
	cmdBefore.Run()

	// Run with --fix
	cmd := exec.CommandContext(ctx, eslintBin, "--fix", absPath)
	cmd.Dir = t.workspace
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()

	// Run again to count remaining
	cmdAfter := exec.CommandContext(ctx, eslintBin, absPath)
	cmdAfter.Dir = t.workspace
	var afterOut bytes.Buffer
	cmdAfter.Stdout = &afterOut
	cmdAfter.Run()

	beforeCount := strings.Count(beforeOut.String(), "\n")
	afterCount := strings.Count(afterOut.String(), "\n")
	fixed := beforeCount - afterCount

	if fixed <= 0 && afterCount == 0 {
		return toolshared.NewToolResult("No fixable issues found.")
	}

	result := fmt.Sprintf("ESLint fix completed.\nIssues before: ~%d\nIssues after: ~%d\nFixed: ~%d",
		beforeCount, afterCount, fixed)

	if afterCount > 0 {
		remaining := afterOut.String()
		if len(remaining) > 10000 {
			remaining = remaining[:10000] + "\n[TRUNCATED]"
		}
		result += fmt.Sprintf("\n\nRemaining issues:\n%s", remaining)
	}

	return toolshared.NewToolResult(result)
}

// --- helpers ---

// findBinary locates a binary: first checks node_modules/.bin, then PATH.
func findBinary(workspace, name, localPath string) (string, error) {
	local := filepath.Join(workspace, localPath)
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}

	// Check if npx is available
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	// Fallback to npx
	if _, err := exec.LookPath("npx"); err == nil {
		return "npx", nil
	}

	return "", fmt.Errorf("%s not found. Install it with: npm install --save-dev %s", name, name)
}

// parserToExt maps Prettier parser names to file extensions.
func parserToExt(parser string) string {
	switch parser {
	case "typescript":
		return "ts"
	case "tsx":
		return "tsx"
	case "javascript":
		return "js"
	case "jsx":
		return "jsx"
	case "json":
		return "json"
	case "css":
		return "css"
	case "scss":
		return "scss"
	case "markdown", "md":
		return "md"
	case "html":
		return "html"
	case "yaml":
		return "yml"
	default:
		return "ts"
	}
}
