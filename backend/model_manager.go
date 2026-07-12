package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ada-love-ai/pkg/config"
	"ada-love-ai/pkg/envutil"
)

func (e *Engine) GetModelList() []ModelConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var list []ModelConfig
	for _, m := range e.adaCfg.ModelList {
		list = append(list, ModelConfig{
			ModelName: m.ModelName,
			Model:     m.Model,
			Provider:  m.Provider,
			APIBase:   m.APIBase,
			Enabled:   m.Enabled,
		})
	}
	return list
}

func (e *Engine) AddModel(m ModelConfig) error {
	e.mu.Lock()
	e.adaCfg.ModelList = append(e.adaCfg.ModelList, &config.ModelConfig{
		ModelName: m.ModelName,
		Model:     m.Model,
		Provider:  m.Provider,
		APIBase:   m.APIBase,
		Enabled:   m.Enabled,
	})
	e.mu.Unlock()
	return e.SaveAdaConfig()
}

func (e *Engine) RemoveModel(name, provider string) {
	e.mu.Lock()
	var newList config.SecureModelList
	for _, m := range e.adaCfg.ModelList {
		if m.ModelName == name && m.Provider == provider {
			continue
		}
		newList = append(newList, m)
	}
	e.adaCfg.ModelList = newList
	e.mu.Unlock()
	e.SaveAdaConfig()
}

func (e *Engine) SetProviderSettings(name, apiBase, apiKey string) {
	e.mu.Lock()
	if e.adaCfg.ProviderBases == nil {
		e.adaCfg.ProviderBases = make(map[string]string)
	}
	if e.adaCfg.ProviderKeys == nil {
		e.adaCfg.ProviderKeys = make(map[string]string)
	}
	e.adaCfg.ProviderBases[name] = apiBase
	e.adaCfg.ProviderKeys[name] = apiKey
	e.mu.Unlock()
	e.SaveAdaConfig()
	// Write-through: atualiza provider no DB se existir em Providers
	e.syncProviderToDB(name)
}

func (e *Engine) GetProviderSettings(name string) (string, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.adaCfg.ProviderBases[name], e.adaCfg.ProviderKeys[name]
}

func (e *Engine) GetProviders() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	providers := make(map[string]bool)
	for _, m := range e.adaCfg.ModelList {
		providers[m.Provider] = true
	}
	for p := range e.adaCfg.ProviderBases {
		providers[p] = true
	}

	var list []string
	for p := range providers {
		list = append(list, p)
	}
	return list
}

func (e *Engine) RemoveProvider(name string) {
	e.mu.Lock()
	delete(e.adaCfg.ProviderBases, name)
	delete(e.adaCfg.ProviderKeys, name)

	var newList config.SecureModelList
	for _, m := range e.adaCfg.ModelList {
		if m.Provider == name {
			continue
		}
		newList = append(newList, m)
	}
	e.adaCfg.ModelList = newList
	// Remove de Providers (DB)
	delete(e.adaCfg.Providers, name)
	e.mu.Unlock()
	e.SaveAdaConfig()
	// Write-through: remove do DB
	if e.db != nil {
		if err := e.db.DeleteProviderFull(name); err != nil {
			fmt.Printf("[Engine] Erro ao remover provider %s do DB: %v\n", name, err)
		}
	}
}

func (e *Engine) GetModelSettings(provider, modelID string) ExtraModelConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", provider, modelID)
	return e.adaCfg.ModelSettings[key]
}

func (e *Engine) SetModelSettings(provider, modelID string, settings ExtraModelConfig) {
	e.mu.Lock()
	if e.adaCfg.ModelSettings == nil {
		e.adaCfg.ModelSettings = make(map[string]ExtraModelConfig)
	}
	key := fmt.Sprintf("%s:%s", provider, modelID)
	e.adaCfg.ModelSettings[key] = settings
	e.mu.Unlock()
	e.SaveAdaConfig()
}

func (e *Engine) PingProvider(name string) error {
	apiBase, _ := e.GetProviderSettings(name)
	if apiBase == "" {
		if name == "lmstudio" {
			apiBase = "http://127.0.0.1:1234/v1"
		} else if name == "ollama" {
			apiBase = "http://127.0.0.1:11434/v1"
		} else {
			return fmt.Errorf("URL base não configurada para %s", name)
		}
	}

	resp, err := http.Get(apiBase + "/models")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}

// FetchProviderModels queries a provider's /models endpoint and returns the
// list enriched with detected capabilities (vision/embedding). The protocol is
// derived from connectionType: "openai" (OpenAI-compatible), "anthropic", or
// "gemini". The provider name is only used to resolve a default apiBase when
// apiBase is empty (falls back to known local servers).
func (e *Engine) FetchProviderModels(name, apiKey, apiBase, connectionType string) ([]ProviderModel, error) {
	apiKey = envutil.ResolveKey(apiKey)
	if apiBase == "" {
		apiBase = defaultAPIBaseFor(name, connectionType)
	}
	if apiBase == "" {
		return nil, fmt.Errorf("URL base não configurada para %s", name)
	}
	return fetchModelsByProtocol(connectionType, apiBase, apiKey)
}

// TestProviderConnection validates that the given API key (which may be a
// literal value or an environment variable reference) authenticates against the
// provider's /models endpoint. It resolves the key via envutil.ResolveKey and
// reuses the per-protocol fetch logic.
func (e *Engine) TestProviderConnection(name, apiKey, apiBase, connectionType string) (ProviderTestResult, error) {
	resolved := envutil.ResolveKey(apiKey)
	if apiBase == "" {
		apiBase = defaultAPIBaseFor(name, connectionType)
	}
	if apiBase == "" {
		return ProviderTestResult{Message: fmt.Sprintf("URL base não configurada para %s", name)}, nil
	}
	if _, err := fetchModelsByProtocol(connectionType, apiBase, resolved); err != nil {
		return ProviderTestResult{Message: fmt.Sprintf("Falha ao conectar: %v", err)}, nil
	}
	return ProviderTestResult{
		Ok:      true,
		Success: true,
		Message: fmt.Sprintf("Conexão validada com %s", name),
	}, nil
}

// defaultAPIBaseFor resolves a sane default apiBase for local providers that
// don't require one to be configured (lmstudio, ollama, vllm). Returns "" for
// providers that genuinely need a configured base.
func defaultAPIBaseFor(name, connectionType string) string {
	switch strings.ToLower(name) {
	case "lmstudio":
		return "http://127.0.0.1:1234/v1"
	case "ollama":
		return "http://127.0.0.1:11434/v1"
	case "vllm":
		return "http://127.0.0.1:8000/v1"
	}
	return ""
}

// fetchModelsByProtocol performs the HTTP fetch and parsing per provider
// protocol, then classifies each model's capabilities.
func fetchModelsByProtocol(connectionType, apiBase, apiKey string) ([]ProviderModel, error) {
	ct := strings.ToLower(strings.TrimSpace(connectionType))
	switch ct {
	case "anthropic":
		return fetchAnthropicModels(apiBase, apiKey)
	case "gemini":
		return fetchGeminiModels(apiBase, apiKey)
	default: // "openai" and anything OpenAI-compatible
		return fetchOpenAIModels(apiBase, apiKey)
	}
}

// fetchOpenAIModels handles the OpenAI-compatible /models shape: { data: [{ id, ... }] }.
// OpenRouter additionally returns architecture.input_modalities and
// architecture.modality, which we use to refine capability detection.
func fetchOpenAIModels(apiBase, apiKey string) ([]ProviderModel, error) {
	body, err := httpGetJSON(apiBase+"/models", map[string]string{
		"Authorization": "Bearer " + apiKey,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []struct {
			ID           string `json:"id"`
			Architecture *struct {
				Modality         string `json:"modality"`
				InputModalities  any    `json:"input_modalities"`
				OutputModalities any    `json:"output_modalities"`
			} `json:"architecture"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta /models: %w", err)
	}

	models := make([]ProviderModel, 0, len(resp.Data))
	for _, m := range resp.Data {
		if m.ID == "" {
			continue
		}
		vision, embedding := detectCapabilities(m.ID, m.Architecture)
		free, thinking := classifyChatModel(m.ID, embedding)
		models = append(models, newProviderModel(m.ID, vision, embedding, free, thinking))
	}
	return models, nil
}

// fetchAnthropicModels handles Anthropic's /v1/models shape: { data: [{ id, ... }] },
// authenticated with x-api-key + anthropic-version headers.
func fetchAnthropicModels(apiBase, apiKey string) ([]ProviderModel, error) {
	body, err := httpGetJSON(apiBase+"/models", map[string]string{
		"x-api-key":         apiKey,
		"anthropic-version": "2023-06-01",
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta /models: %w", err)
	}

	models := make([]ProviderModel, 0, len(resp.Data))
	for _, m := range resp.Data {
		if m.ID == "" {
			continue
		}
		vision, embedding := detectCapabilities(m.ID, nil)
		free, thinking := classifyChatModel(m.ID, embedding)
		models = append(models, newProviderModel(m.ID, vision, embedding, free, thinking))
	}
	return models, nil
}

// fetchGeminiModels handles Google's Generative Language /v1beta/models shape:
// { models: [{ name: "models/gemini-1.5-flash", supportedGenerationMethods: [...] }] },
// authenticated via ?key=. Names are normalized to strip the "models/" prefix.
func fetchGeminiModels(apiBase, apiKey string) ([]ProviderModel, error) {
	url := apiBase + "/models?key=" + apiKey
	body, err := httpGetJSON(url, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Models []struct {
			Name                       string   `json:"name"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta /models: %w", err)
	}

	models := make([]ProviderModel, 0, len(resp.Models))
	for _, m := range resp.Models {
		id := strings.TrimPrefix(m.Name, "models/")
		if id == "" {
			continue
		}
		// Gemini exposes explicit generation methods; use them as ground truth
		// when available (embedContent / generateContent).
		vision, embedding := detectCapabilities(id, nil)
		for _, method := range m.SupportedGenerationMethods {
			if method == "embedContent" {
				embedding = true
			}
		}
		free, thinking := classifyChatModel(id, embedding)
		models = append(models, newProviderModel(id, vision, embedding, free, thinking))
	}
	return models, nil
}

// newProviderModel builds a ProviderModel with the given capabilities.
// Chat models (non-embedding) default to Tools=true (most support function calling).
func newProviderModel(id string, vision, embedding, free, thinking bool) ProviderModel {
	return ProviderModel{
		ID:        id,
		Name:      id,
		Vision:    vision,
		Embedding: embedding,
		Tools:     !embedding,
		Free:      free,
		Thinking:  thinking,
	}
}

// archInfo is the subset of OpenRouter-style architecture metadata we inspect.
type archInfo struct {
	Modality        string
	InputModalities []string
}

// detectCapabilities infers whether a model supports vision (image input) and
// embeddings. It combines explicit architecture metadata (when provided by the
// upstream payload, e.g. OpenRouter) with id-based heuristics as a fallback.
func detectCapabilities(modelID string, arch *struct {
	Modality         string `json:"modality"`
	InputModalities  any    `json:"input_modalities"`
	OutputModalities any    `json:"output_modalities"`
}) (vision, embedding bool) {
	id := strings.ToLower(modelID)

	// --- Embedding ---
	if arch != nil {
		if strings.Contains(strings.ToLower(arch.Modality), "embedding") {
			embedding = true
		}
		if !embedding {
			for _, im := range toStringSlice(arch.InputModalities) {
				if strings.EqualFold(im, "embedding") {
					embedding = true
					break
				}
			}
		}
	}
	if !embedding && (strings.Contains(id, "embedding") || strings.Contains(id, "embed")) {
		embedding = true
	}

	// --- Vision (image input) ---
	if arch != nil {
		for _, im := range toStringSlice(arch.InputModalities) {
			if strings.EqualFold(im, "image") {
				vision = true
				break
			}
		}
	}
	if !vision {
		vision = looksLikeVisionModel(id)
	}
	return vision, embedding
}

// looksLikeVisionModel applies id-based heuristics for multimodal/vision models.
func looksLikeVisionModel(id string) bool {
	if strings.Contains(id, "vision") || strings.Contains(id, "-vl") || strings.HasPrefix(id, "vl") {
		return true
	}
	// Well-known model families that accept image input.
	known := []string{
		"gpt-4o", "gpt-4-vision", "gpt-4-turbo", "gpt-4o-mini",
		"claude-3", "claude-sonnet", "claude-opus", "claude-haiku",
		"gemini-", "gemini-pro-vision",
		"pixtral", "llava", "bakllava",
		"qwen-vl", "qwen2-vl", "qwen2.5-vl",
		"internvl", "cogvlm",
	}
	for _, k := range known {
		if strings.Contains(id, k) {
			return true
		}
	}
	return false
}

// classifyChatModel detects whether a chat model is free and/or thinking.
// These are independent: e.g. deepseek-r1 is both free AND thinking.
// Embedding models return (false, false) since they are not chat models.
func classifyChatModel(modelID string, embedding bool) (free, thinking bool) {
	if embedding {
		return false, false
	}
	id := strings.ToLower(modelID)

	// Known reasoning/thinking models.
	if strings.Contains(id, "o1") || strings.Contains(id, "o3") ||
		strings.Contains(id, "thinking") || strings.Contains(id, "reason") ||
		strings.Contains(id, "r1") || strings.Contains(id, "-reasoner") {
		thinking = true
	}

	// Known genuinely free models (open-weight, community, or free-tier).
	freePatterns := []string{
		// Meta Llama
		"llama-3", "llama3", "llama-4",
		// Mistral open
		"mistral-7b", "mistral-small",
		// Qwen open
		"qwen2.5", "qwen2-", "qwen-2",
		// Google Gemma
		"gemma-2", "gemma-3",
		// Microsoft Phi
		"phi-3", "phi-4",
		// Google Gemini Flash (free tier)
		"gemini-2.0-flash", "gemini-1.5-flash",
		// DeepSeek
		"deepseek-v3", "deepseek-chat", "deepseek-r1",
		// Anthropic Claude Haiku
		"claude-3-5-haiku", "claude-3-haiku",
	}
	for _, p := range freePatterns {
		if strings.Contains(id, p) {
			free = true
			break
		}
	}

	return free, thinking
}

// toStringSlice normalizes a JSON-decoded modalities field (which may arrive as
// []string, []any, or a comma-separated string) into []string.
func toStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		out := make([]string, 0, len(val))
		for _, s := range val {
			if str, ok := s.(string); ok {
				out = append(out, str)
			}
		}
		return out
	case string:
		parts := strings.Split(val, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts
	}
	return nil
}

// httpGetJSON performs a GET request with optional headers and returns the raw
// JSON body. Non-200 responses produce an error that includes the status.
func httpGetJSON(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return body, nil
}

// --- Provider DB write-through ---

// syncProviderToDB writes the current in-memory provider config to DB.
// Called after mutations that update ProviderBases/ProviderKeys/Providers.
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

// SaveDBProvider saves a single provider to DB and updates in-memory config.
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

// DeleteDBProvider removes a provider from DB and in-memory config.
func (e *Engine) DeleteDBProvider(name string) error {
	e.mu.Lock()
	delete(e.adaCfg.Providers, name)
	e.mu.Unlock()
	if e.db != nil {
		if err := e.db.DeleteProviderFull(name); err != nil {
			return fmt.Errorf("erro ao remover provider %s do DB: %w", name, err)
		}
	}
	return e.SaveAdaConfig()
}
