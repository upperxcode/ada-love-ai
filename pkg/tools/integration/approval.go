package integrationtools

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ApprovalDecision is the user's response to a tool approval request.
type ApprovalDecision struct {
	Approved bool
	Reason   string
}

// ApprovalRequest carries the details of a tool call that needs user approval.
type ApprovalRequest struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Tool      string `json:"tool"`
	Args      string `json:"args"`
}

// ToolCategory represents the category of a tool for approval purposes.
type ToolCategory string

const (
	ToolCategoryRead  ToolCategory = "read"
	ToolCategoryWrite ToolCategory = "write"
)

// ApprovalRegistry manages pending tool approval requests and their response channels.
type ApprovalRegistry struct {
	mu        sync.Mutex
	pending   map[string]chan ApprovalDecision // requestID -> response channel
	onApprove func(req ApprovalRequest)
	resolver  func(opaqueKey string) string

	// cachedApprovals stores previously approved tools by session and category
	// Key: "sessionId:toolName" for read tools, "sessionId:write" for write tools in current iteration
	cachedApprovals map[string]bool
	// writeApprovalIteration tracks the current iteration for write approvals
	currentWriteIteration        int
	lastApprovedToolForIteration string
}

// NewApprovalRegistry creates a new ApprovalRegistry.
func NewApprovalRegistry() *ApprovalRegistry {
	return &ApprovalRegistry{
		pending:         make(map[string]chan ApprovalDecision),
		cachedApprovals: make(map[string]bool),
	}
}

// SetResolver sets the function used to resolve opaque session keys to frontend sessionIDs.
func (ar *ApprovalRegistry) SetResolver(fn func(opaqueKey string) string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.resolver = fn
}

// ResolveSessionKey converts a session key (opaque or ada:-prefixed) to a frontend sessionID.
func (ar *ApprovalRegistry) ResolveSessionKey(sessionKey string) string {
	ar.mu.Lock()
	resolver := ar.resolver
	ar.mu.Unlock()
	if resolver != nil {
		return resolver(sessionKey)
	}
	return sessionKeyToID(sessionKey)
}

// OnApprove registers a callback for when an approval is requested.
func (ar *ApprovalRegistry) OnApprove(fn func(req ApprovalRequest)) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.onApprove = fn
}

// WaitForApproval blocks until the user responds or the timeout/context expires.
func (ar *ApprovalRegistry) WaitForApproval(done <-chan struct{}, req ApprovalRequest, timeout time.Duration) (ApprovalDecision, bool) {
	ch := make(chan ApprovalDecision, 1)

	ar.mu.Lock()
	ar.pending[req.ID] = ch
	onApprove := ar.onApprove
	ar.mu.Unlock()

	if onApprove != nil {
		onApprove(req)
	}

	defer func() {
		ar.mu.Lock()
		delete(ar.pending, req.ID)
		ar.mu.Unlock()
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case decision := <-ch:
		return decision, true
	case <-timer.C:
		return ApprovalDecision{Approved: false, Reason: "approval timed out"}, false
	case <-done:
		return ApprovalDecision{Approved: false, Reason: "interrupted"}, false
	}
}

// Respond delivers the user's approval decision to the waiting hook.
func (ar *ApprovalRegistry) Respond(requestID string, approved bool, reason string) bool {
	ar.mu.Lock()
	ch, ok := ar.pending[requestID]
	ar.mu.Unlock()

	if !ok {
		return false
	}

	select {
	case ch <- ApprovalDecision{Approved: approved, Reason: reason}:
		return true
	default:
		return false
	}
}

// IsReadOnlyTool returns true for tools that don't modify state and don't need approval.
func IsReadOnlyTool(tool string) bool {
	switch tool {
	case "read_file", "list_dir", "grep", "web", "web_search", "web_fetch",
		"memory_search", "memory_list", "ask_user", "message", "reaction",
		"find_skills", "load_image":
		return true
	default:
		return false
	}
}

// GetToolCategory returns the category of a tool (read or write).
func GetToolCategory(tool string) ToolCategory {
	if IsReadOnlyTool(tool) {
		return ToolCategoryRead
	}
	// Write tools: write_file, edit_file, etc.
	return ToolCategoryWrite
}

// IsCachedApproval checks if a tool was previously approved in this session.
// For read tools: checks if the specific tool was approved
// For write tools: checks if write permission was granted for this iteration
func (ar *ApprovalRegistry) IsCachedApproval(sessionID, tool string) bool {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	key := fmt.Sprintf("%s:%s", sessionID, tool)
	return ar.cachedApprovals[key]
}

// CacheApproval stores an approval decision for a tool in this session.
func (ar *ApprovalRegistry) CacheApproval(sessionID, tool string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	key := fmt.Sprintf("%s:%s", sessionID, tool)
	ar.cachedApprovals[key] = true
}

// StartNewWriteIteration resets write approvals for a new question/iteration.
// Read tool approvals are cached for the entire session.
func (ar *ApprovalRegistry) StartNewWriteIteration(sessionID string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Clear write tool approvals for this session (they expire per iteration)
	ar.currentWriteIteration++
	ar.lastApprovedToolForIteration = ""
}

// IsWriteToolApproved checks if write permission was granted for this iteration.
// Returns true if any write tool was approved in the current iteration.
func (ar *ApprovalRegistry) IsWriteToolApproved(sessionID string) bool {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// If we have a cached approval for a write tool in this iteration
	if ar.lastApprovedToolForIteration != "" {
		return true
	}
	return false
}

// CacheWriteApproval marks that write permission was granted for this iteration.
func (ar *ApprovalRegistry) CacheWriteApproval(sessionID string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	ar.currentWriteIteration++
	ar.lastApprovedToolForIteration = sessionID + ":write"
}

// ClearCachedApprovals clears all cached approvals for a session.
func (ar *ApprovalRegistry) ClearCachedApprovals(sessionID string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	ar.cachedApprovals = make(map[string]bool)
}

// FormatArgs converts the tool arguments map to a compact string representation.
func FormatArgs(args map[string]any) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, 0, len(args))
	for k, v := range args {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// GenerateApprovalID creates a random hex ID for approval requests.
func GenerateApprovalID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
