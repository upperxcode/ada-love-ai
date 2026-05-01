package components

import (
	"ada-love-ai/backend"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ToolsHub struct {
	engine *backend.Engine

	container    *fyne.Container
	cards        *fyne.Container
	search       *widget.Entry
	statusFilter string
}

func NewToolsHub(engine *backend.Engine) *ToolsHub {
	h := &ToolsHub{
		engine:       engine,
		statusFilter: "All Status",
	}

	h.search = widget.NewEntry()
	h.search.SetPlaceHolder("Search tools...")
	h.search.OnChanged = func(s string) {
		h.Refresh()
	}

	h.cards = container.New(layout.NewVBoxLayout())

	statusSelect := widget.NewSelect([]string{"All Status", "Enabled", "Disabled"}, func(s string) {
		h.statusFilter = s
		h.Refresh()
	})
	
	header := container.NewVBox(
		container.NewHBox(
			canvas.NewText("Tool Library", color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		),
		canvas.NewText("Browse and manage the toolset available to your AI agents.", color.NRGBA{R: 150, G: 150, B: 150, A: 255}),
		container.NewPadded(container.NewGridWithColumns(2, h.search, statusSelect)),
	)

	h.container = container.NewBorder(header, nil, nil, nil, container.NewVScroll(h.cards))
	
	// Adiciona um fundo sólido para evitar sobreposições visuais
	bg := canvas.NewRectangle(color.NRGBA{R: 18, G: 18, B: 24, A: 255})
	h.container = container.NewStack(bg, container.NewPadded(h.container))
	
	// Define o valor inicial sem disparar o Refresh precocemente
	statusSelect.SetSelected("All Status")
	h.Refresh()

	return h
}

func (h *ToolsHub) Refresh() {
	if h.cards == nil || h.engine == nil {
		return
	}
	h.cards.Objects = nil
	tools := h.engine.GetAvailableTools()

	// Agrupar por categoria
	categories := make(map[string][]backend.ToolUIInfo)
	var catOrder []string
	for _, t := range tools {
		if _, ok := categories[t.Category]; !ok {
			catOrder = append(catOrder, t.Category)
		}
		categories[t.Category] = append(categories[t.Category], t)
	}

	searchTerm := strings.ToLower(h.search.Text)

	for _, catName := range catOrder {
		catTools := categories[catName]
		
		// Filtrar ferramentas
		var filtered []backend.ToolUIInfo
		for _, t := range catTools {
			// Filtro de busca
			matchSearch := searchTerm == "" || strings.Contains(strings.ToLower(t.Name), searchTerm) || strings.Contains(strings.ToLower(t.Description), searchTerm)
			
			// Filtro de status
			matchStatus := true
			if h.statusFilter == "Enabled" && !t.Enabled {
				matchStatus = false
			} else if h.statusFilter == "Disabled" && t.Enabled {
				matchStatus = false
			}

			if matchSearch && matchStatus {
				filtered = append(filtered, t)
			}
		}

		if len(filtered) == 0 {
			continue
		}

		// Adicionar título da categoria
		catTitle := canvas.NewText(catName, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
		catTitle.TextStyle = fyne.TextStyle{Bold: true}
		h.cards.Add(container.NewPadded(catTitle))

		// Adicionar Grid de cards (2 colunas)
		grid := container.New(layout.NewGridLayout(2))
		for _, t := range filtered {
			grid.Add(h.createToolCard(t))
		}
		h.cards.Add(grid)
	}
	h.cards.Refresh()
}

func (h *ToolsHub) createToolCard(t backend.ToolUIInfo) fyne.CanvasObject {
	nameLabel := canvas.NewText(t.Name, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	nameLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	statusBadge := canvas.NewText("Enabled", color.NRGBA{R: 50, G: 200, B: 100, A: 255})
	if !t.Enabled {
		statusBadge.Text = "Disabled"
		statusBadge.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	}
	statusBadge.TextSize = 10

	descLabel := widget.NewLabel(t.Description)
	descLabel.Wrapping = fyne.TextWrapWord

	sw := widget.NewCheck("", func(enabled bool) {
		h.engine.ToggleTool(t.Name, enabled)
		h.Refresh()
	})
	sw.Checked = t.Enabled

	topRow := container.NewHBox(nameLabel, container.NewPadded(statusBadge), layout.NewSpacer(), sw)
	
	cardContent := container.NewVBox(topRow, descLabel)
	
	// Fundo escuro com bordas arredondadas
	bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 30, A: 255})
	bg.StrokeWidth = 1
	bg.StrokeColor = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	bg.CornerRadius = 8

	return container.NewMax(bg, container.NewPadded(cardContent))
}

func (h *ToolsHub) CanvasObject() fyne.CanvasObject {
	return h.container
}
