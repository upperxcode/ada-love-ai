package ui

import (
	"ada-love-ai/backend"
	"ada-love-ai/frontend/components"
	adaTheme "ada-love-ai/frontend/theme"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func CreateMainLayout(engine *backend.Engine) fyne.CanvasObject {
	// 1. Sidebar com Título Estilizado e Menu de Ícones
	sidebarTitle := components.NewSidebarTitle()
	navMenu := components.NewNavMenu(sidebarTitle)

	// 2. Painel de Detalhes (agora integrado na esquerda)
	fmt.Println("[Layout] Criando DetailsPanel...")
	detailsPanel := components.NewDetailsPanel()

	// O histórico deve ficar em um container que possa crescer
	historyContainer := container.NewStack(detailsPanel.CanvasObject)

	// Container para o cabeçalho dinâmico (ChatHeader, etc)
	headerContainer := container.NewStack()

	// Barra Superior (Estilo YouTube) - Agora com duas colunas e largura alinhada
	logoArea := container.NewStack(
		adaTheme.NewClickableButton(func() {
			navMenu.Toggle()
		}),
		sidebarTitle,
	)
	// Força uma largura mínima para alinhar com o sidebar expandido
	logoBg := canvas.NewRectangle(color.Transparent)
	logoBg.SetMinSize(fyne.NewSize(278, 0))

	logoAreaContainer := container.NewStack(
		logoBg,
		container.NewPadded(logoArea),
	)

	topBar := container.NewStack(
		canvas.NewRectangle(adaTheme.SidebarColor),
		container.NewBorder(
			nil, nil,
			logoAreaContainer,
			nil,
			headerContainer, // O cabeçalho da página fica aqui
		),
	)

	sidebar := container.NewStack(
		canvas.NewRectangle(adaTheme.SidebarColor),
		container.NewBorder(
			navMenu.Container,
			nil, nil, nil,
			historyContainer,
		),
	)

	// Sincroniza a visibilidade do histórico e o offset do split
	var split *container.Split
	navMenu.OnToggle = func(collapsed bool) {
		if collapsed {
			historyContainer.Hide()
			if split != nil {
				split.Offset = 0.05
			}
		} else {
			historyContainer.Show()
			if split != nil {
				split.Offset = 0.2
			}
		}
		sidebar.Refresh()
		if split != nil {
			split.Refresh()
		}
	}

	// 3. Central: Componentes de Conteúdo
	var workspaceDashboard fyne.CanvasObject
	currentEditIndex := engine.GetAdaConfig().ActiveWorkspaceIndex

	var refreshDashboard func()
	var updateDashboard func(int)

	updateDashboard = func(i int) {
		currentEditIndex = i
		refreshDashboard()
	}

	fmt.Println("[Layout] Criando WorkspaceDashboard...")
	workspaceDashboard = components.NewWorkspaceDashboard(engine, currentEditIndex, updateDashboard, func() {
		if refreshDashboard != nil {
			refreshDashboard()
		}
	})
	fmt.Println("[Layout] Criando WorkspaceHub...")
	workspaceHub := components.NewWorkspaceHub()
	fmt.Println("[Layout] Criando SkillsHub...")
	skillsHub := components.NewSkillsHub(engine)
	fmt.Println("[Layout] Criando ToolsLibrary...")
	toolsHub := components.NewToolsHub(engine)

	var agentsHub fyne.CanvasObject
	var settingsView fyne.CanvasObject
	var chatController *ChatController

	fmt.Println("[Layout] Criando AgentsHub...")
	agentsHub = components.NewAgentsHub(engine, func(a *backend.AgentConfig) {
		if chatController != nil {
			chatController.SetAgent(a)
		}
	})

	fmt.Println("[Layout] Criando SettingsView...")
	settingsView = components.NewSettingsView(engine, func() {
		if refreshDashboard != nil {
			refreshDashboard()
		}
	})

	fmt.Println("[Layout] Criando Chat UI Components...")
	msgList := container.NewVBox()
	scroll := ChatScrollContainer(container.NewPadded(container.NewPadded(msgList)))

	// Cabeçalho do Chat
	fmt.Println("[Layout] Criando ChatHeader...")
	chatHeader := components.NewChatHeader("Nova Conversa", func() {
		if chatController != nil {
			chatController.NewChat()
		}
	})
	fmt.Println("[Layout] ChatHeader OK")

	// Inicializa o Controller de Chat
	fmt.Println("[Layout] Criando ChatController...")
	chatController = NewChatController(engine, msgList, scroll, chatHeader, detailsPanel)
	fmt.Println("[Layout] ChatController OK")

	fmt.Println("[Layout] Criando SmartInput...")
	smartInput := components.NewSmartInput(engine, chatController.SendMessage)
	fmt.Println("[Layout] SmartInput OK")

	// Área de Chat Completa (SEM cabeçalho interno agora)
	chatArea := container.NewBorder(
		nil, // Cabeçalho movido para o topBar
		container.NewPadded(container.NewPadded(smartInput)),
		nil, nil,
		scroll,
	)

	// Pilha central de telas
	centralStack := container.NewStack(
		chatArea,
		workspaceDashboard,
		workspaceHub,
		agentsHub,
		skillsHub.CanvasObject,
		toolsHub.CanvasObject(),
		settingsView,
	)
	chatArea.Show()
	workspaceDashboard.Hide()
	workspaceHub.Hide()
	agentsHub.Hide()
	skillsHub.CanvasObject.Hide()
	toolsHub.CanvasObject().Hide()
	settingsView.Hide()

	centralArea := BackgroundContainer(centralStack, adaTheme.BgColor)
	fmt.Println("[Layout] CentralArea OK")

	var currentView string = "Chat"

	// Agora definimos a implementação real do refresh
	refreshDashboard = func() {
		cfg := engine.GetAdaConfig()

		// Sincroniza o cabeçalho do chat com o workspace ativo
		if cfg.ActiveWorkspaceIndex < len(cfg.Workspaces) {
			activeWS := cfg.Workspaces[cfg.ActiveWorkspaceIndex]
			if chatController != nil {
				chatController.SetWorkspaceName(activeWS.Title)
				// Se mudamos o workspace ativo, precisamos atualizar a lista de sessões
				chatController.refreshSidebar("")
			}
		}

		// Reconstrói o Dashboard
		workspaceDashboard = components.NewWorkspaceDashboard(engine, currentEditIndex, updateDashboard, refreshDashboard)
		centralStack.Objects[1] = workspaceDashboard

		// Reconstrói as Configurações para refletir mudanças (ex: modelos apagados)
		settingsView = components.NewSettingsView(engine, refreshDashboard)
		centralStack.Objects[6] = settingsView

		// Sincroniza a visibilidade para evitar sobreposição
		chatArea.Hide()
		workspaceDashboard.Hide()
		workspaceHub.Hide()
		agentsHub.Hide()
		skillsHub.CanvasObject.Hide()
		toolsHub.CanvasObject().Hide()
		settingsView.Hide()

		switch currentView {
		case "Workspace":
			workspaceDashboard.Show()
		case "Chat":
			chatArea.Show()
		case "Agentes":
			agentsHub.Show()
		case "Skills":
			skillsHub.CanvasObject.Show()
		case "Ferramentas":
			toolsHub.CanvasObject().Show()
		case "Configurações":
			settingsView.Show()
		default:
			chatArea.Show()
		}

		centralArea.Refresh()
	}


	// Helper para criar headers padronizados
	createPageHeader := func(icon, title string) fyne.CanvasObject {
		iconObj := adaTheme.NewIcon(icon, adaTheme.SizeMenuSmall)
		label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

		// Pequena margem para alinhar com o conteúdo
		leftMargin := canvas.NewRectangle(color.Transparent)
		leftMargin.SetMinSize(fyne.NewSize(8, 0))

		return container.NewHBox(leftMargin, iconObj, label)
	}

	// Navegação do menu lateral
	navMenu.OnSelect = func(label string) {
		fmt.Printf("[Layout] NavSelect: %s\n", label)
		currentView = label

		// Limpa o cabeçalho dinâmico
		headerContainer.Objects = nil

		switch label {
		case "Workspace":
			updateDashboard(currentEditIndex)
			headerContainer.Add(createPageHeader(adaTheme.IconStorage, "Workspace Dashboard"))
		case "Chat":
			headerContainer.Add(chatHeader.Container)
		case "Agentes":
			headerContainer.Add(createPageHeader(adaTheme.IconRobot, "Agentes Disponíveis"))
		case "Skills":
			headerContainer.Add(createPageHeader(adaTheme.IconTools, "Biblioteca de Skills"))
		case "Ferramentas":
			toolsHub.Refresh()
			headerContainer.Add(createPageHeader(adaTheme.IconHammer, "Ferramentas e Scripts"))
		case "Configurações":
			headerContainer.Add(createPageHeader(adaTheme.IconSettings, "Configurações do Sistema"))
		default:
			headerContainer.Add(chatHeader.Container)
		}

		refreshDashboard()
		headerContainer.Refresh()
		centralArea.Refresh()
	}

	// 5. Callbacks do painel de detalhes
	detailsPanel.OnChatSelect = func(id string) {
		chatController.LoadSession(id)
		if currentView != "Chat" {
			navMenu.OnSelect("Chat")
		}
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

	// Popula inicial
	fmt.Println("[Layout] Populando dashboard de workspaces...")
	refreshDashboard()
	fmt.Println("[Layout] Dashboard OK")

	// Garante que iniciamos no CHAT e nada mais está visível
	navMenu.OnSelect("Chat")

	// Layout Final
	split = container.NewHSplit(sidebar, centralArea)
	split.Offset = 0.2

	fmt.Println("[Layout] CreateMainLayout finalizado.")
	return container.NewBorder(topBar, nil, nil, nil, split)
}
