package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "experts.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	return path
}

func TestLoadExperts_PopulaStore(t *testing.T) {
	yamlContent := `
experts:
  - id: "go-expert"
    name: "Go Expert"
    description: "Specialist in Go applications"
    language: "go"
    start_command: "./expert"
    triggers: ["go.mod"]
`
	path := writeTempYAML(t, yamlContent)
	plugins, err := LoadExperts(path)
	if err != nil {
		t.Fatalf("LoadExperts returned error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	p := plugins[0]
	if p.ID != "go-expert" || p.Language != "go" || p.StartCommand != "./expert" {
		t.Errorf("plugin parse wrong: %+v", p)
	}
}

func TestLoadExperts_MesclaDuplicatasPorID(t *testing.T) {
	yamlA := `
experts:
  - id: "go-expert"
    name: "Go Expert"
    language: "go"
    start_command: "./a"
`
	yamlB := `
experts:
  - id: "go-expert"
    name: "Go Expert v2"
    language: "go"
    start_command: "./b"
`
	pathA := writeTempYAML(t, yamlA)
	pathB := writeTempYAML(t, yamlB)

	plugins, err := LoadExperts(pathA, pathB)
	if err != nil {
		t.Fatalf("merge err: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected dedupe to 1, got %d", len(plugins))
	}
	if plugins[0].StartCommand != "./b" {
		t.Errorf("expected last-wins, got %s", plugins[0].StartCommand)
	}
}

func TestLoadExperts_MissingFileIsOK(t *testing.T) {
	plugins, err := LoadExperts("/nonexistent/experts.yaml")
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins for missing file, got %d", len(plugins))
	}
}

func TestFindExpertByLanguage(t *testing.T) {
	yamlContent := `
experts:
  - id: "go-expert"
    name: "Go Expert"
    language: "go"
  - id: "py-expert"
    name: "Python Expert"
    language: "python"
`
	path := writeTempYAML(t, yamlContent)
	plugins, err := LoadExperts(path)
	if err != nil {
		t.Fatal(err)
	}

	got, err := FindExpertByLanguage("python", plugins)
	if err != nil {
		t.Fatalf("FindExpertByLanguage: %v", err)
	}
	if got.ID != "py-expert" {
		t.Errorf("expected py-expert, got %s", got.ID)
	}

	_, err = FindExpertByLanguage("rust", plugins)
	if err == nil {
		t.Error("expected error for rust")
	}
}
