package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (e *Engine) SearchSkills(query string) ([]SearchResult, error) {
	if e.skillReg == nil {
		return nil, fmt.Errorf("skill registry n\u00e3o inicializado")
	}
	results, err := e.skillReg.SearchAll(context.Background(), query, 20)
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

func (e *Engine) InstallSkill(registryName, slug, version string) error {
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

	_, err = reg.DownloadAndInstall(context.Background(), slug, version, fullTargetDir)
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

	// 1. Read SKILL.md
	mdPath := skillDir + "/SKILL.md"
	if data, err := os.ReadFile(mdPath); err == nil {
		info.Markdown = string(data)
		info.Raw = info.Markdown
		info.CharCount = len(data)
		info.LineCount = strings.Count(info.Raw, "\n") + 1

			// Extract frontmatter fields (name, description, tags)
			fmName, fmDesc, fmTags := extractFrontmatter(data)
			if fmName != "" {
				info.Name = fmName
			}
			if fmDesc != "" {
				info.Description = fmDesc
			}
			if len(fmTags) > 0 {
				info.Tags = fmTags
			}
	}

	// 2. Read SKILL.json (Picoclaw manifest) — overrides frontmatter
	jsonPath := skillDir + "/SKILL.json"
	if data, err := os.ReadFile(jsonPath); err == nil {
		var manifest struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Version     string `json:"version"`
			Repository  string `json:"repository"`
			Tags        string `json:"tags"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			if manifest.Name != "" {
				info.Name = manifest.Name
			}
			if manifest.Description != "" {
				info.Description = manifest.Description
			}
			if manifest.Version != "" {
				info.Version = manifest.Version
			}
			if manifest.Repository != "" {
				info.URL = manifest.Repository
			}
			if manifest.Tags != "" {
				info.Tags = splitTags(manifest.Tags)
			}
		}
	}

	// 3. Fallback: extract description from first non-heading line
	if info.Description == "" && info.Markdown != "" {
		lines := strings.Split(info.Markdown, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "#") && !strings.HasPrefix(l, "---") {
				info.Description = l
				break
			}
		}
	}

	return info, nil
}

// extractFrontmatter parses YAML-style frontmatter from SKILL.md content.
// Returns name, description, and tags. Returns empty strings/nil if not found.
func extractFrontmatter(data []byte) (name, description string, tags []string) {
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", nil
	}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			break
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"`)
		switch key {
		case "name":
			name = val
		case "description":
			description = val
		case "tags":
			tags = splitTags(val)
		}
	}
	return name, description, tags
}

// splitTags splits a comma-separated tags string into a cleaned slice.
func splitTags(s string) []string {
	var result []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

// SaveCustomSkill creates or updates a skill's SKILL.md with frontmatter.
func (e *Engine) SaveCustomSkill(name, description, tagsCSV, content string) error {
	workspace := e.cfg.Agents.Defaults.Workspace
	skillDir := strings.Join([]string{workspace, "skills", name}, "/")

	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", name))
	sb.WriteString(fmt.Sprintf("description: %s\n", description))
	sb.WriteString(fmt.Sprintf("tags: %s\n", tagsCSV))
	sb.WriteString("---\n\n")
	sb.WriteString(content)
	sb.WriteString("\n")

	return os.WriteFile(skillDir+"/SKILL.md", []byte(sb.String()), 0o644)
}
