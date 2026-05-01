package backend

import (
	"context"
	"slices"
	"strings"
	"sync"
)

type Message = ChatMessage

type PromptLayer string

const (
	PromptLayerKernel      PromptLayer = "kernel"
	PromptLayerInstruction PromptLayer = "instruction"
	PromptLayerCapability  PromptLayer = "capability"
	PromptLayerContext     PromptLayer = "context"
	PromptLayerTurn        PromptLayer = "turn"
)

type PromptSlot string

const (
	PromptSlotIdentity     PromptSlot = "identity"
	PromptSlotHierarchy    PromptSlot = "hierarchy"
	PromptSlotWorkspace    PromptSlot = "workspace"
	PromptSlotTooling      PromptSlot = "tooling"
	PromptSlotSkillCatalog PromptSlot = "skill_catalog"
	PromptSlotActiveSkill  PromptSlot = "active_skill"
	PromptSlotMemory       PromptSlot = "memory"
	PromptSlotRuntime      PromptSlot = "runtime"
	PromptSlotSummary      PromptSlot = "summary"
	PromptSlotMessage      PromptSlot = "message"
	PromptSlotSteering     PromptSlot = "steering"
	PromptSlotSubTurn      PromptSlot = "subturn"
	PromptSlotInterrupt    PromptSlot = "interrupt"
	PromptSlotOutput       PromptSlot = "output"
)

type PromptSourceID string

const (
	PromptSourceKernel       PromptSourceID = "runtime.kernel"
	PromptSourceWorkspace    PromptSourceID = "workspace.definition"
	PromptSourceRuntime      PromptSourceID = "runtime.context"
	PromptSourceSummary      PromptSourceID = "context.summary"
	PromptSourceMemory       PromptSourceID = "memory:workspace"
	PromptSourceSkillCatalog PromptSourceID = "skill:index"
	PromptSourceToolRegistry PromptSourceID = "tool_registry:native"
	PromptSourceOutputPolicy PromptSourceID = "runtime.output"
	PromptSourceUserMessage  PromptSourceID = "turn:user_message"
)

type PromptCachePolicy string

const (
	PromptCacheDefault   PromptCachePolicy = ""
	PromptCacheEphemeral PromptCachePolicy = "ephemeral"
	PromptCacheNone      PromptCachePolicy = "none"
)

type PromptPlacement struct {
	Layer PromptLayer
	Slot  PromptSlot
}

type PromptSourceDescriptor struct {
	ID              PromptSourceID
	Owner           string
	Description     string
	Allowed         []PromptPlacement
	StableByDefault bool
}

type PromptSource struct {
	ID   PromptSourceID
	Name string
}

type PromptPart struct {
	ID      string
	Layer   PromptLayer
	Slot    PromptSlot
	Source  PromptSource
	Title   string
	Content string
	Stable  bool
	Cache   PromptCachePolicy
}

type PromptBuildRequest struct {
	History []Message
	Summary string

	CurrentMessage string
	Media          []string

	ActiveSkills []string
	Overlays     []PromptPart
}

type PromptContributor interface {
	PromptSource() PromptSourceDescriptor
	ContributePrompt(ctx context.Context, req PromptBuildRequest) ([]PromptPart, error)
}

type PromptRegistry struct {
	mu           sync.RWMutex
	sources      map[PromptSourceID]PromptSourceDescriptor
	contributors []PromptContributor
}

func NewPromptRegistry() *PromptRegistry {
	r := &PromptRegistry{
		sources: make(map[PromptSourceID]PromptSourceDescriptor),
	}
	for _, desc := range builtinPromptSources() {
		r.RegisterSource(desc)
	}
	return r
}

func builtinPromptSources() []PromptSourceDescriptor {
	return []PromptSourceDescriptor{
		{
			ID:              PromptSourceKernel,
			Owner:           "engine",
			Description:     "Core Ada Love identity and rules",
			Allowed:         []PromptPlacement{{Layer: PromptLayerKernel, Slot: PromptSlotIdentity}},
			StableByDefault: true,
		},
		{
			ID:              PromptSourceWorkspace,
			Owner:           "workspace",
			Description:     "Workspace and agent definition files",
			Allowed:         []PromptPlacement{{Layer: PromptLayerInstruction, Slot: PromptSlotWorkspace}},
			StableByDefault: true,
		},
		{
			ID:              PromptSourceToolRegistry,
			Owner:           "tools",
			Description:     "Native tool definitions",
			Allowed:         []PromptPlacement{{Layer: PromptLayerCapability, Slot: PromptSlotTooling}},
			StableByDefault: true,
		},
		{
			ID:              PromptSourceSkillCatalog,
			Owner:           "skills",
			Description:     "Installed skill catalog",
			Allowed:         []PromptPlacement{{Layer: PromptLayerCapability, Slot: PromptSlotSkillCatalog}},
			StableByDefault: true,
		},
		{
			ID:              PromptSourceMemory,
			Owner:           "memory",
			Description:     "Workspace memory context",
			Allowed:         []PromptPlacement{{Layer: PromptLayerContext, Slot: PromptSlotMemory}},
			StableByDefault: true,
		},
		{
			ID:              PromptSourceSummary,
			Owner:           "context_manager",
			Description:     "Conversation summary context",
			Allowed:         []PromptPlacement{{Layer: PromptLayerContext, Slot: PromptSlotSummary}},
			StableByDefault: false,
		},
		{
			ID:              PromptSourceUserMessage,
			Owner:           "turn",
			Description:     "Current user message for this turn",
			Allowed:         []PromptPlacement{{Layer: PromptLayerTurn, Slot: PromptSlotMessage}},
			StableByDefault: false,
		},
	}
}

func (r *PromptRegistry) RegisterSource(desc PromptSourceDescriptor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[desc.ID] = desc
}

func (r *PromptRegistry) RegisterContributor(contributor PromptContributor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contributors = append(r.contributors, contributor)
}

func (r *PromptRegistry) Collect(ctx context.Context, req PromptBuildRequest) ([]PromptPart, error) {
	r.mu.RLock()
	contributors := append([]PromptContributor(nil), r.contributors...)
	r.mu.RUnlock()

	var parts []PromptPart
	for _, contributor := range contributors {
		contributed, err := contributor.ContributePrompt(ctx, req)
		if err != nil {
			return nil, err
		}
		parts = append(parts, contributed...)
	}
	return parts, nil
}

type PromptStack struct {
	parts []PromptPart
}

func NewPromptStack() *PromptStack {
	return &PromptStack{}
}

func (s *PromptStack) Add(part PromptPart) {
	if strings.TrimSpace(part.Content) == "" {
		return
	}
	s.parts = append(s.parts, part)
}

func (s *PromptStack) Render() string {
	textParts := make([]string, 0, len(s.parts))
	for _, part := range sortPromptParts(s.parts) {
		textParts = append(textParts, part.Content)
	}
	return strings.Join(textParts, "\n\n---\n\n")
}

func sortPromptParts(parts []PromptPart) []PromptPart {
	sorted := append([]PromptPart(nil), parts...)
	slices.SortStableFunc(sorted, func(a, b PromptPart) int {
		if d := layerPriority(b.Layer) - layerPriority(a.Layer); d != 0 {
			return d
		}
		if d := slotPriority(b.Slot) - slotPriority(a.Slot); d != 0 {
			return d
		}
		return strings.Compare(string(a.Source.ID), string(b.Source.ID))
	})
	return sorted
}

func layerPriority(layer PromptLayer) int {
	switch layer {
	case PromptLayerKernel:
		return 100
	case PromptLayerInstruction:
		return 80
	case PromptLayerCapability:
		return 60
	case PromptLayerContext:
		return 40
	case PromptLayerTurn:
		return 20
	default:
		return 0
	}
}

func slotPriority(slot PromptSlot) int {
	switch slot {
	case PromptSlotIdentity:
		return 1000
	case PromptSlotWorkspace:
		return 900
	case PromptSlotTooling:
		return 800
	case PromptSlotSkillCatalog:
		return 780
	case PromptSlotMemory:
		return 700
	case PromptSlotSummary:
		return 680
	case PromptSlotMessage:
		return 600
	default:
		return 0
	}
}
