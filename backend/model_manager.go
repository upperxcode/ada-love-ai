package backend

import (
	"fmt"
	"net/http"

	"ada-love-ai/pkg/config"
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
	e.mu.Unlock()
	e.SaveAdaConfig()
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
