// Ada-Love - Ultra-lightweight personal AI agent

package agent

import (
	"context"
	"fmt"

	"ada-love-ai/pkg/audio/asr"
	"ada-love-ai/pkg/channels"
	"ada-love-ai/pkg/config"
	"ada-love-ai/pkg/media"
	"ada-love-ai/pkg/tools"
)

func (al *AgentLoop) RegisterTool(tool tools.Tool) {
	registry := al.GetRegistry()
	for _, agentID := range registry.ListAgentIDs() {
		if agent, ok := registry.GetAgent(agentID); ok {
			agent.Tools.Register(tool)
		}
	}
}

func (al *AgentLoop) SetChannelManager(cm *channels.Manager) {
	al.channelManager = cm
}

func (al *AgentLoop) GetRegistry() *AgentRegistry {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return al.registry
}

// RunAgentByID runs a single user turn against the specified agent ID.
// If sessionKey is empty, it will use the agent's main session key.
// This is a thin wrapper that prepares processOptions and delegates to runAgentLoop.
func (al *AgentLoop) RunAgentByID(ctx context.Context, targetAgentID string, content string, sessionKey string) (string, error) {
	if err := al.ensureHooksInitialized(ctx); err != nil {
		return "", err
	}
	if err := al.ensureMCPInitialized(ctx); err != nil {
		return "", err
	}

	registry := al.GetRegistry()
	if registry == nil {
		return "", fmt.Errorf("agent registry not initialized")
	}

	agent, ok := registry.GetAgent(targetAgentID)
	if !ok || agent == nil {
		return "", fmt.Errorf("agent %s not found", targetAgentID)
	}

	// If no explicit session key provided, let runAgentLoop choose a main session key
	dispatch := DispatchRequest{
		SessionKey:  sessionKey,
		UserMessage: content,
	}
	opts := processOptions{
		Dispatch:        dispatch,
		DefaultResponse: defaultResponse,
		EnableSummary:   true,
		SendResponse:    false,
	}
	return al.runAgentLoop(ctx, agent, opts)
}

func (al *AgentLoop) GetConfig() *config.Config {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return al.cfg
}

func (al *AgentLoop) SetMediaStore(s media.MediaStore) {
	al.mediaStore = s

	// Propagate store to all registered tools that can emit media.
	registry := al.GetRegistry()
	for _, agentID := range registry.ListAgentIDs() {
		if agent, ok := registry.GetAgent(agentID); ok {
			agent.Tools.SetMediaStore(s)
		}
	}
	registry.ForEachTool("send_tts", func(t tools.Tool) {
		if st, ok := t.(*tools.SendTTSTool); ok {
			st.SetMediaStore(s)
		}
	})
}

func (al *AgentLoop) SetTranscriber(t asr.Transcriber) {
	al.transcriber = t
}

func (al *AgentLoop) SetReloadFunc(fn func() error) {
	al.reloadFunc = fn
}

func (al *AgentLoop) SetHealthFunc(fn func() (string, error)) {
	al.healthFunc = fn
}

func (al *AgentLoop) SetTestConnFunc(fn func() (string, error)) {
	al.testConnFunc = fn
}

func (al *AgentLoop) RecordLastChannel(channel string) error {
	if al.state == nil {
		return nil
	}
	return al.state.SetLastChannel(channel)
}

func (al *AgentLoop) RecordLastChatID(chatID string) error {
	if al.state == nil {
		return nil
	}
	return al.state.SetLastChatID(chatID)
}

func (al *AgentLoop) GetStartupInfo() map[string]any {
	info := make(map[string]any)

	registry := al.GetRegistry()
	agent := registry.GetDefaultAgent()
	if agent == nil {
		return info
	}

	// Tools info
	toolsList := agent.Tools.List()
	info["tools"] = map[string]any{
		"count": len(toolsList),
		"names": toolsList,
	}

	// Skills info
	info["skills"] = agent.ContextBuilder.GetSkillsInfo()

	// Agents info
	info["agents"] = map[string]any{
		"count": len(registry.ListAgentIDs()),
		"ids":   registry.ListAgentIDs(),
	}

	return info
}
