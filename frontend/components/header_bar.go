package components

import (
	"fmt"
	"image/color"
	adaTheme "ada-love-ai/frontend/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// HeaderBar é um componente base para barras horizontais no topo de painéis
type HeaderBar struct {
	fyne.CanvasObject
	Background *canvas.Rectangle
	Content    *fyne.Container
}

func NewHeaderBar(left, center, right fyne.CanvasObject) *HeaderBar {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))

	// Usamos Border para posicionar os elementos sem preenchimento vertical excessivo
	content := container.NewBorder(nil, nil, left, right, center)

	// Stack do fundo com o conteúdo
	stack := container.NewStack(bg, content)

	return &HeaderBar{
		CanvasObject: stack,
		Background:   bg,
		Content:      content,
	}
}

// ChatHeader representa o cabeçalho da área de chat
type ChatHeader struct {
	Container  *fyne.Container
	Header     *HeaderBar
	TitleLabel  *widget.Label
	AgentLabel  *widget.Label
	StatusLabel *widget.Label

	currentTitle     string
	currentWorkspace string
}

// NewChatHeader cria o cabeçalho específico para a tela de chat
func NewChatHeader(title string, onNewChat func()) *ChatHeader {
	btnNew := adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, onNewChat)

	lblTitle := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	lblAgent := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	lblAgent.Importance = widget.LowImportance

	lblStatus := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	lblStatus.Importance = widget.MediumImportance

	// Adiciona um pequeno recuo para o ícone não ficar colado na borda esquerda
	leftMargin := canvas.NewRectangle(color.Transparent)
	leftMargin.SetMinSize(fyne.NewSize(8, 0))

	leftSide := container.NewHBox(leftMargin, btnNew, lblTitle, widget.NewSeparator(), lblAgent, widget.NewSeparator(), lblStatus)
	rightSide := container.NewHBox()

	header := NewHeaderBar(leftSide, nil, rightSide)
	header.Background.FillColor = theme.Color(theme.ColorNameHeaderBackground)

	c := container.NewStack(header.CanvasObject)

	h := &ChatHeader{
		Container:    c,
		Header:       header,
		TitleLabel:   lblTitle,
		AgentLabel:   lblAgent,
		StatusLabel:  lblStatus,
		currentTitle: title,
	}
	return h
}

func (h *ChatHeader) SetTitle(title string) {
	h.currentTitle = title
	h.refreshTitle()
}

func (h *ChatHeader) SetWorkspaceName(name string) {
	h.currentWorkspace = name
	h.refreshTitle()
}

func (h *ChatHeader) refreshTitle() {
	if h.currentWorkspace != "" {
		h.TitleLabel.SetText(fmt.Sprintf("%s (%s)", h.currentWorkspace, h.currentTitle))
	} else {
		h.TitleLabel.SetText(h.currentTitle)
	}
}

func (h *ChatHeader) SetAgentInfo(name, icon string) {
	if name == "" {
		h.AgentLabel.SetText("")
	} else {
		h.AgentLabel.SetText(fmt.Sprintf("%s %s", icon, name))
	}
}

func (h *ChatHeader) SetStatus(status string) {
	h.StatusLabel.SetText(status)
}
