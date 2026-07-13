package orchestrator

import (
	"context"
	"testing"

	"ada-love-ai/pkg/providers"
)

// mockSubAgent is a lightweight test implementation of SubAgent.
type mockSubAgent struct {
	typ  AgentType
	caps []AgentCapability
	mdl  string
	sys  string
}

func (m *mockSubAgent) Type() AgentType      { return m.typ }
func (m *mockSubAgent) Model() string        { return m.mdl }
func (m *mockSubAgent) SystemPrompt() string { return m.sys }
func (m *mockSubAgent) Capabilities() []AgentCapability {
	return append([]AgentCapability(nil), m.caps...)
}
func (m *mockSubAgent) Tools() []string                     { return nil }
func (m *mockSubAgent) SetProvider(_ providers.LLMProvider) {}
func (m *mockSubAgent) Provider() providers.LLMProvider     { return nil }
func (m *mockSubAgent) SetModel(_ string)                   {}
func (m *mockSubAgent) Execute(_ context.Context, _ string, _ PromptLayers) (*AgentResult, error) {
	return &AgentResult{Success: true, Output: "ok"}, nil
}
func (m *mockSubAgent) Close() error { return nil }

func TestDecideCapsFromTask(t *testing.T) {
	cases := []struct {
		in   string
		want []AgentCapability
	}{
		{"Create a React component with hooks and UI", []AgentCapability{CapabilityReactFrontend, CapabilityUIComponents}},
		{"Implement REST API and Postgres database", []AgentCapability{CapabilityAPI, CapabilityDatabase}},
		{"Write unit tests and validate behavior", []AgentCapability{CapabilityTesting, CapabilityQualityAssurance}},
		{"Optimize goroutines and concurrency", []AgentCapability{CapabilityConcurrency}},
	}

	for _, c := range cases {
		got := decideCapsFromTask(c.in)
		for _, want := range c.want {
			found := false
			for _, g := range got {
				if g == want {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("decideCapsFromTask(%q) missing capability %v; got=%v", c.in, want, got)
			}
		}
	}
}

func TestChooseAgentByCapabilities(t *testing.T) {
	cfg := DefaultOrchestratorConfig()
	o := NewOrchestrator(cfg, "", "")
	// Replace registry with a clean one for deterministic test registration
	o.registry = NewAgentRegistry()

	// Register mock agents
	goAgent := &mockSubAgent{typ: AgentTypeGoLang, caps: []AgentCapability{CapabilityGoBackend, CapabilityAPI, CapabilityDatabase}, mdl: "m"}
	reactAgent := &mockSubAgent{typ: AgentTypeReact, caps: []AgentCapability{CapabilityReactFrontend, CapabilityUIComponents}, mdl: "r"}
	testerAgent := &mockSubAgent{typ: AgentTypeTester, caps: []AgentCapability{CapabilityTesting, CapabilityQualityAssurance}, mdl: "t"}

	o.registry.Register(goAgent)
	o.registry.Register(reactAgent)
	o.registry.Register(testerAgent)

	// Case 1: prefer API+Database -> Go agent
	desired := []AgentCapability{CapabilityAPI, CapabilityDatabase}
	if sub, ok := o.chooseAgentByCapabilities(desired); !ok {
		t.Fatalf("expected a match for %v, got none", desired)
	} else if sub.Type() != AgentTypeGoLang {
		t.Fatalf("expected golang agent, got %v", sub.Type())
	}

	// Case 2: prefer React frontend -> React agent
	desired = []AgentCapability{CapabilityReactFrontend}
	if sub, ok := o.chooseAgentByCapabilities(desired); !ok {
		t.Fatalf("expected a match for %v, got none", desired)
	} else if sub.Type() != AgentTypeReact {
		t.Fatalf("expected react agent, got %v", sub.Type())
	}

	// Case 3: capability not present -> no match
	desired = []AgentCapability{"nonexistent_capability"}
	if _, ok := o.chooseAgentByCapabilities(desired); ok {
		t.Fatalf("expected no match for nonexistent capability, but found one")
	}

	// Case 4: using agentTypeToCapabilities hint (golang)
	des := agentTypeToCapabilities(AgentTypeGoLang)
	if sub, ok := o.chooseAgentByCapabilities(des); !ok {
		t.Fatalf("expected an agent for hint golang, got none")
	} else if sub.Type() != AgentTypeGoLang {
		t.Fatalf("expected golang agent from hint, got %v", sub.Type())
	}
}
