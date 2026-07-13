package tools

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

const maxKBResults = 20

// SearchKnowledgeBaseTool performs text search over local documentation files.
type SearchKnowledgeBaseTool struct {
	workspace string
}

func NewSearchKnowledgeBaseTool(workspace string) *SearchKnowledgeBaseTool {
	return &SearchKnowledgeBaseTool{workspace: workspace}
}

func (t *SearchKnowledgeBaseTool) Name() string { return "search_knowledge_base" }
func (t *SearchKnowledgeBaseTool) Description() string {
	return "Search local documentation and knowledge base files (markdown, text, docs/) for relevant information. Useful for finding answers in project docs, READMEs, wikis, or private documentation."
}
func (t *SearchKnowledgeBaseTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query. Matches against file content using keyword search.",
			},
			"docs_path": map[string]any{
				"type":        "string",
				"description": "Subdirectory to search in (e.g. 'docs/', 'wiki/'). Defaults to workspace root.",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results. Default: 10.",
				"default":     10,
			},
		},
		"required": []string{"query"},
	}
}

func (t *SearchKnowledgeBaseTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return toolshared.ErrorResult("query is required")
	}

	searchDir := t.workspace
	if dp, ok := args["docs_path"].(string); ok && dp != "" {
		searchDir = filepath.Join(t.workspace, dp)
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
		if limit > maxKBResults {
			limit = maxKBResults
		}
	}

	// Build search terms from query
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return toolshared.ErrorResult("query must contain at least one term")
	}

	type kbResult struct {
		file    string
		score   int
		snippet string
	}

	var results []kbResult
	docExts := map[string]bool{
		".md": true, ".txt": true, ".rst": true, ".adoc": true,
		".mdx": true, ".rdoc": true,
	}

	filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !docExts[ext] {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		relPath, _ := filepath.Rel(t.workspace, path)
		score := 0
		var bestSnippet string

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.ToLower(scanner.Text())

			// Count term matches
			lineScore := 0
			for _, term := range terms {
				if strings.Contains(line, term) {
					lineScore++
				}
			}

			if lineScore > 0 {
				score += lineScore
				if bestSnippet == "" || lineScore > 1 {
					bestSnippet = fmt.Sprintf("  L%d: %s", lineNum, strings.TrimSpace(scanner.Text()))
				}
			}
		}

		if score > 0 {
			results = append(results, kbResult{
				file:    relPath,
				score:   score,
				snippet: bestSnippet,
			})
		}

		return nil
	})

	if len(results) == 0 {
		return toolshared.NewToolResult(fmt.Sprintf("No documentation found matching: %s", query))
	}

	// Sort by score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("Knowledge base results for '%s' (%d match(es)):\n\n", query, len(results)))
	for i, r := range results {
		out.WriteString(fmt.Sprintf("%d. %s (relevance: %d)\n", i+1, r.file, r.score))
		if r.snippet != "" {
			out.WriteString(r.snippet + "\n")
		}
		out.WriteString("\n")
	}

	return toolshared.NewToolResult(out.String())
}

// Ensure we satisfy the interface
var _ = regexp.Compile
