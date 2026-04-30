package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Centralização de ícones Unicode (Emojis/Símbolos)
// Isso facilita a manutenção e garante consistência visual em todo o app.
const (
	IconSearch   = "󰍉"
	IconSettings = "󰒓"
	IconDelete   = "󰆴"
	IconInfo     = "󰋽"
	IconView     = "󰈈"
	IconAdd      = "󰐕"
	IconClose    = "󰅖"
	IconCheck    = "󰄬"
	IconStorage  = "󰋊"
	IconDocument = "󰈙"
	IconFolder   = "󰉋"
	IconMail     = "󰇰"
	IconTerminal = "󰞷"
	IconWarning  = "󰀪"
	IconUser     = "󰙄"
	IconRobot    = "󰚩"
	IconStats    = "󰏘"
	IconHistory  = "󰄉"
	IconTools    = "󰓠"
	IconStar     = "󰓎"
	IconCloud    = "󰅟"
)

// Ícones de Menu (específicos para itens de lista e contextos)
const (
	MenuShareIcon   = "🔗"
	MenuPinIcon     = "📌"
	MenuEditIcon    = "✏️"
	MenuProjectIcon = "📁"
	MenuDeleteIcon  = "🗑️"
)

// Constantes de tamanho para padronização da UI
const (
	SizeCardSmall    = 24
	SizeCardBig      = 40
	SizeMenuSmall    = 22
	SizeMenuBig      = 30
	SizeControlSmall = 28
	SizeControlBig   = 24
)

// NewIcon cria um ícone Unicode (canvas.Text) com tamanho personalizado.
// Se size for 0, usa um tamanho padrão (18).
func NewIcon(icon string, size float32) *canvas.Text {
	t := canvas.NewText(icon, TextColor)
	if size > 0 {
		t.TextSize = size
	} else {
		t.TextSize = 18 // Tamanho default para ícones
	}
	return t
}

// NewIconButton cria um botão de ícone customizado com tamanho de ícone garantido e cursor de mão.
func NewIconButton(icon string, size float32, tapped func()) fyne.CanvasObject {
	iconObj := NewIcon(icon, size)
	btn := NewClickableButton(tapped)
	// GhostTheme garante transparência, tamanho correto e cursor de mão
	styledBtn := container.NewThemeOverride(btn, GhostTheme{TextSize: size})
	return container.NewStack(container.NewCenter(iconObj), styledBtn)
}

// NewTextIconButton cria um botão com ícone e texto, cursor de mão e estilo ghost.
func NewTextIconButton(icon, label string, size float32, tapped func()) fyne.CanvasObject {
	btn := NewClickableButton(tapped)
	btn.Text = icon + " " + label
	return container.NewThemeOverride(btn, GhostTheme{TextSize: size})
}

// NewClickableButton cria um botão transparente que mostra o cursor de mão (pointer)
func NewClickableButton(tapped func()) *ClickableButton {
	b := &ClickableButton{}
	b.ExtendBaseWidget(b)
	b.OnTapped = tapped
	b.Importance = widget.LowImportance
	return b
}

// ClickableButton é um botão que mostra o cursor de mão (pointer)
type ClickableButton struct {
	widget.Button
}

func (b *ClickableButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
