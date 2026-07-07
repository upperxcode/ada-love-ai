import { useState, useEffect } from 'react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Switch } from '../ui/switch';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import { Combobox } from '../ui/combobox';
import { BaseCard } from '../BaseCard';
import { EditDialog } from '../EditDialog';
import { Icon } from '../Icon';
import * as api from '../../api';

const knownApiUrls = [
  { name: 'OpenAI', url: 'https://api.openai.com/v1' },
  {
    name: 'Cloudflare',
    url: 'https://api.cloudflare.com/client/v4/accounts/{account_id}/ai/v1',
  },
  { name: 'OpenRouter', url: 'https://openrouter.ai/api/v1' },
  { name: 'Anthropic', url: 'https://api.anthropic.com/v1' },
  {
    name: 'Google Gemini',
    url: 'https://generativelanguage.googleapis.com/v1beta',
  },
  { name: 'Ollama', url: 'http://localhost:11434/v1' },
];

function ModelsSection() {
  const [adaConfig, setAdaConfig] = useState<api.backend.AdaConfig | null>(
    null,
  );
  const [showProviderDialog, setShowProviderDialog] = useState(false);
  const [editingProvider, setEditingProvider] = useState<{
    name: string;
    config: api.backend.ProviderConfig;
  } | null>(null);
  const [collapsed, setCollapsed] = useState<{
    apiUrl: boolean;
    apiKeys: boolean;
    models: boolean;
  }>({
    apiUrl: false,
    apiKeys: false,
    models: true,
  });
  const [newModelForm, setNewModelForm] = useState({
    name: '',
    type: '' as api.backend.ModelType | '',
  });
  const [providerForm, setProviderForm] = useState({
    name: '',
    api_url: '',
    api_keys: [] as api.backend.ProviderApiKey[],
    type_connection: '',
    models: {} as Record<string, api.backend.ModelSettings>,
  });

  // Fetch models dialog state
  const [showFetchModelsDialog, setShowFetchModelsDialog] = useState(false);
  const [fetchedModels, setFetchedModels] = useState<
    api.backend.ProviderModel[]
  >([]);
  const [selectedModelsToAdd, setSelectedModelsToAdd] = useState<string[]>([]);
  // Multi-select filter: set of active type/capability filters. Model is shown if
  // it matches ALL active filters (pure AND). Each filter restricts the list;
  // toggling a filter off relaxes that restriction. Empty set = show all.
  const [activeFilters, setActiveFilters] = useState<Set<string>>(new Set());
  const [fetchingModels, setFetchingModels] = useState(false);

  const toggleFilter = (filter: string) => {
    setActiveFilters((prev) => {
      const next = new Set(prev);
      if (next.has(filter)) {
        next.delete(filter);
      } else {
        next.add(filter);
      }
      return next;
    });
  };

  // Returns true if a fetched model passes the current active filter set.
  // Pure AND: every active filter must match. Empty set = show all.
  // Each of free/thinking/tools/embedding/vision is an independent capability:
  // clicking a filter adds a restriction, un-clicking removes it.
  const matchesFilter = (model: api.backend.ProviderModel) => {
    if (activeFilters.size === 0) return true;

    const isFree = !!model.free;
    const isThinking = !!model.thinking;
    const isTools = model.tools !== false;
    const isEmbedding = !!model.embedding;
    const isVision = !!model.vision;

    if (activeFilters.has('free') && !isFree) return false;
    if (activeFilters.has('thinking') && !isThinking) return false;
    if (activeFilters.has('tools') && !isTools) return false;
    if (activeFilters.has('embedding') && !isEmbedding) return false;
    if (activeFilters.has('vision') && !isVision) return false;

    return true;
  };

  // API key validation
  const [validatingApiKey, setValidatingApiKey] = useState<number | null>(null);
  const [apiKeyValidationResults, setApiKeyValidationResults] = useState<
    Record<number, api.backend.ProviderTestResult>
  >({});

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    const cfg = await api.getAdaConfig();
    setAdaConfig(cfg || new api.backend.AdaConfig());
  };

  const saveConfig = async () => {
    if (adaConfig) {
      await api.setAdaConfig(adaConfig);
    }
  };

  const handleOpenProvider = (
    name: string,
    config: api.backend.ProviderConfig,
  ) => {
    setEditingProvider({ name, config });
    setProviderForm({
      name,
      api_url: config.api_url,
      api_keys: config.api_keys || [],
      type_connection: config.type_connection,
      models: config.models || {},
    });
    setShowProviderDialog(true);
  };

  const handleAddModel = () => {
    if (!newModelForm.name || !newModelForm.type) return;
    setProviderForm({
      ...providerForm,
      models: {
        ...providerForm.models,
        [newModelForm.name]: new api.backend.ModelSettings({
          type: newModelForm.type,
        }),
      },
    });
    setNewModelForm({ name: '', type: '' });
  };

  const handleRemoveModel = (modelName: string) => {
    const newModels = { ...providerForm.models };
    delete newModels[modelName];
    setProviderForm({ ...providerForm, models: newModels });
  };

  const handleAddApiKey = () => {
    setProviderForm({
      ...providerForm,
      api_keys: [...providerForm.api_keys, { key: '', user_key: '' }],
    });
  };

  const handleUpdateApiKey = (
    index: number,
    field: 'key' | 'user_key',
    value: string,
  ) => {
    const newKeys = [...providerForm.api_keys];
    newKeys[index] = { ...newKeys[index], [field]: value };
    setProviderForm({ ...providerForm, api_keys: newKeys });
  };

  const handleRemoveApiKey = (index: number) => {
    const newKeys = providerForm.api_keys.filter((_, i) => i !== index);
    setProviderForm({ ...providerForm, api_keys: newKeys });
    // Clear validation results for this index
    setApiKeyValidationResults((prev) => {
      const newResults = { ...prev };
      delete newResults[index];
      return newResults;
    });
  };

  const handleValidateApiKey = async (index: number) => {
    const apiKeyEntry = providerForm.api_keys[index];
    if (
      !apiKeyEntry ||
      !apiKeyEntry.key ||
      !providerForm.api_url ||
      !providerForm.type_connection
    )
      return;

    setValidatingApiKey(index);
    const result = await api.testProviderConnection(
      providerForm.name,
      apiKeyEntry.key,
      providerForm.api_url,
      providerForm.type_connection,
    );
    setApiKeyValidationResults((prev) => ({ ...prev, [index]: result }));
    setValidatingApiKey(null);
  };

  const handleFetchModels = async () => {
    // Use first API key if available
    const firstApiKey = providerForm.api_keys.find((k) => k.key.trim() !== '');
    if (
      !firstApiKey ||
      !providerForm.api_url ||
      !providerForm.type_connection ||
      !providerForm.name
    )
      return;

    setFetchingModels(true);
    const models = await api.fetchProviderModels(
      providerForm.name,
      firstApiKey.key,
      providerForm.api_url,
      providerForm.type_connection,
    );
    setFetchedModels(models);
    setSelectedModelsToAdd([]);
    setShowFetchModelsDialog(true);
    setFetchingModels(false);
  };

  const handleAddSelectedModels = () => {
    const newModels = { ...providerForm.models };
    selectedModelsToAdd.forEach((modelName) => {
      const model = fetchedModels.find((m) => m.name === modelName);
      if (model) {
        newModels[modelName] = new api.backend.ModelSettings({
          vision: model.vision,
          embedding: model.embedding,
          tools: model.tools,
          free: model.free,
          thinking: model.thinking,
        });
      }
    });
    setProviderForm({ ...providerForm, models: newModels });
    setShowFetchModelsDialog(false);
    setSelectedModelsToAdd([]);
    setFetchedModels([]);
  };

  const handleSaveProvider = async () => {
    if (!adaConfig || !providerForm.name) return;

    // Create a proper AdaConfig instance
    const newConfig = new api.backend.AdaConfig({
      ...adaConfig,
      providers: {
        ...(adaConfig.providers || {}),
        [providerForm.name]: new api.backend.ProviderConfig({
          api_url: providerForm.api_url,
          api_keys: providerForm.api_keys.filter((k) => k.key.trim() !== ''),
          type_connection: providerForm.type_connection,
          models: providerForm.models,
        }),
      },
    });

    setAdaConfig(newConfig);
    await api.setAdaConfig(newConfig);
    setShowProviderDialog(false);
    setEditingProvider(null);
  };

  const handleDeleteProvider = async (name: string) => {
    if (!adaConfig) return;

    // Create proper AdaConfig instance
    const newProviders = { ...(adaConfig.providers || {}) };
    delete newProviders[name];

    const newConfig = new api.backend.AdaConfig({
      ...adaConfig,
      providers: newProviders,
    });

    setAdaConfig(newConfig);
    await api.setAdaConfig(newConfig);
  };

  // Modelos disponíveis para os selects de Embedding/Image, derivados dos
  // modelos cadastrados no provider escolhido e filtrados por capability.
  const embeddingModels = (() => {
    const provider = adaConfig?.embedding_provider;
    const models = provider
      ? adaConfig?.providers?.[provider]?.models
      : undefined;
    if (!models) return [];
    return Object.entries(models)
      .filter(([, s]) => s?.embedding)
      .map(([name]) => ({ value: name, label: name }));
  })();

  const imageModels = (() => {
    const provider = adaConfig?.image_provider;
    const models = provider
      ? adaConfig?.providers?.[provider]?.models
      : undefined;
    if (!models) return [];
    return Object.entries(models)
      .filter(([, s]) => s?.vision)
      .map(([name]) => ({ value: name, label: name }));
  })();

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-foreground">Models</h3>
        <p className="text-sm text-muted-foreground">
          Configure providers, API keys, and models.
        </p>
      </div>

      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Embedding Provider</label>
            <Select
              value={adaConfig?.embedding_provider || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    embedding_provider: v,
                  });
                  setAdaConfig(newCfg);
                }
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select provider" />
              </SelectTrigger>
              <SelectContent onCloseAutoFocus={() => saveConfig()}>
                {adaConfig &&
                  Object.keys(adaConfig.providers || {}).map((name) => (
                    <SelectItem key={name} value={name}>
                      {name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Embedding Model</label>
            <Combobox
              value={adaConfig?.embedding_model || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    embedding_model: v,
                  });
                  setAdaConfig(newCfg);
                  saveConfig();
                }
              }}
              options={embeddingModels}
              placeholder="e.g., text-embedding-3-small"
              emptyText="Nenhum modelo de embedding neste provider — digite um valor custom"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Image Provider</label>
            <Select
              value={adaConfig?.image_provider || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    image_provider: v,
                  });
                  setAdaConfig(newCfg);
                }
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select provider" />
              </SelectTrigger>
              <SelectContent onCloseAutoFocus={() => saveConfig()}>
                {adaConfig &&
                  Object.keys(adaConfig.providers || {}).map((name) => (
                    <SelectItem key={name} value={name}>
                      {name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Image Model</label>
            <Combobox
              value={adaConfig?.image_model || ''}
              onValueChange={(v) => {
                if (adaConfig) {
                  const newCfg = new api.backend.AdaConfig({
                    ...adaConfig,
                    image_model: v,
                  });
                  setAdaConfig(newCfg);
                  saveConfig();
                }
              }}
              options={imageModels}
              placeholder="e.g., dall-e-3"
              emptyText="Nenhum modelo com visão neste provider — digite um valor custom"
            />
          </div>
        </div>
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h4 className="text-sm font-medium text-foreground">Providers</h4>
          <Button
            size="sm"
            onClick={() => {
              setEditingProvider(null);
              setProviderForm({
                name: '',
                api_url: '',
                api_keys: [],
                type_connection: '',
                models: {},
              });
              setShowProviderDialog(true);
            }}
          >
            <Icon name="Plus" size={16} />
            Add Provider
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {adaConfig &&
            Object.entries(adaConfig.providers || {}).map(([name, config]) => (
              <BaseCard
                key={name}
                color="#6b7280"
                headerLeft={
                  <span className="text-xs text-white opacity-90">
                    Provider
                  </span>
                }
                headerRight={
                  <div className="flex gap-1">
                    <button
                      className="base-card-btn"
                      onClick={() => handleOpenProvider(name, config)}
                    >
                      <Icon name="Edit" className="w-3 h-3" />
                    </button>
                    <button
                      className="base-card-btn"
                      onClick={() => handleDeleteProvider(name)}
                    >
                      <Icon name="Trash2" className="w-3 h-3" />
                    </button>
                  </div>
                }
                icon="🔌"
                title={name}
                small
              >
                <div className="text-xs text-muted-foreground truncate">
                  {config.api_url || 'No API URL'}
                </div>
              </BaseCard>
            ))}
        </div>
      </div>

      <EditDialog
        open={showProviderDialog}
        onOpenChange={setShowProviderDialog}
        title={editingProvider ? 'Edit Provider' : 'New Provider'}
        onSave={handleSaveProvider}
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Name</label>
            <Input
              value={providerForm.name}
              onChange={(e) =>
                setProviderForm({ ...providerForm, name: e.target.value })
              }
              placeholder="e.g., OpenAI"
              disabled={!!editingProvider}
            />
          </div>
          <div className="space-y-2">
            <div
              className="flex items-center justify-between cursor-pointer"
              onClick={() =>
                setCollapsed({ ...collapsed, apiUrl: !collapsed.apiUrl })
              }
            >
              <label className="text-sm font-medium">API URL</label>
              <Icon
                name={collapsed.apiUrl ? 'ChevronDown' : 'ChevronRight'}
                size={16}
                className="text-muted-foreground"
              />
            </div>
            {!collapsed.apiUrl && (
              <div className="space-y-1">
                <Select
                  value={
                    knownApiUrls.find((u) => u.url === providerForm.api_url)
                      ? providerForm.api_url
                      : 'custom'
                  }
                  onValueChange={(v) => {
                    if (v !== 'custom') {
                      setProviderForm({ ...providerForm, api_url: v });
                    }
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select known API" />
                  </SelectTrigger>
                  <SelectContent>
                    {knownApiUrls.map((u) => (
                      <SelectItem key={u.url} value={u.url}>
                        {u.name}
                      </SelectItem>
                    ))}
                    <SelectItem value="custom">Custom URL</SelectItem>
                  </SelectContent>
                </Select>
                <Input
                  value={providerForm.api_url}
                  onChange={(e) =>
                    setProviderForm({
                      ...providerForm,
                      api_url: e.target.value,
                    })
                  }
                  placeholder="Enter custom API URL"
                />
              </div>
            )}
          </div>
          <div className="space-y-2">
            <div
              className="flex items-center justify-between cursor-pointer"
              onClick={() =>
                setCollapsed({ ...collapsed, apiKeys: !collapsed.apiKeys })
              }
            >
              <div className="flex items-center gap-2">
                <label className="text-sm font-medium">API Keys</label>
                {providerForm.api_keys.length > 0 && (
                  <span className="text-xs text-muted-foreground">
                    ({providerForm.api_keys.length})
                  </span>
                )}
              </div>
              <Icon
                name={collapsed.apiKeys ? 'ChevronDown' : 'ChevronRight'}
                size={16}
                className="text-muted-foreground"
              />
            </div>
            {!collapsed.apiKeys && (
              <div className="space-y-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleAddApiKey}
                  className="w-full"
                >
                  <Icon name="Plus" size={14} className="mr-1" /> Add API Key
                </Button>
                {providerForm.api_keys.map((apiKey, index) => (
                  <div
                    key={index}
                    className="space-y-1 p-2 border border-input rounded-md"
                  >
                    <div className="flex gap-2">
                      <Input
                        type="password"
                        value={apiKey.key}
                        onChange={(e) =>
                          handleUpdateApiKey(index, 'key', e.target.value)
                        }
                        placeholder="Valor da chave ou nome de variável (ex: OPENROUTER_API_KEY)"
                        className="flex-1"
                      />
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleValidateApiKey(index)}
                        disabled={validatingApiKey === index}
                      >
                        {validatingApiKey === index ? (
                          <span className="flex items-center gap-1">
                            <span className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full"></span>
                          </span>
                        ) : (
                          <span>Validate</span>
                        )}
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleRemoveApiKey(index)}
                      >
                        <Icon name="Trash2" size={14} />
                      </Button>
                    </div>
                    <Input
                      value={apiKey.user_key}
                      onChange={(e) =>
                        handleUpdateApiKey(index, 'user_key', e.target.value)
                      }
                      placeholder="User Key (label or Cloudflare user ID)"
                    />
                    {apiKeyValidationResults[index] && (
                      <div
                        className={`text-xs p-1 rounded ${
                          apiKeyValidationResults[index].success
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {apiKeyValidationResults[index].success ? '✓ ' : '✗ '}
                        {apiKeyValidationResults[index].message}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
          <div className="space-y-2">
            <div
              className="flex items-center justify-between cursor-pointer"
              onClick={() =>
                setCollapsed({ ...collapsed, models: !collapsed.models })
              }
            >
              <div className="flex items-center gap-2">
                <label className="text-sm font-medium">Models</label>
                {Object.keys(providerForm.models).length > 0 && (
                  <span className="text-xs text-muted-foreground">
                    ({Object.keys(providerForm.models).length})
                  </span>
                )}
              </div>
              <Icon
                name={collapsed.models ? 'ChevronDown' : 'ChevronRight'}
                size={16}
                className="text-muted-foreground"
              />
            </div>
            {!collapsed.models && (
              <div className="space-y-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleFetchModels}
                  disabled={
                    fetchingModels ||
                    providerForm.api_keys.filter((k) => k.key.trim() !== '')
                      .length === 0
                  }
                  className="w-full"
                >
                  {fetchingModels ? (
                    <span className="flex items-center gap-1">
                      <span className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full"></span>
                      Fetching...
                    </span>
                  ) : (
                    <span>
                      <Icon name="Plus" size={14} className="mr-1" /> Fetch
                      Models
                    </span>
                  )}
                </Button>
                <div className="flex gap-2">
                  <Input
                    value={newModelForm.name}
                    onChange={(e) =>
                      setNewModelForm({ ...newModelForm, name: e.target.value })
                    }
                    placeholder="Model name"
                    className="flex-1"
                  />
                  <Select
                    value={newModelForm.type}
                    onValueChange={(v) =>
                      setNewModelForm({
                        ...newModelForm,
                        type: v as api.backend.ModelType,
                      })
                    }
                  >
                    <SelectTrigger className="w-32">
                      <SelectValue placeholder="Type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="free">Free</SelectItem>
                      <SelectItem value="thinking">Thinking</SelectItem>
                      <SelectItem value="tools">Tools</SelectItem>
                    </SelectContent>
                  </Select>
                  <Button
                    size="sm"
                    onClick={handleAddModel}
                    disabled={!newModelForm.name || !newModelForm.type}
                  >
                    <Icon name="Plus" size={14} />
                  </Button>
                </div>
                <div className="space-y-1">
                  {Object.entries(providerForm.models).map(
                    ([modelName, settings]) => (
                      <div
                        key={modelName}
                        className="flex items-center justify-between p-2 border border-input rounded-md"
                      >
                        <div className="flex items-center gap-2 min-w-0">
                          <span className="text-sm font-medium truncate">
                            {modelName}
                          </span>
                          {/* Clickable capability toggles */}
                          <div className="flex items-center gap-1 shrink-0">
                            <button
                              type="button"
                              className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                                settings?.free
                                  ? 'bg-green-500/15 text-green-600 dark:text-green-400'
                                  : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                              }`}
                              title="Free"
                              onClick={(e) => {
                                e.stopPropagation();
                                const updated = new api.backend.ModelSettings({
                                  ...settings,
                                  free: !settings?.free,
                                });
                                setProviderForm({
                                  ...providerForm,
                                  models: {
                                    ...providerForm.models,
                                    [modelName]: updated,
                                  },
                                });
                              }}
                            >
                              <Icon name="MessageSquare" size={12} />
                            </button>
                            <button
                              type="button"
                              className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                                settings?.thinking
                                  ? 'bg-purple-500/15 text-purple-600 dark:text-purple-400'
                                  : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                              }`}
                              title="Thinking"
                              onClick={(e) => {
                                e.stopPropagation();
                                const updated = new api.backend.ModelSettings({
                                  ...settings,
                                  thinking: !settings?.thinking,
                                });
                                setProviderForm({
                                  ...providerForm,
                                  models: {
                                    ...providerForm.models,
                                    [modelName]: updated,
                                  },
                                });
                              }}
                            >
                              <Icon name="Sparkles" size={12} />
                            </button>
                            <button
                              type="button"
                              className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                                settings?.tools !== false
                                  ? 'bg-blue-500/15 text-blue-600 dark:text-blue-400'
                                  : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                              }`}
                              title="Tools"
                              onClick={(e) => {
                                e.stopPropagation();
                                const updated = new api.backend.ModelSettings({
                                  ...settings,
                                  tools:
                                    settings?.tools === false ? true : false,
                                });
                                setProviderForm({
                                  ...providerForm,
                                  models: {
                                    ...providerForm.models,
                                    [modelName]: updated,
                                  },
                                });
                              }}
                            >
                              <Icon name="Wrench" size={12} />
                            </button>
                            <button
                              type="button"
                              className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                                settings?.embedding
                                  ? 'bg-orange-500/15 text-orange-600 dark:text-orange-400'
                                  : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                              }`}
                              title="Embedding"
                              onClick={(e) => {
                                e.stopPropagation();
                                const updated = new api.backend.ModelSettings({
                                  ...settings,
                                  embedding: !settings?.embedding,
                                });
                                setProviderForm({
                                  ...providerForm,
                                  models: {
                                    ...providerForm.models,
                                    [modelName]: updated,
                                  },
                                });
                              }}
                            >
                              <Icon name="Network" size={12} />
                            </button>
                            <button
                              type="button"
                              className={`inline-flex items-center justify-center h-5 w-5 rounded transition-colors ${
                                settings?.vision
                                  ? 'bg-cyan-500/15 text-cyan-600 dark:text-cyan-400'
                                  : 'text-muted-foreground/30 hover:text-muted-foreground/60'
                              }`}
                              title="Vision"
                              onClick={(e) => {
                                e.stopPropagation();
                                const updated = new api.backend.ModelSettings({
                                  ...settings,
                                  vision: !settings?.vision,
                                });
                                setProviderForm({
                                  ...providerForm,
                                  models: {
                                    ...providerForm.models,
                                    [modelName]: updated,
                                  },
                                });
                              }}
                            >
                              <Icon name="Eye" size={12} />
                            </button>
                          </div>
                        </div>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleRemoveModel(modelName)}
                        >
                          <Icon name="Trash2" size={14} />
                        </Button>
                      </div>
                    ),
                  )}
                </div>
              </div>
            )}
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Connection Type</label>
            <Select
              value={providerForm.type_connection}
              onValueChange={(v) =>
                setProviderForm({ ...providerForm, type_connection: v })
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="openai">OpenAI Compatible</SelectItem>
                <SelectItem value="anthropic">Anthropic</SelectItem>
                <SelectItem value="gemini">Google Gemini</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </EditDialog>

      {/* Fetch Models Dialog */}
      <Dialog
        open={showFetchModelsDialog}
        onOpenChange={setShowFetchModelsDialog}
      >
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Select Models to Add</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            {/* Filter Checkboxes */}
            <div className="flex flex-wrap gap-2">
              {(
                [
                  { key: 'free', label: 'Free', icon: 'MessageSquare' },
                  { key: 'thinking', label: 'Thinking', icon: 'Sparkles' },
                  { key: 'tools', label: 'Tools', icon: 'Wrench' },
                  { key: 'embedding', label: 'Embedding', icon: 'Network' },
                  { key: 'vision', label: 'Vision', icon: 'Eye' },
                ] as const
              ).map(({ key, label, icon }) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => toggleFilter(key)}
                  className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md border text-xs transition-colors ${
                    activeFilters.has(key)
                      ? 'border-primary bg-primary/10 text-primary'
                      : 'border-border text-muted-foreground hover:border-primary/50 hover:text-foreground'
                  }`}
                >
                  <Icon name={icon} size={12} />
                  <span>{label}</span>
                </button>
              ))}
            </div>

            {/* Models List — fixed height */}
            <div className="h-80 overflow-y-auto space-y-1">
              {fetchedModels
                .filter(matchesFilter)
                .filter((model) => !providerForm.models[model.name]) // Exclude already added models
                .map((model) => {
                  const isSelected = selectedModelsToAdd.includes(model.name);
                  const isFree = !!model.free;
                  const isThinking = !!model.thinking;
                  const isTools = model.tools !== false;
                  const isEmbedding = !!model.embedding;
                  const isVision = !!model.vision;
                  return (
                    <div
                      key={model.name}
                      className={`flex items-center gap-2.5 p-2 border rounded-md cursor-pointer transition-colors ${
                        isSelected
                          ? 'border-primary bg-primary/5'
                          : 'border-input hover:bg-accent'
                      }`}
                      onClick={() => {
                        setSelectedModelsToAdd((prev) => {
                          if (prev.includes(model.name)) {
                            return prev.filter((n) => n !== model.name);
                          } else {
                            return [...prev, model.name];
                          }
                        });
                      }}
                    >
                      {/* Themed checkbox */}
                      <span
                        className={`inline-flex items-center justify-center h-4 w-4 rounded border transition-colors shrink-0 ${
                          isSelected
                            ? 'bg-primary border-primary text-primary-foreground'
                            : 'border-muted-foreground/40'
                        }`}
                      >
                        {isSelected && <Icon name="Check" size={10} />}
                      </span>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium truncate">
                          {model.name}
                        </div>
                      </div>
                      {/* Capability icons — show all applicable */}
                      <div className="flex items-center gap-1 text-muted-foreground shrink-0">
                        {isFree && (
                          <span title="Free">
                            <Icon name="MessageSquare" size={14} />
                          </span>
                        )}
                        {isThinking && (
                          <span title="Thinking">
                            <Icon name="Sparkles" size={14} />
                          </span>
                        )}
                        {isTools && (
                          <span title="Tools">
                            <Icon name="Wrench" size={14} />
                          </span>
                        )}
                        {isEmbedding && (
                          <span title="Embedding">
                            <Icon name="Network" size={14} />
                          </span>
                        )}
                        {isVision && (
                          <span title="Vision">
                            <Icon name="Eye" size={14} />
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              {fetchedModels
                .filter(matchesFilter)
                .filter((m) => !providerForm.models[m.name]).length === 0 && (
                <div className="flex items-center justify-center h-full text-sm text-muted-foreground">
                  No models available for the selected filter.
                </div>
              )}
            </div>
          </div>
          <DialogFooter className="sm:justify-between">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                const visible = fetchedModels
                  .filter(matchesFilter)
                  .filter((m) => !providerForm.models[m.name]);
                const allSelected =
                  visible.length > 0 &&
                  visible.every((m) => selectedModelsToAdd.includes(m.name));
                if (allSelected) {
                  setSelectedModelsToAdd([]);
                } else {
                  setSelectedModelsToAdd(visible.map((m) => m.name));
                }
              }}
            >
              {(() => {
                const visible = fetchedModels
                  .filter(matchesFilter)
                  .filter((m) => !providerForm.models[m.name]);
                const allSelected =
                  visible.length > 0 &&
                  visible.every((m) => selectedModelsToAdd.includes(m.name));
                return allSelected ? 'Uncheck All' : 'Check All';
              })()}
            </Button>
            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setShowFetchModelsDialog(false);
                  setSelectedModelsToAdd([]);
                  setFetchedModels([]);
                }}
              >
                Cancel
              </Button>
              <Button
                onClick={handleAddSelectedModels}
                disabled={selectedModelsToAdd.length === 0}
              >
                Add Selected ({selectedModelsToAdd.length})
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default ModelsSection;
