package fstools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toolshared "ada-love-ai/pkg/tools/shared"
)

// LocateFilesTool searches for files matching a glob pattern within the workspace.
type LocateFilesTool struct {
	workspace string
	restrict  bool
	patterns  []*regexp.Regexp
}

func NewLocateFilesTool(workspace string, restrict bool, allowPaths ...[]*regexp.Regexp) *LocateFilesTool {
	var patterns []*regexp.Regexp
	if len(allowPaths) > 0 {
		patterns = allowPaths[0]
	}
	return &LocateFilesTool{workspace: workspace, restrict: restrict, patterns: patterns}
}

func (t *LocateFilesTool) Name() string {
	return "locate_files"
}

func (t *LocateFilesTool) Description() string {
	return "Search for files matching a glob pattern (e.g. '*.py', '**/*.go', 'src/**/*.ts'). Returns matching file paths sorted by modification time. Use '**' to search recursively."
}

func (t *LocateFilesTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Glob pattern to match files (e.g. '*.go', '**/*.test.ts', 'src/**/*.py'). Use '**' for recursive search.",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Directory to search in. Defaults to workspace root.",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results to return. Default: 100.",
				"default":     100,
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *LocateFilesTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
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

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
	}

	// Validate the search directory is within workspace when restricted
	if t.restrict {
		absSearchDir, err := validatePathWithAllowPaths(searchDir, t.workspace, true, t.patterns)
		if err != nil {
			return toolshared.ErrorResult(err.Error())
		}
		searchDir = absSearchDir
	}

	// Build the full glob pattern
	fullPattern := filepath.Join(searchDir, pattern)

	// Use filepath.Glob for non-recursive patterns
	// For recursive (**), we need manual walk
	var matches []string
	var err error

	if strings.Contains(pattern, "**") {
		matches, err = recursiveGlob(searchDir, pattern, maxResults)
	} else {
		matches, err = filepath.Glob(fullPattern)
	}

	if err != nil {
		return toolshared.ErrorResult(fmt.Sprintf("glob search failed: %v", err))
	}

	if len(matches) == 0 {
		return toolshared.NewToolResult("No files found matching pattern: " + pattern)
	}

	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	// Make paths relative to workspace for readability
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d file(s) matching '%s':\n", len(matches), pattern))
	for _, m := range matches {
		rel, err := filepath.Rel(t.workspace, m)
		if err != nil {
			rel = m
		}
		result.WriteString(rel + "\n")
	}

	return toolshared.NewToolResult(result.String())
}

// recursiveGlob handles ** patterns by walking the directory tree.
func recursiveGlob(root, pattern string, maxResults int) ([]string, error) {
	// Split pattern on ** to get prefix and suffix
	parts := strings.SplitN(pattern, "**", 2)
	prefix := strings.TrimSuffix(parts[0], string(os.PathSeparator))
	suffix := ""
	if len(parts) > 1 {
		suffix = strings.TrimPrefix(parts[1], string(os.PathSeparator))
	}

	var matches []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if len(matches) >= maxResults {
			return filepath.SkipAll
		}
		if d.IsDir() {
			// Skip hidden directories and common non-essential dirs
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// If there's a prefix, check the path starts correctly
		if prefix != "" {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}
			if !strings.HasPrefix(rel, prefix) && prefix != "." {
				return nil
			}
		}

		// If there's a suffix, match it
		if suffix != "" {
			matched, err := filepath.Match(suffix, filepath.Base(path))
			if err != nil || !matched {
				// Also try matching the relative path against the suffix pattern
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return nil
				}
				if prefix != "" {
					rel = strings.TrimPrefix(rel, prefix)
					rel = strings.TrimPrefix(rel, string(os.PathSeparator))
				}
				matched, _ = filepath.Match(suffix, rel)
				if !matched {
					return nil
				}
			}
		}

		matches = append(matches, path)
		return nil
	})

	return matches, err
}
