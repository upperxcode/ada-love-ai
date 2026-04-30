package components

import (
	"image/color"

	adaTheme "ada-love-ai/frontend/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// tightLayout é um layout que empilha dois elementos verticalmente sem nenhum espaçamento (padding)
type tightLayout struct{}

func (t *tightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	header := objects[0]
	body := objects[1]

	headerHeight := header.MinSize().Height
	header.Move(fyne.NewPos(0, 0))
	header.Resize(fyne.NewSize(size.Width, headerHeight))

	body.Move(fyne.NewPos(0, headerHeight))
	body.Resize(fyne.NewSize(size.Width, fyne.Max(0, size.Height-headerHeight)))
}

func (t *tightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}
	hMin := objects[0].MinSize()
	bMin := objects[1].MinSize()
	return fyne.NewSize(fyne.Max(hMin.Width, bMin.Width), hMin.Height+bMin.Height)
}

// listLayout é um layout de lista vertical com espaçamento customizável
type listLayout struct {
	spacing float32
}

func (l *listLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		childHeight := child.MinSize().Height
		child.Move(fyne.NewPos(0, y))
		child.Resize(fyne.NewSize(size.Width, childHeight))
		y += childHeight + l.spacing
	}
}

func (l *listLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	visibleCount := 0
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		ms := child.MinSize()
		w = fyne.Max(w, ms.Width)
		h += ms.Height
		visibleCount++
	}
	if visibleCount > 1 {
		h += float32(visibleCount-1) * l.spacing
	}
	return fyne.NewSize(w, h)
}

// hoverableRow é um componente que mostra o botão de deletar e highlight apenas no hover ou seleção
type hoverableRow struct {
	widget.BaseWidget
	content   fyne.CanvasObject
	deleteBtn fyne.CanvasObject
	hoverBg   *canvas.Rectangle
	selected  bool
	onTap     func()
}

func newHoverableRow(content fyne.CanvasObject, deleteBtn fyne.CanvasObject, onTap func()) *hoverableRow {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = 4
	h := &hoverableRow{content: content, deleteBtn: deleteBtn, hoverBg: bg, onTap: onTap}
	h.ExtendBaseWidget(h)
	deleteBtn.Hide()
	return h
}

func (h *hoverableRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(
		h.hoverBg,
		container.NewBorder(nil, nil, canvas.NewRectangle(color.Transparent), h.deleteBtn, h.content),
	))
}

func (h *hoverableRow) MouseIn(*desktop.MouseEvent) {
	if !h.selected {
		h.hoverBg.FillColor = theme.Color(theme.ColorNameHover)
		h.deleteBtn.Show()
		h.Refresh()
	}
}

func (h *hoverableRow) MouseOut() {
	if !h.selected {
		h.hoverBg.FillColor = color.Transparent
		h.deleteBtn.Hide()
		h.Refresh()
	}
}

func (h *hoverableRow) Tapped(*fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

func (h *hoverableRow) SetSelected(selected bool) {
	h.selected = selected
	if h.selected {
		h.hoverBg.FillColor = theme.Color(theme.ColorNameSelection)
		h.deleteBtn.Show()
	} else {
		h.hoverBg.FillColor = color.Transparent
		h.deleteBtn.Hide()
	}
	h.Refresh()
}

// ConfigSection define uma seção de configuração com card e lista
type ConfigSection struct {
	Title         string
	Items         []string
	HeaderActions []fyne.CanvasObject
	OnDel         func(int)
	OnEdit        func(int, string)
}

func (s ConfigSection) Render() fyne.CanvasObject {
	list := container.New(&listLayout{spacing: 2})
	var rows []*hoverableRow

	for i, item := range s.Items {
		idx := i
		txt := widget.NewLabel(item)
		txt.Truncation = fyne.TextTruncateEllipsis

		delBtn := adaTheme.NewIconButton(adaTheme.IconDelete, adaTheme.SizeControlSmall, func() { s.OnDel(idx) })

		var row *hoverableRow
		row = newHoverableRow(txt, delBtn, func() {
			// Desmarcar todos os outros
			for _, r := range rows {
				r.SetSelected(false)
			}
			row.SetSelected(true)

			if s.OnEdit != nil {
				s.OnEdit(idx, item)
			}
		})
		rows = append(rows, row)
		list.Add(row)
	}

	var actions fyne.CanvasObject
	if len(s.HeaderActions) > 0 {
		actions = container.NewHBox(s.HeaderActions...)
	}

	return createConfigCard(s.Title, list, actions)
}

// createConfigCard cria um container padronizado para itens de configuração com Header escuro e Borda
func createConfigCard(title string, content fyne.CanvasObject, actions fyne.CanvasObject) fyne.CanvasObject {
	titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	headerButtons := actions
	if headerButtons == nil {
		headerButtons = layout.NewSpacer()
	}

	headerBg := canvas.NewRectangle(theme.Color(theme.ColorNameHeaderBackground))
	headerBg.CornerRadius = 12
	headerContent := container.NewPadded(container.NewBorder(nil, nil, nil, headerButtons, titleLbl))
	header := container.NewStack(headerBg, headerContent)

	bodyBg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bodyBg.CornerRadius = 12
	body := container.NewStack(bodyBg, container.NewPadded(content))

	// Overlay de borda para dar o efeito de card único
	border := canvas.NewRectangle(color.Transparent)
	border.CornerRadius = 12
	border.StrokeWidth = 1
	border.StrokeColor = theme.Color(theme.ColorNameSeparator)

	mainBox := container.New(&tightLayout{}, header, body)
	return container.NewStack(border, mainBox)
}
