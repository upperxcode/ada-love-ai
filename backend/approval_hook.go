package backend

import (
	"context"
	"time"

	"ada-love-ai/pkg/agent"
	integration "ada-love-ai/pkg/tools/integration"
)

// FrontendApprovalHook is a ToolApprover that blocks on user approval from the frontend.
type FrontendApprovalHook struct {
	registry *integration.ApprovalRegistry
	timeout  time.Duration
}

// NewFrontendApprovalHook creates a new FrontendApprovalHook.
func NewFrontendApprovalHook(registry *integration.ApprovalRegistry, timeout time.Duration) *FrontendApprovalHook {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	return &FrontendApprovalHook{
		registry: registry,
		timeout:  timeout,
	}
}

// ApproveTool implements agent.ToolApprover.
func (h *FrontendApprovalHook) ApproveTool(ctx context.Context, req *agent.ToolApprovalRequest) (agent.ApprovalDecision, error) {
	if req == nil {
		return agent.ApprovalDecision{Approved: true}, nil
	}

	// Read-only tools pass through without approval
	if integration.IsReadOnlyTool(req.Tool) {
		return agent.ApprovalDecision{Approved: true}, nil
	}

	requestID := integration.GenerateApprovalID()
	sessionID := h.registry.ResolveSessionKey(req.Meta.SessionKey)

	approvalReq := integration.ApprovalRequest{
		ID:        requestID,
		SessionID: sessionID,
		Tool:      req.Tool,
		Args:      integration.FormatArgs(req.Arguments),
	}

	// We use ctx.Done() as the interrupt channel; HardAbort cancels the turn context.
	decision, ok := h.registry.WaitForApproval(ctx.Done(), approvalReq, h.timeout)
	if !ok {
		return agent.ApprovalDecision{Approved: false, Reason: "User did not respond in time"}, nil
	}
	return agent.ApprovalDecision{Approved: decision.Approved, Reason: decision.Reason}, nil
}

// sessionKeyToID strips the "ada:" prefix from a session key.
func sessionKeyToID(key string) string {
	if len(key) > 4 && key[:4] == "ada:" {
		return key[4:]
	}
	return key
}
