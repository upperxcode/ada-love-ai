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

func TestGetPatternsForLanguage(t *testing.T) {
	repo := NewRepository()

	// Valid language: should return the expected list for "go".
	got := repo.GetPatternsForLanguage("go")
	if got == nil {
		t.Fatal("expected non-nil slice for known language 'go'")
	}
	if len(got) == 0 {
		t.Fatal("expected non-empty pattern list for 'go'")
	}
	if !containsPattern(got, "clean_architecture", "Architecture") {
		t.Errorf("expected 'clean_architecture' in Architecture category for go")
	}

	// Unknown language: must return non-nil empty slice (not nil).
	// Returning nil breaks Wails JSON serialization (nil -> null), but the
	// frontend TypeScript expects Pattern[] (i.e. []). Pin this contract.
	got = repo.GetPatternsForLanguage("nonexistent")
	if got == nil {
		t.Fatal("expected non-nil empty slice for unknown language")
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice for unknown language, got %d items", len(got))
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
