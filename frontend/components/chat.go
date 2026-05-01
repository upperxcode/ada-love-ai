package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
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

type smartEntry struct {
	widget.Entry
	onSend    func()
	isSending bool
}

func (e *smartEntry) TypedKey(k *fyne.KeyEvent) {
	if k.Name == fyne.KeyReturn {
		// No desktop podemos verificar modificadores
		drv, isDesktop := fyne.CurrentApp().Driver().(desktop.Driver)
		
		isModifierPressed := false
		if isDesktop {
			mods := drv.CurrentKeyModifiers()
			isModifierPressed = (mods&fyne.KeyModifierControl) != 0 || 
			                    (mods&fyne.KeyModifierShift) != 0 || 
								(mods&fyne.KeyModifierSuper) != 0
		}

		// Se pressionar Modificador + Enter, insere nova linha
		if isModifierPressed {
			e.Entry.TypedKey(k)
			return
		}

		// Caso contrário, Enter sozinho envia
		if e.onSend != nil && e.Text != "" {
			e.onSend()
		}
		return
	}
	e.Entry.TypedKey(k)
}

func newSmartEntry(onSend func()) *smartEntry {
	e := &smartEntry{onSend: onSend}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapWord
	e.PlaceHolder = "Digite sua mensagem... (Enter para enviar, Ctrl+Enter para nova linha)"
	e.ExtendBaseWidget(e)
	return e
}

func NewSmartInput(engine *backend.Engine, onSend func(string)) fyne.CanvasObject {
	initializing := true

	var input *smartEntry
	sendFunc := func() {
		if input.Text != "" && !input.isSending {
			input.isSending = true
			onSend(input.Text)
			input.SetText("")
			input.isSending = false
		}
	}

	input = newSmartEntry(sendFunc)
	input.SetPlaceHolder("Envie uma mensagem ou execute uma skill...")

	btnSend := adaTheme.NewIconButton(adaTheme.IconMail, 0, sendFunc)

	btnAttach := adaTheme.NewIconButton(adaTheme.IconDocument, 0, func() {})

	// Configuração dinâmica de Provedores e Modelos
	allModels := engine.GetModelList()
	providers := engine.GetProviders()

	comboModel := widget.NewSelect([]string{}, func(s string) {
		if initializing || s == "" {
			return
		}
		engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
			c.TinyBrain.ModelName = s
		})
		// engine.ReloadAgentLoop() // Removido: UpdateWorkspaceConfig já dispara assincronamente
	})

	comboProvider := widget.NewSelect(providers, func(s string) {
		if initializing || s == "" {
			return
		}
		// Atualiza modelos disponíveis para este provedor
		var models []string
		for _, m := range allModels {
			if m.Provider == s {
				models = append(models, m.ModelName)
			}
		}
		comboModel.Options = models
		if len(models) > 0 {
			comboModel.SetSelected(models[0])
		} else {
			comboModel.SetSelected("")
		}

		engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
			c.TinyBrain.Provider = s
		})
	})

	// Inicializa com os valores atuais do AdaConfig
	currentAda := engine.GetAdaConfig()
	comboProvider.SetSelected(currentAda.TinyBrain.Provider)

	// Popula modelos para o provider atual para permitir o SetSelected inicial
	var initialModels []string
	for _, m := range allModels {
		if m.Provider == currentAda.TinyBrain.Provider {
			initialModels = append(initialModels, m.ModelName)
		}
	}
	comboModel.Options = initialModels
	comboModel.SetSelected(currentAda.TinyBrain.ModelName)

	initializing = false

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
// ThoughtPanel exibe os passos de pensamento e execução da IA
type ThoughtPanel struct {
	fyne.CanvasObject
	accordion *widget.Accordion
	container *fyne.Container
	steps     []string
}

func NewThoughtPanel() *ThoughtPanel {
	content := container.NewVBox()
	
	// Usamos Accordion para permitir que o usuário oculte se quiser
	acc := widget.NewAccordion(
		widget.NewAccordionItem("Pensamento e Ações", content),
	)
	
	tp := &ThoughtPanel{
		CanvasObject: acc,
		accordion:    acc,
		container:    content,
	}
	tp.Hide()
	return tp
}

func (tp *ThoughtPanel) AddStep(text string) {
	icon := adaTheme.NewIcon(adaTheme.IconRobot, 16)
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	
	// Usamos Border para garantir que o Label ocupe o espaço restante e respeite o wrapping
	step := container.NewBorder(nil, nil, icon, nil, label)
	tp.container.Add(container.NewPadded(step))
	tp.container.Refresh()
	tp.Show()
}

func (tp *ThoughtPanel) Clear() {
	tp.container.Objects = nil
	tp.Hide()
}
