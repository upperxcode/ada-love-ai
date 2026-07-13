package toolshared

import (
	"testing"
)

func TestMessageFields(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "hello",
	}
	if msg.Role != "user" {
		t.Errorf("Role = %q, want user", msg.Role)
	}
	if msg.Content != "hello" {
		t.Errorf("Content = %q, want hello", msg.Content)
	}
}

func TestToolCallFields(t *testing.T) {
	tc := ToolCall{
		ID:   "call_123",
		Type: "function",
		Function: &FunctionCall{
			Name:      "test_func",
			Arguments: `{"arg1":"value1"}`,
		},
	}
	if tc.ID != "call_123" {
		t.Errorf("ID = %q, want call_123", tc.ID)
	}
	if tc.Function == nil {
		t.Fatal("Function should not be nil")
	}
	if tc.Function.Name != "test_func" {
		t.Errorf("Function.Name = %q, want test_func", tc.Function.Name)
	}
}

func TestLLMResponseFields(t *testing.T) {
	resp := LLMResponse{
		Content:      "response text",
		FinishReason: "stop",
		Usage: &UsageInfo{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}
	if resp.Content != "response text" {
		t.Errorf("Content = %q, want response text", resp.Content)
	}
	if resp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if resp.Usage.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want 150", resp.Usage.TotalTokens)
	}
}

func TestToolDefinitionFields(t *testing.T) {
	def := ToolDefinition{
		Type: "function",
		Function: ToolFunctionDefinition{
			Name:        "test_tool",
			Description: "A test tool",
		},
	}
	if def.Type != "function" {
		t.Errorf("Type = %q, want function", def.Type)
	}
	if def.Function.Name != "test_tool" {
		t.Errorf("Function.Name = %q, want test_tool", def.Function.Name)
	}
}
