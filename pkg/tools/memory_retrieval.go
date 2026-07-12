package tools

import (
	"context"
	"fmt"
	"strings"

	"ada-love-ai/pkg/agent/interfaces"
	toolshared "ada-love-ai/pkg/tools/shared"
)

// GetAgentMemoryTool retrieves saved memories from the long-term memory store.
type GetAgentMemoryTool struct {
	workspace string
	memStore  interfaces.MemoryStore
}

func NewGetAgentMemoryTool(workspace string, memStore interfaces.MemoryStore) *GetAgentMemoryTool {
	return &GetAgentMemoryTool{
		workspace: workspace,
		memStore:  memStore,
	}
}

func (t *GetAgentMemoryTool) Name() string { return "get_agent_memory" }
func (t *GetAgentMemoryTool) Description() string {
	return "Retrieve previously saved memories (facts, preferences, decisions) from the long-term memory database. Returns memories sorted by importance."
}
func (t *GetAgentMemoryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Optional search term to filter memories. If omitted, returns all memories.",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of memories to return. Default: 20.",
				"default":     20,
			},
		},
	}
}

func (t *GetAgentMemoryTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	if t.memStore == nil {
		return toolshared.ErrorResult("Memory store not initialized")
	}

	query := ""
	if q, ok := args["query"].(string); ok {
		query = strings.ToLower(q)
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	memories, err := t.memStore.GetMemories(t.workspace)
	if err != nil {
		return toolshared.ErrorResult(fmt.Sprintf("failed to retrieve memories: %v", err))
	}

	if len(memories) == 0 {
		return toolshared.NewToolResult("No memories saved yet.")
	}

	// Filter by query if provided
	var filtered []interfaces.MemoryEntry
	for _, m := range memories {
		if query == "" || strings.Contains(strings.ToLower(m.Content), query) {
			filtered = append(filtered, m)
		}
	}

	if len(filtered) == 0 {
		return toolshared.NewToolResult(fmt.Sprintf("No memories matching query: %s", query))
	}

	// Sort by importance (descending)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Importance > filtered[i].Importance {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d memory(ies):\n\n", len(filtered)))
	for i, m := range filtered {
		result.WriteString(fmt.Sprintf("%d. [importance: %d] %s", i+1, m.Importance, m.Content))
		if !m.CreatedAt.IsZero() {
			result.WriteString(fmt.Sprintf(" (saved: %s)", m.CreatedAt.Format("2006-01-02")))
		}
		result.WriteString("\n")
	}

	return toolshared.NewToolResult(result.String())
}
