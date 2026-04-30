package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Use sempre iniciais MAIÚSCULAS para exportar variáveis entre pacotes
var (
	BgColor      = color.NRGBA{R: 11, G: 12, B: 16, A: 255}    // Deepest Midnight
	SidebarColor = color.NRGBA{R: 18, G: 20, B: 29, A: 255}    // Slightly lighter surface
	AccentColor  = color.NRGBA{R: 129, G: 140, B: 248, A: 255} // Modern Indigo
	UserMsgColor = color.NRGBA{R: 30, G: 41, B: 59, A: 255}    // Slate 800
	AIMsgColor   = color.NRGBA{R: 15, G: 23, B: 42, A: 255}    // Slate 900
	TextColor    = color.NRGBA{R: 248, G: 250, B: 252, A: 245} // Slate 50 with high alpha
)

type MyTheme struct{}

// Obrigatório: Define as cores customizadas

func (m MyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return BgColor
	case theme.ColorNameInputBackground:
		return SidebarColor // Usar a cor de superfície para inputs
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 24, G: 27, B: 38, A: 255} // Tom intermediário
	case theme.ColorNamePrimary:
		return AccentColor
	case theme.ColorNameForeground:
		return TextColor
	case theme.ColorNameButton:
		return UserMsgColor
	case theme.ColorNameHover:
		return color.NRGBA{R: 129, G: 140, B: 248, A: 30} // Accent com transparência
	case theme.ColorNameSelection:
		return color.NRGBA{R: 129, G: 140, B: 248, A: 60}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 20}
	}
	return theme.DefaultTheme().Color(n, v)
}

// Obrigatório: Retornar o padrão para evitar erro de compilação
func (m MyTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (m MyTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (m MyTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }

// GhostTheme é um tema que torna o fundo dos botões transparente
type GhostTheme struct {
	MyTheme
	TextSize float32
}

func (g GhostTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameButton {
		return color.Transparent
	}
	if n == theme.ColorNameHover {
		return color.Transparent
	}
	if n == theme.ColorNamePressed {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 50} // Feedback de clique
	}
	return g.MyTheme.Color(n, v)
}

func (g GhostTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText && g.TextSize > 0 {
		return g.TextSize
	}
	return g.MyTheme.Size(n)
}
