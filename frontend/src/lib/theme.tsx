import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import registry from '../themes/registry.json'
import iconRegistry from '../themes/icon-themes.json'

export interface ThemeDef {
  id: string
  name: string
  author?: string
  url?: string
  light: Record<string, string>
  dark: Record<string, string>
}

export interface IconThemeDef {
  id: string
  name: string
  vars: Record<string, string>
}

export interface IconSetDef {
  id: string
  name: string
}

export const iconSetMeta: IconSetDef[] = [
  { id: "modern", name: "Lucide (Modern)" },
  { id: "classic", name: "FontAwesome (Classic)" },
  { id: "minimal", name: "Heroicons (Minimal)" },
  { id: "material", name: "Material Design" },
  { id: "rounded", name: "Remix Rounded" },
]

interface ThemeContextType {
  themes: ThemeDef[]
  currentTheme: string
  setTheme: (id: string) => void
  dark: boolean
  setDark: (v: boolean) => void
  iconThemes: IconThemeDef[]
  currentIconTheme: string
  setIconTheme: (id: string) => void
  iconSets: IconSetDef[]
  currentIconSet: string
  setIconSet: (v: string) => void
}

const ThemeContext = createContext<ThemeContextType>({
  themes: [],
  currentTheme: 'default',
  setTheme: () => {},
  dark: true,
  setDark: () => {},
  iconThemes: [],
  currentIconTheme: 'default',
  setIconTheme: () => {},
  currentIconSet: 'modern',
  setIconSet: () => {},
  iconSets: [],
})

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [themes] = useState<ThemeDef[]>(registry as ThemeDef[])
  const [iconThemes] = useState<IconThemeDef[]>(iconRegistry as IconThemeDef[])
  const [iconSets] = useState<IconSetDef[]>(iconSetMeta as IconSetDef[])
  const [currentTheme, setCurrentThemeState] = useState(() =>
    localStorage.getItem('ada-theme') || 'default'
  )
  const [dark, setDarkState] = useState(() =>
    localStorage.getItem('ada-dark') !== 'false'
  )
  const [currentIconTheme, setIconThemeState] = useState(() =>
    localStorage.getItem('ada-icon-theme') || 'default'
  )
  const [currentIconSet, setIconSetState] = useState<string>(() =>
    localStorage.getItem('ada-icon-set') || 'modern'
  )

  const applyVars = useCallback((id: string, isDark: boolean) => {
    const theme = registry.find((t: any) => t.id === id)
    if (!theme) return
    const vars = isDark ? theme.dark : theme.light
    const root = document.documentElement
    for (const [key, value] of Object.entries(vars)) {
      root.style.setProperty(`--${key}`, value)
    }
    root.classList.toggle('dark', isDark)
    root.setAttribute('data-theme', id)
  }, [])

  const applyIconVars = useCallback((id: string) => {
    const theme = iconRegistry.find((t: any) => t.id === id)
    if (!theme) return
    const root = document.documentElement
    for (const [key, value] of Object.entries(theme.vars)) {
      root.style.setProperty(`--${key}`, value)
    }
  }, [])

  const setTheme = useCallback((id: string) => {
    setCurrentThemeState(id)
    localStorage.setItem('ada-theme', id)
    applyVars(id, dark)
  }, [dark, applyVars])

  const setDark = useCallback((v: boolean) => {
    setDarkState(v)
    localStorage.setItem('ada-dark', String(v))
    applyVars(currentTheme, v)
  }, [currentTheme, applyVars])

  const setIconTheme = useCallback((id: string) => {
    setIconThemeState(id)
    localStorage.setItem('ada-icon-theme', id)
    applyIconVars(id)
  }, [applyIconVars])

  const setIconSet = useCallback((v: string) => {
    setIconSetState(v)
    localStorage.setItem('ada-icon-set', v)
  }, [])

  useEffect(() => {
    applyVars(currentTheme, dark)
    applyIconVars(currentIconTheme)
  }, [])

  return (
    <ThemeContext.Provider value={{
      themes, currentTheme, setTheme, dark, setDark,
      iconThemes, currentIconTheme, setIconTheme,
      iconSets, currentIconSet, setIconSet,
    }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  return useContext(ThemeContext)
}
