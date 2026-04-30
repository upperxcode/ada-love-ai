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

// ChatBubble representa a estrutura visual de uma mensagem
type ChatBubble struct {
	fyne.CanvasObject
	label *widget.Label
}

// UpdateText atualiza o conteúdo da bolha dinamicamente (útil para streaming)
func (b *ChatBubble) UpdateText(text string) {
	b.label.SetText(text)
}

// NewChatBubble cria uma bolha base com alinhamento e cores específicas
func NewChatBubble(text string, isUser bool) *ChatBubble {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord

	// Define a cor de fundo baseado no autor
	bgColor := myTheme.AIMsgColor
	if isUser {
		bgColor = myTheme.UserMsgColor
	}

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 16 // Cantos mais suaves

	// Botão de cópia discreto
	copyBtn := widget.NewButton("󰆏", func() {
		app := fyne.CurrentApp()
		if len(app.Driver().AllWindows()) > 0 {
			win := app.Driver().AllWindows()[0]
			win.Clipboard().SetContent(label.Text)
		}
	})
	copyBtn.Importance = widget.LowImportance

	// Conteúdo interno com preenchimento mais generoso
	content := container.NewPadded(label)

	// Layout para o botão ficar no topo direito para não atrapalhar o texto
	copyOverlay := container.NewHBox(layout.NewSpacer(), container.NewVBox(copyBtn, layout.NewSpacer()))

	bubble := container.NewStack(bg, content, container.NewPadded(copyOverlay))

	// Aplicamos restrições de tamanho (Min: 120px, Max: 750px)
	constrained := container.New(&bubbleLayout{
		label:    label,
		minWidth: 120,
		maxWidth: 750,
	}, bubble)

	var alignment fyne.CanvasObject
	if isUser {
		alignment = container.NewHBox(layout.NewSpacer(), constrained)
	} else {
		alignment = container.NewHBox(constrained, layout.NewSpacer())
	}

	// Adiciona uma margem externa para as bolhas não grudarem nas bordas
	outer := container.NewPadded(alignment)

	return &ChatBubble{
		CanvasObject: outer,
		label:        label,
	}
}

// bubbleLayout impõe limites de largura para as bolhas de chat
type bubbleLayout struct {
	label              *widget.Label
	minWidth, maxWidth float32
}

func (l *bubbleLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, child := range objects {
		child.Resize(size)
	}
}

func (l *bubbleLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(l.minWidth, 0)
	}

	// Mede o texto de forma precisa usando as configurações do tema
	ts := fyne.CurrentApp().Settings().Theme().Size(theme.SizeNameText)
	textSize := fyne.MeasureText(l.label.Text, ts, l.label.TextStyle)

	// Adicionamos padding (aprox 40px horizontais e 20px verticais)
	paddingH := float32(60)
	paddingV := float32(24)

	width := textSize.Width + paddingH

	if width < l.minWidth {
		width = l.minWidth
	}
	if width > l.maxWidth {
		width = l.maxWidth
	}

	// Forçamos o label a assumir a largura disponível (descontando o botão de cópia se necessário)
	// O botão de cópia ocupa aprox 30px
	availableWidth := width - paddingH
	if availableWidth < 20 {
		availableWidth = 20
	}

	l.label.Resize(fyne.NewSize(availableWidth, 0))
	h := l.label.MinSize().Height + paddingV

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
