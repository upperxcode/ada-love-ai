package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var lastSelectedTab string

func NewSettingsView(engine *backend.Engine, onRefresh ...func()) fyne.CanvasObject {
	var onRefreshCallback func()
	if len(onRefresh) > 0 {
		onRefreshCallback = onRefresh[0]
	}

	adaCfg := engine.GetAdaConfig()
	modelList := engine.GetModelList()

	// 1. Seção de Embeddings
	embedProviderEntry := widget.NewSelect([]string{"ollama", "lmstudio", "openai"}, nil)
	embedProviderEntry.SetSelected(adaCfg.TinyBrain.EmbeddingProvider)
	if embedProviderEntry.Selected == "" {
		embedProviderEntry.SetSelected("ollama")
	}

	embedModelEntry := widget.NewEntry()
	embedModelEntry.SetPlaceHolder("Ex: nomic-embed-text")
	embedModelEntry.Text = adaCfg.TinyBrain.EmbeddingModel

	embedCard := widget.NewCard("Configuração de Embedding", "Modelo usado para memória semântica (RAG)",
		container.NewVBox(
			widget.NewLabel("Provedor de Embedding:"),
			embedProviderEntry,
			widget.NewLabel("Modelo de Embedding:"),
			embedModelEntry,
		),
	)

	// 2. Seção de Provedores e Modelos
	providersMap := make(map[string][]backend.ModelConfig)
	for _, m := range modelList {
		providersMap[m.Provider] = append(providersMap[m.Provider], backend.ModelConfig{
			ModelName: m.ModelName,
			Model:     m.Model,
			Provider:  m.Provider,
			APIBase:   m.APIBase,
			Enabled:   m.Enabled,
		})
	}

	providerTabs := container.NewAppTabs()
	providerTabs.OnSelected = func(ti *container.TabItem) {
		lastSelectedTab = ti.Text
	}

	refresh := func() {
		if onRefreshCallback != nil {
			onRefreshCallback()
		}
	}

	saveAll := func() {
		engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
			c.TinyBrain.EmbeddingModel = embedModelEntry.Text
			c.TinyBrain.EmbeddingProvider = embedProviderEntry.Selected
		})
		dialog.ShowInformation("Configurações", "Configurações de sistema salvas!", fyne.CurrentApp().Driver().AllWindows()[0])
		refresh()
	}

	showAddProviderDialog := func() {
		pEntry := widget.NewEntry()
		pEntry.SetPlaceHolder("Ex: my-local-ai")

		bEntry := widget.NewEntry()
		bEntry.SetPlaceHolder("Ex: http://127.0.0.1:1234/v1")

		kEntry := widget.NewPasswordEntry()
		kEntry.SetPlaceHolder("Opcional")

		items := []*widget.FormItem{
			{Text: "Nome do Provedor", Widget: pEntry},
			{Text: "API Base URL", Widget: bEntry},
			{Text: "API Key", Widget: kEntry},
		}

		d := dialog.NewForm("Adicionar Novo Provedor", "Salvar", "Cancelar", items, func(ok bool) {
			if ok {
				engine.SetProviderSettings(pEntry.Text, bEntry.Text, kEntry.Text)
				refresh()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
		d.Resize(fyne.NewSize(500, 350))
		d.Show()
	}

	showAddModelDialog := func(provider string) {
		idEntry := widget.NewEntry()
		idEntry.SetPlaceHolder("Ex: llama-3.1")

		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Ex: Llama 3.1 8B")

		ctxEntry := widget.NewEntry()
		ctxEntry.SetPlaceHolder("Default: 4096")
		ctxEntry.SetText("4096")

		tempEntry := widget.NewEntry()
		tempEntry.SetPlaceHolder("Default: 0.7")
		tempEntry.SetText("0.7")

		maxEntry := widget.NewEntry()
		maxEntry.SetPlaceHolder("Default: 2048")
		maxEntry.SetText("2048")

		topPEntry := widget.NewEntry()
		topPEntry.SetPlaceHolder("Default: 0.9")
		topPEntry.SetText("0.9")

		items := []*widget.FormItem{
			{Text: "ID do Modelo", Widget: idEntry},
			{Text: "Nome de Exibição", Widget: nameEntry},
			{Text: "Contexto (Tokens)", Widget: ctxEntry},
			{Text: "Temperatura", Widget: tempEntry},
			{Text: "Max Output Tokens", Widget: maxEntry},
			{Text: "Top P", Widget: topPEntry},
		}

		d := dialog.NewForm("Adicionar Modelo a "+strings.ToUpper(provider), "Adicionar", "Cancelar", items, func(ok bool) {
			if ok {
				apiBase, _ := engine.GetProviderSettings(provider)
				if apiBase == "" {
					if provider == "lmstudio" {
						apiBase = "http://127.0.0.1:1234/v1"
					} else if provider == "ollama" {
						apiBase = "http://127.0.0.1:11434/v1"
					}
				}

				// Nome do modelo (se vazio usa o ID)
				displayName := nameEntry.Text
				if displayName == "" {
					displayName = idEntry.Text
				}

				err := engine.AddModel(backend.ModelConfig{
					ModelName: displayName,
					Model:     idEntry.Text,
					Provider:  provider,
					APIBase:   apiBase,
					Enabled:   true,
				})

				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
					return
				}

				// Salva configurações extras com defaults se falhar o parse
				ctxVal := 4096
				maxVal := 2048
				tempVal := 0.7
				topPVal := 0.9
				fmt.Sscanf(ctxEntry.Text, "%d", &ctxVal)
				fmt.Sscanf(tempEntry.Text, "%f", &tempVal)
				fmt.Sscanf(maxEntry.Text, "%d", &maxVal)
				fmt.Sscanf(topPEntry.Text, "%f", &topPVal)

				engine.SetModelSettings(provider, idEntry.Text, backend.ExtraModelConfig{
					ContextSize: ctxVal,
					Temperature: tempVal,
					MaxTokens:   maxVal,
					TopP:        topPVal,
				})

				refresh()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
		d.Resize(fyne.NewSize(550, 480))
		d.Show()
	}

	// Cria abas para cada provedor
	allProviders := engine.GetProviders()
	for _, p := range allProviders {
		providerName := p
		models := providersMap[providerName]

		var modelItems []fyne.CanvasObject
		for _, m := range models {
			mName := m.ModelName
			mID := m.Model
			status := "❌"
			if m.Enabled {
				status = "✅"
			}

			settings := engine.GetModelSettings(providerName, mID)
			infoText := ""
			if settings.ContextSize > 0 {
				infoText = fmt.Sprintf(" (Ctx: %d, Temp: %.1f)", settings.ContextSize, settings.Temperature)
			}

			label := widget.NewLabel(fmt.Sprintf("%s %s%s", status, mName, infoText))

			deleteBtn := adaTheme.NewIconButton(adaTheme.IconDelete, 20, func() {
				dialog.ShowConfirm("Remover Modelo", fmt.Sprintf("Deseja remover o modelo %s?", mName), func(ok bool) {
					if ok {
						engine.RemoveModel(mName, providerName)
						refresh()
					}
				}, fyne.CurrentApp().Driver().AllWindows()[0])
			})

			modelItems = append(modelItems, container.NewHBox(label, layout.NewSpacer(), deleteBtn))
		}

		if len(models) == 0 {
			modelItems = append(modelItems, widget.NewLabel("Nenhum modelo configurado."))
		}

		addBtn := adaTheme.NewIconButton(adaTheme.IconAdd, 22, func() {
			showAddModelDialog(providerName)
		})

		deleteProviderBtn := adaTheme.NewIconButton(adaTheme.IconDelete, 20, func() {
			dialog.ShowConfirm("Remover Provedor", fmt.Sprintf("Deseja remover o provedor %s e todos os seus modelos?", providerName), func(ok bool) {
				if ok {
					engine.RemoveProvider(providerName)
					refresh()
				}
			}, fyne.CurrentApp().Driver().AllWindows()[0])
		})

		scrollModels := container.NewVScroll(container.NewVBox(modelItems...))
		scrollModels.SetMinSize(fyne.NewSize(0, 300))

		pingBtn := adaTheme.NewIconButton(adaTheme.IconCloud, 20, func() {
			go func() {
				err := engine.PingProvider(providerName)
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				} else {
					dialog.ShowInformation("Conexão OK", fmt.Sprintf("Conexão com o provedor %s estabelecida com sucesso!", providerName), fyne.CurrentApp().Driver().AllWindows()[0])
				}
			}()
		})

		tabContent := container.NewBorder(
			container.NewHBox(
				widget.NewLabelWithStyle("Modelos:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				layout.NewSpacer(),
				pingBtn,
				addBtn,
				deleteProviderBtn,
			),
			nil, nil, nil,
			scrollModels,
		)
		providerTabs.Append(container.NewTabItem(strings.ToUpper(providerName), tabContent))
	}

	if lastSelectedTab != "" {
		for _, item := range providerTabs.Items {
			if item.Text == lastSelectedTab {
				providerTabs.Select(item)
				break
			}
		}
	}

	// Botão global para adicionar novo provedor
	addProviderBtn := widget.NewButtonWithIcon("Novo Provedor", nil, func() {
		showAddProviderDialog()
	})

	saveBtn := adaTheme.NewClickableButton(saveAll)
	saveBtn.Text = "Salvar Todas as Configurações"
	saveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		container.NewHBox(adaTheme.NewIcon(adaTheme.IconSettings, adaTheme.SizeMenuSmall), widget.NewLabelWithStyle("CONFIGURAÇÕES GERAIS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewSeparator(),
		container.NewPadded(embedCard),
		container.NewHBox(widget.NewLabelWithStyle("PROVEDORES DE IA", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), addProviderBtn),
		container.NewPadded(providerTabs),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), saveBtn),
	)

	return container.NewPadded(container.NewVScroll(content))
}
