package providers

import (
	"context"
	"fmt"
	"sync"

	"ada-love-ai/pkg/providers/openai_compat"
)

// FailoverProvider wraps an LLMProvider with multiple API keys.
// On HTTP error, it rotates to the next key and retries transparently.
type FailoverProvider struct {
	apiKeys      []string
	apiBase      string
	proxy        string
	opts         []openai_compat.Option
	providerName string
	currentIdx   int
	mu           sync.Mutex
}

// NewFailoverProvider creates a provider that tries multiple API keys.
func NewFailoverProvider(apiKeys []string, apiBase, proxy string, opts ...openai_compat.Option) *FailoverProvider {
	return &FailoverProvider{
		apiKeys: apiKeys,
		apiBase: apiBase,
		proxy:   proxy,
		opts:    opts,
	}
}

// SetProviderName sets the provider name on the underlying providers.
func (f *FailoverProvider) SetProviderName(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providerName = name
}

func (f *FailoverProvider) createProvider(idx int) *openai_compat.Provider {
	p := openai_compat.NewProvider(f.apiKeys[idx], f.apiBase, f.proxy, f.opts...)
	if f.providerName != "" {
		p.SetProviderName(f.providerName)
	}
	return p
}

// Chat tries each API key in order, failing over on HTTP error.
func (f *FailoverProvider) Chat(
	ctx context.Context,
	messages []Message,
	tools []ToolDefinition,
	model string,
	options map[string]any,
) (*LLMResponse, error) {
	f.mu.Lock()
	startIdx := f.currentIdx
	f.mu.Unlock()

	keys := len(f.apiKeys)
	var lastErr error

	for i := 0; i < keys; i++ {
		idx := (startIdx + i) % keys
		p := f.createProvider(idx)

		resp, err := p.Chat(ctx, messages, tools, model, options)
		if err == nil {
			f.mu.Lock()
			f.currentIdx = (idx + 1) % keys
			f.mu.Unlock()
			return resp, nil
		}

		lastErr = err
		fmt.Printf("[FailoverProvider] key %d/%d failed for %s: %v\n", i+1, keys, f.apiBase, err)
	}

	return nil, fmt.Errorf("all %d API keys exhausted for %s: %w", keys, f.apiBase, lastErr)
}

// ChatStream tries each API key in order, failing over on HTTP error.
func (f *FailoverProvider) ChatStream(
	ctx context.Context,
	messages []Message,
	tools []ToolDefinition,
	model string,
	options map[string]any,
	onChunk func(accumulated string),
) (*LLMResponse, error) {
	f.mu.Lock()
	startIdx := f.currentIdx
	f.mu.Unlock()

	keys := len(f.apiKeys)
	var lastErr error

	for i := 0; i < keys; i++ {
		idx := (startIdx + i) % keys
		p := f.createProvider(idx)

		resp, err := p.ChatStream(ctx, messages, tools, model, options, onChunk)
		if err == nil {
			f.mu.Lock()
			f.currentIdx = (idx + 1) % keys
			f.mu.Unlock()
			return resp, nil
		}

		lastErr = err
		fmt.Printf("[FailoverProvider] stream key %d/%d failed for %s: %v\n", i+1, keys, f.apiBase, err)
	}

	return nil, fmt.Errorf("all %d API keys exhausted for %s: %w", keys, f.apiBase, lastErr)
}

// GetDefaultModel returns the default model.
func (f *FailoverProvider) GetDefaultModel() string {
	return ""
}