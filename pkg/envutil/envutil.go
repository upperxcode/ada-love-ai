// Package envutil resolves provider API keys that are stored as environment
// variable references (e.g. OPENROUTER_API_KEY) rather than literal values, and
// loads a project-local .env file so those references resolve even when the
// variable isn't present in the real process environment.
package envutil

import (
	"ada-love-ai/pkg/logger"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// envVarName matches strings that look like an environment variable name:
// starts with an uppercase letter, followed by uppercase letters, digits or
// underscores. This intentionally does NOT match real API key values such as
// "sk-or-v1-..." (which contain lowercase and hyphens), so there is no
// ambiguity between a literal key and a variable reference.
var envVarName = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// IsEnvVarName reports whether s looks like an environment variable name
// (i.e. it should be resolved rather than used as a literal API key).
func IsEnvVarName(s string) bool {
	return envVarName.MatchString(s)
}

// ResolveKey resolves a stored API key value to its final form:
//   - Empty input returns empty.
//   - If the value looks like an environment variable name, it is resolved via
//     os.Getenv. If the variable is unset, the original value is returned as a
//     graceful fallback (so a literal key that coincidentally matches the
//     pattern still works).
//   - Anything else is returned verbatim (it's a literal key).
//
// This is the single choke point for env-var resolution of provider keys.
func ResolveKey(raw string) string {
	if raw == "" {
		return ""
	}
	if IsEnvVarName(raw) {
		if v, ok := os.LookupEnv(raw); ok {
			return v
		}
	}
	return raw
}

// LoadEnvFiles loads variables from project-local .env files into the process
// environment. Variables that are already set in the real environment take
// precedence and are NOT overwritten (the .env file is a fallback).
//
// It searches, in order: the working directory, ./config, and the OS-specific
// config directory (e.g. ~/.config/ada-love-ai on Linux). The first .env file that
// exists and parses wins. If none exists, creates a template at the OS config dir.
func LoadEnvFiles() {
	found := false
	for _, p := range envCandidatePaths() {
		if loadEnvFile(p) {
			logger.DebugCF("envutil", "Carregado .env",
				map[string]any{"path": p})
			found = true
			break
		}
	}
	if !found {
		configDir := osConfigDir()
		fmt.Printf("[envutil] DEBUG: configDir=%q found=%v\n", configDir, found)
		if configDir != "" {
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				logger.WarnCF("envutil", "Falha ao criar dir de config",
					map[string]any{"dir": configDir, "error": err.Error()})
			} else {
				templatePath := filepath.Join(configDir, ".env")
				fmt.Printf("[envutil] DEBUG: templatePath=%q\n", templatePath)
				if err := createEnvTemplate(templatePath); err != nil {
					logger.WarnCF("envutil", "Falha ao criar .env template",
						map[string]any{"path": templatePath, "error": err.Error()})
				} else {
					logger.InfoCF("envutil", "Criado .env template - preencha suas chaves",
						map[string]any{"path": templatePath})
				}
				_ = loadEnvFile(templatePath)
			}
		} else {
			logger.WarnCF("envutil", "Nenhum arquivo .env encontrado e não foi possível criar template",
				map[string]any{"searched_paths": envCandidatePaths()})
		}
	}
}

// envCandidatePaths returns the locations where a .env file may live.
func envCandidatePaths() []string {
	candidates := []string{
		".env",
		filepath.Join("config", ".env"),
	}
	if dir := osConfigDir(); dir != "" {
		candidates = append(candidates, filepath.Join(dir, ".env"))
	}
	return candidates
}

// osConfigDir mirrors backend.getOSConfigDir so the .env file is looked for in
// the same place the rest of the app keeps its config.
func osConfigDir() string {
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "ada-love-ai")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ada-love-ai")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "ada-love-ai")
	}
	return ""
}

// loadEnvFile parses a .env file and sets non-overridden variables. Returns
// false if the file does not exist or could not be read.
func loadEnvFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := parseEnvLine(line)
		if !ok {
			continue
		}
		// Real environment wins over .env.
		if _, set := os.LookupEnv(key); set {
			continue
		}
		os.Setenv(key, value)
	}
	return true
}

// parseEnvLine splits a "KEY=VALUE" line, stripping optional surrounding quotes
// from the value. Inline comments after an unquoted value are preserved as part
// of the value (only quoted values have a well-defined end).
func parseEnvLine(line string) (key, value string, ok bool) {
	idx := strings.IndexByte(line, '=')
	if idx <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}
	value = strings.TrimSpace(line[idx+1:])
	value = stripQuotes(value)
	return key, value, true
}

// stripQuotes removes a matching pair of surrounding single or double quotes
// from the value, if present.
func stripQuotes(v string) string {
	if len(v) >= 2 {
		first, last := v[0], v[len(v)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}

// createEnvTemplate creates a .env file with example API keys at the given path.
func createEnvTemplate(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	template := `# Ada Love AI - Environment Variables
# Preencha as chaves abaixo. Variáveis de ambiente reais têm precedência sobre este arquivo.
# Obtenha suas chaves em: https://openrouter.ai/keys, https://platform.openai.com/api-keys, etc.

OPENROUTER_API_KEY=
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GEMINI_API_KEY=
GROQ_API_KEY=
TOGETHER_API_KEY=
DEEPSEEK_API_KEY=
OLLAMA_HOST=
`
	return os.WriteFile(path, []byte(template), 0o600)
}
