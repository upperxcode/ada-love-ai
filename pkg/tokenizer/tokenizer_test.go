package tokenizer

import (
	"testing"

	"ada-love-ai/pkg/providers"
)

func TestEstimateMessageTokens(t *testing.T) {
	t.Run("empty message", func(t *testing.T) {
		msg := providers.Message{}
		got := EstimateMessageTokens(msg)
		want := 12 * 2 / 5
		if got != want {
			t.Errorf("EstimateMessageTokens() = %d, want %d", got, want)
		}
	})

	t.Run("simple text message", func(t *testing.T) {
		msg := providers.Message{
			Content: "hello world",
		}
		got := EstimateMessageTokens(msg)
		if got <= 0 {
			t.Error("should return positive value")
		}
	})

	t.Run("message with reasoning", func(t *testing.T) {
		msg := providers.Message{
			Content:          "answer",
			ReasoningContent: "thinking step by step",
		}
		got := EstimateMessageTokens(msg)
		if got <= 10 {
			t.Errorf("EstimateMessageTokens() = %d, want > 10", got)
		}
	})
}

func TestEstimateMessageTokensWithToolCallID(t *testing.T) {
	msg := providers.Message{
		Content:    "result",
		ToolCallID: "call_abc",
	}
	got := EstimateMessageTokens(msg)
	if got <= 5 {
		t.Errorf("EstimateMessageTokens() = %d, want > 5", got)
	}
}
