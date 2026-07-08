import { useState } from 'react';
import SettingsPage from './components/SettingsPage';
import { MainLayout } from './components/MainLayout';
import { Button } from './components/ui/button';
import { Icon } from './components/Icon';
import { ThemeProvider, useTheme } from './lib/theme';
import { SnackbarProvider } from './components/Snackbar';

function ThemeToggle() {
  const { dark, setDark } = useTheme();
  return (
    <Button
      variant="ghost"
      size="sm"
      className="fixed top-3 right-3 z-[100] h-8 w-8 p-0"
      onClick={() => setDark(!dark)}
    >
      {dark ? (
        <Icon name="Sun" className="w-full h-full" />
      ) : (
        <Icon name="Moon" className="w-full h-full" />
      )}
    </Button>
  );
}

function AppContent() {
  const [showSettings, setShowSettings] = useState(false);

  return (
    <div className="relative min-h-screen">
      <ThemeToggle />
      {showSettings ? (
        <SettingsPage onClose={() => setShowSettings(false)} />
      ) : (
        <MainLayout onOpenSettings={() => setShowSettings(true)} />
      )}
    </div>
  );
}

function App() {
  return (
    <ThemeProvider>
      <SnackbarProvider>
        <AppContent />
      </SnackbarProvider>
    </ThemeProvider>
  );
}

export default App;
