package maixcam

import (
	"encoding/json"
	"testing"

	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/channels"
	"ada-love-ai/pkg/config"
)

func TestNewMaixCamChannel(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		bc := &config.Channel{
			Enabled:   true,
			Type:      "maixcam",
			AllowFrom: []string{"*"},
		}
		cfg := &config.MaixCamSettings{
			Host: "localhost",
			Port: 8080,
		}
		ch, err := NewMaixCamChannel(bc, cfg, bus.NewMessageBus())
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

func TestMaixCamMaxMessageLength(t *testing.T) {
	ch := &MaixCamChannel{
		BaseChannel: channels.NewBaseChannel("maixcam", config.MaixCamSettings{}, bus.NewMessageBus(), nil),
	}
	maxLen := ch.MaxMessageLength()
	if maxLen != 0 {
		t.Errorf("MaxMessageLength() = %d, want 0 (no limit)", maxLen)
	}
}

func TestMaixCamMessageParsing(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		wantType string
		wantTips string
		wantOK   bool
	}{
		{
			name:     "valid text message",
			jsonStr:  `{"type":"text","tips":"hello","timestamp":1234567890}`,
			wantType: "text",
			wantTips: "hello",
			wantOK:   true,
		},
		{
			name:     "valid button message",
			jsonStr:  `{"type":"button","tips":"press","timestamp":1234567890}`,
			wantType: "button",
			wantTips: "press",
			wantOK:   true,
		},
		{
			name:   "invalid json",
			jsonStr: `{invalid`,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg MaixCamMessage
			err := json.Unmarshal([]byte(tt.jsonStr), &msg)
			if tt.wantOK {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if msg.Type != tt.wantType {
					t.Errorf("Type = %q, want %q", msg.Type, tt.wantType)
				}
				if msg.Tips != tt.wantTips {
					t.Errorf("Tips = %q, want %q", msg.Tips, tt.wantTips)
				}
			} else {
				if err == nil {
					t.Error("expected error for invalid json")
				}
			}
		})
	}
}