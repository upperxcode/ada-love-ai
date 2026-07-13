package orchestrator

import "testing"

func TestResolveIntentCandidates_Defaults(t *testing.T) {
	cases := []struct {
		label   string
		wantMin int
	}{
		{"React", 1},
		{"flutter", 1},
		{"Go", 1},
		{"Tester", 1},
		{"unknown", 1},
	}

	for _, c := range cases {
		res := ResolveIntentCandidates(c.label)
		if res == nil || len(res) < c.wantMin {
			t.Fatalf("ResolveIntentCandidates(%q) returned %v, want at least %d entries", c.label, res, c.wantMin)
		}
	}
}
