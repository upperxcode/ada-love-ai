package components

import (
	adaTheme "ada-love-ai/frontend/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ChatMessage struct {
	*ChatBubble
}

func NewChatMessage(text string, isUser bool) *ChatMessage {
	var bubble *ChatBubble
	if isUser {
		bubble = NewUserBubble(text)
	} else {
		bubble = NewAIBubble(text)
	}

	return &ChatMessage{
		ChatBubble: bubble,
	}
}

func NewSmartInput(onSend func(string)) fyne.CanvasObject {
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Envie uma mensagem ou execute uma skill...")
	input.Wrapping = fyne.TextWrapWord
	input.OnChanged = func(s string) {
		// O widget MultiLineEntry já expande automaticamente no Fyne
		// quando colocado em um container que permite (como VBox)
	}

	btnSend := adaTheme.NewIconButton(adaTheme.IconMail, 0, func() {
		if input.Text != "" {
			onSend(input.Text)
			input.SetText("")
		}
	})

	btnAttach := adaTheme.NewIconButton(adaTheme.IconDocument, 0, func() {})
	comboProvider := widget.NewSelect([]string{"Ollama", "OpenAI", "Anthropic"}, func(s string) {})
	comboProvider.SetSelected("Ollama")

	comboModel := widget.NewSelect([]string{"qwen2.5-coder", "deepseek-v3"}, func(s string) {})
	comboModel.SetSelected("qwen2.5-coder")

	checkPlan := widget.NewCheck("Plan", func(b bool) {})
	checkPlan.SetChecked(true)

	toolBar := container.NewHBox(
		btnAttach,
		container.NewPadded(comboProvider),
		container.NewPadded(comboModel),
		checkPlan,
		layout.NewSpacer(),
		btnSend,
	)

	// Unificando em um bloco visual único com cantos mais modernos
	bg := canvas.NewRectangle(adaTheme.AIMsgColor) // Usar a mesma cor das bolhas da IA para harmonia
	bg.CornerRadius = 16
	bg.StrokeColor = adaTheme.AccentColor
	bg.StrokeWidth = 0.5 // Borda sutil

	content := container.NewPadded(container.NewVBox(
		input,
		widget.NewSeparator(),
		toolBar,
	))

	return container.NewStack(bg, content)
}
