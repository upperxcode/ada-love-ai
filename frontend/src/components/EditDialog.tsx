import { ReactNode, useState } from 'react';
import { Dialog, DialogContent } from './ui/dialog';
import { Button } from './ui/button';
import { Icon } from './Icon';
import { Popover, PopoverTrigger, PopoverContent } from './ui/popover';
import { IconPicker } from './IconPicker';

interface EditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: () => void;
  onBack?: () => void;
  onNext?: () => void;
  children: ReactNode;
  color?: string;
  icon?: string;
  title?: string;
  description?: string;
  showNext?: boolean;
  showBack?: boolean;
  onColorChange?: (color: string) => void;
  onIconChange?: (icon: string) => void;
}

const COLOR_PRESETS = [
  '#3b82f6', // blue
  '#10b981', // emerald
  '#f59e0b', // amber
  '#ef4444', // red
  '#8b5cf6', // violet
  '#ec4899', // pink
  '#14b8a6', // teal
  '#f97316', // orange
  '#06b6d4', // cyan
  '#84cc16', // lime
  '#6366f1', // indigo
  '#a855f7', // purple
];

export function EditDialog({
  open,
  onOpenChange,
  onSave,
  onBack,
  onNext,
  children,
  color,
  icon,
  title,
  description,
  showNext,
  showBack,
  onColorChange,
  onIconChange,
}: EditDialogProps) {
  const [currentColor, setCurrentColor] = useState(color || '#3b82f6');
  const [currentIcon, setCurrentIcon] = useState(icon || '📝');
  const [colorPickerOpen, setColorPickerOpen] = useState(false);
  const [iconPickerOpen, setIconPickerOpen] = useState(false);

  const handleColorChange = (newColor: string) => {
    setCurrentColor(newColor);
    onColorChange?.(newColor);
  };

  const handleIconChange = (newIcon: string) => {
    setCurrentIcon(newIcon);
    onIconChange?.(newIcon);
  };

  const handleClose = () => {
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        topAligned
        className="max-w-2xl p-0 gap-0 fixed-height-dialog flex flex-col"
        hideClose
        onInteractOutside={(e: any) => e.preventDefault()}
        onEscapeKeyDown={(e: any) => e.preventDefault()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-3 border-b border-border fixed-header">
          <div className="flex items-center gap-3">
            <span className="text-sm font-semibold text-foreground">
              {title || 'Spec Wizard'}
            </span>
            {description && (
              <span className="text-xs text-muted-foreground">
                {description}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            {/* Color button */}
            <Popover open={colorPickerOpen} onOpenChange={setColorPickerOpen}>
              <PopoverTrigger asChild>
                <button
                  className="inline-flex items-center gap-1.5 h-7 px-2.5 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground text-xs font-medium"
                  title="Choose color"
                >
                  <span
                    className="w-3.5 h-3.5 rounded border border-border"
                    style={{ backgroundColor: currentColor }}
                  />
                  Color
                </button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-3" align="end">
                <div className="space-y-2">
                  <div className="text-xs font-medium text-muted-foreground">
                    Color Presets
                  </div>
                  <div className="grid grid-cols-6 gap-1.5">
                    {COLOR_PRESETS.map((preset) => (
                      <button
                        key={preset}
                        className={`w-6 h-6 rounded border-2 transition-all ${
                          currentColor === preset
                            ? 'border-foreground scale-110'
                            : 'border-transparent hover:scale-110'
                        }`}
                        style={{ backgroundColor: preset }}
                        onClick={() => {
                          handleColorChange(preset);
                          setColorPickerOpen(false);
                        }}
                        title={preset}
                      />
                    ))}
                  </div>
                  <div className="flex items-center gap-2 pt-1 border-t">
                    <input
                      type="color"
                      value={currentColor}
                      onChange={(e) =>
                        handleColorChange(e.target.value)
                      }
                      className="h-7 w-10 rounded border border-input cursor-pointer"
                    />
                    <input
                      type="text"
                      value={currentColor}
                      onChange={(e) => handleColorChange(e.target.value)}
                      className="flex-1 h-7 px-2 rounded border border-input bg-background text-xs font-mono"
                      placeholder="#000000"
                    />
                  </div>
                </div>
              </PopoverContent>
            </Popover>

            {/* Icon button */}
            <Popover open={iconPickerOpen} onOpenChange={setIconPickerOpen}>
              <PopoverTrigger asChild>
                <button
                  className="inline-flex items-center gap-1.5 h-7 px-2.5 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground text-xs font-medium"
                  title="Choose icon"
                >
                  <span className="text-sm leading-none">{currentIcon}</span>
                  Icon
                </button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-1" align="end">
                <div className="grid grid-cols-8 gap-0.5">
                  {[
                    '📂',
                    '🔧',
                    '⚡',
                    '🚀',
                    '🎯',
                    '💻',
                    '🔍',
                    '📝',
                    '🧪',
                    '🛠️',
                    '📦',
                    '🎨',
                    '🔐',
                    '📊',
                    '🗂️',
                    '🌐',
                    '🤖',
                    '🧠',
                    '💡',
                    '🔥',
                    '⭐',
                    '🎮',
                    '📱',
                    '⚙️',
                    '🔑',
                    '📁',
                    '🔨',
                    '📋',
                    '📌',
                    '💎',
                    '📈',
                    '💾',
                    '🖥️',
                    '🔬',
                    '🎵',
                    '📷',
                    '🎬',
                  ].map((emoji) => (
                    <button
                      key={emoji}
                      className={`w-6 h-6 flex items-center justify-center rounded text-sm hover:bg-accent transition-colors ${
                        currentIcon === emoji
                          ? 'bg-accent ring-1 ring-ring'
                          : ''
                      }`}
                      onClick={() => {
                        handleIconChange(emoji);
                        setIconPickerOpen(false);
                      }}
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
              </PopoverContent>
            </Popover>

            <span
              className="text-muted-foreground/30 select-none"
              aria-hidden
            >
              |
            </span>

            {/* Save button */}
            <Button size="sm" className="h-7 gap-1 px-3" onClick={onSave}>
              <Icon name="Save" size={14} /> Save
            </Button>

            {/* Close (x) button */}
            <button
              className="inline-flex items-center justify-center h-7 w-7 rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground"
              onClick={handleClose}
              title="Close"
            >
              <Icon name="X" size={16} />
            </button>
          </div>
        </div>

        {/* Content - scrolls if needed, always aligned to top */}
        <div className="px-6 py-4 flex-1 min-h-0 overflow-y-auto flex flex-col items-stretch justify-start">
          {children}
        </div>

        {/* Footer - navigation */}
        <div className="flex items-center justify-end gap-2 px-6 py-3 border-t border-border bg-background/50">
          {onBack && showBack !== false && (
            <Button variant="ghost" size="sm" onClick={onBack}>
              <Icon name="ArrowLeft" size={16} /> Back
            </Button>
          )}
          {onNext && showNext !== false && (
            <Button size="sm" onClick={onNext}>
              Next <Icon name="ArrowRight" size={16} />
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
