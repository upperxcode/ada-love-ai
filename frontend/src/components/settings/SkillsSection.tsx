import { useState, useEffect, useMemo } from 'react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import { Card, CardContent } from '../ui/card';
import { BaseCard } from '../BaseCard';
import { Icon } from '../Icon';
import * as api from '../../api';

const tagColors: Record<string, string> = {
  coding: '#3b82f6',
  writing: '#22c55e',
  analysis: '#f59e0b',
  communication: '#a855f7',
  reasoning: '#ef4444',
  creative: '#ec4899',
  research: '#06b6d4',
  planning: '#f97316',
  debugging: '#8b5cf6',
  testing: '#14b8a6',
};

function getTagColor(tag: string): string {
  const lower = tag.toLowerCase();
  for (const [key, color] of Object.entries(tagColors)) {
    if (lower.includes(key)) return color;
  }
  // Deterministic color from string hash
  let hash = 0;
  for (let i = 0; i < tag.length; i++) {
    hash = tag.charCodeAt(i) + ((hash << 5) - hash);
  }
  const palette = [
    '#3b82f6',
    '#22c55e',
    '#f59e0b',
    '#a855f7',
    '#ef4444',
    '#06b6d4',
    '#ec4899',
    '#f97316',
    '#8b5cf6',
    '#14b8a6',
  ];
  return palette[Math.abs(hash) % palette.length];
}

function SkillsSection() {
  const [installedSkills, setInstalledSkills] = useState<
    api.backend.SkillFullInfo[]
  >([]);
  const [searchResults, setSearchResults] = useState<
    api.backend.SearchResult[]
  >([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState<'installed' | 'browse'>(
    'installed',
  );
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [detailSkill, setDetailSkill] =
    useState<api.backend.SkillFullInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [searching, setSearching] = useState(false);
  const [installingSlug, setInstallingSlug] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  // Create/edit form state
  const [formName, setFormName] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formTags, setFormTags] = useState('');
  const [formContent, setFormContent] = useState('');

  const loadInstalledSkills = async () => {
    setLoading(true);
    try {
      const names = await api.getInstalledSkills();
      const infos = await Promise.all((names ?? []).map((n) => api.getSkillFullInfo(n)));
      setInstalledSkills(
        infos.filter((s): s is api.backend.SkillFullInfo => s !== null),
      );
    } catch {
      setInstalledSkills([]);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadInstalledSkills();
  }, []);

  // Group installed skills by their first tag (or "Untagged")
  const groupedByTag = useMemo(() => {
    const groups: Record<string, api.backend.SkillFullInfo[]> = {};
    for (const skill of installedSkills) {
      const tag =
        skill.tags && skill.tags.length > 0 ? skill.tags[0] : 'Untagged';
      if (!groups[tag]) groups[tag] = [];
      groups[tag].push(skill);
    }
    return groups;
  }, [installedSkills]);

  // Group search results by registry
  const groupedByRegistry = useMemo(() => {
    const groups: Record<string, api.backend.SearchResult[]> = {};
    for (const r of searchResults) {
      const reg = r.registry_name || 'Other';
      if (!groups[reg]) groups[reg] = [];
      groups[reg].push(r);
    }
    return groups;
  }, [searchResults]);

  // Create a set of installed skill names/slugs for quick lookup
  const installedSlugs = useMemo(() => {
    return new Set(installedSkills.map((s) => s.name));
  }, [installedSkills]);

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;
    setSearching(true);
    const results = await api.searchSkills(searchQuery.trim());
    setSearchResults(results ?? []);
    setSearching(false);
  };

  const handleInstall = async (
    registryName: string,
    slug: string,
    version: string,
  ) => {
    setInstallingSlug(slug);
    await api.installSkill(registryName, slug, version);
    loadInstalledSkills();
    setInstallingSlug(null);
  };

  const handleUninstall = async (name: string) => {
    await api.uninstallSkill(name);
    setDetailSkill(null);
    loadInstalledSkills();
  };

  const openCreateDialog = (skill?: api.backend.SkillFullInfo) => {
    if (skill) {
      setFormName(skill.name);
      setFormDescription(skill.description);
      setFormTags(skill.tags?.join(', ') || '');
      // Strip frontmatter from markdown to get raw content
      const md = skill.markdown || '';
      const lines = md.split('\n');
      const endIdx = lines.findIndex((l, i) => i > 0 && l.trim() === '---');
      setFormContent(
        endIdx > 0
          ? lines
              .slice(endIdx + 1)
              .join('\n')
              .trim()
          : md,
      );
    } else {
      setFormName('');
      setFormDescription('');
      setFormTags('');
      setFormContent('');
    }
    setShowCreateDialog(true);
  };

  const handleSave = async () => {
    if (!formName.trim()) return;
    setSaving(true);
    await api.saveCustomSkill(
      formName.trim(),
      formDescription.trim(),
      formTags.trim(),
      formContent.trim(),
    );
    setShowCreateDialog(false);
    loadInstalledSkills();
    setSaving(false);
  };

  const installedCount = installedSkills.length;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-foreground">Skills</h3>
          <Button size="sm" onClick={() => openCreateDialog()}>
            <Icon name="Plus" size={14} className="mr-1" /> Create Skill
          </Button>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 p-1 bg-muted rounded-lg w-fit">
          <button
            className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
              activeTab === 'installed'
                ? 'bg-background shadow-sm text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
            onClick={() => setActiveTab('installed')}
          >
            Installed{installedCount > 0 && ` (${installedCount})`}
          </button>
          <button
            className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
              activeTab === 'browse'
                ? 'bg-background shadow-sm text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
            onClick={() => setActiveTab('browse')}
          >
            Browse Registry
          </button>
        </div>
      </div>

      {/* === Installed Tab === */}
      {activeTab === 'installed' && (
        <>
          {loading ? (
            <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
              Loading skills...
            </div>
          ) : installedSkills.length === 0 ? (
            <Card>
              <CardContent className="py-12">
                <div className="text-center space-y-2">
                  <p className="text-sm text-muted-foreground">
                    No skills installed yet.
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Create a custom skill or browse the registry to get started.
                  </p>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-6">
              {Object.entries(groupedByTag).map(([tag, skills]) => (
                <div key={tag}>
                  <div className="flex items-center gap-2 mb-2">
                    <h4 className="text-sm font-medium text-foreground">
                      {tag}
                    </h4>
                    <span className="text-xs text-muted-foreground">
                      ({skills.length})
                    </span>
                    {tag !== 'Untagged' && (
                      <span
                        className="text-[10px] px-1.5 py-0.5 rounded text-white"
                        style={{ backgroundColor: getTagColor(tag) }}
                      >
                        {tag}
                      </span>
                    )}
                  </div>
                  <div className="skills-grid">
                    {skills.map((skill) => (
                      <BaseCard
                        key={skill.name}
                        color={getTagColor(tag)}
                        headerLeft={
                          <span className="text-xs text-white opacity-90">
                            {skill.version || 'local'}
                          </span>
                        }
                        headerRight={
                          <div className="flex gap-1">
                            <button
                              className="base-card-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                setDetailSkill(skill);
                              }}
                            >
                              <Icon name="Eye" size={10} />
                            </button>
                            <button
                              className="base-card-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                openCreateDialog(skill);
                              }}
                            >
                              <Icon name="Edit" size={10} />
                            </button>
                            <button
                              className="base-card-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleUninstall(skill.name);
                              }}
                            >
                              <Icon name="Trash2" size={10} />
                            </button>
                          </div>
                        }
                        icon="🧠"
                        title={skill.name}
                        onClick={() => setDetailSkill(skill)}
                        className="cursor-pointer"
                        small
                      >
                        <div className="base-card-desc">
                          {skill.description || 'No description'}
                        </div>
                        {skill.tags && skill.tags.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-1">
                            {skill.tags.map((t) => (
                              <span
                                key={t}
                                className="text-[9px] px-1 py-0.5 rounded text-white/80"
                                style={{
                                  backgroundColor: getTagColor(t) + '99',
                                }}
                              >
                                {t}
                              </span>
                            ))}
                          </div>
                        )}
                        {installedSlugs.has(skill.name) && (
                          <span className="text-xs text-green-400 mt-1">
                            ✓ Installed
                          </span>
                        )}
                      </BaseCard>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {/* === Browse Registry Tab === */}
      {activeTab === 'browse' && (
        <div className="space-y-4">
          <div className="flex gap-2">
            <Input
              placeholder="Search skills from registries..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="max-w-md"
            />
            <Button
              onClick={handleSearch}
              disabled={searching || !searchQuery.trim()}
            >
              {searching ? (
                <span className="flex items-center gap-1">
                  <Icon name="Settings" size={14} className="animate-spin" />{' '}
                  Searching...
                </span>
              ) : (
                <span className="flex items-center gap-1">
                  <Icon name="Search" size={14} /> Search
                </span>
              )}
            </Button>
          </div>

          {searching ? (
            <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
              Searching registries...
            </div>
          ) : searchResults.length === 0 && searchQuery ? (
            <Card>
              <CardContent className="py-8">
                <p className="text-sm text-muted-foreground text-center">
                  No results found for &ldquo;{searchQuery}&rdquo;
                </p>
              </CardContent>
            </Card>
          ) : searchResults.length === 0 ? (
            <Card>
              <CardContent className="py-8">
                <div className="text-center space-y-2">
                  <p className="text-sm text-muted-foreground">
                    Search for skills from ClawHub and GitHub registries.
                  </p>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-6">
              {Object.entries(groupedByRegistry).map(([registry, results]) => (
                <div key={registry}>
                  <div className="flex items-center gap-2 mb-2">
                    <h4 className="text-sm font-medium text-foreground">
                      {registry}
                    </h4>
                    <span className="text-xs text-muted-foreground">
                      ({results.length})
                    </span>
                  </div>
                  <div className="skills-grid">
                    {results.map((result) => {
                      const isInstalling = installingSlug === result.slug;
                      const isAlreadyInstalled = installedSlugs.has(
                        result.slug,
                      );
                      return (
                        <BaseCard
                          key={result.slug}
                          color={isAlreadyInstalled ? '#22c55e' : '#374151'}
                          headerLeft={
                            <span className="text-xs text-white opacity-90">
                              {result.version || 'latest'}
                            </span>
                          }
                          headerRight={
                            <button
                              className="base-card-btn"
                              disabled={isInstalling || isAlreadyInstalled}
                              onClick={(e) => {
                                e.stopPropagation();
                                if (!isAlreadyInstalled) {
                                  handleInstall(
                                    result.registry_name,
                                    result.slug,
                                    result.version,
                                  );
                                }
                              }}
                            >
                              {isInstalling ? (
                                <Icon
                                  name="Settings"
                                  size={10}
                                  className="animate-spin"
                                />
                              ) : isAlreadyInstalled ? (
                                <Icon name="Check" size={10} />
                              ) : (
                                <Icon name="Plus" size={10} />
                              )}
                            </button>
                          }
                          icon="📦"
                          title={result.display_name || result.name}
                          small
                        >
                          <div className="base-card-desc">
                            {result.summary ||
                              result.description ||
                              'No description'}
                          </div>
                          {isAlreadyInstalled && (
                            <span className="text-xs text-green-400 mt-1">
                              ✓ Installed
                            </span>
                          )}
                        </BaseCard>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* === Create/Edit Skill Dialog === */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="max-w-lg max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>
              {formName ? 'Edit Skill' : 'Create Skill'}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2 overflow-y-auto flex-1">
            <div className="grid grid-cols-4 items-center gap-3">
              <span className="text-right text-sm">Name</span>
              <Input
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder="my-skill-name"
                className="col-span-3"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-3">
              <span className="text-right text-sm">Description</span>
              <Input
                value={formDescription}
                onChange={(e) => setFormDescription(e.target.value)}
                placeholder="What this skill does"
                className="col-span-3"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-3">
              <span className="text-right text-sm">Tags</span>
              <Input
                value={formTags}
                onChange={(e) => setFormTags(e.target.value)}
                placeholder="coding, analysis (comma-separated)"
                className="col-span-3"
              />
            </div>
            <div className="space-y-1.5">
              <span className="text-sm">Content (Markdown)</span>
              <textarea
                className="w-full h-48 rounded-md border border-border bg-transparent px-3 py-2 text-sm resize-none font-mono"
                value={formContent}
                onChange={(e) => setFormContent(e.target.value)}
                placeholder="# My Skill

Instructions for the LLM..."
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowCreateDialog(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={saving || !formName.trim()}>
              {saving ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* === Skill Detail Dialog === */}
      <Dialog open={!!detailSkill} onOpenChange={() => setDetailSkill(null)}>
        <DialogContent className="max-w-lg max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{detailSkill?.name || 'Skill Details'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2 overflow-y-auto flex-1">
            {detailSkill && (
              <>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span className="text-muted-foreground">Version:</span>{' '}
                    {detailSkill.version || '—'}
                  </div>
                  <div>
                    <span className="text-muted-foreground">Source:</span>{' '}
                    {detailSkill.registry || 'local'}
                  </div>
                  <div>
                    <span className="text-muted-foreground">Lines:</span>{' '}
                    {detailSkill.line_count}
                  </div>
                  <div>
                    <span className="text-muted-foreground">Chars:</span>{' '}
                    {detailSkill.char_count}
                  </div>
                </div>
                {detailSkill.description && (
                  <p className="text-sm text-muted-foreground">
                    {detailSkill.description}
                  </p>
                )}
                {detailSkill.tags && detailSkill.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1.5">
                    {detailSkill.tags.map((t) => (
                      <span
                        key={t}
                        className="text-xs px-2 py-0.5 rounded text-white"
                        style={{ backgroundColor: getTagColor(t) }}
                      >
                        {t}
                      </span>
                    ))}
                  </div>
                )}
                <div className="space-y-1.5">
                  <span className="text-sm font-medium">SKILL.md</span>
                  <pre className="text-xs bg-muted rounded-md p-3 overflow-auto max-h-60 whitespace-pre-wrap font-mono">
                    {detailSkill.markdown || 'No content'}
                  </pre>
                </div>
              </>
            )}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => handleUninstall(detailSkill!.name)}
              className="text-destructive hover:text-destructive"
            >
              <Icon name="Trash2" size={14} className="mr-1" /> Uninstall
            </Button>
            <Button
              onClick={() => {
                const s = detailSkill;
                setDetailSkill(null);
                openCreateDialog(s!);
              }}
            >
              <Icon name="Edit" size={14} className="mr-1" /> Edit
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default SkillsSection;
