package onebot

import (
	"encoding/json"
	"testing"

	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/channels"
	"ada-love-ai/pkg/config"
)

func TestNewOneBotChannel(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		bc := &config.Channel{
			Enabled:   true,
			Type:      "onebot",
			AllowFrom: []string{"*"},
		}
		cfg := &config.OneBotSettings{
			WSUrl: "ws://localhost:8080/ws",
		}
		ch, err := NewOneBotChannel(bc, cfg, bus.NewMessageBus())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ch == nil {
			t.Fatal("channel should not be nil")
		}
		if ch.IsRunning() {
			t.Error("new channel should not be running")
		}
	})
}

func TestOneBotMaxMessageLength(t *testing.T) {
	ch := &OneBotChannel{
		BaseChannel: channels.NewBaseChannel("onebot", config.OneBotSettings{}, bus.NewMessageBus(), nil),
	}
	maxLen := ch.MaxMessageLength()
	if maxLen != 0 {
		t.Errorf("MaxMessageLength() = %d, want 0 (no limit)", maxLen)
	}
}

func TestIsAPIResponse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "ok string",
			input: `"ok"`,
			want:  true,
		},
		{
			name:  "failed string",
			input: `"failed"`,
			want:  true,
		},
		{
			name:  "online status",
			input: `{"online":true}`,
			want:  true,
		},
		{
			name:  "good status",
			input: `{"good":true}`,
			want:  true,
		},
		{
			name:  "empty",
			input: ``,
			want:  false,
		},
		{
			name:  "random string",
			input: `"hello"`,
			want:  false,
		},
		{
			name:  "random object",
			input: `{"foo":"bar"}`,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAPIResponse(json.RawMessage(tt.input))
			if got != tt.want {
				t.Errorf("isAPIResponse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestOneBotSplitMessage(t *testing.T) {
	tests := []struct {
		name    string
		content string
		maxLen  int
		want    int
	}{
		{
			name:    "short message",
			content: "hello",
			maxLen:  5000,
			want:    1,
		},
		{
			name:    "needs split",
			content: string(make([]byte, 6000)),
			maxLen:  5000,
			want:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := channels.SplitMessage(tt.content, tt.maxLen)
			if len(got) != tt.want {
				t.Errorf("SplitMessage() got %d parts, want %d parts", len(got), tt.want)
			}
		})
	}
}