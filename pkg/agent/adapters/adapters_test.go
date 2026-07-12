package adapters

import (
	"context"
	"testing"

	"ada-love-ai/pkg/bus"
)

func TestNewMessageBus(t *testing.T) {
	msgBus := bus.NewMessageBus()
	adapter := NewMessageBus(msgBus)
	if adapter == nil {
		t.Error("NewMessageBus() should not return nil")
	}
}

func TestMessageBusAdapterPublishInbound(t *testing.T) {
	msgBus := bus.NewMessageBus()
	adapter := NewMessageBus(msgBus)

	ctx := context.Background()
	msg := bus.InboundMessage{
		Channel:    "test",
		ChatID:     "chat1",
		SenderID:   "user1",
		Content:    "hello",
		SessionKey: "session1",
	}

	err := adapter.PublishInbound(ctx, msg)
	if err != nil {
		t.Errorf("PublishInbound() error = %v", err)
	}
}

func TestMessageBusAdapterPublishOutbound(t *testing.T) {
	msgBus := bus.NewMessageBus()
	adapter := NewMessageBus(msgBus)

	ctx := context.Background()
	msg := bus.OutboundMessage{
		Channel: "test",
		ChatID:  "chat1",
		Content: "hello",
	}

	err := adapter.PublishOutbound(ctx, msg)
	if err != nil {
		t.Errorf("PublishOutbound() error = %v", err)
	}
}

func TestMessageBusAdapterInboundChan(t *testing.T) {
	msgBus := bus.NewMessageBus()
	adapter := NewMessageBus(msgBus)

	ch := adapter.InboundChan()
	if ch == nil {
		t.Error("InboundChan() should not return nil")
	}
}
