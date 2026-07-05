import SettingsPage from './components/SettingsPage'
import { Button } from './components/ui/button'
import { Icon } from './components/Icon'
import { ThemeProvider, useTheme } from './lib/theme'

function ThemeToggle() {
  const { dark, setDark } = useTheme()
  return (
    <Button
      variant="ghost"
      size="sm"
      className="fixed top-3 right-3 z-[100] h-8 w-8 p-0"
      onClick={() => setDark(!dark)}
    >
      {dark ? <Icon name="Sun" className="w-full h-full" /> : <Icon name="Moon" className="w-full h-full" />}
    </Button>
  )
}

function App() {
  return (
    <ThemeProvider>
      <div className="relative min-h-screen">
        <ThemeToggle />
        <SettingsPage />
      </div>
    </ThemeProvider>
  )
}

export default App
