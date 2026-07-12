package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ada-love-ai/pkg/providers"
)

// ContextLogEntry represents a single log entry for what a chat sent as context
type ContextLogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	SessionID   string    `json:"session_id"`
	WorkspaceID string    `json:"workspace_id"`
	WorkerName  string    `json:"worker_name"`
	Model       string    `json:"model"`
	Provider    string    `json:"provider"`
	Mode        string    `json:"mode"`
	Thinking    string    `json:"thinking"`
	UserMessage string    `json:"user_message"`
	FullPrompt  string    `json:"full_prompt,omitempty"`
	// Complete context actually sent to the LLM
	SystemPrompt string           `json:"system_prompt,omitempty"`
	Messages     []ContextMessage `json:"messages,omitempty"`
	ToolCount    int              `json:"tool_count,omitempty"`
	Tools        []string         `json:"tools,omitempty"`
	HistoryCount int              `json:"history_count"`
	History      []ChatMessage    `json:"history,omitempty"`
	MessageCount int              `json:"message_count"`
	Error        string           `json:"error,omitempty"`
}

// ContextMessage mirrors providers.Message for JSON logging
type ContextMessage struct {
	Role         string `json:"role"`
	Content      string `json:"content"`
	HasToolCalls bool   `json:"has_tool_calls,omitempty"`
}

// ContextLogger manages logging of context sent to LLMs
type ContextLogger struct {
	mu            sync.Mutex
	logPath       string
	enabled       bool
	file          *os.File
	buffered      bool
	lastEntryTime time.Time
}

var (
	globalContextLogger *ContextLogger
	loggerOnce          sync.Once
)

// InitContextLogger initializes the global context logger
func InitContextLogger(logPath string, enabled bool) *ContextLogger {
	loggerOnce.Do(func() {
		globalContextLogger = &ContextLogger{
			logPath:  logPath,
			enabled:  enabled,
			buffered: true,
		}
		if enabled && logPath != "" {
			if err := globalContextLogger.initFile(); err != nil {
				fmt.Printf("[ContextLogger] Error initializing log file: %v\n", err)
				globalContextLogger.enabled = false
			}
		}
	})
	return globalContextLogger
}

func (cl *ContextLogger) initFile() error {
	if cl.logPath == "" || !cl.enabled {
		return nil
	}

	dir := filepath.Dir(cl.logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(cl.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	cl.mu.Lock()
	cl.file = file
	cl.mu.Unlock()

	// Write header if file is empty
	info, _ := file.Stat()
	if info.Size() == 0 {
		cl.writeLine([]byte("[\n"))
	}

	return nil
}

// LogContext logs a context entry
func (cl *ContextLogger) LogContext(entry ContextLogEntry) {
	if cl == nil || !cl.enabled {
		return
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.file == nil {
		return
	}

	// Add comma separator for subsequent entries
	if entry.Timestamp.After(time.Time{}) && cl.lastEntryTime.IsZero() == false {
		cl.file.Write([]byte(",\n"))
	}
	cl.lastEntryTime = entry.Timestamp

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		fmt.Printf("[ContextLogger] Error marshaling entry: %v\n", err)
		return
	}

	// Write to file
	_, err = cl.file.Write(data)
	if err != nil {
		fmt.Printf("[ContextLogger] Error writing to file: %v\n", err)
	}

	// Also print to console for debugging
	fmt.Printf("[ContextLogger] Session=%s Worker=%s Model=%s Messages=%d UserMsg=%q...\n",
		entry.SessionID, entry.WorkerName, entry.Model, entry.MessageCount, truncateString(entry.UserMessage, 50))
}

var lastEntryTime time.Time

func (cl *ContextLogger) writeLine(data []byte) {
	if cl.file != nil {
		cl.file.Write(data)
	}
}

// Close closes the log file
func (cl *ContextLogger) Close() error {
	if cl == nil || cl.file == nil {
		return nil
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.writeLine([]byte("\n]"))
	err := cl.file.Close()
	cl.file = nil
	return err
}

// IsEnabled returns whether logging is enabled
func (cl *ContextLogger) IsEnabled() bool {
	return cl != nil && cl.enabled
}

// SetEnabled enables or disables logging
func (cl *ContextLogger) SetEnabled(enabled bool) {
	if cl == nil {
		return
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	if enabled && cl.file == nil {
		cl.initFile()
	}
	cl.enabled = enabled
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetContextLogger returns the global context logger instance
func GetContextLogger() *ContextLogger {
	return globalContextLogger
}

// LogChatContext logs what context is being sent to a chat
// This is a convenience function that creates and logs a ContextLogEntry
func LogChatContext(
	sessionID string,
	workspaceID string,
	workerName string,
	model string,
	provider string,
	mode string,
	thinking string,
	userMessage string,
	fullPrompt string,
	history []ChatMessage,
) {
	cl := GetContextLogger()
	if cl == nil || !cl.IsEnabled() {
		return
	}

	entry := ContextLogEntry{
		Timestamp:    time.Now(),
		SessionID:    sessionID,
		WorkspaceID:  workspaceID,
		WorkerName:   workerName,
		Model:        model,
		Provider:     provider,
		Mode:         mode,
		Thinking:     thinking,
		UserMessage:  userMessage,
		FullPrompt:   fullPrompt,
		HistoryCount: len(history),
		History:      history,
		MessageCount: len(history),
	}

	cl.LogContext(entry)
}

// LogFullContext logs the COMPLETE context sent to the LLM: system prompt,
// full message list (history + current turn), and tool definitions.
// This is the real payload the model receives.
func LogFullContext(
	sessionID string,
	agentID string,
	model string,
	mode string,
	messages []providers.Message,
	toolDefs []providers.ToolDefinition,
	userMessage string,
) {
	cl := GetContextLogger()
	if cl == nil || !cl.IsEnabled() {
		return
	}

	// Extract system prompt (first system message)
	var systemPrompt string
	var ctxMessages []ContextMessage
	for _, m := range messages {
		if m.Role == "system" && systemPrompt == "" {
			systemPrompt = m.Content
		}
		hasToolCalls := len(m.ToolCalls) > 0
		ctxMessages = append(ctxMessages, ContextMessage{
			Role:         m.Role,
			Content:      m.Content,
			HasToolCalls: hasToolCalls,
		})
	}

	toolNames := make([]string, 0, len(toolDefs))
	for _, td := range toolDefs {
		if td.Function.Name != "" {
			toolNames = append(toolNames, td.Function.Name)
		}
	}

	entry := ContextLogEntry{
		Timestamp:    time.Now(),
		SessionID:    sessionID,
		WorkerName:   agentID,
		Model:        model,
		Mode:         mode,
		UserMessage:  userMessage,
		SystemPrompt: systemPrompt,
		Messages:     ctxMessages,
		ToolCount:    len(toolDefs),
		Tools:        toolNames,
		MessageCount: len(messages),
	}

	cl.LogContext(entry)
}

// LogChatContextWithHistory logs context with history, truncating if too large
func LogChatContextWithHistory(
	sessionID string,
	workspaceID string,
	workerName string,
	model string,
	provider string,
	mode string,
	thinking string,
	userMessage string,
	fullPrompt string,
	history []ChatMessage,
	maxHistoryEntries int,
) {
	cl := GetContextLogger()
	if cl == nil || !cl.IsEnabled() {
		return
	}

	// Truncate history if too large
	historyToLog := history
	if maxHistoryEntries > 0 && len(history) > maxHistoryEntries {
		historyToLog = make([]ChatMessage, maxHistoryEntries)
		copy(historyToLog, history[len(history)-maxHistoryEntries:])
	}

	entry := ContextLogEntry{
		Timestamp:    time.Now(),
		SessionID:    sessionID,
		WorkspaceID:  workspaceID,
		WorkerName:   workerName,
		Model:        model,
		Provider:     provider,
		Mode:         mode,
		Thinking:     thinking,
		UserMessage:  userMessage,
		FullPrompt:   fullPrompt,
		HistoryCount: len(history),
		History:      historyToLog,
		MessageCount: len(history),
	}

	cl.LogContext(entry)
}

// LogChatError logs an error that occurred during chat processing
func LogChatError(sessionID string, workerName string, userMessage string, err error) {
	cl := GetContextLogger()
	if cl == nil || !cl.IsEnabled() {
		return
	}

	entry := ContextLogEntry{
		Timestamp:   time.Now(),
		SessionID:   sessionID,
		WorkerName:  workerName,
		UserMessage: userMessage,
		Error:       err.Error(),
	}

	cl.LogContext(entry)
}
