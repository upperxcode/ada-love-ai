# Provider API Key Enhancements

**Date:** 2026-07-10
**Status:** Draft

## Problem

Two related issues with provider API key management:

1. **Env var references not persisted.** When a user types `OPENROUTER_API_KEY` in the API key field, the system correctly resolves the environment variable at runtime via `envutil.ResolveKey()`, but the **name of the env var** (`OPENROUTER_API_KEY`) is saved to the database rather than the **resolved value** (`sk-or-v1-...`). On restart, if the env var is no longer set, the key is lost.

2. **No automatic failover for multiple API keys.** The `ProviderConfig` already supports multiple `ApiKeys []ProviderApiKey` in the struct, but only the **first key** is ever used (`GetAPIKey()` returns `ApiKeys[0].Key`). There is no mechanism to try subsequent keys when the first one fails.

## Solution

### Part 1: Resolve Env Vars on Save

**Location:** `backend/model_manager.go` — `SaveDBProvider`

When saving a provider, iterate through all `cfg.ApiKeys` and resolve any environment variable references **before persisting to the database**.

**Algorithm:**

```go
func resolveAPIKeys(keys []ProviderApiKey) []ProviderApiKey {
    for i, k := range keys {
        if k.Key == "" {
            continue
        }
        resolved := envutil.ResolveKey(k.Key)
        if resolved != k.Key {
            // It was an env var — save the resolved value
            keys[i].Key = resolved
            keys[i].UserKey = k.Key  // store original env var name for reference
        }
    }
    return keys
}
```

- `envutil.IsEnvVarName()` already detects strings like `OPENROUTER_API_KEY` (all uppercase + digits + underscores)
- `envutil.ResolveKey()` resolves via `os.Getenv()` with `.env` fallback
- The existing `ProviderApiKey.UserKey` field (currently unused) stores the original env var name for display/reference
- The `ProviderApiKey.Key` field stores the resolved literal value

**Entry points to patch:**
- `Engine.SaveDBProvider(name, cfg)` → resolve keys before persisting
- `Engine.syncProviderToDB(name)` → same resolution before sync

### Part 2: Failover Across Multiple API Keys

**Location:** New file `pkg/providers/failover_provider.go`

A wrapper type that implements `LLMProvider` (and optionally `StreamingProvider`) and cycles through multiple API keys on HTTP errors.

#### Interface

```go
package providers

// FailoverProvider wraps an LLMProvider with multiple API keys
// and retries with the next key on any HTTP error.
type FailoverProvider struct {
    apiKeys    []string
    apiBase    string
    proxy      string
    opts       []openai_compat.Option  // or a generic factory
    currentIdx int
}
```

It stores the configuration needed to create a fresh `openai_compat.Provider` (or any protocol-specific provider) for each key.

#### Chat Flow

```
Chat(ctx, messages, tools, model, options):
  1. startIdx = currentIdx
  2. loop:
     a. create provider with apiKeys[currentIdx]
     b. call provider.Chat(ctx, messages, tools, model, options)
     c. if success → return response
     d. if HTTP error (any status != 2xx):
        - increment currentIdx (wrap around, but stop after full cycle)
        - if currentIdx == startIdx → break (all keys tried)
        - log "failover: key {i} failed, trying next"
        - continue loop
  3. return consolidated error "all API keys failed for {apiBase}"
```

#### Streaming Support

Same logic for `ChatStream` — on HTTP error, rotate key and restart the stream.

#### Where to Wire In

In `pkg/providers/factory_provider.go`:

```go
func createProviderWithFailover(keys []string, apiBase, proxy string, opts ...openai_compat.Option) LLMProvider {
    if len(keys) <= 1 {
        return openai_compat.NewProvider(keys[0], apiBase, proxy, opts...)
    }
    return NewFailoverProvider(keys, apiBase, proxy, opts...)
}
```

This is called from the `case "openai"` branch of `CreateProviderFromConfig` (and any other protocol that uses API keys — Anthropic, Gemini, etc.).

For simplicity, the **first release** will only implement failover for the OpenAI-compatible protocol (which covers OpenRouter, Groq, DeepSeek, Together, etc.). Other protocols can be added later.

#### Key Resolution for Failover

The keys stored in the database will already be resolved (Part 1). When loading from DB on startup, the `ProviderConfig.Models` map is rebuilt via `deadaptProviderConfig`, and the keys come from `sp.APIKeys`. No additional env var resolution needed at runtime — the values are already literal.

### Data Model

No schema changes. The existing `provider_apikeys` table already stores `(provider_id, apikey)` — one row per key. The `ProviderApiKey.UserKey` field is used to store the original env var name for reference.

```go
type ProviderApiKey struct {
    Key     string `json:"key"`               // resolved literal value (or literal if not an env var)
    UserKey string `json:"user_key,omitempty"` // original env var name, e.g. "OPENROUTER_API_KEY"
}
```

### Error Handling

- **Single key failure:** Transparent — system tries next key automatically
- **All keys fail:** Consolidated error message: `"all {N} API keys exhausted for {provider} ({apiBase}): last error: {error}"`
- **Logging:** Each failover attempt is logged at debug level with key index and status code
- **Frontend:** No immediate change needed. The existing error display already shows backend errors

### Frontend

**Minimal changes.** The `ModelsSection.tsx` already supports:
- Adding multiple API keys via `handleAddApiKey`
- Editing keys via `handleUpdateApiKey`
- Removing keys via `handleRemoveApiKey`
- The `api_keys: []ProviderApiKey` form field

No UI changes needed for this release. The env var resolution and failover happen transparently in the backend.

**Future enhancement:** Show a badge/label on each key indicating "from env var" when `user_key` is set.

### Files Changed

| File | Change |
|------|--------|
| `pkg/providers/failover_provider.go` | **NEW** — FailoverProvider wrapper |
| `backend/model_manager.go` | Resolve env vars in `SaveDBProvider` and `syncProviderToDB` |
| `pkg/providers/factory_provider.go` | Wire `createProviderWithFailover` into OpenAI-compatible provider creation |
| `backend/types.go` | Minor: ensure `ProviderApiKey.UserKey` is properly serialized |

### Out of Scope

- **Encryption:** future PR (the user explicitly mentioned this)
- **Model-level failover:** not yet implemented (the user explicitly mentioned this is not ready)
- **Rate-limit backoff:** only basic retry on HTTP error, no exponential backoff
- **Frontend UX for key status (active/failed):** future enhancement
- **Protocols beyond OpenAI-compatible:** limited to OpenAI-compatible for v1