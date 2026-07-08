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

// ApprovalRegistry manages pending tool approval requests and their response channels.
type ApprovalRegistry struct {
	mu        sync.Mutex
	pending   map[string]chan ApprovalDecision // requestID -> response channel
	onApprove func(req ApprovalRequest)
	resolver  func(opaqueKey string) string
}

// NewApprovalRegistry creates a new ApprovalRegistry.
func NewApprovalRegistry() *ApprovalRegistry {
	return &ApprovalRegistry{
		pending: make(map[string]chan ApprovalDecision),
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
