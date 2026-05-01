package components

import (
	"image/color"

	adaTheme "ada-love-ai/frontend/theme"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// tappableBG é um fundo reativo a cliques que não interfere nos filhos
type tappableBG struct {
	widget.BaseWidget
	onTapped func()
	fill     color.Color
	radius   float32
}

func newTappableBG(onTapped func()) *tappableBG {
	t := &tappableBG{onTapped: onTapped, fill: color.Transparent, radius: 8}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableBG) Tapped(_ *fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func (t *tappableBG) TappedSecondary(_ *fyne.PointEvent) {}

func (t *tappableBG) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(t.fill)
	r.CornerRadius = t.radius
	return widget.NewSimpleRenderer(r)
}

// tightLayout é um layout que empilha dois elementos verticalmente sem nenhum espaçamento (padding)
type tightLayout struct{}

func (t *tightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		for _, o := range objects {
			if o != nil {
				o.Resize(size)
				o.Move(fyne.NewPos(0, 0))
			}
		}
		return
	}
	header := objects[0]
	body := objects[1]

	if header == nil || body == nil {
		return
	}

	headerHeight := header.MinSize().Height
	header.Move(fyne.NewPos(0, 0))
	header.Resize(fyne.NewSize(size.Width, headerHeight))

	body.Move(fyne.NewPos(0, headerHeight))
	body.Resize(fyne.NewSize(size.Width, fyne.Max(0, size.Height-headerHeight)))
}

func (t *tightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		if len(objects) == 1 && objects[0] != nil {
			return objects[0].MinSize()
		}
		return fyne.NewSize(0, 0)
	}
	h := objects[0]
	b := objects[1]
	if h == nil || b == nil {
		return fyne.NewSize(0, 0)
	}

	hMin := h.MinSize()
	bMin := b.MinSize()
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
	content     fyne.CanvasObject
	titleLabel  *widget.Label
	subLabel    *widget.Label
	deleteBtn   fyne.CanvasObject
	hoverBg     *canvas.Rectangle
	activeBar   *canvas.Rectangle
	selected    bool
	onTap       func()
	bg          *tappableBG
}

func newHoverableRow(content fyne.CanvasObject, deleteBtn fyne.CanvasObject, onTap func()) *hoverableRow {
	bg := newTappableBG(onTap)
	bg.radius = 8

	bar := canvas.NewRectangle(color.Transparent)
	bar.CornerRadius = 2

	h := &hoverableRow{
		content:   content,
		deleteBtn: deleteBtn,
		hoverBg:   canvas.NewRectangle(color.Transparent), // Somente visual
		activeBar: bar,
		onTap:     onTap,
		bg:        bg,
	}
	h.hoverBg.CornerRadius = 8

	// Tenta extrair os labels para controle de cor/importância (recursivo simples)
	var findLabels func(fyne.CanvasObject)
	findLabels = func(obj fyne.CanvasObject) {
		if lbl, ok := obj.(*widget.Label); ok {
			if h.titleLabel == nil {
				h.titleLabel = lbl
			} else {
				h.subLabel = lbl
			}
			return
		}
		if container, ok := obj.(*fyne.Container); ok {
			for _, child := range container.Objects {
				findLabels(child)
			}
		}
	}
	findLabels(content)

	h.ExtendBaseWidget(h)
	deleteBtn.Hide()
	return h
}

func (h *hoverableRow) MinSize() fyne.Size {
	return h.BaseWidget.MinSize()
}

func (h *hoverableRow) CreateRenderer() fyne.WidgetRenderer {
	// A barra ativa fica na esquerda
	h.activeBar.Resize(fyne.NewSize(4, 0)) // Largura de 4px

	return widget.NewSimpleRenderer(container.NewStack(
		h.hoverBg,
		h.bg, // O sensor de clique fica aqui
		container.NewBorder(nil, nil, h.activeBar, h.deleteBtn, h.content),
	))
}

func (h *hoverableRow) MouseIn(*desktop.MouseEvent) {
	h.hoverBg.FillColor = theme.Color(theme.ColorNameHover)
	h.deleteBtn.Show()
	h.Refresh()
}

func (h *hoverableRow) MouseOut() {
	if !h.selected {
		h.hoverBg.FillColor = color.Transparent
		h.deleteBtn.Hide()
	}
	h.Refresh()
}

func (h *hoverableRow) Tapped(_ *fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

func (h *hoverableRow) SetSelected(selected bool) {
	h.selected = selected
	if h.selected {
		// Destaque de seleção para edição (cor de fundo mais forte)
		accent := theme.Color(theme.ColorNamePrimary)
		h.activeBar.FillColor = accent

		r, g, b, _ := accent.RGBA()
		h.hoverBg.FillColor = color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 45} // Aumento de opacidade

		if h.titleLabel != nil {
			h.titleLabel.Importance = widget.HighImportance
		}
		if h.subLabel != nil {
			h.subLabel.Importance = widget.MediumImportance
		}

		h.deleteBtn.Show()
	} else {
		h.activeBar.FillColor = color.Transparent
		h.hoverBg.FillColor = color.Transparent

		if h.titleLabel != nil {
			h.titleLabel.Importance = widget.MediumImportance
		}
		if h.subLabel != nil {
			h.subLabel.Importance = widget.LowImportance
		}

		h.deleteBtn.Hide()
	}
	h.Refresh()
}

type ConfigItem struct {
	Title      string
	Subtitle   string
	Value      string // Valor original/físico
	IsChecked  bool   // Indica se é o item PADRÃO GLOBAL (ícone de check)
	IsSelected bool   // Indica se é o item SELECIONADO PARA EDIÇÃO (highlight)
}

// ConfigSection define uma seção de configuração com card e lista
type ConfigSection struct {
	Title         string
	Items         []ConfigItem
	HeaderActions []fyne.CanvasObject
	OnDel         func(int)
	OnEdit        func(int, ConfigItem)
	OnSelect      func(int)
	OnCheck       func(int) // Para tornar o workspace ativo
}

func (s ConfigSection) Render() fyne.CanvasObject {
	list := container.New(&listLayout{spacing: 8})
	var rows []*hoverableRow

	for i, item := range s.Items {
		idx := i
		it := item

		var rowContent fyne.CanvasObject
		titleLbl := widget.NewLabel(it.Title)
		titleLbl.Truncation = fyne.TextTruncateEllipsis

		if it.Subtitle != "" {
			// Remove quebras de linha para evitar o "fantasma" expandindo o card
			cleanSubtitle := strings.ReplaceAll(it.Subtitle, "\n", " ")
			subLbl := widget.NewLabelWithStyle(cleanSubtitle, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
			subLbl.Truncation = fyne.TextTruncateEllipsis
			rowContent = container.NewVBox(titleLbl, subLbl)
		} else {
			rowContent = titleLbl
		}

		// Adiciona ícone de check se estiver marcado, usando Border para não esmagar o texto
		if it.IsChecked {
			rowContent = container.NewBorder(nil, nil, adaTheme.NewIcon(adaTheme.IconCheck, adaTheme.SizeIconTiny), nil, rowContent)
		}

		var actions fyne.CanvasObject
		if !it.IsChecked {
			// Apenas workspaces que NÃO são o padrão podem ser gerenciados
			delBtn := adaTheme.NewIconButton(adaTheme.IconDelete, adaTheme.SizeControlSmall, func() {
				if s.OnDel != nil {
					s.OnDel(idx)
				}
			})
			checkBtn := adaTheme.NewIconButton(adaTheme.IconCheck, adaTheme.SizeControlSmall, func() {
				if s.OnCheck != nil {
					s.OnCheck(idx)
				}
			})
			actions = container.NewHBox(checkBtn, delBtn)
		} else {
			// O workspace padrão é protegido: não pode ser deletado nem "tornado padrão" (já é)
			// Espaçador para manter o alinhamento do layout
			actions = layout.NewSpacer()
		}

		var row *hoverableRow
		row = newHoverableRow(rowContent, actions, func() {
			// Desmarcar todos os outros
			for _, r := range rows {
				r.SetSelected(false)
			}
			row.SetSelected(true)

			if s.OnEdit != nil {
				s.OnEdit(idx, it)
			}
			if s.OnSelect != nil {
				s.OnSelect(idx)
			}
		})
		row.SetSelected(it.IsSelected)
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
