import { useTheme } from '../../lib/theme';

function GeneralSection() {
  const {
    themes,
    currentTheme,
    setTheme,
    iconThemes,
    currentIconTheme,
    setIconTheme,
    iconSets,
    currentIconSet,
    setIconSet,
  } = useTheme();

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-foreground">General</h3>
        <p className="text-sm text-muted-foreground">
          Theme and application settings.
        </p>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Theme</h4>
        <div className="flex flex-wrap gap-3">
          {themes.map((t) => (
            <button
              key={t.id}
              className={`flex flex-col gap-2 p-3 rounded-lg border text-sm transition-colors w-48 ${
                currentTheme === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setTheme(t.id)}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium">{t.name}</span>
                {t.author && (
                  <span className="text-xs text-muted-foreground">
                    {t.author}
                  </span>
                )}
              </div>
              <div className="flex gap-1.5">
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.primary }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.secondary }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.accent }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.muted }}
                />
                <span
                  className="w-5 h-5 rounded-sm"
                  style={{ backgroundColor: t.light?.destructive }}
                />
              </div>
              <div className="flex gap-1.5 mt-1">
                <span
                  className="px-2 py-0.5 text-[10px] rounded-sm text-white"
                  style={{ backgroundColor: t.light?.primary }}
                >
                  Button
                </span>
                <span
                  className="px-2 py-0.5 text-[10px] rounded-sm"
                  style={{
                    backgroundColor: t.light?.secondary,
                    color: t.light?.['secondary-foreground'],
                  }}
                >
                  Button
                </span>
                <span
                  className="px-2 py-0.5 text-[10px] rounded-sm border"
                  style={{
                    borderColor: t.light?.border,
                    color: t.light?.['secondary-foreground'],
                  }}
                >
                  Outline
                </span>
              </div>
            </button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground">
          Install:{' '}
          <code className="bg-muted px-1 rounded">
            npm run theme:add &lt;tweakcn-url&gt; "Theme Name"
          </code>
        </p>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Icon Style</h4>
        <div className="flex flex-wrap gap-2">
          {iconThemes.map((t) => (
            <button
              key={t.id}
              className={`px-3 py-1.5 rounded-lg border text-xs transition-colors ${
                currentIconTheme === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setIconTheme(t.id)}
            >
              {t.name}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium text-foreground">Icon Set</h4>
        <div className="flex flex-wrap gap-2">
          {iconSets.map((t) => (
            <button
              key={t.id}
              className={`px-3 py-1.5 rounded-lg border text-xs transition-colors ${
                currentIconSet === t.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setIconSet(t.id)}
            >
              {t.name}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

export default GeneralSection;
