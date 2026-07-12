package messageutil

import (
	"testing"

	"ada-love-ai/pkg/providers/protocoltypes"
)

func TestIsTransientAssistantThoughtMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  protocoltypes.Message
		want bool
	}{
		{
			name: "transient thought message",
			msg: protocoltypes.Message{
				Role:             "assistant",
				Content:          "",
				ReasoningContent: "thinking...",
			},
			want: true,
		},
		{
			name: "normal assistant message with content",
			msg: protocoltypes.Message{
				Role:    "assistant",
				Content: "hello",
			},
			want: false,
		},
		{
			name: "user message",
			msg: protocoltypes.Message{
				Role:    "user",
				Content: "hi",
			},
			want: false,
		},
		{
			name: "assistant with content and reasoning",
			msg: protocoltypes.Message{
				Role:             "assistant",
				Content:          "answer",
				ReasoningContent: "thinking...",
			},
			want: false,
		},
		{
			name: "assistant with tool calls",
			msg: protocoltypes.Message{
				Role:             "assistant",
				Content:          "",
				ReasoningContent: "thinking...",
				ToolCalls:        []protocoltypes.ToolCall{{Name: "tool1"}},
			},
			want: false,
		},
		{
			name: "assistant with tool call id",
			msg: protocoltypes.Message{
				Role:             "assistant",
				Content:          "",
				ReasoningContent: "thinking...",
				ToolCallID:       "call_123",
			},
			want: false,
		},
		{
			name: "assistant with only whitespace content",
			msg: protocoltypes.Message{
				Role:             "assistant",
				Content:          "   ",
				ReasoningContent: "thinking...",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTransientAssistantThoughtMessage(tt.msg)
			if got != tt.want {
				t.Errorf("IsTransientAssistantThoughtMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterInvalidHistoryMessages(t *testing.T) {
	tests := []struct {
		name    string
		history []protocoltypes.Message
		wantLen int
	}{
		{
			name:    "empty history",
			history: []protocoltypes.Message{},
			wantLen: 0,
		},
		{
			name: "no transient messages",
			history: []protocoltypes.Message{
				{Role: "user", Content: "hello"},
				{Role: "assistant", Content: "hi"},
			},
			wantLen: 2,
		},
		{
			name: "filter transient message",
			history: []protocoltypes.Message{
				{Role: "user", Content: "hello"},
				{Role: "assistant", Content: "", ReasoningContent: "thinking..."},
				{Role: "assistant", Content: "answer"},
			},
			wantLen: 2,
		},
		{
			name: "all transient messages",
			history: []protocoltypes.Message{
				{Role: "assistant", Content: "", ReasoningContent: "thinking..."},
				{Role: "assistant", Content: "", ReasoningContent: "more thinking..."},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterInvalidHistoryMessages(tt.history)
			if len(got) != tt.wantLen {
				t.Errorf("FilterInvalidHistoryMessages() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}
