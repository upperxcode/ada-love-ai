package whatsapp

import (
	"context"
	"testing"

	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/channels"
	"ada-love-ai/pkg/config"
)

func TestHandleIncomingMessage_DoesNotConsumeGenericCommandsLocally(t *testing.T) {
	messageBus := bus.NewMessageBus()
	ch := &WhatsAppChannel{
		BaseChannel: channels.NewBaseChannel("whatsapp", config.WhatsAppSettings{}, messageBus, nil),
		ctx:         context.Background(),
	}

	ch.handleIncomingMessage(map[string]any{
		"type":    "message",
		"id":      "mid1",
		"from":    "user1",
		"chat":    "chat1",
		"content": "/help",
	})

	inbound, ok := <-messageBus.InboundChan()
	if !ok {
		t.Fatal("expected inbound message to be forwarded")
	}
	if inbound.Channel != "whatsapp" {
		t.Fatalf("channel=%q", inbound.Channel)
	}
	if inbound.Content != "/help" {
		t.Fatalf("content=%q", inbound.Content)
	}
}

func TestWhatsAppMaxMessageLength(t *testing.T) {
	ch := &WhatsAppChannel{
		BaseChannel: channels.NewBaseChannel("whatsapp", config.WhatsAppSettings{}, bus.NewMessageBus(), nil,
			channels.WithMaxMessageLength(65536)),
	}
	maxLen := ch.MaxMessageLength()
	if maxLen != 65536 {
		t.Errorf("MaxMessageLength() = %d, want 65536", maxLen)
	}
}

func TestWhatsAppSplitMessage(t *testing.T) {
	tests := []struct {
		name    string
		content string
		maxLen  int
		want    int
	}{
		{
			name:    "short message",
			content: "hello",
			maxLen:  65536,
			want:    1,
		},
		{
			name:    "needs split",
			content: string(make([]byte, 70000)),
			maxLen:  65536,
			want:    2,
		},
		{
			name:    "exact length",
			content: string(make([]byte, 65536)),
			maxLen:  65536,
			want:    1,
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

func TestNewWhatsAppChannel(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		bc := &config.Channel{
			Enabled:   true,
			Type:      "whatsapp",
			AllowFrom: []string{"*"},
		}
		cfg := &config.WhatsAppSettings{
			BridgeURL: "ws://localhost:8080",
		}
		ch, err := NewWhatsAppChannel(bc, cfg, bus.NewMessageBus())
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
