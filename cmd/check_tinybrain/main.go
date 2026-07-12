package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"ada-love-ai/backend"
)

func main() {
	eng, err := backend.NewEngine()
	if err != nil {
		fmt.Printf("NewEngine error: %v\n", err)
		return
	}
	cfg := eng.GetAdaConfig()
	fmt.Printf("TinyBrain config: provider=%q model=%q\n", cfg.TinyBrain.Provider, cfg.TinyBrain.ModelName)

	apiBase := ""
	apiKey := ""
	for _, m := range cfg.ModelList {
		if m != nil && m.Provider == cfg.TinyBrain.Provider && m.APIBase != "" {
			apiBase = m.APIBase
			apiKey = m.APIKey()
			break
		}
	}
	if apiBase == "" {
		switch cfg.TinyBrain.Provider {
		case "lmstudio":
			apiBase = "http://127.0.0.1:1234/v1"
		case "ollama":
			apiBase = "http://127.0.0.1:11434/v1"
		}
	}

	modelToUse := strings.TrimSpace(cfg.TinyBrain.ModelName)
	if modelToUse == "" {
		prov := strings.TrimSpace(cfg.TinyBrain.Provider)
		for _, m := range cfg.ModelList {
			if m == nil {
				continue
			}
			if prov != "" && strings.EqualFold(m.Provider, prov) {
				modelToUse = strings.TrimSpace(m.ModelName)
				break
			}
		}
	}
	if modelToUse == "" {
		fmt.Printf("No model resolved for TinyBrain provider %q\n", cfg.TinyBrain.Provider)
		return
	}

	requestBody := map[string]interface{}{
		"model": modelToUse,
		"messages": []map[string]string{{"role": "user", "content": "hello from check"}},
		"temperature": 0.3,
	}
	b, _ := json.MarshalIndent(requestBody, "", "  ")
	fmt.Printf("api_base=%q api_key_set=%v\nRequest body:\n%s\n", apiBase, apiKey != "", string(b))
}
