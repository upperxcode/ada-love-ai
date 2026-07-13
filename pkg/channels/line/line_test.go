package line

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/channels"
	"ada-love-ai/pkg/config"
)

func TestWebhookRejectsOversizedBody(t *testing.T) {
	ch := &LINEChannel{}

	oversized := bytes.Repeat([]byte("A"), maxWebhookBodySize+1)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(oversized))
	rec := httptest.NewRecorder()

	ch.webhookHandler(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
}

func TestWebhookAcceptsMaxBodySize(t *testing.T) {
	ch := &LINEChannel{}

	body := bytes.Repeat([]byte("A"), maxWebhookBodySize)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	ch.webhookHandler(rec, req)

	// Missing signature should be rejected, but the body size should not trigger 413.
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestWebhookRejectsOversizedBodyBeforeSignatureCheck(t *testing.T) {
	ch := &LINEChannel{}

	oversized := bytes.Repeat([]byte("A"), maxWebhookBodySize+1)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(oversized))
	req.Header.Set("X-Line-Signature", "invalidsignature")
	rec := httptest.NewRecorder()

	ch.webhookHandler(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
}

func TestWebhookRejectsNonPostMethod(t *testing.T) {
	ch := &LINEChannel{}

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rec := httptest.NewRecorder()

	ch.webhookHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestWebhookRejectsInvalidSignature(t *testing.T) {
	ch := &LINEChannel{
		config: &config.LINESettings{},
	}

	body := `{"events":[]}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	req.Header.Set("X-Line-Signature", "invalidsignature")
	rec := httptest.NewRecorder()

	ch.webhookHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestResolveChatID(t *testing.T) {
	tests := []struct {
		name     string
		source   lineSource
		expected string
	}{
		{
			name: "user source",
			source: lineSource{
				Type:   "user",
				UserID: "U123456",
			},
			expected: "U123456",
		},
		{
			name: "group source",
			source: lineSource{
				Type:    "group",
				UserID:  "U123456",
				GroupID: "C987654",
			},
			expected: "C987654",
		},
		{
			name: "room source",
			source: lineSource{
				Type:   "room",
				UserID: "U123456",
				RoomID: "R111111",
			},
			expected: "R111111",
		},
		{
			name: "unknown source falls back to user",
			source: lineSource{
				Type:   "unknown",
				UserID: "U123456",
			},
			expected: "U123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &LINEChannel{}
			result := ch.resolveChatID(tt.source)
			if result != tt.expected {
				t.Errorf("resolveChatID() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStripBotMention(t *testing.T) {
	ch := &LINEChannel{
		botUserID:      "U123456",
		botDisplayName: "MyBot",
	}

	tests := []struct {
		name     string
		text     string
		msg      lineMessage
		expected string
	}{
		{
			name:     "no mention",
			text:     "Hello world",
			msg:      lineMessage{},
			expected: "Hello world",
		},
		{
			name:     "strip @mention by display name",
			text:     "@MyBot hello",
			msg:      lineMessage{Text: "@MyBot hello"},
			expected: "hello",
		},
		{
			name:     "mention at end",
			text:     "hello @MyBot",
			msg:      lineMessage{Text: "hello @MyBot"},
			expected: "hello",
		},
		{
			name:     "mention in middle",
			text:     "hello @MyBot how are you",
			msg:      lineMessage{Text: "hello @MyBot how are you"},
			expected: "hello  how are you",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ch.stripBotMention(tt.text, tt.msg)
			if result != tt.expected {
				t.Errorf("stripBotMention() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsBotMentioned(t *testing.T) {
	ch := &LINEChannel{
		botUserID:      "U123456",
		botDisplayName: "MyBot",
	}

	tests := []struct {
		name     string
		msg      lineMessage
		expected bool
	}{
		{
			name: "no mention",
			msg: lineMessage{
				Text: "Hello world",
			},
			expected: false,
		},
		{
			name: "mention by user id",
			msg: lineMessage{
				Text: "Hello",
				Mention: &struct {
					Mentionees []lineMentionee `json:"mentionees"`
				}{
					Mentionees: []lineMentionee{
						{Index: 0, Length: 5, Type: "user", UserID: "U123456"},
					},
				},
			},
			expected: true,
		},
		{
			name: "mention all",
			msg: lineMessage{
				Text: "Hello everyone",
				Mention: &struct {
					Mentionees []lineMentionee `json:"mentionees"`
				}{
					Mentionees: []lineMentionee{
						{Index: 0, Length: 17, Type: "all", UserID: ""},
					},
				},
			},
			expected: true,
		},
		{
			name: "text mention by display name",
			msg: lineMessage{
				Text: "@MyBot hello",
			},
			expected: true,
		},
		{
			name: "no mention other user",
			msg: lineMessage{
				Text: "Hello",
				Mention: &struct {
					Mentionees []lineMentionee `json:"mentionees"`
				}{
					Mentionees: []lineMentionee{
						{Index: 0, Length: 5, Type: "user", UserID: "U999999"},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ch.isBotMentioned(tt.msg)
			if result != tt.expected {
				t.Errorf("isBotMentioned() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	ch := &LINEChannel{
		config: &config.LINESettings{
			ChannelSecret: *config.NewSecureString("test_secret"),
		},
	}

	body := []byte(`{"events":[]}`)

	t.Run("empty signature", func(t *testing.T) {
		if ch.verifySignature(body, "") {
			t.Error("expected false for empty signature")
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		if ch.verifySignature(body, "invalid_signature") {
			t.Error("expected false for invalid signature")
		}
	})
}

func TestWebhookPath(t *testing.T) {
	t.Run("custom path", func(t *testing.T) {
		ch := &LINEChannel{
			config: &config.LINESettings{
				WebhookPath: "/custom/webhook",
			},
		}
		if path := ch.WebhookPath(); path != "/custom/webhook" {
			t.Errorf("WebhookPath() = %q, want /custom/webhook", path)
		}
	})

	t.Run("default path", func(t *testing.T) {
		ch := &LINEChannel{
			config: &config.LINESettings{},
		}
		if path := ch.WebhookPath(); path != "/webhook/line" {
			t.Errorf("WebhookPath() = %q, want /webhook/line", path)
		}
	})
}

func TestBuildTextMessage(t *testing.T) {
	t.Run("without quote token", func(t *testing.T) {
		msg := buildTextMessage("hello", "")
		if msg["type"] != "text" || msg["text"] != "hello" {
			t.Errorf("buildTextMessage() = %v, want type=text, text=hello", msg)
		}
		if _, ok := msg["quoteToken"]; ok {
			t.Error("should not have quoteToken when empty")
		}
	})

	t.Run("with quote token", func(t *testing.T) {
		msg := buildTextMessage("hello", "quote_token_123")
		if msg["quoteToken"] != "quote_token_123" {
			t.Errorf("buildTextMessage() quoteToken = %q, want quote_token_123", msg["quoteToken"])
		}
	})
}

func TestNewLINEChannel(t *testing.T) {
	t.Run("missing channel_secret", func(t *testing.T) {
		bc := &config.Channel{
			Enabled: true,
			Type:    "line",
		}
		cfg := &config.LINESettings{
			ChannelAccessToken: *config.NewSecureString("token"),
		}
		_, err := NewLINEChannel(bc, cfg, bus.NewMessageBus())
		if err == nil {
			t.Error("expected error for missing channel_secret")
		}
	})

	t.Run("missing channel_access_token", func(t *testing.T) {
		bc := &config.Channel{
			Enabled: true,
			Type:    "line",
		}
		cfg := &config.LINESettings{
			ChannelSecret: *config.NewSecureString("secret"),
		}
		_, err := NewLINEChannel(bc, cfg, bus.NewMessageBus())
		if err == nil {
			t.Error("expected error for missing channel_access_token")
		}
	})

	t.Run("valid config", func(t *testing.T) {
		bc := &config.Channel{
			Enabled: true,
			Type:    "line",
		}
		cfg := &config.LINESettings{
			ChannelSecret:      *config.NewSecureString("secret"),
			ChannelAccessToken: *config.NewSecureString("token"),
		}
		ch, err := NewLINEChannel(bc, cfg, bus.NewMessageBus())
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

func TestLineMaxMessageLength(t *testing.T) {
	bc := &config.Channel{
		Enabled: true,
		Type:    "line",
	}
	cfg := &config.LINESettings{
		ChannelSecret:      *config.NewSecureString("secret"),
		ChannelAccessToken: *config.NewSecureString("token"),
	}
	ch, err := NewLINEChannel(bc, cfg, bus.NewMessageBus())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	maxLen := ch.MaxMessageLength()
	if maxLen != 5000 {
		t.Errorf("MaxMessageLength() = %d, want 5000", maxLen)
	}
}

func TestLineSplitMessage(t *testing.T) {
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
		{
			name:    "exact length",
			content: string(make([]byte, 5000)),
			maxLen:  5000,
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
