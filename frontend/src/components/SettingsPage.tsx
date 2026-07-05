import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Switch } from './ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from './ui/dialog'
import { Card, CardContent } from './ui/card'
import { Separator } from './ui/separator'
import { BaseCard } from './BaseCard'
import { EditDialog } from './EditDialog'
import { ExpandableEditor } from './ExpandableEditor'
import { IconPicker } from './IconPicker'
import { Icon } from './Icon'
import * as api from '../api'
import { useTheme } from '../lib/theme'

const categoryColors: Record<string, string> = {
  'File System': '#3b82f6',
  'Git': '#f59e0b',
  'Code Search': '#10b981',
  'Testing': '#8b5cf6',
  'Build': '#ec4899',
  'Web': '#22c55e',
  'Communication': '#8b5cf6',
  'Other': '#6b7280',
}

const settingsSections = [
  { id: 'general', icon: 'Settings', label: 'General' },
  { id: 'workspace', icon: 'Folder', label: 'Workspace' },
  { id: 'agents', icon: 'User', label: 'Agents' },
  { id: 'skills', icon: 'Brain', label: 'Skills' },
  { id: 'tools', icon: 'Wrench', label: 'Tools' },
  { id: 'models', icon: 'Cpu', label: 'Models' },
]

function GeneralSection() {
  const { themes, currentTheme, setTheme, iconThemes, currentIconTheme, setIconTheme, iconSets, currentIconSet, setIconSet } = useTheme()

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-foreground">General</h3>
        <p className="text-sm text-muted-foreground">Theme and application settings.</p>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Theme</h4>
        <div className="flex flex-wrap gap-3">
          {themes.map(t => (
            <button
              key={t.id}
              className={`flex flex-col gap-2 p-3 rounded-lg border text-sm transition-colors w-48 ${
                currentTheme === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setTheme(t.id)}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium">{t.name}</span>
                {t.author && <span className="text-xs text-muted-foreground">{t.author}</span>}
              </div>
              <div className="flex gap-1.5">
                <span className="w-5 h-5 rounded-sm" style={{backgroundColor: t.light?.primary}} />
                <span className="w-5 h-5 rounded-sm" style={{backgroundColor: t.light?.secondary}} />
                <span className="w-5 h-5 rounded-sm" style={{backgroundColor: t.light?.accent}} />
                <span className="w-5 h-5 rounded-sm" style={{backgroundColor: t.light?.muted}} />
                <span className="w-5 h-5 rounded-sm" style={{backgroundColor: t.light?.destructive}} />
              </div>
              <div className="flex gap-1.5 mt-1">
                <span className="px-2 py-0.5 text-[10px] rounded-sm text-white" style={{backgroundColor: t.light?.primary}}>Button</span>
                <span className="px-2 py-0.5 text-[10px] rounded-sm" style={{backgroundColor: t.light?.secondary, color: t.light?.['secondary-foreground']}}>Button</span>
                <span className="px-2 py-0.5 text-[10px] rounded-sm border" style={{borderColor: t.light?.border, color: t.light?.['secondary-foreground']}}>Outline</span>
              </div>
            </button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground">
          Install: <code className="bg-muted px-1 rounded">npm run theme:add &lt;tweakcn-url&gt; "Theme Name"</code>
        </p>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Icon Style</h4>
        <div className="flex flex-wrap gap-2">
          {iconThemes.map(t => (
            <button
              key={t.id}
              className={`px-3 py-1.5 rounded-lg border text-xs transition-colors ${
                currentIconTheme === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setIconTheme(t.id)}
            >
              {t.name}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Icon Set</h4>
        <div className="flex flex-wrap gap-2">
          {iconSets.map(t => (
            <button
              key={t.id}
              className={`px-3 py-1.5 rounded-lg border text-xs transition-colors ${
                currentIconSet === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setIconSet(t.id)}
            >
              {t.name}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
function WorkspacesSection() {
  const [workspaces, setWorkspaces] = useState<api.backend.WorkspaceConfig[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [showEdit, setShowEdit] = useState(false)
  const [editing, setEditing] = useState<api.backend.WorkspaceConfig | null>(null)
  const [knownTools, setKnownTools] = useState<api.backend.ToolUIInfo[]>([])
  const [knownSkills, setKnownSkills] = useState<string[]>([])
  const [E, setE] = useState({
    title: '', description: '', color: '#3b82f6', icon: '📂',
    personality: '', folders: [] as string[], knowledge: [] as string[],
    workspace_agents: [] as string[], skills: [] as string[], tools: [] as string[],
    enabled: true, maxPromptSend: 0, commitChanges: true, maxContextLength: 0,
  })
  const [A, setA] = useState({ title: '', personality: '' })
  const [selectedField, setSelectedField] = useState('folders')
  const fieldMap: Record<string, { state: string[], set: (v: string[]) => void }> = {
    folders: { state: E.folders, set: (v) => setE(prev => ({ ...prev, folders: v })) },
    knowledge: { state: E.knowledge, set: (v) => setE(prev => ({ ...prev, knowledge: v })) },
    workspace_agents: { state: E.workspace_agents, set: (v) => setE(prev => ({ ...prev, workspace_agents: v })) },
    skills: { state: E.skills, set: (v) => setE(prev => ({ ...prev, skills: v })) },
    tools: { state: E.tools, set: (v) => setE(prev => ({ ...prev, tools: v })) },
  }

  const computeTokens = (ws: api.backend.WorkspaceConfig) => {
    let t = 2000
    for (const f of ws.folders || []) t += f.length * 10
    for (const a of ws.workspace_agents || []) t += 500
    for (const s of ws.skills || []) t += 300
    return t.toLocaleString()
  }

  const load = () => {
    api.getWorkspaces().then(setWorkspaces)
    api.getAvailableTools().then(setKnownTools)
  }
  useEffect(() => { load(); api.getAgentCategories().then(setKnownSkills) }, [])

  const handleAdd = async () => {
    if (!A.title.trim()) return
    await api.addWorkspace(A.title.trim(), '', A.personality.trim())
    setShowAdd(false)
    setA({ title: '', personality: '' })
    load()
  }

  const handleDelete = (title: string) => api.deleteWorkspace(title).then(load)
  const handleToggle = (title: string) => api.toggleWorkspace(title).then(load)

  const toggleArrayItem = (arr: string[], item: string) =>
    arr.includes(item) ? arr.filter(x => x !== item) : [...arr, item]

  const openEdit = (ws: api.backend.WorkspaceConfig) => {
    setEditing(ws)
    setE({
      title: ws.title, description: ws.description,
      color: ws.color || '#3b82f6', icon: ws.icon || '📂',
      personality: ws.personality,
      folders: ws.folders || [],
      knowledge: ws.knowledge || [],
      workspace_agents: ws.workspace_agents || [],
      skills: ws.skills || [],
      tools: ws.tools || [],
      enabled: ws.enabled,
      maxPromptSend: ws.max_prompt_send || 0,
      commitChanges: ws.commit_changes !== false,
      maxContextLength: ws.max_context_length || 0,
    })
    setShowEdit(true)
  }

  const handleSaveEdit = async () => {
    if (!editing) return
    await api.updateWorkspace(editing.title, {
      ...editing,
      title: E.title, description: E.description,
      color: E.color, icon: E.icon,
      personality: E.personality,
      folders: E.folders, knowledge: E.knowledge,
      workspace_agents: E.workspace_agents,
      skills: E.skills, tools: E.tools,
      enabled: E.enabled,
      max_prompt_send: E.maxPromptSend,
      commit_changes: E.commitChanges,
      max_context_length: E.maxContextLength,
    })
    setShowEdit(false)
    setEditing(null)
    load()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">Workspaces</h3>
          <p className="text-sm text-muted-foreground">Manage your workspaces.</p>
        </div>
        <Button size="sm" onClick={() => setShowAdd(true)} className="gap-1">
          <Icon name="Plus" size={16} /> New Workspace
        </Button>
      </div>

      <div className="agents-grid">
        {workspaces.map(ws => (
          <BaseCard
            key={ws.title}
            color={ws.enabled ? (ws.color || '#3b82f6') : '#6b7280'}
            inactive={!ws.enabled}
            headerLeft={ws.enabled ? <span className="text-xs text-green-400">✔ Active</span> : null}
            headerRight={
              <>
                <button className="base-card-btn" onClick={() => openEdit(ws)} title="Edit"><Icon name="Edit" className="w-3 h-3" /></button>
                <button className="base-card-btn" onClick={() => handleToggle(ws.title)} title="Toggle"><Icon name={ws.enabled ? 'Check' : 'X'} className="w-3 h-3" /></button>
                <button className="base-card-btn" onClick={() => handleDelete(ws.title)} title="Delete"><Icon name="Trash2" className="w-3 h-3" /></button>
              </>
            }
            icon={ws.icon || '📂'}
            title={ws.title}
          >
            <div className="base-card-desc">{ws.description || 'No description'}</div>
            <div className="text-xs text-muted-foreground mt-1">
              {ws.folders?.length || 0} folders · {ws.workspace_agents?.length || 0} agents · {ws.skills?.length || 0} skills
            </div>
          </BaseCard>
        ))}
        {workspaces.length === 0 && (
          <div className="col-span-full text-center py-12 text-muted-foreground text-sm">No workspaces yet.</div>
        )}
      </div>

      <Dialog open={showAdd} onOpenChange={setShowAdd}>
        <DialogContent>
          <DialogHeader><DialogTitle>New Workspace</DialogTitle></DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Title</label>
              <Input value={A.title} onChange={e => setA({ ...A, title: e.target.value })} />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Personality</label>
              <Input value={A.personality} onChange={e => setA({ ...A, personality: e.target.value })} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAdd(false)}>Cancel</Button>
            <Button onClick={handleAdd}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <EditDialog
        open={showEdit}
        onOpenChange={setShowEdit}
        title="Edit Workspace"
        description={editing ? `${computeTokens(editing)} tokens` : undefined}
        onSave={handleSaveEdit}
      >
        <div className="grid grid-cols-2 gap-x-4 gap-y-3">

          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Title</label>
            <Input value={E.title} onChange={e => setE({ ...E, title: e.target.value })} />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Description</label>
            <Input value={E.description} onChange={e => setE({ ...E, description: e.target.value })} />
          </div>

          <div className="col-span-2 flex items-center gap-3 flex-wrap">
            <div className="flex items-center gap-1">
              <div className="w-7 h-7 rounded border cursor-pointer" style={{ backgroundColor: E.color }}
                onClick={() => document.getElementById('pick-color')?.click()} />
              <input id="pick-color" type="color" value={E.color}
                onChange={e => setE({ ...E, color: e.target.value })} className="w-0 h-0 opacity-0" />
            </div>
            <IconPicker value={E.icon} onChange={v => setE({ ...E, icon: v })} />
            <Separator orientation="vertical" className="h-6" />
            <div className="flex items-center gap-1.5">
              <label className="text-xs text-muted-foreground whitespace-nowrap">Max Prompt</label>
              <Input type="number" value={E.maxPromptSend}
                onChange={e => setE({ ...E, maxPromptSend: parseInt(e.target.value) || 0 })} className="h-7 w-16" />
            </div>
            <div className="flex items-center gap-1.5">
              <label className="text-xs text-muted-foreground whitespace-nowrap">Max Context</label>
              <Input type="number" value={E.maxContextLength}
                onChange={e => setE({ ...E, maxContextLength: parseInt(e.target.value) || 0 })} className="h-7 w-16" />
            </div>
            <div className="flex items-center gap-1.5">
              <Switch checked={E.commitChanges} onCheckedChange={c => setE({ ...E, commitChanges: c })} />
              <label className="text-xs text-muted-foreground cursor-pointer">Commit</label>
            </div>
          </div>

          <div className="col-span-2">
            <ExpandableEditor
              label="Personality"
              value={E.personality}
              onChange={v => setE({ ...E, personality: v })}
            />
          </div>

          <div className="col-span-2 flex gap-4 min-h-[150px]">
            <div className="w-20 shrink-0 flex flex-col gap-0.5 border-r border-border pr-2">
              {['folders', 'knowledge', 'workspace_agents', 'skills', 'tools'].map(key => (
                <button
                  key={key}
                  className={`text-left px-2 py-1 rounded text-xs transition-colors ${
                    selectedField === key ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                  }`}
                  onClick={() => setSelectedField(key)}
                >
                  {{'folders':'Folders','knowledge':'Knowledge','workspace_agents':'Agents','skills':'Skills','tools':'Tools'}[key]}
                </button>
              ))}
            </div>
            <div className="flex-1 min-h-[28px] max-h-[120px] overflow-y-auto p-1.5 rounded border border-border">
              <div className="flex gap-1 overflow-x-auto">
                {fieldMap[selectedField]?.state.map((item: string) => (
                  <span key={item} className="flex items-center gap-1 px-2 py-0.5 text-xs rounded bg-muted text-muted-foreground whitespace-nowrap">
                    {item}
                    <button onClick={() => fieldMap[selectedField].set(fieldMap[selectedField].state.filter((x: string) => x !== item))}
                      className="text-destructive hover:text-destructive/80">✕</button>
                  </span>
                ))}
                <input
                  placeholder="Add..."
                  className="h-6 border-0 p-0 text-xs bg-transparent outline-none min-w-[80px]"
                  onKeyDown={e => {
                    if (e.key === 'Enter' && (e.target as HTMLInputElement).value.trim()) {
                      const f = fieldMap[selectedField]
                      f.set([...f.state, (e.target as HTMLInputElement).value.trim()])
                      ;(e.target as HTMLInputElement).value = ''
                    }
                  }} />
              </div>
            </div>
          </div>
        </div>
      </EditDialog>
    </div>
  )
}

function AgentsSection() {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-foreground">Agents</h3>
      <p className="text-sm text-muted-foreground">Configure your agents.</p>
    </div>
  )
}

function SkillsSection() {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-foreground">Skills</h3>
      <p className="text-sm text-muted-foreground">Manage skills and capabilities.</p>
    </div>
  )
}

function ToolsSection() {
  const [tools, setTools] = useState<api.backend.ToolUIInfo[]>([])
  const [profiles, setProfiles] = useState<api.backend.ToolProfile[]>([])
  const [selectedProfileID, setSelectedProfileID] = useState<number | null>(null)
  const [showNewProfileDialog, setShowNewProfileDialog] = useState(false)
  const [newProfileName, setNewProfileName] = useState('')
  const [newProfileColor, setNewProfileColor] = useState('#6b7280')
  const [newProfileIcon, setNewProfileIcon] = useState('🔧')

  useEffect(() => {
    api.getAvailableTools().then(setTools)
    api.getToolProfiles().then(profiles => {
      setProfiles(profiles)
      const defaultProfile = profiles.find(p => p.name === 'Default')
      if (defaultProfile) setSelectedProfileID(defaultProfile.id)
    })
  }, [])

  const handleToggle = async (toolName: string, enabled: boolean) => {
    if (selectedProfileID) {
      await api.toggleProfileTool(selectedProfileID, toolName, enabled)
      setProfiles(prev => prev.map(p => {
        if (p.id === selectedProfileID) {
          return {
            ...p,
            tools: enabled 
              ? [...p.tools, toolName] 
              : p.tools.filter(t => t !== toolName)
          }
        }
        return p
      }))
    } else {
      await api.toggleTool(toolName, enabled)
      setTools(prev => prev.map(t => t.name === toolName ? { ...t, enabled } : t))
    }
  }

  const handleCreateProfile = async () => {
    const profile = await api.createToolProfile(newProfileName, newProfileColor, newProfileIcon)
    if (profile) {
      setProfiles(prev => [...prev, profile])
      setSelectedProfileID(profile.id)
      setShowNewProfileDialog(false)
      setNewProfileName('')
      setNewProfileColor('#6b7280')
      setNewProfileIcon('🔧')
    }
  }

  const handleDeleteProfile = async (id: number) => {
    const success = await api.deleteToolProfile(id)
    if (success) {
      setProfiles(prev => prev.filter(p => p.id !== id))
      if (selectedProfileID === id) {
        const defaultProfile = profiles.find(p => p.name === 'Default')
        setSelectedProfileID(defaultProfile?.id || null)
      }
    }
  }

  const groupedTools = tools.reduce((acc, tool) => {
    const cat = tool.category || 'Other'
    if (!acc[cat]) acc[cat] = []
    acc[cat].push(tool)
    return acc
  }, {} as Record<string, api.backend.ToolUIInfo[]>)

  const selectedProfile = profiles.find(p => p.id === selectedProfileID)
  const isProfileActive = selectedProfile && selectedProfile.name !== 'Default'

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-foreground">Tools</h3>
        <p className="text-sm text-muted-foreground">Configure agent tools for the active workspace.</p>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Profiles</h4>
        <div className="profiles-grid">
          {profiles.map(profile => (
            <BaseCard
              key={profile.id}
              color={selectedProfileID === profile.id ? profile.color : '#4b5563'}
              headerLeft={<span className="text-xs text-white opacity-90">Profile</span>}
              headerRight={
                profile.name !== 'Default' ? (
                  <button
                    className="base-card-btn"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDeleteProfile(profile.id)
                    }}
                  >
                    <Icon name="Trash2" className="w-3 h-3" />
                  </button>
                ) : null
              }
              icon={profile.icon}
              title={profile.name}
              selected={selectedProfileID === profile.id}
              onClick={() => setSelectedProfileID(profile.id)}
              className="cursor-pointer"
              small
            >
              <div className="base-card-desc">{profile.tools.length} tools</div>
            </BaseCard>
          ))}
          <BaseCard
            color="#374151"
            headerLeft={<span className="text-xs text-white opacity-90">Action</span>}
            icon="+"
            title="Add Profile"
            onClick={() => setShowNewProfileDialog(true)}
            className="cursor-pointer"
            small
          >
            <div className="base-card-desc">Create new profile</div>
          </BaseCard>
        </div>
      </div>

      {Object.entries(groupedTools).map(([category, categoryTools]) => (
        <div key={category}>
          <h4 className="text-sm font-medium text-foreground mb-2">{category}</h4>
          <div className="tools-grid">
            {categoryTools.map(tool => {
              const isEnabled = isProfileActive 
                ? selectedProfile?.tools.includes(tool.name) ?? false
                : tool.enabled
              return (
                <BaseCard
                  key={tool.name}
                  color={isEnabled ? categoryColors[category] : '#6b7280'}
                  headerLeft={<span className="text-xs text-white opacity-90">{category}</span>}
                  headerRight={
                    <Switch 
                      checked={isEnabled}
                      onCheckedChange={(checked) => handleToggle(tool.name, checked)}
                      className="data-[state=checked]:bg-white/30"
                    />
                  }
                  icon="🔧"
                  title={tool.name}
                >
                  <div className="base-card-desc">{tool.description}</div>
                </BaseCard>
              )
            })}
          </div>
        </div>
      ))}
      
      {tools.length === 0 && (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">No tools available</p>
          </CardContent>
        </Card>
      )}

      <Dialog open={showNewProfileDialog} onOpenChange={setShowNewProfileDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Profile</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <span className="text-right text-sm">Name</span>
              <Input 
                value={newProfileName}
                onChange={(e) => setNewProfileName(e.target.value)}
                className="col-span-3"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <span className="text-right text-sm">Color</span>
              <div className="col-span-3 flex items-center gap-2">
                <div 
                  className="w-8 h-8 rounded border border-border cursor-pointer"
                  style={{ backgroundColor: newProfileColor }}
                />
                <Input 
                  type="color"
                  value={newProfileColor}
                  onChange={(e) => setNewProfileColor(e.target.value)}
                  className="w-16 h-8"
                />
              </div>
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <span className="text-right text-sm">Icon</span>
              <IconPicker value={newProfileIcon} onChange={setNewProfileIcon} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowNewProfileDialog(false)}>Cancel</Button>
            <Button onClick={handleCreateProfile}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function ModelsSection() {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-foreground">Models</h3>
      <p className="text-sm text-muted-foreground">Configure providers, API keys, and models.</p>
    </div>
  )
}

export default function SettingsPage() {
  const [activeSection, setActiveSection] = useState('general')

  return (
    <div className="fixed inset-0 z-50 flex bg-background">
      <div className="settings-sidebar">
        <div className="settings-sidebar-header">
          <span className="text-sm font-semibold text-foreground">Settings</span>
        </div>
        <div className="settings-sidebar-nav">
          {settingsSections.map((section) => (
            <button
              key={section.id}
              className={`settings-nav-item ${activeSection === section.id ? 'active' : ''}`}
              onClick={() => setActiveSection(section.id)}
            >
              <span className="settings-nav-icon"><Icon name={section.icon} className="w-4 h-4" /></span>
              <span>{section.label}</span>
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto">
        <div className="settings-content">
          {activeSection === 'general' && <GeneralSection />}
          {activeSection === 'workspace' && <WorkspacesSection />}
          {activeSection === 'agents' && <AgentsSection />}
          {activeSection === 'skills' && <SkillsSection />}
          {activeSection === 'tools' && <ToolsSection />}
          {activeSection === 'models' && <ModelsSection />}
        </div>
      </div>
    </div>
  )
}
