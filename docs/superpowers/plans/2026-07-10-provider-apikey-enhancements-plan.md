# Provider API Key Enhancements — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Resolve env var API key references on save and implement automatic failover across multiple API keys.

**Architecture:** Two independent changes: (1) resolve env var references via `envutil.ResolveKey()` before persisting to DB, storing the literal value; (2) wrap the OpenAI-compatible provider in a `FailoverProvider` that cycles through API keys on HTTP error. Both changes are backend-only with no UI changes needed.

**Tech Stack:** Go, SQLite, OpenAI-compatible HTTP provider

---

### Task 1: Resolve env vars in SaveDBProvider

**Files:**
- Modify: `backend/model_manager.go` (add `resolveAPIKeys` helper, call from `SaveDBProvider` and `syncProviderToDB`)

**Goal:** When saving a provider, resolve any env var references (e.g. `OPENROUTER_API_KEY`) to their literal values before persisting to the database. Store the original env var name in `UserKey` for reference.

- [ ] **Step 1: Add resolveAPIKeys helper**

Add to `backend/model_manager.go` (before `SaveDBProvider`):

```go
// resolveAPIKeys resolves environment variable references in API keys to their
// literal values. The original env var name is preserved in UserKey.
func resolveAPIKeys(keys []ProviderApiKey) []ProviderApiKey {
	for i, k := range keys {
		if k.Key == "" {
			continue
		}
		if resolved := envutil.ResolveKey(k.Key); resolved != k.Key {
			keys[i].Key = resolved
			keys[i].UserKey = k.Key // store original env var name
		}
	}
	return keys
}
```

- [ ] **Step 2: Hook into SaveDBProvider**

Modify `Engine.SaveDBProvider` in `backend/model_manager.go` to call `resolveAPIKeys` before persisting:

```go
func (e *Engine) SaveDBProvider(name string, cfg ProviderConfig) error {
	e.mu.Lock()
	if e.adaCfg.Providers == nil {
		e.adaCfg.Providers = make(map[string]ProviderConfig)
	}
	cfg.ApiKeys = resolveAPIKeys(cfg.ApiKeys)
	e.adaCfg.Providers[name] = cfg
	e.mu.Unlock()
	if e.db != nil {
		if err := e.db.SaveProviderFull(adaptProviderConfig(name, cfg)); err != nil {
			return fmt.Errorf("erro ao salvar provider %s no DB: %w", name, err)
		}
	}
	return e.SaveAdaConfig()
}
```

- [ ] **Step 3: Hook into syncProviderToDB**

Modify `Engine.syncProviderToDB` to also resolve keys:

```go
func (e *Engine) syncProviderToDB(name string) {
	e.mu.RLock()
	cfg, ok := e.adaCfg.Providers[name]
	e.mu.RUnlock()
	if !ok || e.db == nil {
		return
	}
	cfg.ApiKeys = resolveAPIKeys(cfg.ApiKeys)
	if err := e.db.SaveProviderFull(adaptProviderConfig(name, cfg)); err != nil {
		fmt.Printf("[Engine] Erro ao sincronizar provider %s no DB: %v\n", name, err)
	}
}
```

- [ ] **Step 4: Build and verify**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && go build ./backend/...
```
Expected: no errors

---

### Task 2: FailoverProvider implementation

**Files:**
- Create: `pkg/providers/failover_provider.go`
- Modify: `pkg/providers/factory_provider.go` (wire failover into OpenAI-compatible creation)

**Goal:** A wrapper that implements `LLMProvider` (and optionally `StreamingProvider`) and cycles through multiple API keys on HTTP error.

- [ ] **Step 1: Create failover_provider.go**

Create `pkg/providers/failover_provider.go`:

```go
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
	apiKeys    []string
	apiBase    string
	proxy      string
	opts       []openai_compat.Option
	currentIdx int
	mu         sync.Mutex
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

// createProvider creates a fresh openai_compat.Provider with the given key index.
func (f *FailoverProvider) createProvider(idx int) *openai_compat.Provider {
	return openai_compat.NewProvider(f.apiKeys[idx], f.apiBase, f.proxy, f.opts...)
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
			// Success — update currentIdx for next call
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

// GetDefaultModel returns the default model (delegated to the first key's provider).
func (f *FailoverProvider) GetDefaultModel() string {
	p := f.createProvider(0)
	return p.GetDefaultModel()
}
```

- [ ] **Step 2: Add StreamingProvider support**

Add to the same file:

```go
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

		// Check if the underlying provider supports streaming
		sp, ok := any(p).(StreamingProvider)
		if !ok {
			// Fallback to non-streaming Chat
			return f.Chat(ctx, messages, tools, model, options)
		}

		resp, err := sp.ChatStream(ctx, messages, tools, model, options, onChunk)
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
```

- [ ] **Step 3: Build and verify**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && go build ./pkg/providers/...
```
Expected: no errors

---

### Task 3: Wire failover into factory provider

**Files:**
- Modify: `pkg/providers/factory_provider.go`

**Goal:** When creating an OpenAI-compatible provider, use `NewFailoverProvider` if multiple API keys are available.

- [ ] **Step 1: Add import for failover provider**

The factory already imports `openai_compat`. Add the `FailoverProvider` import (it's in the same package `providers`, so no import needed).

- [ ] **Step 2: Modify the "openai" case in CreateProviderFromConfig**

In `pkg/providers/factory_provider.go`, find the `case "openai"` block (around line 164). After building the provider options, replace the direct `NewHTTPProvider` call with a conditional that wraps in `FailoverProvider` when multiple keys exist:

```go
case "openai":
	// OpenAI with OAuth/token auth (Codex-style)
	if cfg.AuthMethod == "oauth" || cfg.AuthMethod == "token" {
		provider, err := createCodexAuthProvider()
		if err != nil {
			return nil, "", err
		}
		return provider, modelID, nil
	}
	// OpenAI with API key
	if cfg.APIKey() == "" && cfg.APIBase == "" {
		return nil, "", fmt.Errorf("api_key or api_base is required for HTTP-based protocol %q", protocol)
	}
	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = getDefaultAPIBase(protocol)
	}

	// Collect all API keys
	allKeys := cfg.APIKeys()
	if len(allKeys) == 0 {
		allKeys = []string{cfg.APIKey()}
	}

	if len(allKeys) <= 1 {
		// Single key — direct provider (existing behavior)
		provider := NewHTTPProviderWithMaxTokensFieldAndRequestTimeout(
			cfg.APIKey(),
			apiBase,
			cfg.Proxy,
			cfg.MaxTokensField,
			userAgent,
			cfg.RequestTimeout,
			cfg.ExtraBody,
			cfg.CustomHeaders,
		)
		provider.SetProviderName(protocol)
		return provider, modelID, nil
	}

	// Multiple keys — wrap in FailoverProvider
	opts := []openai_compat.Option{
		openai_compat.WithMaxTokensField(cfg.MaxTokensField),
		openai_compat.WithUserAgent(userAgent),
		openai_compat.WithRequestTimeout(time.Duration(cfg.RequestTimeout) * time.Second),
	}
	if cfg.ExtraBody != nil {
		opts = append(opts, openai_compat.WithExtraBody(cfg.ExtraBody))
	}
	if cfg.CustomHeaders != nil {
		opts = append(opts, openai_compat.WithCustomHeaders(cfg.CustomHeaders))
	}

	provider := NewFailoverProvider(allKeys, apiBase, cfg.Proxy, opts...)
	provider.SetProviderName(protocol)
	return provider, modelID, nil
```

- [ ] **Step 3: Add APIKeys() method to ModelConfig**

Check if `config.ModelConfig` already has a method to return all API keys. If not, add it to `pkg/config/config.go`:

```go
// APIKeys returns all API keys as a string slice.
func (m *ModelConfig) APIKeys() []string {
	keys := make([]string, 0, len(m.APIKeys))
	for _, k := range m.APIKeys {
		keys = append(keys, k.String())
	}
	return keys
}
```

- [ ] **Step 4: Build and verify**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && go build ./pkg/...
```
Expected: no errors

---

### Task 4: Verify full build

- [ ] **Step 1: Build the entire project**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && go build ./... 2>&1 | grep -v feishu
```
Expected: no errors (only feishu package errors which are pre-existing)

- [ ] **Step 2: Run go vet**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai && go vet ./pkg/providers/... ./backend/...
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/data/aux/dev/projects/go/ada-love-ai
git add pkg/providers/failover_provider.go backend/model_manager.go pkg/providers/factory_provider.go pkg/config/config.go
git commit -m "feat(provider): resolve env vars on save and add API key failover

- Resolve OPENROUTER_API_KEY env var references to literal values on save
- Store original env var name in UserKey field for reference
- New FailoverProvider wraps multiple API keys with automatic failover
- Wire failover into OpenAI-compatible provider creation"
```