package gittools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	toolshared "ada-love-ai/pkg/tools/shared"
)

// gitRunner executes git commands in a given workspace directory.
type gitRunner struct {
	workspace string
}

func (r *gitRunner) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.workspace
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// --- git_init ---

type GitInitTool struct {
	workspace string
}

func NewGitInitTool(workspace string) *GitInitTool {
	return &GitInitTool{workspace: workspace}
}

func (t *GitInitTool) Name() string        { return "git_init" }
func (t *GitInitTool) Description() string {
	return "Initialize a new Git repository in the workspace directory."
}
func (t *GitInitTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *GitInitTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	r := &gitRunner{workspace: t.workspace}
	out, err := r.run("init")
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	return toolshared.NewToolResult(out)
}

// --- git_status ---

type GitStatusTool struct {
	workspace string
}

func NewGitStatusTool(workspace string) *GitStatusTool {
	return &GitStatusTool{workspace: workspace}
}

func (t *GitStatusTool) Name() string        { return "git_status" }
func (t *GitStatusTool) Description() string {
	return "Show the working tree status: modified, added, deleted, and untracked files."
}
func (t *GitStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *GitStatusTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	r := &gitRunner{workspace: t.workspace}
	out, err := r.run("status", "--porcelain")
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	if out == "" {
		return toolshared.NewToolResult("Working tree is clean. No modifications.")
	}

	// Parse porcelain format for a cleaner output
	var modified, added, deleted, untracked, renamed []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		status := line[:2]
		name := strings.TrimSpace(line[3:])
		switch {
		case status == "??":
			untracked = append(untracked, name)
		case strings.Contains(status, "D"):
			deleted = append(deleted, name)
		case strings.Contains(status, "A"):
			added = append(added, name)
		case strings.Contains(status, "R"):
			renamed = append(renamed, name)
		default:
			modified = append(modified, name)
		}
	}

	var result strings.Builder
	result.WriteString("Git Status:\n")
	if len(modified) > 0 {
		result.WriteString(fmt.Sprintf("\nModified (%d):\n", len(modified)))
		for _, f := range modified {
			result.WriteString("  M " + f + "\n")
		}
	}
	if len(added) > 0 {
		result.WriteString(fmt.Sprintf("\nAdded (%d):\n", len(added)))
		for _, f := range added {
			result.WriteString("  A " + f + "\n")
		}
	}
	if len(deleted) > 0 {
		result.WriteString(fmt.Sprintf("\nDeleted (%d):\n", len(deleted)))
		for _, f := range deleted {
			result.WriteString("  D " + f + "\n")
		}
	}
	if len(renamed) > 0 {
		result.WriteString(fmt.Sprintf("\nRenamed (%d):\n", len(renamed)))
		for _, f := range renamed {
			result.WriteString("  R " + f + "\n")
		}
	}
	if len(untracked) > 0 {
		result.WriteString(fmt.Sprintf("\nUntracked (%d):\n", len(untracked)))
		for _, f := range untracked {
			result.WriteString("  ? " + f + "\n")
		}
	}

	return toolshared.NewToolResult(result.String())
}

// --- git_diff ---

type GitDiffTool struct {
	workspace string
}

func NewGitDiffTool(workspace string) *GitDiffTool {
	return &GitDiffTool{workspace: workspace}
}

func (t *GitDiffTool) Name() string        { return "git_diff" }
func (t *GitDiffTool) Description() string {
	return "Show the diff of changes. By default shows unstaged changes. Set staged=true to see staged changes."
}
func (t *GitDiffTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"staged": map[string]any{
				"type":        "boolean",
				"description": "Show staged changes instead of unstaged. Default: false.",
				"default":     false,
			},
			"file": map[string]any{
				"type":        "string",
				"description": "Show diff for a specific file only.",
			},
		},
	}
}

func (t *GitDiffTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	r := &gitRunner{workspace: t.workspace}
	staged, _ := args["staged"].(bool)

	gitArgs := []string{"diff"}
	if staged {
		gitArgs = append(gitArgs, "--cached")
	}
	if file, ok := args["file"].(string); ok && file != "" {
		gitArgs = append(gitArgs, "--", file)
	}

	out, err := r.run(gitArgs...)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	if out == "" {
		if staged {
			return toolshared.NewToolResult("No staged changes.")
		}
		return toolshared.NewToolResult("No unstaged changes.")
	}

	// Truncate very large diffs
	const maxDiffLen = 30000
	if len(out) > maxDiffLen {
		out = out[:maxDiffLen] + "\n\n[TRUNCATED - diff exceeds 30KB. Use 'file' parameter to scope to a specific file.]"
	}

	return toolshared.NewToolResult(out)
}

// --- git_create_branch ---

type GitCreateBranchTool struct {
	workspace string
}

func NewGitCreateBranchTool(workspace string) *GitCreateBranchTool {
	return &GitCreateBranchTool{workspace: workspace}
}

func (t *GitCreateBranchTool) Name() string        { return "git_create_branch" }
func (t *GitCreateBranchTool) Description() string {
	return "Create a new Git branch and optionally switch to it."
}
func (t *GitCreateBranchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the new branch.",
			},
			"switch": map[string]any{
				"type":        "boolean",
				"description": "Switch to the new branch after creating it. Default: true.",
				"default":     true,
			},
		},
		"required": []string{"name"},
	}
}

func (t *GitCreateBranchTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return toolshared.ErrorResult("name is required")
	}

	switchTo, _ := args["switch"].(bool)
	if _, exists := args["switch"]; !exists {
		switchTo = true
	}

	r := &gitRunner{workspace: t.workspace}

	if switchTo {
		_, err := r.run("checkout", "-b", name)
		if err != nil {
			return toolshared.ErrorResult(err.Error())
		}
		return toolshared.NewToolResult(fmt.Sprintf("Created and switched to branch: %s", name))
	}

	_, err := r.run("branch", name)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	return toolshared.NewToolResult(fmt.Sprintf("Created branch: %s", name))
}

// --- git_switch_branch ---

type GitSwitchBranchTool struct {
	workspace string
}

func NewGitSwitchBranchTool(workspace string) *GitSwitchBranchTool {
	return &GitSwitchBranchTool{workspace: workspace}
}

func (t *GitSwitchBranchTool) Name() string        { return "git_switch_branch" }
func (t *GitSwitchBranchTool) Description() string {
	return "Switch to an existing Git branch."
}
func (t *GitSwitchBranchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the branch to switch to.",
			},
		},
		"required": []string{"name"},
	}
}

func (t *GitSwitchBranchTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return toolshared.ErrorResult("name is required")
	}

	r := &gitRunner{workspace: t.workspace}
	_, err := r.run("checkout", name)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	return toolshared.NewToolResult(fmt.Sprintf("Switched to branch: %s", name))
}

// --- git_commit ---

type GitCommitTool struct {
	workspace string
}

func NewGitCommitTool(workspace string) *GitCommitTool {
	return &GitCommitTool{workspace: workspace}
}

func (t *GitCommitTool) Name() string        { return "git_commit" }
func (t *GitCommitTool) Description() string {
	return "Stage all changes and create a commit with the given message."
}
func (t *GitCommitTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"message": map[string]any{
				"type":        "string",
				"description": "The commit message.",
			},
			"add_all": map[string]any{
				"type":        "boolean",
				"description": "Stage all changes before committing (git add -A). Default: true.",
				"default":     true,
			},
		},
		"required": []string{"message"},
	}
}

func (t *GitCommitTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return toolshared.ErrorResult("message is required")
	}

	addAll := true
	if aa, exists := args["add_all"]; exists {
		addAll, _ = aa.(bool)
	}

	r := &gitRunner{workspace: t.workspace}

	if addAll {
		_, err := r.run("add", "-A")
		if err != nil {
			return toolshared.ErrorResult(fmt.Sprintf("git add failed: %v", err))
		}
	}

	out, err := r.run("commit", "-m", message)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}

	return toolshared.NewToolResult(out)
}

// --- git_log ---

type GitLogTool struct {
	workspace string
}

func NewGitLogTool(workspace string) *GitLogTool {
	return &GitLogTool{workspace: workspace}
}

func (t *GitLogTool) Name() string        { return "git_log" }
func (t *GitLogTool) Description() string {
	return "Show recent commit history with hash, author, date, and message."
}
func (t *GitLogTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"count": map[string]any{
				"type":        "integer",
				"description": "Number of commits to show. Default: 10.",
				"default":     10,
			},
			"file": map[string]any{
				"type":        "string",
				"description": "Show history for a specific file only.",
			},
		},
	}
}

func (t *GitLogTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	count := 10
	if c, ok := args["count"].(float64); ok && c > 0 {
		count = int(c)
		if count > 100 {
			count = 100
		}
	}

	r := &gitRunner{workspace: t.workspace}
	gitArgs := []string{"log", fmt.Sprintf("-%d", count), "--oneline", "--no-merges"}
	if file, ok := args["file"].(string); ok && file != "" {
		gitArgs = append(gitArgs, "--", file)
	}

	out, err := r.run(gitArgs...)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	if out == "" {
		return toolshared.NewToolResult("No commits found.")
	}

	return toolshared.NewToolResult(out)
}

// --- git_reset ---

type GitResetTool struct {
	workspace string
}

func NewGitResetTool(workspace string) *GitResetTool {
	return &GitResetTool{workspace: workspace}
}

func (t *GitResetTool) Name() string        { return "git_reset" }
func (t *GitResetTool) Description() string {
	return "Reset the repository state. Use 'mixed' (default) to unstage changes, 'soft' to keep changes staged, or 'hard' to discard all changes. Use 'commit_ref' to reset to a specific commit."
}
func (t *GitResetTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"mode": map[string]any{
				"type":        "string",
				"description": "Reset mode: 'soft', 'mixed' (default), or 'hard'.",
				"enum":        []string{"soft", "mixed", "hard"},
				"default":     "mixed",
			},
			"commit_ref": map[string]any{
				"type":        "string",
				"description": "Commit reference to reset to (e.g. 'HEAD~1', a hash). Default: HEAD.",
				"default":     "HEAD",
			},
		},
	}
}

func (t *GitResetTool) Execute(ctx context.Context, args map[string]any) *toolshared.ToolResult {
	mode := "mixed"
	if m, ok := args["mode"].(string); ok && m != "" {
		mode = m
	}

	commitRef := "HEAD"
	if cr, ok := args["commit_ref"].(string); ok && cr != "" {
		commitRef = cr
	}

	r := &gitRunner{workspace: t.workspace}
	_, err := r.run("reset", "--"+mode, commitRef)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}

	return toolshared.NewToolResult(fmt.Sprintf("Git reset --%s %s completed.", mode, commitRef))
}
