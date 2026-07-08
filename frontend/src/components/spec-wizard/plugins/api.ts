// Tipos exportados pelo backend (espelho pkg/patterns + pkg/registry).
// Os bindings Wails (frontend/wailsjs/go/main/App.d.ts) já declaram
// window.go.main.App.GetPatterns / GetArchitectures / GetExperts.
// Aqui apenas re-declamos as formas tipadas para consumo local.
export interface Pattern {
  id: string;
  name: string;
  category: string;
  group: string;
  scope: string;
  description: string;
}

export interface Architecture {
  id: string;
  name: string;
  description: string;
  best_for: string[];
  aliases: string[];
}

export interface ExpertPlugin {
  id: string;
  name: string;
  description: string;
  language: string;
  start_command: string;
  triggers: string[];
}

// TTL do cache: 5 minutos.
const CACHE_TTL_MS = 5 * 60 * 1000;

interface CacheEntry<T> {
  data: T;
  fetchedAt: number;
}
const patternsCache = new Map<string, CacheEntry<Pattern[]>>();
let architecturesCache: CacheEntry<Architecture[]> | null = null;
let expertsCache: CacheEntry<ExpertPlugin[]> | null = null;

export async function fetchPatterns(lang: string): Promise<Pattern[]> {
  if (!lang) return [];
  const cached = patternsCache.get(lang);
  if (cached && Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
    return cached.data;
  }
  if (!window.go?.main?.App?.GetPatterns) {
    throw new Error('Wails bindings not available — app not running');
  }
  const data = (await window.go.main.App.GetPatterns(
    lang,
  )) as unknown as Pattern[];
  patternsCache.set(lang, { data, fetchedAt: Date.now() });
  return data;
}

export async function fetchArchitectures(): Promise<Architecture[]> {
  if (architecturesCache && Date.now() - architecturesCache.fetchedAt < CACHE_TTL_MS) {
    return architecturesCache.data;
  }
  if (!window.go?.main?.App?.GetArchitectures) {
    throw new Error('Wails bindings not available');
  }
  const data = (await window.go.main.App.GetArchitectures()) as unknown as Architecture[];
  architecturesCache = { data, fetchedAt: Date.now() };
  return data;
}

export async function fetchExperts(): Promise<ExpertPlugin[]> {
  if (expertsCache && Date.now() - expertsCache.fetchedAt < CACHE_TTL_MS) {
    return expertsCache.data;
  }
  if (!window.go?.main?.App?.GetExperts) {
    throw new Error('Wails bindings not available');
  }
  const data = (await window.go.main.App.GetExperts()) as unknown as ExpertPlugin[];
  expertsCache = { data, fetchedAt: Date.now() };
  return data;
}

export function clearPluginsCache(): void {
  patternsCache.clear();
  architecturesCache = null;
  expertsCache = null;
}