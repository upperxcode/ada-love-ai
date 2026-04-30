package components

import (
	"fmt"
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
	bg.SetMinSize(fyne.NewSize(0, 32)) // Altura reduzida para 32px

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
	TitleLabel *widget.Label
	AgentLabel *widget.Label
}

// NewChatHeader cria o cabeçalho específico para a tela de chat
func NewChatHeader(title string, onNewChat func()) *ChatHeader {
	btnNew := adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, onNewChat)

	lblTitle := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	lblAgent := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	lblAgent.Importance = widget.LowImportance

	leftSide := container.NewHBox(btnNew, lblTitle, widget.NewSeparator(), lblAgent)
	rightSide := container.NewHBox()

	header := NewHeaderBar(leftSide, nil, rightSide)
	header.Background.FillColor = theme.Color(theme.ColorNameHeaderBackground)

	c := container.NewStack(header.CanvasObject)

	return &ChatHeader{
		Container:  c,
		Header:     header,
		TitleLabel: lblTitle,
		AgentLabel: lblAgent,
	}
}

func (h *ChatHeader) SetTitle(title string) {
	h.TitleLabel.SetText(title)
}

func (h *ChatHeader) SetAgentInfo(name, icon string) {
	if name == "" {
		h.AgentLabel.SetText("")
	} else {
		h.AgentLabel.SetText(fmt.Sprintf("%s %s", icon, name))
	}
}
