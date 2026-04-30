package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showAgentDialog(engine *backend.Engine, existing *backend.AgentConfig, onSave func(backend.AgentConfig)) {
	w := fyne.CurrentApp().Driver().AllWindows()[0]

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nome do Agente (ex: Arquiteto)")

	personaEntry := widget.NewMultiLineEntry()
	personaEntry.SetPlaceHolder("Defina a personalidade e objetivos deste agente...")
	personaEntry.SetMinRowsVisible(8)
	personaEntry.Wrapping = fyne.TextWrapWord

	cfg := engine.GetAdaConfig()
	categories := cfg.AgentCategories
	if len(categories) == 0 {
		categories = []string{"Geral"}
	}
	categorySelect := widget.NewSelect(categories, nil)
	categorySelect.PlaceHolder = "Selecione uma Categoria"

	addCategoryBtn := adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeControlSmall, func() {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Nome da nova categoria")
		dialog.ShowCustomConfirm("Nova Categoria", "Adicionar", "Cancelar", entry, func(ok bool) {
			if ok && entry.Text != "" {
				newCat := entry.Text
				cfg := engine.GetAdaConfig()
				cfg.AgentCategories = append(cfg.AgentCategories, newCat)
				engine.SetAdaConfig(cfg)

				categorySelect.Options = cfg.AgentCategories
				categorySelect.SetSelected(newCat)
				categorySelect.Refresh()
			}
		}, w)
	})

	modelList := engine.GetModelList()
	providersMap := make(map[string][]string)
	var providerNames []string

	for _, m := range modelList {
		if _, ok := providersMap[m.Provider]; !ok {
			providerNames = append(providerNames, m.Provider)
		}
		providersMap[m.Provider] = append(providersMap[m.Provider], m.ModelName)
	}

	providerSelect := widget.NewSelect(providerNames, nil)
	modelSelect := widget.NewSelect([]string{}, nil)

	providerSelect.OnChanged = func(p string) {
		modelSelect.Options = providersMap[p]
		modelSelect.SetSelectedIndex(0)
		modelSelect.Refresh()
	}

	colorOptions := []string{"Padrão", "Roxo", "Azul", "Ciano", "Verde", "Laranja", "Amarelo", "Vermelho", "Rosa", "Cinza", "Branco", "Preto"}
	colorMap := map[string]string{
		"Padrão":   "#7B61FF",
		"Roxo":     "#9D7BFF",
		"Azul":     "#4DA1FF",
		"Ciano":    "#4DFFFF",
		"Verde":    "#4DFF91",
		"Laranja":  "#FF9D4D",
		"Amarelo":  "#FFD64D",
		"Vermelho": "#FF4D4D",
		"Rosa":     "#FF4D9D",
		"Cinza":    "#8E8E93",
		"Branco":   "#FFFFFF",
		"Preto":    "#1C1C1E",
	}
	colorSelect := widget.NewSelect(colorOptions, nil)
	colorSelect.SetSelected("Padrão")

	// Configuração de Ícones
	iconOptions := []string{"🤖 Robô", "🧠 Cérebro", "💻 Código", "⚙️ Sistema", "🧪 Pesquisa", "🎨 Criativo", "🔍 Analista", "📈 Dados", "💬 Chat", "🔒 Segurança", "Outro..."}
	iconMap := map[string]string{
		"🤖 Robô": "🤖", "🧠 Cérebro": "🧠", "💻 Código": "💻", "⚙️ Sistema": "⚙️", "🧪 Pesquisa": "🧪",
		"🎨 Criativo": "🎨", "🔍 Analista": "🔍", "📈 Dados": "📈", "💬 Chat": "💬", "🔒 Segurança": "🔒",
	}

	iconSelect := widget.NewSelect(iconOptions, nil)
	iconEntry := widget.NewEntry()
	iconEntry.SetPlaceHolder("Digite o emoji customizado...")
	iconEntry.Hide()

	iconSelect.OnChanged = func(s string) {
		if s == "Outro..." {
			iconEntry.Show()
		} else {
			iconEntry.Hide()
			if val, ok := iconMap[s]; ok {
				iconEntry.Text = val
			}
		}
	}

	if existing != nil {
		nameEntry.Text = existing.Name
		personaEntry.Text = existing.Persona
		providerSelect.SetSelected(existing.Provider)
		modelSelect.SetSelected(existing.Model)
		categorySelect.SetSelected(existing.Category)

		// Tentar encontrar o ícone nas opções
		foundIcon := false
		for label, emoji := range iconMap {
			if emoji == existing.Icon {
				iconSelect.SetSelected(label)
				foundIcon = true
				break
			}
		}
		if !foundIcon && existing.Icon != "" {
			iconSelect.SetSelected("Outro...")
			iconEntry.Text = existing.Icon
			iconEntry.Show()
		}

		// Encontrar o nome da cor pelo valor hex
		for name, val := range colorMap {
			if val == existing.Color {
				colorSelect.SetSelected(name)
				break
			}
		}
	}

	form := widget.NewForm(
		widget.NewFormItem("Nome", nameEntry),
		widget.NewFormItem("Categoria", container.NewBorder(nil, nil, nil, addCategoryBtn, categorySelect)),
		widget.NewFormItem("Cor da Borda", colorSelect),
		widget.NewFormItem("Persona", personaEntry),
		widget.NewFormItem("Provedor", providerSelect),
		widget.NewFormItem("Modelo", modelSelect),
		widget.NewFormItem("Ícone", container.NewVBox(iconSelect, iconEntry)),
	)

	title := "Novo Agente"
	if existing != nil {
		title = "Editar Agente"
	}

	// Usamos um scroll aqui para o caso de a persona ser muito longa
	scrollContent := container.NewVScroll(container.NewPadded(form))
	scrollContent.SetMinSize(fyne.NewSize(580, 420))

	d := dialog.NewCustomConfirm(title, "Salvar", "Cancelar", scrollContent, func(ok bool) {
		if ok && nameEntry.Text != "" {
			onSave(backend.AgentConfig{
				Name:     nameEntry.Text,
				Persona:  personaEntry.Text,
				Category: categorySelect.Selected,
				Icon:     iconEntry.Text,
				Color:    colorMap[colorSelect.Selected],
				Provider: providerSelect.Selected,
				Model:    modelSelect.Selected,
			})
		}
	}, w)

	d.Resize(fyne.NewSize(600, 500))
	d.Show()
}
