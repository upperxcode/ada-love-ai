import { useEffect, useState, useCallback } from 'react';
import {
  Architecture,
  ExpertPlugin,
  Pattern,
  fetchArchitectures,
  fetchExperts,
  fetchPatterns,
} from './api';

export interface GroupedPatterns {
  architectures: Pattern[];
  philosophies: Pattern[];
  designPatterns: Pattern[];
  dataPatterns: Pattern[];
  stateManagements: Pattern[];
  dataStrategies: Pattern[];
}

interface HookState {
  patterns: GroupedPatterns;
  architectures: Architecture[];
  experts: ExpertPlugin[];
  isLoading: boolean;
  error: string | null;
}

const EMPTY_GROUPED: GroupedPatterns = {
  architectures: [],
  philosophies: [],
  designPatterns: [],
  dataPatterns: [],
  stateManagements: [],
  dataStrategies: [],
};

function groupPatterns(patterns: Pattern[]): GroupedPatterns {
  return {
    architectures: patterns.filter((p) => p.category === 'Architecture'),
    philosophies: patterns.filter((p) => p.category === 'Philosophy'),
    designPatterns: patterns.filter((p) => p.category === 'DesignPattern'),
    dataPatterns: patterns.filter((p) => p.category === 'Data'),
    stateManagements: patterns.filter((p) => p.category === 'StateManagement'),
    dataStrategies: patterns.filter((p) => p.category === 'DataStrategy'),
  };
}

export function usePatterns(language: string | null): HookState {
  const [state, setState] = useState<HookState>({
    patterns: EMPTY_GROUPED,
    architectures: [],
    experts: [],
    isLoading: false,
    error: null,
  });

  const loadAll = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      // experts e architectures são independentes da linguagem (catálogo global).
      const [architectures, experts] = await Promise.all([
        fetchArchitectures(),
        fetchExperts(),
      ]);
      // patterns dependem da linguagem selecionada.
      const patternsData = language ? await fetchPatterns(language) : [];
      const grouped = groupPatterns(patternsData);
      setState({
        patterns: grouped,
        architectures,
        experts,
        isLoading: false,
        error: null,
      });
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : String(err),
      }));
    }
  }, [language]);

  useEffect(() => {
    loadAll();
  }, [loadAll]);

  return state;
}
