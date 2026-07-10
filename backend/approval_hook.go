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

	// Resolve session ID
	sessionID := h.registry.ResolveSessionKey(req.Meta.SessionKey)

	// Check if this is a read-only tool that was already approved in this session
	if integration.IsReadOnlyTool(req.Tool) {
		if h.registry.IsCachedApproval(sessionID, req.Tool) {
			return agent.ApprovalDecision{Approved: true, Reason: "Read tool previously approved"}, nil
		}
	}

	// Check if this is a write tool and write permission was granted for this iteration
	if integration.GetToolCategory(req.Tool) == integration.ToolCategoryWrite {
		if h.registry.IsWriteToolApproved(sessionID) {
			return agent.ApprovalDecision{Approved: true, Reason: "Write permission granted for this iteration"}, nil
		}
	}

	// Read-only tools that haven't been approved yet - ask user
	if integration.IsReadOnlyTool(req.Tool) {
		requestID := integration.GenerateApprovalID()
		approvalReq := integration.ApprovalRequest{
			ID:        requestID,
			SessionID: sessionID,
			Tool:      req.Tool,
			Args:      integration.FormatArgs(req.Arguments),
		}

		decision, ok := h.registry.WaitForApproval(ctx.Done(), approvalReq, h.timeout)
		if !ok {
			return agent.ApprovalDecision{Approved: false, Reason: "User did not respond in time"}, nil
		}
		
		// Cache the approval for this read tool
		if decision.Approved {
			h.registry.CacheApproval(sessionID, req.Tool)
		}
		
		return agent.ApprovalDecision{Approved: decision.Approved, Reason: decision.Reason}, nil
	}

	// Write tools - ask user about writing to file
	requestID := integration.GenerateApprovalID()
	
	// Build a more descriptive message for write tools
	toolDesc := req.Tool
	if req.Tool == "write_file" {
		toolDesc = "write_file"
	} else if req.Tool == "edit_file" {
		toolDesc = "edit_file"
	}

	approvalReq := integration.ApprovalRequest{
		ID:        requestID,
		SessionID: sessionID,
		Tool:      toolDesc,
		Args:      integration.FormatArgs(req.Arguments),
	}

	decision, ok := h.registry.WaitForApproval(ctx.Done(), approvalReq, h.timeout)
	if !ok {
		return agent.ApprovalDecision{Approved: false, Reason: "User did not respond in time"}, nil
	}
	
	// If approved, cache write permission for this iteration
	if decision.Approved {
		h.registry.CacheWriteApproval(sessionID)
	}
	
	return agent.ApprovalDecision{Approved: decision.Approved, Reason: decision.Reason}, nil
}