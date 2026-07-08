package patterns

import "testing"

func TestSetupDefaults_ContemGoPatterns(t *testing.T) {
	repo := NewRepository()
	langs, ok := repo.Store["go"]
	if !ok {
		t.Fatalf("expected 'go' language patterns, missing")
	}
	if len(langs) == 0 {
		t.Fatal("expected non-empty pattern list for 'go'")
	}
	if !containsPattern(langs, "clean_architecture", "Architecture") {
		t.Errorf("expected 'clean_architecture' in Architecture category for go")
	}
	if !containsPattern(langs, "kiss", "Philosophy") {
		t.Errorf("expected 'kiss' philosophy for go")
	}
}

func TestSetupDefaults_MobilePossuiRiverpod(t *testing.T) {
	repo := NewRepository()
	flutterPatterns := repo.Store["flutter"]
	if !containsPattern(flutterPatterns, "riverpod", "StateManagement") {
		t.Error("expected 'riverpod' in StateManagement for flutter")
	}
}

func TestSetupDefaults_BackendSemRiverpod(t *testing.T) {
	repo := NewRepository()
	goPatterns := repo.Store["go"]
	for _, p := range goPatterns {
		if p.ID == "riverpod" {
			t.Error("riverpod deve ser mobile-scope, nao deve aparecer em 'go'")
		}
	}
}

func containsPattern(list []Pattern, id, category string) bool {
	for _, p := range list {
		if p.ID == id && p.Category == category {
			return true
		}
	}
	return false
}
