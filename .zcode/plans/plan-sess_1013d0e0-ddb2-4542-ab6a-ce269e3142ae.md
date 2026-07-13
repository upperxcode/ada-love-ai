Goal: Replace the ad-hoc hardcoded intent→agent mapping in backend/engine.go with a centralized, testable mapping component in pkg/orchestrator (Option A: quick, testable), and wire Engine to use it. Keep behavior compatible and add unit tests.

Phase 1: Read-only exploration (already completed)
- Located static mapping and classifier in backend/engine.go (ProcessOrchestrated). Found switch on labels and candidate list around lines 1766–1796 and lookup + RunAgentByID at 1798–1850.
- Found orchestrator capability selection and registry in pkg/orchestrator/orchestrator.go (chooseAgentByCapabilities, agentTypeToCapabilities, decideCapsFromTask).
- Found routing helpers in pkg/routing for normalization (NormalizeAgentID) and in pkg/tinybrain for classifier.

Phase 2: Design (concrete plan to implement Option A)
1) New package/file: pkg/orchestrator/intent_map.go
   - Provide public functions:
     - func CandidateLabels() []string
       - returns the list of candidate labels to pass into tinybrain.ClassifyWithCandidates; default list: {"React","Go","Tester","Geral"}
       - reads override from orchestrator.Config.AvailableLabels or from adaCfg if present (for future extensibility) but for now a simple default with optional env/config hook.
     - func ResolveIntentCandidates(label string, adaCfg *backend.AdaConfig) []string
       - normalize label and return aliases slice (e.g. "go" → {"go","golang","go_agent","golang_agent","galang"})
       - support adaCfg override via a map (if adaCfg has IntentMap field) but default to built-in map
   - Keep function small and deterministic; include unit tests for expected mappings.

2) Refactor backend/engine.go ProcessOrchestrated
   - Replace existing hardcoded candidateLabels and switch with calls to orchestrator.CandidateLabels() and orchestrator.ResolveIntentCandidates(topLabel, &e.adaCfg)
   - Preserve behavior for GENERAL label: still bypass orchestrator via agentLoop.ProcessDirect.
   - Keep fallback sequence intact: foundAgentID via registry.GetAgent (IDs normalized) -> match by agent.Name -> session.WorkerName -> orchestrator LLM route.
   - Ensure logging remains informative.

3) Tests
   - New unit tests for pkg/orchestrator/intent_map_test.go covering: label normalization, alias lists, behavior on unknown labels.
   - Add Engine-level test to verify ProcessOrchestrated uses orchestrator.CandidateLabels() indirectly (mock tinybrain.ClassifyWithCandidates to return a label and assert RunAgentByID is called). If mocking is tricky, add a test for the helper integration.

4) Backwards compatibility and rollout
   - Feature is opt-in in code; behavior remains same unless intent_map is changed. No DB changes in Phase 1.
   - Add TODO comments where hook to adaCfg/db should be implemented if later desired.

Phase 3: Implementation steps (what I will change during coding)
- Create file pkg/orchestrator/intent_map.go with mapping functions and tests.
- Update imports and call sites in backend/engine.go: replace candidateLabels + switch with new helper calls.
- Run unit tests: pkg/orchestrator tests and pkg/agent tests (quick run of affected packages).
- Commit changes on feature branch (feat/intent-map-centralize) with descriptive commit message.

Safety and testing
- Keep the GENERAL bypass behavior identical to current code.
- Add unit tests for mapping logic to catch regressions.
- Ensure no DB schema changes in this phase.

Deliverable
- Clean refactor with a single source-of-truth for intent→agent mapping, unit-tested and easy to extend to DB overrides in Phase 2.

I request approval to implement this plan (make code edits and run tests).