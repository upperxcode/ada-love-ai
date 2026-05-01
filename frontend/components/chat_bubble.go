package components

import (
	myTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	UserBubbleMaxWidthPercent = 0.45
	AIBubbleMaxWidthPercent   = 0.75
	BubbleMinWidth            = 120.0
	BubblePaddingH            = 60.0
	BubblePaddingV            = 24.0
)

// ChatBubble representa a estrutura visual de uma mensagem
type ChatBubble struct {
	fyne.CanvasObject
	label fyne.CanvasObject
	text  string
}

// UpdateText atualiza o conteúdo da bolha dinamicamente (útil para streaming)
func (b *ChatBubble) UpdateText(text string) {
	b.text = text
	if rt, ok := b.label.(*widget.RichText); ok {
		rt.ParseMarkdown(text)
	} else if lbl, ok := b.label.(*widget.Label); ok {
		lbl.SetText(text)
	}
}

// NewChatBubble cria uma bolha base com alinhamento e cores específicas
func NewChatBubble(text string, isUser bool) *ChatBubble {
	b := &ChatBubble{text: text}

	var display fyne.CanvasObject
	if isUser {
		lbl := widget.NewLabel(text)
		lbl.Wrapping = fyne.TextWrapWord
		display = lbl
	} else {
		rt := widget.NewRichTextFromMarkdown(text)
		rt.Wrapping = fyne.TextWrapWord
		display = rt
	}
	b.label = display

	// Define a cor de fundo baseado no autor
	bgColor := myTheme.AIMsgColor
	if isUser {
		bgColor = myTheme.UserMsgColor
	}

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 16

	// Botão de cópia discreto - agora usa b.text que é atualizado pelo UpdateText
	copyBtn := widget.NewButton("󰆏 ", func() {
		app := fyne.CurrentApp()
		if len(app.Driver().AllWindows()) > 0 {
			win := app.Driver().AllWindows()[0]
			textToCopy := b.text
			if textToCopy == "" {
				// Fallback para o label se b.text estiver vazio
				if lbl, ok := b.label.(*widget.Label); ok {
					textToCopy = lbl.Text
				} else if rt, ok := b.label.(*widget.RichText); ok {
					// Extrai texto bruto do RichText
					for _, seg := range rt.Segments {
						if tseg, ok := seg.(*widget.TextSegment); ok {
							textToCopy += tseg.Text
						}
					}
				}
			}
			win.Clipboard().SetContent(textToCopy)
		}
	})
	copyBtn.Importance = widget.LowImportance

	content := container.NewPadded(display)
	copyOverlay := container.NewHBox(layout.NewSpacer(), container.NewVBox(copyBtn, layout.NewSpacer()))
	bubble := container.NewStack(bg, content, container.NewPadded(copyOverlay))

	// Calculamos o MaxWidth baseado na largura da janela/painel
	percent := AIBubbleMaxWidthPercent
	if isUser {
		percent = UserBubbleMaxWidthPercent
	}

	constrained := container.New(&bubbleLayout{
		content: display,
		isUser:  isUser,
		percent: percent,
	}, bubble)

	var alignment fyne.CanvasObject
	if isUser {
		alignment = container.NewHBox(layout.NewSpacer(), constrained)
	} else {
		alignment = container.NewHBox(constrained, layout.NewSpacer())
	}

	outer := container.NewPadded(alignment)
	b.CanvasObject = outer

	return b
}

// bubbleLayout impõe limites de largura para as bolhas de chat
type bubbleLayout struct {
	content fyne.CanvasObject
	isUser  bool
	percent float64
}

func (l *bubbleLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, child := range objects {
		child.Resize(size)
	}
}

func (l *bubbleLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(BubbleMinWidth, 0)
	}

	// Obtemos a largura total disponível do sistema (janela)
	// Isso é uma aproximação para o painel de chat
	windowWidth := float32(1000) // Valor padrão de segurança
	app := fyne.CurrentApp()
	if len(app.Driver().AllWindows()) > 0 {
		windowWidth = app.Driver().AllWindows()[0].Canvas().Size().Width
	}
	
	// O painel de chat ocupa aprox 70% da janela (descontando sidebar)
	chatPanelWidth := windowWidth * 0.7 
	maxWidth := chatPanelWidth * float32(l.percent)

	// Mede o texto de forma precisa para evitar "atrofia"
	var textSize fyne.Size
	ts := fyne.CurrentApp().Settings().Theme().Size(theme.SizeNameText)
	
	if lbl, ok := l.content.(*widget.Label); ok {
		textSize = fyne.MeasureText(lbl.Text, ts, lbl.TextStyle)
	} else if rt, ok := l.content.(*widget.RichText); ok {
		// Para evitar atrofia no RichText, medimos o texto bruto
		rawText := ""
		for _, seg := range rt.Segments {
			if tseg, ok := seg.(*widget.TextSegment); ok {
				rawText += tseg.Text
			}
		}
		textSize = fyne.MeasureText(rawText, ts, fyne.TextStyle{})
	}

	width := textSize.Width + BubblePaddingH
	if width < BubbleMinWidth {
		width = BubbleMinWidth
	}
	if width > maxWidth {
		width = maxWidth
	}

	// Forçamos o conteúdo a assumir a largura disponível
	availableWidth := width - BubblePaddingH
	if availableWidth < 20 {
		availableWidth = 20
	}

	l.content.Resize(fyne.NewSize(availableWidth, 0))
	h := l.content.MinSize().Height + BubblePaddingV

	return fyne.NewSize(width, h)
}

// NewUserBubble estende o ChatBubble para mensagens do usuário (Perguntas)
func NewUserBubble(text string) *ChatBubble {
	return NewChatBubble(text, true)
}

// NewAIBubble estende o ChatBubble para mensagens da inteligência (Respostas)
func NewAIBubble(text string) *ChatBubble {
	return NewChatBubble(text, false)
}
