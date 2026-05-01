package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (e *Engine) SearchSkills(ctx context.Context, query string) ([]SearchResult, error) {
	if e.skillReg == nil {
		return nil, fmt.Errorf("skill registry não inicializado")
	}
	results, err := e.skillReg.SearchAll(ctx, query, 20)
	if err != nil {
		return nil, err
	}

	final := make([]SearchResult, len(results))
	for i, res := range results {
		final[i] = SearchResult{
			Name:         res.DisplayName,
			DisplayName:  res.DisplayName,
			RegistryName: res.RegistryName,
			Summary:      res.Summary,
			Description:  res.Summary,
			Slug:         res.Slug,
			Version:      res.Version,
			Score:        res.Score,
		}
	}
	return final, nil
}

func (e *Engine) InstallSkill(ctx context.Context, registryName, slug, version string) error {
	reg := e.skillReg.GetRegistry(registryName)
	if reg == nil {
		return fmt.Errorf("registry %s not found", registryName)
	}

	// Define o diretório de destino (workspace atual)
	workspace := e.cfg.Agents.Defaults.Workspace
	targetDir := strings.Join([]string{workspace, "skills"}, "/")

	// Resolve o nome do diretório da skill
	dirName, err := reg.ResolveInstallDirName(slug)
	if err != nil {
		return err
	}
	fullTargetDir := strings.Join([]string{targetDir, dirName}, "/")

	_, err = reg.DownloadAndInstall(ctx, slug, version, fullTargetDir)
	return err
}

func (e *Engine) GetInstalledSkills() ([]string, error) {
	workspace := e.cfg.Agents.Defaults.Workspace
	targetDir := strings.Join([]string{workspace, "skills"}, "/")

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var skillsList []string
	for _, entry := range entries {
		if entry.IsDir() {
			skillsList = append(skillsList, entry.Name())
		}
	}
	return skillsList, nil
}

func (e *Engine) UninstallSkill(name string) error {
	workspace := e.cfg.Agents.Defaults.Workspace
	targetDir := strings.Join([]string{workspace, "skills", name}, "/")
	return os.RemoveAll(targetDir)
}

func (e *Engine) GetSkillDetails(name string) (string, error) {
	workspace := e.cfg.Agents.Defaults.Workspace
	filePath := strings.Join([]string{workspace, "skills", name, "SKILL.md"}, "/")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (e *Engine) GetSkillFullInfo(name string) (*SkillFullInfo, error) {
	workspace := e.cfg.Agents.Defaults.Workspace
	skillDir := strings.Join([]string{workspace, "skills", name}, "/")

	info := &SkillFullInfo{
		Name:     name,
		Registry: "local",
		Version:  "0.0.1",
	}

	// 1. Tentar ler SKILL.md
	mdPath := skillDir + "/SKILL.md"
	if data, err := os.ReadFile(mdPath); err == nil {
		info.Markdown = string(data)
		info.Raw = info.Markdown
		info.CharCount = len(data)
		info.LineCount = strings.Count(info.Raw, "\n") + 1
	}

	// 2. Tentar ler SKILL.json (manifesto oficial do Picoclaw)
	jsonPath := skillDir + "/SKILL.json"
	if data, err := os.ReadFile(jsonPath); err == nil {
		var manifest struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Version     string `json:"version"`
			Repository  string `json:"repository"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			if manifest.Description != "" {
				info.Description = manifest.Description
			}
			if manifest.Version != "" {
				info.Version = manifest.Version
			}
			if manifest.Repository != "" {
				info.URL = manifest.Repository
			}
		}
	}

	// 3. Como fallback para a descrição, tentar extrair do MD se ainda estiver vazia
	if info.Description == "" && info.Markdown != "" {
		lines := strings.Split(info.Markdown, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "#") {
				info.Description = l
				break
			}
		}
	}

	// 3. Simulação de metadados técnicos (em uma implementação real, leríamos o manifest da skill)
	info.URL = "https://github.com/sipeed/picoclaw/tree/main/skills/" + name
	
	return info, nil
}
