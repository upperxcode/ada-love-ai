import { useState, useEffect } from 'react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Switch } from '../ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import { Card, CardContent } from '../ui/card';
import { BaseCard } from '../BaseCard';
import { IconPicker } from '../IconPicker';
import { Icon } from '../Icon';
import * as api from '../../api';

const categoryColors: Record<string, string> = {
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

function ToolsSection() {
  const [tools, setTools] = useState<api.backend.ToolUIInfo[]>([]);
  const [profiles, setProfiles] = useState<api.backend.ToolProfile[]>([]);
  const [selectedProfileID, setSelectedProfileID] = useState<number | null>(
    null,
  );
  const [showNewProfileDialog, setShowNewProfileDialog] = useState(false);
  const [newProfileName, setNewProfileName] = useState('');
  const [newProfileColor, setNewProfileColor] = useState('#6b7280');
  const [newProfileIcon, setNewProfileIcon] = useState('🔧');
  const [filterType, setFilterType] = useState<'all' | 'active' | 'inactive'>(
    'all',
  );
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    api.getAvailableTools().then(setTools).catch(() => setTools([]));
    api.getToolProfiles().then((profiles) => {
      setProfiles(profiles ?? []);
      const defaultProfile = (profiles ?? []).find((p) => p.name === 'Default');
      if (defaultProfile) setSelectedProfileID(defaultProfile.id);
    }).catch(() => {});
  }, []);

  const handleToggle = async (toolName: string, enabled: boolean) => {
    const profile = profiles.find((p) => p.id === selectedProfileID);
    if (!profile) return;
    await api.toggleProfileTool(profile.id, toolName, enabled);
    setProfiles((prev) =>
      prev.map((p) => {
        if (p.id === profile.id) {
          return {
            ...p,
            tools: enabled
              ? [...p.tools, toolName]
              : p.tools.filter((t) => t !== toolName),
          };
        }
        return p;
      }),
    );
  };

  const handleCreateProfile = async () => {
    const profile = await api.createToolProfile(
      newProfileName,
      newProfileColor,
      newProfileIcon,
    );
    if (profile) {
      setProfiles((prev) => [...prev, profile]);
      setSelectedProfileID(profile.id);
      setShowNewProfileDialog(false);
      setNewProfileName('');
      setNewProfileColor('#6b7280');
      setNewProfileIcon('🔧');
    }
  };

  const handleDeleteProfile = async (id: number) => {
    const success = await api.deleteToolProfile(id);
    if (success) {
      setProfiles((prev) => prev.filter((p) => p.id !== id));
      if (selectedProfileID === id) {
        const defaultProfile = profiles.find((p) => p.name === 'Default');
        setSelectedProfileID(defaultProfile?.id || null);
      }
    }
  };

  const selectedProfile = profiles.find((p) => p.id === selectedProfileID);

  const selectedTools = tools.filter((tool) => {
    // Search filter
    if (searchQuery.trim() !== '') {
      const query = searchQuery.toLowerCase();
      if (!tool.name.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Status filter — always check against the selected profile's tools list
    const isEnabled = selectedProfile?.tools.includes(tool.name) ?? false;

    if (filterType === 'active' && !isEnabled) return false;
    if (filterType === 'inactive' && isEnabled) return false;

    return true;
  });

  const groupedTools = tools.filter(filteredTools => filteredTools).reduce(
    (acc, tool) => {
      const cat = tool.category || 'Other';
      if (!acc[cat]) acc[cat] = [];
      acc[cat].push(tool);
      return acc;
    },
    {} as Record<string, api.backend.ToolUIInfo[]>,
  );

  return (
    <div className="space-y-6">
      <div>
        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-foreground">Tools</h3>
            <div className="flex items-center gap-3">
              <Input
                placeholder="Search tools by name..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="max-w-sm"
              />
              <Select
                value={filterType}
                onValueChange={(value: 'all' | 'active' | 'inactive') =>
                  setFilterType(value)
                }
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Filter" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="inactive">Inactive</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <p className="text-sm text-muted-foreground">
            Configure agent tools for the active workspace.
          </p>
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Profiles</h4>
        <div className="profiles-grid">
          {profiles.map((profile) => (
            <BaseCard
              key={profile.id}
              color={
                selectedProfileID === profile.id ? profile.color : '#4b5563'
              }
              headerLeft={
                <span className="text-xs text-white opacity-90">Profile</span>
              }
              headerRight={
                profile.name !== 'Default' ? (
                  <button
                    className="base-card-btn"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDeleteProfile(profile.id);
                    }}
                  >
                    <Icon name="Trash2" className="w-3 h-3" />
                  </button>
                ) : null
              }
              icon={profile.icon}
              title={profile.name}
              selected={selectedProfileID === profile.id}
              className="cursor-pointer"
              small
            >
              <div className="base-card-desc">{profile.tools.length} tools</div>
            </BaseCard>
          ))}
          <BaseCard
            color="#374151"
            headerLeft={
              <span className="text-xs text-white opacity-90">Action</span>
            }
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
          <h4 className="text-sm font-medium text-foreground mb-2">
            {category}
          </h4>
          <div className="tools-grid">
            {categoryTools.map((tool) => {
              const isEnabled = selectedProfile?.tools.includes(tool.name) ?? false;
              return (
                <BaseCard
                  key={tool.name}
                  color={isEnabled ? categoryColors[category] : '#6b7280'}
                  headerLeft={
                    <span className="text-xs text-white opacity-90">
                      {category}
                    </span>
                  }
                  headerRight={
                    <Switch
                      checked={isEnabled}
                      onCheckedChange={(checked) =>
                        handleToggle(tool.name, checked)
                      }
                      className="data-[state=checked]:bg-white/30"
                    />
                  }
                  icon="🔧"
                  title={tool.name}
                >
                  <div className="base-card-desc">{tool.description}</div>
                </BaseCard>
              );
            })}
          </div>
        </div>
      ))}

      {tools.length === 0 && (
        <Card>
          <CardContent className="py-8">
            <p className="text-sm text-muted-foreground text-center">
              No tools available
            </p>
          </CardContent>
        </Card>
      )}

      <Dialog
        open={showNewProfileDialog}
        onOpenChange={setShowNewProfileDialog}
      >
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
              <div className="col-span-3 flex items-center gap-1">
                <div
                  className="w-10 h-10 rounded border cursor-pointer"
                  style={{ backgroundColor: newProfileColor }}
                  onClick={() =>
                    document.getElementById('profile-pick-color')?.click()
                  }
                />
                <input
                  id="profile-pick-color"
                  type="color"
                  value={newProfileColor}
                  onChange={(e) => setNewProfileColor(e.target.value)}
                  className="w-0 h-0 opacity-0"
                />
              </div>
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <span className="text-right text-sm">Icon</span>
              <IconPicker value={newProfileIcon} onChange={setNewProfileIcon} />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowNewProfileDialog(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleCreateProfile}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default ToolsSection;
