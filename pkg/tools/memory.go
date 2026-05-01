package tools

import (
	"context"
	"fmt"

	"ada-love-ai/pkg/agent/interfaces"
	toolshared "ada-love-ai/pkg/tools/shared"
)

type SaveMemoryTool struct {
	workspace string
	memStore  interfaces.MemoryStore
}

func NewSaveMemoryTool(workspace string, memStore interfaces.MemoryStore) *SaveMemoryTool {
	return &SaveMemoryTool{
		workspace: workspace,
		memStore:  memStore,
	}
}

func (t *SaveMemoryTool) Name() string {
	return "tool_save_memory"
}

func (t *SaveMemoryTool) Description() string {
	return "Saves an important piece of information to the long-term memory database. Use this for facts, user preferences, or project decisions that should persist across sessions."
}

func (t *SaveMemoryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"content": map[string]any{
				"type":        "string",
				"description": "The information to remember.",
			},
			"importance": map[string]any{
				"type":        "integer",
				"description": "Importance level (1-10). Default is 5.",
				"minimum":     1,
				"maximum":     10,
			},
		},
		"required": []string{"content"},
	}
}

func (t *SaveMemoryTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	content, ok := args["content"].(string)
	if !ok || content == "" {
		return toolshared.ErrorResult("content is required")
	}

	importance := 5
	if imp, ok := args["importance"].(float64); ok {
		importance = int(imp)
	}

	if t.memStore == nil {
		return toolshared.NewToolResult("Memory store not initialized")
	}

	err := t.memStore.SaveMemory(t.workspace, content, importance)
	if err != nil {
		return toolshared.ErrorResult(fmt.Sprintf("failed to save memory: %v", err))
	}

	return toolshared.NewToolResult(fmt.Sprintf("Memory saved successfully to database: %s", content))
}
