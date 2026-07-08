export interface LanguagePlugin {
  name: string;
  engineeringPhilosophies: string[];
  designPatterns: string[];
  dataPatterns?: string[];
}

export interface PluginRegistry {
  [language: string]: LanguagePlugin;
}
