export const categoryColors: Record<string, string> = {
  'File System': '#3b82f6',
  Git: '#f59e0b',
  'Code Search': '#10b981',
  Testing: '#8b5cf6',
  Build: '#ec4899',
  Web: '#22c55e',
  Media: '#eab308',
  MCP: '#a855f7',
  Communication: '#8b5cf6',
  Shell: '#ef4444',
  'Scheduled Tasks': '#14b8a6',
  Memory: '#f43f5e',
  Knowledge: '#06b6d4',
  Hardware: '#06b6d4',
  Skills: '#a855f7',
  Agent: '#f97316',
  Other: '#6b7280',
};

export const settingsSections = [
  { id: 'general', icon: 'Settings', label: 'General' },
  { id: 'workspace', icon: 'Folder', label: 'Workspace' },
  { id: 'workers', icon: 'Bot', label: 'Workers' },
  { id: 'agents', icon: 'User', label: 'Agents' },
  { id: 'skills', icon: 'Brain', label: 'Skills' },
  { id: 'tools', icon: 'Wrench', label: 'Tools' },
  { id: 'models', icon: 'Cpu', label: 'Models' },
];

export const knownApiUrls = [
  { name: 'OpenAI', url: 'https://api.openai.com/v1' },
  { name: 'Cloudflare', url: 'https://api.cloudflare.com/client/v4/accounts/{account_id}/ai/v1' },
  { name: 'OpenRouter', url: 'https://openrouter.ai/api/v1' },
  { name: 'Anthropic', url: 'https://api.anthropic.com/v1' },
  { name: 'Google Gemini', url: 'https://generativelanguage.googleapis.com/v1beta' },
  { name: 'Ollama', url: 'http://localhost:11434/v1' },
];
