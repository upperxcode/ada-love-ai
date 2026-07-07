import { useState, useCallback, useEffect, useRef } from 'react';

// ---------------------------------------------------------------------------
// Model Health Scoring
// ---------------------------------------------------------------------------
// Each model has a health score in [0, 1].
//   1.0 = verde perfeito
//   0.5 = amarelo / instável
//   0.0 = vermelho / inutilizável
//
// Sucesso incrementa 0.05 (teto 1.0).
// Falha decrementa 0.25 (piso 0.0).
// A cada 5 minutos sem uso, recupera 0.02 (slow heal).
// ---------------------------------------------------------------------------

const STORAGE_KEY = 'ada:modelHealth';
const HEAL_INTERVAL_MS = 5 * 60 * 1000; // 5 min

interface HealthEntry {
  score: number;
  lastUpdate: number; // timestamp ms
}

type HealthMap = Record<string, HealthEntry>;

function loadHealth(): HealthMap {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch {}
  return {};
}

function persistHealth(map: HealthMap) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(map));
  } catch {}
}

export function getHealthColor(score: number): string {
  if (score >= 0.75) return '#22c55e'; // green-500
  if (score >= 0.5) return '#eab308';  // yellow-500
  if (score >= 0.25) return '#f97316'; // orange-500
  return '#ef4444';                     // red-500
}

export function getHealthLabel(score: number): string {
  if (score >= 0.75) return 'Ótimo';
  if (score >= 0.5) return 'Instável';
  if (score >= 0.25) return 'Lento';
  return 'Com problemas';
}

export function useModelHealth() {
  const [health, setHealth] = useState<HealthMap>(loadHealth);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Slow heal interval
  useEffect(() => {
    timerRef.current = setInterval(() => {
      setHealth((prev) => {
        const now = Date.now();
        let changed = false;
        const next: HealthMap = {};
        for (const [key, entry] of Object.entries(prev)) {
          const elapsed = now - entry.lastUpdate;
          const heals = Math.floor(elapsed / HEAL_INTERVAL_MS);
          if (heals > 0 && entry.score < 1) {
            const newScore = Math.min(1, entry.score + heals * 0.02);
            next[key] = { score: newScore, lastUpdate: now };
            changed = true;
          } else {
            next[key] = entry;
          }
        }
        if (changed) persistHealth(next);
        return changed ? next : prev;
      });
    }, HEAL_INTERVAL_MS);
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  const recordSuccess = useCallback((modelKey: string) => {
    setHealth((prev) => {
      const entry = prev[modelKey] ?? { score: 1, lastUpdate: Date.now() };
      const newScore = Math.min(1, entry.score + 0.05);
      const next = { ...prev, [modelKey]: { score: newScore, lastUpdate: Date.now() } };
      persistHealth(next);
      return next;
    });
  }, []);

  const recordFailure = useCallback((modelKey: string) => {
    setHealth((prev) => {
      const entry = prev[modelKey] ?? { score: 1, lastUpdate: Date.now() };
      const newScore = Math.max(0, entry.score - 0.25);
      const next = { ...prev, [modelKey]: { score: newScore, lastUpdate: Date.now() } };
      persistHealth(next);
      return next;
    });
  }, []);

  const getScore = useCallback(
    (modelKey: string): number => {
      return health[modelKey]?.score ?? 1;
    },
    [health],
  );

  return { health, recordSuccess, recordFailure, getScore };
}
