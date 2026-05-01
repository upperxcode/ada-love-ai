package vk

import (
	"ada-love-ai/pkg/bus"
	"ada-love-ai/pkg/channels"
	"ada-love-ai/pkg/config"
)

func init() {
	channels.RegisterFactory(
		config.ChannelVK,
		func(channelName, channelType string, cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
			bc := cfg.Channels[channelName]
			if bc == nil {
				return nil, channels.ErrSendFailed
			}
			return NewVKChannel(channelName, bc, b)
		},
	)
}
