import { useState } from 'react';
import { Icon } from './Icon';
import { Button } from './ui/button';
import GeneralSection from './settings/GeneralSection';
import WorkspacesSection from './settings/WorkspacesSection';
import AgentsSection from './settings/AgentsSection';
import WorkersSection from './settings/WorkersSection';
import SkillsSection from './settings/SkillsSection';
import ToolsSection from './settings/ToolsSection';
import ModelsSection from './settings/ModelsSection';
import MCPSection from './settings/MCPSection';

interface SettingsPageProps {
  onClose?: () => void;
}

const settingsSections = [
  { id: 'general', icon: 'Settings', label: 'General' },
  { id: 'workspace', icon: 'Folder', label: 'Workspace' },
  { id: 'workers', icon: 'Bot', label: 'Workers' },
  { id: 'agents', icon: 'User', label: 'Agents' },
  { id: 'skills', icon: 'Brain', label: 'Skills' },
  { id: 'tools', icon: 'Wrench', label: 'Tools' },
  { id: 'mcp', icon: 'Plug', label: 'MCP' },
  { id: 'models', icon: 'Cpu', label: 'Models' },
];

function SettingsPage({ onClose }: SettingsPageProps) {
  const [activeSection, setActiveSection] = useState('general');

  return (
    <div className="fixed inset-0 z-50 flex bg-background">
      <div className="settings-sidebar">
        <div className="settings-sidebar-header">
          <span className="text-sm font-semibold text-foreground">
            Settings
          </span>
          {onClose && (
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 p-0"
              onClick={onClose}
              title="Fechar configurações"
            >
              <Icon name="X" className="w-4 h-4" />
            </Button>
          )}
        </div>
        <div className="settings-sidebar-nav">
          {settingsSections.map((section) => (
            <button
              key={section.id}
              className={`settings-nav-item ${activeSection === section.id ? 'active' : ''}`}
              onClick={() => setActiveSection(section.id)}
            >
              <span className="settings-nav-icon">
                <Icon name={section.icon} className="w-4 h-4" />
              </span>
              <span>{section.label}</span>
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto">
        <div className="settings-content">
          {activeSection === 'general' && <GeneralSection />}
          {activeSection === 'workspace' && <WorkspacesSection />}
          {activeSection === 'workers' && <WorkersSection />}
          {activeSection === 'agents' && <AgentsSection />}
          {activeSection === 'skills' && <SkillsSection />}
          {activeSection === 'tools' && <ToolsSection />}
          {activeSection === 'mcp' && <MCPSection />}
          {activeSection === 'models' && <ModelsSection />}
        </div>
      </div>
    </div>
  );
}

export default SettingsPage;
