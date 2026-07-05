package fstools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toolshared "ada-love-ai/pkg/tools/shared"
)

const (
	maxGrepResults      = 50
	maxGrepFileBytes    = 1024 * 1024 // 1MB per file
	defaultGrepContext  = 0
	maxGrepContextLines = 5
)

// GrepSearchTool searches for text or regex patterns in files within the workspace.
type GrepSearchTool struct {
	workspace string
	restrict  bool
	patterns  []*regexp.Regexp
}

func NewGrepSearchTool(workspace string, restrict bool, allowPaths ...[]*regexp.Regexp) *GrepSearchTool {
	var patterns []*regexp.Regexp
	if len(allowPaths) > 0 {
		patterns = allowPaths[0]
	}
	return &GrepSearchTool{workspace: workspace, restrict: restrict, patterns: patterns}
}

func (t *GrepSearchTool) Name() string {
	return "grep_search"
}

func (t *GrepSearchTool) Description() string {
	return "Search for text or regex patterns in files within the workspace. Returns matching lines with file paths and line numbers. Useful for finding usages, definitions, or references across the codebase."
}

func (t *GrepSearchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Text or regex pattern to search for.",
			},
			"include": map[string]any{
				"type":        "string",
				"description": "File pattern to filter (e.g. '*.go', '*.py', '*.{ts,tsx}').",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Subdirectory to search in. Defaults to workspace root.",
			},
			"context_lines": map[string]any{
				"type":        "integer",
				"description": "Number of context lines before and after each match (0-5). Default: 0.",
				"default":     0,
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of matching lines to return. Default: 50.",
				"default":     50,
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepSearchTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return toolshared.ErrorResult("pattern is required")
	}

	searchDir := t.workspace
	if dir, ok := args["path"].(string); ok && dir != "" {
		searchDir = dir
		if !filepath.IsAbs(searchDir) {
			searchDir = filepath.Join(t.workspace, searchDir)
		}
	}

	include := ""
	if inc, ok := args["include"].(string); ok {
		include = inc
	}

	contextLines := defaultGrepContext
	if cl, ok := args["context_lines"].(float64); ok {
		contextLines = int(cl)
		if contextLines > maxGrepContextLines {
			contextLines = maxGrepContextLines
		}
	}

	maxResults := maxGrepResults
	if mr, ok := args["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
	}

	// Compile the search pattern as regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fall back to literal string search
		re = regexp.MustCompile(regexp.QuoteMeta(pattern))
	}

	// Validate path
	if t.restrict {
		absPath, err := validatePathWithAllowPaths(searchDir, t.workspace, true, t.patterns)
		if err != nil {
			return toolshared.ErrorResult(err.Error())
		}
		searchDir = absPath
	}

	type matchResult struct {
		file    string
		lineNum int
		line    string
		context []string
	}

	var results []matchResult
	totalMatches := 0

	err = filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if totalMatches >= maxResults {
			return filepath.SkipAll
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check include filter
		if include != "" {
			matched, _ := filepath.Match(include, filepath.Base(path))
			if !matched {
				// Try matching against extension patterns like *.{ts,tsx}
				if !matchIncludePattern(include, filepath.Base(path)) {
					return nil
				}
			}
		}

		// Skip binary files by extension
		ext := strings.ToLower(filepath.Ext(path))
		if isBinaryExt(ext) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		// Check file size
		info, err := file.Stat()
		if err != nil || info.Size() > maxGrepFileBytes {
			return nil
		}

		// Read all lines for context support
		var lines []string
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		relPath, _ := filepath.Rel(t.workspace, path)

		for i, line := range lines {
			if totalMatches >= maxResults {
				break
			}
			if re.MatchString(line) {
				totalMatches++
				result := matchResult{
					file:    relPath,
					lineNum: i + 1,
					line:    line,
				}
				// Add context lines
				if contextLines > 0 {
					start := i - contextLines
					if start < 0 {
						start = 0
					}
					end := i + contextLines + 1
					if end > len(lines) {
						end = len(lines)
					}
					var ctx []string
					for j := start; j < end; j++ {
						prefix := "  "
						if j == i {
							prefix = "> "
						}
						ctx = append(ctx, fmt.Sprintf("%s%d|%s", prefix, j+1, lines[j]))
					}
					result.context = ctx
				}
				results = append(results, result)
			}
		}

		return nil
	})

	if err != nil {
		return toolshared.ErrorResult(fmt.Sprintf("grep search failed: %v", err))
	}

	if len(results) == 0 {
		return toolshared.NewToolResult(fmt.Sprintf("No matches found for pattern: %s", pattern))
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("Found %d match(es) for '%s':\n\n", totalMatches, pattern))
	for _, r := range results {
		if len(r.context) > 0 {
			out.WriteString(fmt.Sprintf("%s:%d:\n", r.file, r.lineNum))
			for _, c := range r.context {
				out.WriteString(c + "\n")
			}
			out.WriteString("\n")
		} else {
			out.WriteString(fmt.Sprintf("%s:%d: %s\n", r.file, r.lineNum, r.line))
		}
	}

	if totalMatches >= maxResults {
		out.WriteString(fmt.Sprintf("\n[TRUNCATED - showing first %d results. Refine your search to see more.]\n", maxResults))
	}

	return toolshared.NewToolResult(out.String())
}

// matchIncludePattern handles complex include patterns like *.{ts,tsx}
func matchIncludePattern(include, name string) bool {
	// Expand {a,b} patterns
	if strings.Contains(include, "{") && strings.Contains(include, "}") {
		start := strings.Index(include, "{")
		end := strings.Index(include, "}")
		if start < end {
			prefix := include[:start]
			suffix := include[end+1:]
			alts := strings.Split(include[start+1:end], ",")
			for _, alt := range alts {
				pattern := prefix + strings.TrimSpace(alt) + suffix
				if matched, _ := filepath.Match(pattern, name); matched {
					return true
				}
			}
		}
	}
	return false
}

func isBinaryExt(ext string) bool {
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".obj": true, ".o": true, ".a": true,
		".zip": true, ".tar": true, ".gz": true, ".bz2": true,
		".7z": true, ".rar": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".bmp": true, ".ico": true, ".webp": true, ".svg": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true,
		".xlsx": true, ".ppt": true, ".pptx": true,
		".pyc": true, ".pyo": true, ".class": true,
		".sqlite": true, ".db": true,
	}
	return binaryExts[ext]
}
