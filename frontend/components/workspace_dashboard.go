package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func renderWorkspaceList(engine *backend.Engine, activeIndex int, onRefresh func(), onSelect func(int)) fyne.CanvasObject {
	cfg := engine.GetAdaConfig()
	workspaces := cfg.Workspaces

	headerActions := []fyne.CanvasObject{
		adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() {
			w := fyne.CurrentApp().Driver().AllWindows()[0]
			d := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
				if err != nil || list == nil {
					return
				}
				path := list.Path()
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.Workspaces = append(c.Workspaces, backend.WorkspaceConfig{
						Title:   "Novo Workspace",
						Folders: []string{path},
						Path:    path,
						Tools:   []string{"read_file", "write_file", "list_dir", "edit_file"},
					})
				})
				newIdx := len(engine.GetAdaConfig().Workspaces) - 1
				onSelect(newIdx)
				onRefresh()
			}, w)
			d.Resize(fyne.NewSize(800, 600))
			d.Show()
		}),
	}

	items := make([]ConfigItem, len(workspaces))
	for i, ws := range workspaces {
		title := ws.Title
		if title == "" {
			title = "Workspace " + fmt.Sprintf("%d", i+1)
		}

		subtitle := ws.Description
		if subtitle == "" {
			subtitle = "Sem descrição definida"
		}

		items[i] = ConfigItem{
			Title:      title,
			Subtitle:   subtitle,
			Value:      fmt.Sprintf("%d", i),
			IsChecked:  i == cfg.ActiveWorkspaceIndex,
			IsSelected: i == activeIndex,
		}
	}

	wsSection := ConfigSection{
		Title:         "MEUS WORKSPACES",
		Items:         items,
		HeaderActions: headerActions,
		OnSelect: func(i int) {
			onSelect(i)
		},
		OnCheck: func(i int) {
			engine.SetActiveWorkspace(workspaces[i].Path)
			onRefresh()
		},
		OnDel: func(i int) {
			w := fyne.CurrentApp().Driver().AllWindows()[0]
			dialog.ShowConfirm("Excluir Workspace", "Tem certeza?", func(ok bool) {
				if ok {
					engine.DeleteWorkspace(workspaces[i].Path)
					onRefresh()
				}
			}, w)
		},
	}

	return wsSection.Render()
}

func NewWorkspaceDashboard(engine *backend.Engine, editIndex int, onSelect func(int), onRefresh ...func()) fyne.CanvasObject {
	var onRefreshCallback func()
	if len(onRefresh) > 0 {
		onRefreshCallback = onRefresh[0]
	}

	cfg := engine.GetAdaConfig()

	// Se não houver workspaces, mostra vazio
	if len(cfg.Workspaces) == 0 {
		return container.NewCenter(widget.NewLabel("Nenhum Workspace encontrado. Crie um no botão '+'.") )
	}

	// Garante que o índice de edição é válido
	if editIndex < 0 || editIndex >= len(cfg.Workspaces) {
		editIndex = cfg.ActiveWorkspaceIndex
	}

	activeWS := cfg.Workspaces[editIndex]

	// Helpers para converter tipos do backend para tipos de UI
	toStringItems := func(list []string) []ConfigItem {
		items := make([]ConfigItem, len(list))
		for i, v := range list {
			items[i] = ConfigItem{Title: v, Value: v, IsChecked: false, IsSelected: false}
		}
		return items
	}

	// 0. Seção de Identificação
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Ex: Meu Workspace de IA")
	titleEntry.Text = activeWS.Title

	descriptionEntry := widget.NewEntry()
	descriptionEntry.SetPlaceHolder("Ex: Descrição detalhada do projeto Ada Love AI")
	descriptionEntry.Text = activeWS.Description

	titleCard := createConfigCard("Título do Workspace", titleEntry, nil)
	descriptionCard := createConfigCard("Descrição do Workspace", descriptionEntry, nil)

	// 1. Seção de Personalidade
	systemEntry := widget.NewMultiLineEntry()
	systemEntry.SetMinRowsVisible(4)
	systemEntry.Text = activeWS.Personality
	systemCard := createConfigCard("Personalidade", systemEntry, nil)

	// Helper para mostrar diálogo de adição
	showAddDialog := func(title string, onConfirm func(string)) {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Digite o value...")
		w := fyne.CurrentApp().Driver().AllWindows()[0]

		form := widget.NewForm(widget.NewFormItem("Value", entry))
		d := dialog.NewCustomConfirm("Adicionar "+title, "Adicionar", "Cancelar", container.NewPadded(form), func(ok bool) {
			if ok && entry.Text != "" {
				onConfirm(entry.Text)
			}
		}, w)
		d.Resize(fyne.NewSize(500, 200))
		d.Show()
	}

	// 2. Seções Padronizadas
	var content *fyne.Container

	refresh := func() {
		if onRefreshCallback != nil {
			onRefreshCallback()
		}
	}

	updateAndSave := func() {
		engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
			if editIndex >= 0 && editIndex < len(c.Workspaces) {
				// Validação: deve ter pelo menos uma pasta
				if len(c.Workspaces[editIndex].Folders) == 0 {
					dialog.ShowError(fmt.Errorf("O workspace deve conter pelo menos uma pasta de projeto"), fyne.CurrentApp().Driver().AllWindows()[0])
					return
				}

				// Sincroniza o path do workspace com a primeira pasta das folders
				// O path é usado para identificar a memória e o contexto do agente
				oldPath := c.Workspaces[editIndex].Path
				newPath := c.Workspaces[editIndex].Folders[0]
				
				c.Workspaces[editIndex].Title = titleEntry.Text
				c.Workspaces[editIndex].Description = descriptionEntry.Text
				c.Workspaces[editIndex].Personality = systemEntry.Text
				c.Workspaces[editIndex].Path = newPath

				// Se este for o workspace ativo, atualizamos o path global
				if editIndex == c.ActiveWorkspaceIndex || c.ActiveWorkspacePath == oldPath {
					c.ActiveWorkspacePath = newPath
				}

				dialog.ShowInformation("Sucesso", "Configurações salvas!", fyne.CurrentApp().Driver().AllWindows()[0])
			}
		})
		refresh()
	}

	sections := []ConfigSection{
		{
			Title: "Pastas do Projeto",
			Items: toStringItems(activeWS.Folders),
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() {
					w := fyne.CurrentApp().Driver().AllWindows()[0]
					d := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
						if err != nil || list == nil {
							return
						}
						path := list.Path()
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Workspaces[editIndex].Folders = append(c.Workspaces[editIndex].Folders, path)
						})
						refresh()
					}, w)
					d.Resize(fyne.NewSize(800, 600))
					d.Show()
				}),
			},
			OnDel: func(i int) {
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.Workspaces[editIndex].Folders = append(c.Workspaces[editIndex].Folders[:i], c.Workspaces[editIndex].Folders[i+1:]...)
				})
				refresh()
			},
		},
		{
			Title: "Knowledge Base (RAG)",
			Items: toStringItems(activeWS.Knowledge),
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconDocument, adaTheme.SizeMenuSmall, func() {
					w := fyne.CurrentApp().Driver().AllWindows()[0]
					d := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
						if err != nil || file == nil {
							return
						}
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Workspaces[editIndex].Knowledge = append(c.Workspaces[editIndex].Knowledge, file.URI().Path())
						})
						refresh()
					}, w)
					d.Resize(fyne.NewSize(800, 600))
					d.Show()
				}),
				adaTheme.NewIconButton(adaTheme.MenuShareIcon, adaTheme.SizeMenuSmall, func() {
					showAddDialog("Link (URL)", func(val string) {
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Workspaces[editIndex].Knowledge = append(c.Workspaces[editIndex].Knowledge, val)
						})
						refresh()
					})
				}),
			},
			OnDel: func(i int) {
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.Workspaces[editIndex].Knowledge = append(c.Workspaces[editIndex].Knowledge[:i], c.Workspaces[editIndex].Knowledge[i+1:]...)
				})
				refresh()
			},
		},
		{
			Title: "Agentes Ativos",
			Items: toStringItems(activeWS.WorkspaceAgents),
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() {
					ShowAgentSelectionDialog(engine, func(selectedName string) {
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Workspaces[editIndex].WorkspaceAgents = append(c.Workspaces[editIndex].WorkspaceAgents, selectedName)
						})
						refresh()
					})
				}),
			},
			OnDel: func(i int) {
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.Workspaces[editIndex].WorkspaceAgents = append(c.Workspaces[editIndex].WorkspaceAgents[:i], c.Workspaces[editIndex].WorkspaceAgents[i+1:]...)
				})
				refresh()
			},
		},
		{
			Title: "Skills Disponíveis",
			Items: toStringItems(activeWS.Skills),
			HeaderActions: []fyne.CanvasObject{
				adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() {
					ShowSkillSelectionDialog(engine, func(selectedName string) {
						engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
							c.Workspaces[editIndex].Skills = append(c.Workspaces[editIndex].Skills, selectedName)
						})
						refresh()
					})
				}),
			},
			OnDel: func(i int) {
				engine.UpdateWorkspaceConfig(func(c *backend.AdaConfig) {
					c.Workspaces[editIndex].Skills = append(c.Workspaces[editIndex].Skills[:i], c.Workspaces[editIndex].Skills[i+1:]...)
				})
				refresh()
			},
		},
	}

	// Botão de Salvar
	saveBtn := adaTheme.NewClickableButton(updateAndSave)
	saveBtn.Text = adaTheme.IconCheck + " Salvar Alterações"
	saveBtn.Importance = widget.HighImportance

	// Montagem do Layout do Painel de Edição (Esquerda)
	headerLabel := "EDITANDO WORKSPACE: " + activeWS.Title
	if editIndex == cfg.ActiveWorkspaceIndex {
		headerLabel += " (PADRÃO GLOBAL)"
	}

	innerContent := container.NewVBox(
		container.NewHBox(adaTheme.NewIcon(adaTheme.IconSettings, adaTheme.SizeMenuSmall), widget.NewLabelWithStyle(headerLabel, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewSeparator(),
		container.NewGridWithColumns(2, container.NewPadded(titleCard), container.NewPadded(descriptionCard)),
		container.NewPadded(systemCard),
		container.NewGridWithColumns(2, container.NewPadded(sections[0].Render()), container.NewPadded(sections[1].Render())),
		container.NewGridWithColumns(2, container.NewPadded(sections[2].Render()), container.NewPadded(sections[3].Render())),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), saveBtn),
	)

	scroll := container.NewVScroll(container.NewPadded(innerContent))

	// Painel de Lista de Workspaces (Direita)
	wsListPanel := container.NewVScroll(container.NewPadded(renderWorkspaceList(engine, editIndex, refresh, func(i int) {
		if onSelect != nil {
			onSelect(i)
		}
	})))

	split := container.NewHSplit(scroll, wsListPanel)
	split.Offset = 0.7 // Dá mais espaço para a edição

	content = container.NewStack(container.NewPadded(split))
	return content
}
