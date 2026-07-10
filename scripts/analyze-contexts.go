// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// ContextLogEntry represents a log entry for what a chat sent as context
type ContextLogEntry struct {
	Timestamp     time.Time `json:"timestamp"`
	SessionID     string    `json:"session_id"`
	WorkspaceID   string    `json:"workspace_id"`
	WorkerName    string    `json:"worker_name"`
	Model         string    `json:"model"`
	Provider      string    `json:"provider"`
	Mode          string    `json:"mode"`
	Thinking      string    `json:"thinking"`
	UserMessage   string    `json:"user_message"`
	FullPrompt    string    `json:"full_prompt"`
	HistoryCount  int       `json:"history_count"`
	MessageCount  int       `json:"message_count"`
	Error         string    `json:"error,omitempty"`
}

// SessionStats contains statistics for a single session
type SessionStats struct {
	SessionID       string
	WorkerName      string
	WorkspaceID     string
	Model           string
	Provider        string
	Mode            string
	TotalMessages   int
	TotalTokens     int
	FirstMessage    time.Time
	LastMessage     time.Time
	UniqueModels    int
	UniqueProviders int
	UserMessages    []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/analyze-contexts.go <log-file>")
		fmt.Println("       go run scripts/analyze-contexts.go /path/to/context_logs.jsonl")
		os.Exit(1)
	}

	logPath := os.Args[1]
	entries, err := readLogEntries(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading log: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Context Analysis Report ===")
	fmt.Printf("Total log entries: %d\n\n", len(entries))

	// Group by session
	sessions := groupBySession(entries)
	
	// Print session summaries
	fmt.Println("--- Sessions Summary ---")
	for _, sess := range sessions {
		fmt.Printf("\nSession: %s\n", sess.SessionID)
		fmt.Printf("  Worker: %s\n", sess.WorkerName)
		fmt.Printf("  Workspace: %s\n", sess.WorkspaceID)
		fmt.Printf("  Model: %s\n", sess.Model)
		fmt.Printf("  Provider: %s\n", sess.Provider)
		fmt.Printf("  Mode: %s\n", sess.Mode)
		fmt.Printf("  Total Messages: %d\n", sess.TotalMessages)
		fmt.Printf("  First: %s\n", sess.FirstMessage.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Last: %s\n", sess.LastMessage.Format("2006-01-02 15:04:05"))
		if len(sess.UserMessages) > 0 {
			fmt.Printf("  User Messages:\n")
			for i, msg := range sess.UserMessages {
				if i >= 5 {
					fmt.Printf("    ... and %d more\n", len(sess.UserMessages)-5)
					break
				}
				fmt.Printf("    - %s\n", truncate(msg, 100))
			}
		}
	}

	// Print model/provider usage
	fmt.Println("\n--- Model/Provider Usage ---")
	modelCounts := make(map[string]int)
	providerCounts := make(map[string]int)
	for _, entry := range entries {
		if entry.Model != "" {
			modelCounts[entry.Model]++
		}
		if entry.Provider != "" {
			providerCounts[entry.Provider]++
		}
	}

	fmt.Println("\nModels used:")
	for model, count := range modelCounts {
		fmt.Printf("  %s: %d times\n", model, count)
	}

	fmt.Println("\nProviders used:")
	for provider, count := range providerCounts {
		fmt.Printf("  %s: %d times\n", provider, count)
	}

	// Print mode distribution
	fmt.Println("\n--- Mode Distribution ---")
	modeCounts := make(map[string]int)
	for _, entry := range entries {
		modeCounts[entry.Mode]++
	}
	for mode, count := range modeCounts {
		fmt.Printf("  %s: %d times\n", mode, count)
	}
}

func readLogEntries(path string) ([]ContextLogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var entries []ContextLogEntry
	
	for {
		var entry ContextLogEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		entries = append(entries, entry)
	}
	
	return entries, nil
}

func groupBySession(entries []ContextLogEntry) []*SessionStats {
	sessionMap := make(map[string]*SessionStats)
	
	for _, entry := range entries {
		sess, ok := sessionMap[entry.SessionID]
		if !ok {
			sess = &SessionStats{
				SessionID:       entry.SessionID,
				WorkerName:      entry.WorkerName,
				WorkspaceID:     entry.WorkspaceID,
				Model:           entry.Model,
				Provider:        entry.Provider,
				Mode:            entry.Mode,
				UserMessages:    make([]string, 0),
			}
			sessionMap[entry.SessionID] = sess
		}
		
		sess.TotalMessages += entry.MessageCount
		sess.FirstMessage = minTime(sess.FirstMessage, entry.Timestamp)
		sess.LastMessage = maxTime(sess.LastMessage, entry.Timestamp)
		if entry.Model != "" {
			sess.UniqueModels++
		}
		if entry.Provider != "" {
			sess.UniqueProviders++
		}
		sess.UserMessages = append(sess.UserMessages, entry.UserMessage)
	}
	
	result := make([]*SessionStats, 0, len(sessionMap))
	for _, sess := range sessionMap {
		result = append(result, sess)
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastMessage.After(result[j].LastMessage)
	})
	
	return result
}

func minTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.After(b) {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}