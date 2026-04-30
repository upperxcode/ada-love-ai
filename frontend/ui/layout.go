package ui

import (
	"ada-love-ai/backend"
	"ada-love-ai/frontend/components"
	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func CreateMainLayout(engine *backend.Engine) fyne.CanvasObject {
	// 1. Sidebar com Título Estilizado e Menu de Ícones
	sidebarTitle := components.NewSidebarTitle()
	navMenu := components.NewNavMenu(sidebarTitle)

	// Botão Chevron para Toggle
	toggleBtn := widget.NewButton("‹", nil)
	styledToggle := container.NewThemeOverride(toggleBtn, adaTheme.GhostTheme{})

	sidebar := container.NewStack(
		canvas.NewRectangle(adaTheme.SidebarColor),
		container.NewVBox(
			container.NewHBox(sidebarTitle, layout.NewSpacer(), styledToggle),
			navMenu.Container,
		),
	)

	// 2. Direita: Painel de Detalhes
	detailsPanel := components.NewDetailsPanel()

	// 3. Central: Componentes de Conteúdo
	workspaceDashboard := components.NewWorkspaceDashboard(engine)
	workspaceHub := components.NewWorkspaceHub()
	skillsHub := components.NewSkillsHub(engine)

	var agentsHub fyne.CanvasObject
	var chatController *ChatController

	agentsHub = components.NewAgentsHub(engine, func(a *backend.AgentConfig) {
		if chatController != nil {
			chatController.SetAgent(a)
		}
	})

	msgList := container.NewVBox()
	scroll := ChatScrollContainer(container.NewPadded(container.NewPadded(msgList)))

	// Cabeçalho do Chat
	chatHeader := components.NewChatHeader("Nova Conversa", func() {
		if chatController != nil {
			chatController.NewChat()
		}
	})

	// Inicializa o Controller de Chat (precisa do scroll e msgList)
	chatController = NewChatController(engine, msgList, scroll, chatHeader, detailsPanel)

	smartInput := components.NewSmartInput(chatController.SendMessage)

	// Área de Chat Completa (com Header e Input)
	chatArea := container.NewBorder(
		chatHeader.Container,
		container.NewPadded(container.NewPadded(smartInput)),
		nil,
		nil,
		scroll,
	)

	// Pilha Central para o corpo do conteúdo (Alterna entre ChatArea, Dashboard e Hub)
	centralStack := container.NewStack(chatArea, workspaceDashboard, workspaceHub, agentsHub, skillsHub.CanvasObject)
	workspaceDashboard.Hide()
	workspaceHub.Hide()
	agentsHub.Hide()
	skillsHub.CanvasObject.Hide()

	// Container Principal com Fundo
	centralArea := BackgroundContainer(centralStack, adaTheme.BgColor)

	// Implementa a navegação do menu lateral
	navMenu.OnSelect = func(label string) {
		chatArea.Hide()
		workspaceDashboard.Hide()
		workspaceHub.Hide()
		agentsHub.Hide()
		skillsHub.CanvasObject.Hide()

		switch label {
		case "Workspaces":
			workspaceDashboard.Show()
		case "Chat":
			chatArea.Show()
		case "Agentes":
			agentsHub.Show()
		case "Skills":
			skillsHub.CanvasObject.Show()
		case "Configurações":
			workspaceDashboard.Show()
		default:
			chatArea.Show()
		}
		centralArea.Refresh()
	}

	// 5. Callbacks do painel de detalhes
	detailsPanel.OnChatSelect = func(id string) {
		navMenu.OnSelect("Chat") // Garante que volta pro chat ao selecionar histórico
		chatController.LoadSession(id)
	}

	detailsPanel.OnPin = func(id string) { chatController.PinSession(id) }
	detailsPanel.OnDelete = func(id string) { chatController.DeleteSession(id) }
	detailsPanel.OnRename = func(id string) {
		w := fyne.CurrentApp().Driver().AllWindows()[0]
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Novo título...")
		dialog.ShowForm("Renomear Conversa", "Salvar", "Cancelar", []*widget.FormItem{
			widget.NewFormItem("Título", entry),
		}, func(ok bool) {
			if ok && entry.Text != "" {
				chatController.RenameSession(id, entry.Text)
			}
		}, w)
	}

	// Layout Final: Sidebar | [Central | Detalhes]
	mainContent := container.NewHSplit(centralArea, detailsPanel.CanvasObject)
	mainContent.Offset = 0.8

	split := container.NewHSplit(sidebar, mainContent)
	split.Offset = 0.2

	toggleBtn.OnTapped = func() {
		navMenu.Toggle()
		if navMenu.Collapsed {
			toggleBtn.SetText("›")
			split.Offset = 0.05
		} else {
			toggleBtn.SetText("‹")
			split.Offset = 0.2
		}
		split.Refresh()
	}

	return split
}
