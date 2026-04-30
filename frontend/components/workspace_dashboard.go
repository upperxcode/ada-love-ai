package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewWorkspaceDashboard(engine *backend.Engine) fyne.CanvasObject {
	cfg := engine.GetAdaConfig()

	// 1. Seção de Personalidade
	systemEntry := widget.NewMultiLineEntry()
	systemEntry.SetMinRowsVisible(4)
	systemEntry.Text = cfg.Personality
	systemCard := createConfigCard("Personalidade", systemEntry, nil)

	// Helper para mostrar diálogo de adição
	showAddDialog := func(title string, onConfirm func(string)) {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Digite o valor...")
		w := fyne.CurrentApp().Driver().AllWindows()[0]

		form := widget.NewForm(widget.NewFormItem("Valor", entry))
		d := dialog.NewCustomConfirm("Adicionar "+title, "Adicionar", "Cancelar", container.NewPadded(form), func(ok bool) {
			if ok && entry.Text != "" {
				onConfirm(entry.Text)
			}
		}, w)
		d.Resize(fyne.NewSize(500, 200))
		d.Show()
	}

	// 2. Seções Padronizadas (usando ponteiros para atualizar a UI localmente)
	var content *fyne.Container

	refresh := func() {
		newDashboard := NewWorkspaceDashboard(engine)
		if content != nil {
			content.Objects = []fyne.CanvasObject{newDashboard}
			content.Refresh()
		}
	}

	// Funções de salvamento imediato ou no botão
	updateAndSave := func() {
		engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
			c.Personality = systemEntry.Text
		})
		dialog.ShowInformation("Sucesso", "Configurações salvas com sucesso!", fyne.CurrentApp().Driver().AllWindows()[0])
	}

	sections := []ConfigSection{
		{
			Title: "Pastas de trabalho",
			Items: cfg.Workspaces,
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() {
					w := fyne.CurrentApp().Driver().AllWindows()[0]
					d := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
						if err != nil || list == nil {
							return
						}
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Workspaces = append(c.Workspaces, list.Path())
						})
						refresh()
					}, w)
					d.Resize(fyne.NewSize(800, 600))
					d.Show()
				}),
			},
			OnDel: func(i int) {
				cfg.Workspaces = append(cfg.Workspaces[:i], cfg.Workspaces[i+1:]...)
				engine.SetAdaConfig(cfg)
				refresh()
			},
		},
		{
			Title: "Knowledge Base (RAG)",
			Items: cfg.Knowledge,
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconDocument, adaTheme.SizeMenuSmall, func() {
					w := fyne.CurrentApp().Driver().AllWindows()[0]
					d := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
						if err != nil || file == nil {
							return
						}
						cfg.Knowledge = append(cfg.Knowledge, file.URI().Path())
						engine.SetAdaConfig(cfg)
						refresh()
					}, w)
					d.Resize(fyne.NewSize(800, 600))
					d.Show()
				}),
				adaTheme.NewIconButton(adaTheme.MenuShareIcon, adaTheme.SizeMenuSmall, func() {
					showAddDialog("Link (URL)", func(val string) {
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Knowledge = append(c.Knowledge, val)
						})
						refresh()
					})
				}),
			},
			OnDel: func(i int) {
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.Knowledge = append(c.Knowledge[:i], c.Knowledge[i+1:]...)
				})
				refresh()
			},
		},
		{
			Title: "Agentes",
			Items: cfg.WorkspaceAgents,
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() {
					ShowAgentSelectionDialog(engine, func(selectedName string) {
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.WorkspaceAgents = append(c.WorkspaceAgents, selectedName)
						})
						refresh()
					})
				}),
			},
			OnDel: func(i int) {
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.WorkspaceAgents = append(c.WorkspaceAgents[:i], c.WorkspaceAgents[i+1:]...)
				})
				refresh()
			},
		},
		{
			Title: "Skills",
			Items: cfg.Skills,
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() {
					ShowSkillSelectionDialog(engine, func(selectedName string) {
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Skills = append(c.Skills, selectedName)
						})
						refresh()
					})
				}),
			},
			OnDel: func(i int) {
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.Skills = append(c.Skills[:i], c.Skills[i+1:]...)
				})
				refresh()
			},
		},
	}

	// Botão de Salvar (para a Personalidade e garantir persistência total)
	saveBtn := adaTheme.NewClickableButton(updateAndSave)
	saveBtn.Text = adaTheme.IconCheck + " Salvar Configurações"
	saveBtn.Importance = widget.HighImportance

	// Montagem do Layout
	innerContent := container.NewVBox(
		container.NewHBox(adaTheme.NewIcon(adaTheme.IconSettings, adaTheme.SizeMenuSmall), widget.NewLabelWithStyle("CONFIGURAÇÃO DO WORKSPACE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewSeparator(),
		systemCard,
		container.NewGridWithColumns(2, container.NewPadded(sections[0].Render()), container.NewPadded(sections[1].Render())),
		container.NewGridWithColumns(2, container.NewPadded(sections[2].Render()), container.NewPadded(sections[3].Render())),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), saveBtn),
	)

	scroll := container.NewVScroll(container.NewPadded(innerContent))

	// Usamos um container.Stack como wrapper para permitir o refresh trocando o conteúdo
	content = container.NewStack(container.NewPadded(scroll))
	return content
}
