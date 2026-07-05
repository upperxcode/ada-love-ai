import { useState } from 'react'
import { Popover, PopoverTrigger, PopoverContent } from './ui/popover'
import { Button } from './ui/button'

const ICONS = ['📂','🔧','⚡','🚀','🎯','💻','🔍','📝','🧪','🛠️','📦','🎨','🔐','📊','🗂️','🌐','🤖','🧠','💡','🔥','⭐','🎮','📱','⚙️','🔑','📁','🔨','📋','📌','💎','🎯','📈','💾','🖥️','🔬','🎵','📷','🎬']

interface IconPickerProps {
  value: string
  onChange: (v: string) => void
}

export function IconPicker({ value, onChange }: IconPickerProps) {
  const [open, setOpen] = useState(false)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" className="h-8 w-10 p-0 text-lg">{value}</Button>
      </PopoverTrigger>
      <PopoverContent className="w-56 p-1" align="start">
        <div className="grid grid-cols-8 gap-0.5">
          {ICONS.map(icon => (
            <button
              key={icon}
              className={`w-6 h-6 flex items-center justify-center rounded text-sm hover:bg-accent ${value === icon ? 'bg-accent' : ''}`}
              onClick={() => { onChange(icon); setOpen(false) }}
            >
              {icon}
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}
