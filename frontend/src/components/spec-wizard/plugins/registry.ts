import { PluginRegistry } from './index';

export const plugins: PluginRegistry = {
  'go': {
    name: 'Go',
    engineeringPhilosophies: ['KISS', 'YAGNI'],
    designPatterns: ['Factory', 'Builder', 'Singleton']
  },
  'java': {
    name: 'Java',
    engineeringPhilosophies: ['SOLID', 'DRY'],
    designPatterns: ['Factory', 'Builder', 'Singleton', 'Observer', 'Strategy']
  },
  'python': {
    name: 'Python',
    engineeringPhilosophies: ['DRY', 'KISS'],
    designPatterns: ['Factory', 'Builder', 'Singleton', 'Adapter']
  }
};
