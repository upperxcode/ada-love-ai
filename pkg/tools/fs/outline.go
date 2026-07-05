package fstools

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	toolshared "ada-love-ai/pkg/tools/shared"
)

// ViewFileOutlineTool extracts the structure of a file (functions, classes, methods)
// without reading the full content. Supports Go, Python, TypeScript, JavaScript, Rust, Java, C/C++.
type ViewFileOutlineTool struct {
	fs fileSystem
}

func NewViewFileOutlineTool(workspace string, restrict bool, allowPaths ...[]*regexp.Regexp) *ViewFileOutlineTool {
	var patterns []*regexp.Regexp
	if len(allowPaths) > 0 {
		patterns = allowPaths[0]
	}
	return &ViewFileOutlineTool{fs: buildFs(workspace, restrict, patterns)}
}

func (t *ViewFileOutlineTool) Name() string {
	return "view_file_outline"
}

func (t *ViewFileOutlineTool) Description() string {
	return "Extract the structure of a file: lists functions, classes, methods, interfaces, types, and constants with their line numbers. Supports Go, Python, TypeScript, JavaScript, Rust, Java, C/C++. Does NOT read the full file content."
}

func (t *ViewFileOutlineTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to outline.",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ViewFileOutlineTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return toolshared.ErrorResult("path is required")
	}

	file, err := t.fs.Open(path)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	entries := extractOutline(file, ext)

	if len(entries) == 0 {
		return toolshared.NewToolResult(fmt.Sprintf("No structural elements found in %s", filepath.Base(path)))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("[file outline: %s]\n", filepath.Base(path)))
	for _, e := range entries {
		result.WriteString(fmt.Sprintf("  L%d | %s\n", e.line, e.text))
	}

	return toolshared.NewToolResult(result.String())
}

type outlineEntry struct {
	line int
	text string
}

// outlinePatterns maps file extensions to their structural element regexes.
var outlinePatterns = map[string][]*regexp.Regexp{
	".go": {
		regexp.MustCompile(`^func\s+(?:\([^)]+\)\s+)?(\w+)`),
		regexp.MustCompile(`^type\s+(\w+)\s+(struct|interface)\s*\{`),
		regexp.MustCompile(`^var\s+(\w+)`),
		regexp.MustCompile(`^const\s+(\w+)`),
	},
	".py": {
		regexp.MustCompile(`^class\s+(\w+)`),
		regexp.MustCompile(`^\s*def\s+(\w+)`),
	},
	".ts": {
		regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
		regexp.MustCompile(`^(?:export\s+)?class\s+(\w+)`),
		regexp.MustCompile(`^(?:export\s+)?interface\s+(\w+)`),
		regexp.MustCompile(`^(?:export\s+)?type\s+(\w+)`),
		regexp.MustCompile(`^(?:export\s+)?enum\s+(\w+)`),
	},
	".tsx": nil, // same as .ts
	".js": {
		regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+(\w+)`),
		regexp.MustCompile(`^(?:export\s+)?class\s+(\w+)`),
	},
	".jsx": nil, // same as .js
	".rs": {
		regexp.MustCompile(`^(?:pub\s+)?fn\s+(\w+)`),
		regexp.MustCompile(`^(?:pub\s+)?struct\s+(\w+)`),
		regexp.MustCompile(`^(?:pub\s+)?enum\s+(\w+)`),
		regexp.MustCompile(`^(?:pub\s+)?trait\s+(\w+)`),
		regexp.MustCompile(`^(?:pub\s+)?impl\s+(?:\w+\s+for\s+)?(\w+)`),
	},
	".java": {
		regexp.MustCompile(`^(?:public\s+)?(?:abstract\s+)?class\s+(\w+)`),
		regexp.MustCompile(`^(?:public\s+)?interface\s+(\w+)`),
		regexp.MustCompile(`^\s+(?:public|private|protected)?\s*(?:static\s+)?(?:\w+\s+)+(\w+)\s*\(`),
	},
	".c": {
		regexp.MustCompile(`^(?:\w+\s+)*(\w+)\s*\([^)]*\)\s*\{`),
	},
	".h":   nil,
	".cpp": nil,
	".hpp": nil,
}

func init() {
	// Share patterns for variant extensions
	outlinePatterns[".tsx"] = outlinePatterns[".ts"]
	outlinePatterns[".jsx"] = outlinePatterns[".js"]
	outlinePatterns[".cpp"] = outlinePatterns[".c"]
	outlinePatterns[".hpp"] = outlinePatterns[".c"]
	outlinePatterns[".h"] = outlinePatterns[".c"]
}

func extractOutline(file io.Reader, ext string) []outlineEntry {
	patterns, ok := outlinePatterns[ext]
	if !ok || len(patterns) == 0 {
		return nil
	}

	var entries []outlineEntry
	lineNum := 0

	// Use a simple line reader approach
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		for _, pat := range patterns {
			if match := pat.FindStringSubmatch(line); match != nil {
				name := match[1]
				// For Go methods, include the receiver
				if ext == ".go" && strings.HasPrefix(line, "func (") {
					receiverEnd := strings.Index(line, ")")
					if receiverEnd > 0 {
						receiver := strings.TrimSpace(line[6:receiverEnd])
						// Clean receiver type
						parts := strings.Fields(receiver)
						if len(parts) >= 2 {
							name = fmt.Sprintf("(%s).%s", parts[len(parts)-1], match[1])
						} else {
							name = fmt.Sprintf("(%s).%s", receiver, match[1])
						}
					}
				}

				// For Go type declarations, include the kind
				if ext == ".go" && len(match) > 2 {
					name = fmt.Sprintf("%s %s", match[2], match[1])
				}

				entries = append(entries, outlineEntry{line: lineNum, text: name})
				break
			}
		}
	}

	return entries
}
