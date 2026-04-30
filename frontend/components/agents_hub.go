package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strings"
)

// Temas customizados para botões de ação
type ActionTheme struct {
	adaTheme.GhostTheme
	ContrastColor color.Color
}

func (t ActionTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameHover {
		return color.NRGBA{R: 129, G: 140, B: 248, A: 100}
	}
	if n == theme.ColorNameForeground && t.ContrastColor != nil {
		return t.ContrastColor
	}
	return t.GhostTheme.Color(n, v)
}

type DangerTheme struct {
	adaTheme.GhostTheme
	ContrastColor color.Color
}

func (t DangerTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameHover {
		return color.NRGBA{R: 239, G: 68, B: 68, A: 100} // Red-500
	}
	if n == theme.ColorNameForeground && t.ContrastColor != nil {
		return t.ContrastColor
	}
	return t.GhostTheme.Color(n, v)
}

// Tema para garantir contraste no header baseado na cor de fundo
type HeaderContrastTheme struct {
	fyne.Theme
	BackgroundColor color.Color
}

func (t HeaderContrastTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameForeground || n == theme.ColorNameButton || n == theme.ColorNamePrimary {
		return getContrastColor(t.BackgroundColor)
	}
	return t.Theme.Color(n, v)
}

func getContrastColor(bg color.Color) color.Color {
	r, g, b, _ := bg.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	lum := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)
	if lum > 140 {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}
	return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
}

func NewAgentsHub(engine *backend.Engine, onSelect func(*backend.AgentConfig)) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Meus Agentes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	cfg := engine.GetAdaConfig()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Buscar agentes...")

	catOptions := []string{"Todas as Categorias"}
	catOptions = append(catOptions, cfg.AgentCategories...)
	categoryFilter := widget.NewSelect(catOptions, nil)
	categoryFilter.SetSelected("Todas as Categorias")

	// Envolver o select em um container que força a largura mínima
	catSpacer := canvas.NewRectangle(color.Transparent)
	catSpacer.SetMinSize(fyne.NewSize(180, 0))
	categoryFilterWrap := container.NewStack(catSpacer, categoryFilter)

	var grid *fyne.Container
	var refreshGrid func()

	addBtn := adaTheme.NewIconButton(adaTheme.IconAdd, 0, func() {
		showAgentDialog(engine, nil, func(newAgent backend.AgentConfig) {
			cfg := engine.GetAdaConfig()
			cfg.Agents = append(cfg.Agents, newAgent)
			engine.SetAdaConfig(cfg)
			refreshGrid()
		})
	})

	// Layout do Header: [Add] [Title] [Spacer] [Search (Expand)] [Filter (Fixed)]
	header := container.NewBorder(nil, nil,
		container.NewHBox(addBtn, title),
		categoryFilterWrap,
		container.NewPadded(searchEntry),
	)

	grid = container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 200)))

	refreshGrid = func() {
		grid.Objects = nil
		cfg = engine.GetAdaConfig()

		searchText := strings.ToLower(searchEntry.Text)
		selectedCat := categoryFilter.Selected

		for i := range cfg.Agents {
			agent := &cfg.Agents[i]
			idx := i

			if searchText != "" && !strings.Contains(strings.ToLower(agent.Name), searchText) {
				continue
			}
			if selectedCat != "Todas as Categorias" && agent.Category != selectedCat {
				continue
			}

			card := createAgentCard(agent, func() {
				if onSelect != nil {
					onSelect(agent)
				}
			}, func() {
				showAgentDialog(engine, agent, func(updated backend.AgentConfig) {
					cfg.Agents[idx] = updated
					engine.SetAdaConfig(cfg)
					refreshGrid()
				})
			}, func() {
				w := fyne.CurrentApp().Driver().AllWindows()[0]
				msg := fmt.Sprintf("Deseja realmente apagar o agente '%s'?", agent.Name)
				dialog.ShowConfirm("Confirmar Exclusão", msg, func(ok bool) {
					if ok {
						newAgents := append(cfg.Agents[:idx], cfg.Agents[idx+1:]...)
						cfg.Agents = newAgents
						engine.SetAdaConfig(cfg)
						refreshGrid()
					}
				}, w)
			})
			grid.Add(card)
		}
		grid.Refresh()
	}

	searchEntry.OnChanged = func(s string) { refreshGrid() }
	categoryFilter.OnChanged = func(s string) { refreshGrid() }

	refreshGrid()

	return container.NewBorder(
		container.NewVBox(container.NewPadded(header), widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(container.NewPadded(grid)),
	)
}

func parseHexColor(s string) color.Color {
	if s == "" {
		return theme.Color(theme.ColorNameSelection)
	}
	var r, g, b uint8
	fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

func createAgentCard(a *backend.AgentConfig, onSelect func(), onEdit func(), onDelete func()) fyne.CanvasObject {
	categoryName := a.Category
	if categoryName == "" {
		categoryName = "Geral"
	}

	category := widget.NewLabelWithStyle(categoryName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	category.Importance = widget.HighImportance

	// Cabeçalho colorido
	headerColor := parseHexColor(a.Color)
	contrastColor := getContrastColor(headerColor)

	editBtn := adaTheme.NewClickableButton(onEdit)
	editBtn.Text = adaTheme.MenuEditIcon
	editWithTheme := container.NewThemeOverride(editBtn, ActionTheme{ContrastColor: contrastColor})

	delBtn := adaTheme.NewClickableButton(onDelete)
	delBtn.Text = adaTheme.IconDelete
	delWithTheme := container.NewThemeOverride(delBtn, DangerTheme{ContrastColor: contrastColor})

	name := widget.NewLabelWithStyle(a.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	name.Wrapping = fyne.TextWrapWord

	providerInfo := widget.NewLabelWithStyle(a.Model, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	providerInfo.Importance = widget.LowImportance
	providerInfo.Wrapping = fyne.TextWrapWord

	var iconObj fyne.CanvasObject
	if a.Icon != "" {
		iconObj = adaTheme.NewIcon(a.Icon, 32)
	} else {
		iconObj = adaTheme.NewIcon(adaTheme.IconRobot, 32)
	}

	headerBg := canvas.NewRectangle(headerColor)
	headerBg.SetMinSize(fyne.NewSize(200, 32))
	headerBg.CornerRadius = 12 // Cantos superiores

	header := container.NewStack(headerBg,
		container.NewThemeOverride(
			container.NewBorder(nil, nil, nil, container.NewHBox(editWithTheme, delWithTheme), container.NewCenter(category)),
			HeaderContrastTheme{Theme: theme.DefaultTheme(), BackgroundColor: headerColor},
		),
	)

	// Conteúdo principal
	mainBody := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(iconObj),
		name,
		providerInfo,
		layout.NewSpacer(),
	)

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 12

	// Borda sutil
	bg.StrokeColor = theme.Color(theme.ColorNameSeparator)
	bg.StrokeWidth = 0.5

	clickable := adaTheme.NewClickableButton(onSelect)
	ghostBtn := container.NewThemeOverride(clickable, adaTheme.GhostTheme{})

	// O botão de seleção (ghostBtn) deve cobrir apenas a área de conteúdo,
	// deixando o cabeçalho (header) com os botões de ação livre para interação.
	content := container.NewStack(container.NewPadded(mainBody), ghostBtn)

	cardLayout := container.NewBorder(header, nil, nil, nil, content)

	return container.NewStack(bg, cardLayout)
}
