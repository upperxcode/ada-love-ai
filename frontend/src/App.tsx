import { useState } from 'react';
import SettingsPage from './components/SettingsPage';
import { MainLayout } from './components/MainLayout';
import { ThemeProvider } from './lib/theme';
import { SnackbarProvider } from './components/Snackbar';

function AppContent() {
  const [showSettings, setShowSettings] = useState(false);

  return (
    <div className="relative min-h-screen">
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
