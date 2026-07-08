package registry

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TestConfig espelha wizard-spec/internal/registry/expert.go.
type TestConfig struct {
	Command    string `yaml:"command" json:"command"`
	FailPrompt string `yaml:"fail_prompt" json:"fail_prompt"`
}

// ExpertPlugin representa um plugin carregado de experts.yaml.
type ExpertPlugin struct {
	ID                 string      `yaml:"id" json:"id"`
	Name               string      `yaml:"name" json:"name"`
	Description        string      `yaml:"description" json:"description"`
	Endpoint           string      `yaml:"endpoint" json:"endpoint"`
	Triggers           []string    `yaml:"triggers" json:"triggers"`
	Language           string      `yaml:"language" json:"language"`
	StartCommand       string      `yaml:"start_command" json:"start_command"`
	DependencyEndpoint string      `yaml:"dependency_endpoint" json:"dependency_endpoint"`
	TestConfig         *TestConfig `yaml:"test_config" json:"test_config"`
}

// LoadExperts carrega e mescla YAMLs por ID (last-wins).
// Arquivos inexistentes são silenciosamente ignorados.
func LoadExperts(paths ...string) ([]*ExpertPlugin, error) {
	expertsMap := make(map[string]*ExpertPlugin)

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg struct {
			Experts []*ExpertPlugin `yaml:"experts"`
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		for _, p := range cfg.Experts {
			if p != nil && p.ID != "" {
				expertsMap[p.ID] = p
			}
		}
	}

	all := make([]*ExpertPlugin, 0, len(expertsMap))
	for _, p := range expertsMap {
		all = append(all, p)
	}
	return all, nil
}

// FindExpertByLanguage retorna o primeiro plugin cuja Language bate (case-insensitive).
func FindExpertByLanguage(language string, plugins []*ExpertPlugin) (*ExpertPlugin, error) {
	for _, p := range plugins {
		if strings.EqualFold(p.Language, language) {
			return p, nil
		}
	}
	return nil, errors.New("MCP nao encontrado para a linguagem especificada")
}
