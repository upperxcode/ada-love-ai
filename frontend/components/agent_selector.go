package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowAgentSelectionDialog exibe uma lista de agentes cadastrados para seleção
func ShowAgentSelectionDialog(engine *backend.Engine, onSelect func(string)) {
	cfg := engine.GetAdaConfig()

	// Filtra agentes que já não estão no workspace ativo
	availableAgents := []backend.AgentConfig{}
	activeMap := make(map[string]bool)
	if len(cfg.Workspaces) > 0 {
		activeWS := cfg.Workspaces[cfg.ActiveWorkspaceIndex]
		for _, name := range activeWS.WorkspaceAgents {
			activeMap[name] = true
		}
	}

	for _, a := range cfg.Agents {
		if !activeMap[a.Name] {
			availableAgents = append(availableAgents, a)
		}
	}

	if len(availableAgents) == 0 {
		w := fyne.CurrentApp().Driver().AllWindows()[0]
		dialog.ShowInformation("Aviso", "Todos os agentes cadastrados já estão no workspace ou não há agentes cadastrados.", w)
		return
	}

	var selectedName string
	list := widget.NewList(
		func() int { return len(availableAgents) },
		func() fyne.CanvasObject {
			placeholder := adaTheme.NewIcon(adaTheme.IconRobot, adaTheme.SizeCardSmall)
			emoji := adaTheme.NewIcon("", adaTheme.SizeCardSmall)
			iconStack := container.NewStack(placeholder, emoji)

			name := widget.NewLabel("Nome do Agente")
			cat := widget.NewLabelWithStyle("Categoria", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})
			cat.Importance = widget.LowImportance

			return container.NewBorder(nil, nil, iconStack, cat, name)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			a := availableAgents[id]
			box := obj.(*fyne.Container)

			var iconStack *fyne.Container
			var nameLabel *widget.Label
			var catLabel *widget.Label

			for _, o := range box.Objects {
				if c, ok := o.(*fyne.Container); ok {
					iconStack = c
				} else if l, ok := o.(*widget.Label); ok {
					if l.TextStyle.Italic {
						catLabel = l
					} else {
						nameLabel = l
					}
				}
			}

			if iconStack != nil && len(iconStack.Objects) >= 2 {
				placeholder := iconStack.Objects[0].(*canvas.Text)
				emoji := iconStack.Objects[1].(*canvas.Text)
				if a.Icon != "" {
					placeholder.Hide()
					emoji.Text = a.Icon
					emoji.Show()
				} else {
					placeholder.Show()
					emoji.Hide()
				}
				emoji.Refresh()
				placeholder.Refresh()
			}

			if nameLabel != nil {
				nameLabel.SetText(a.Name)
			}
			if catLabel != nil {
				catLabel.SetText(a.Category)
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedName = availableAgents[id].Name
	}

	w := fyne.CurrentApp().Driver().AllWindows()[0]
	d := dialog.NewCustomConfirm("Adicionar Agente ao Workspace", "Adicionar", "Cancelar",
		container.NewGridWrap(fyne.NewSize(500, 400), list),
		func(ok bool) {
			if ok && selectedName != "" {
				onSelect(selectedName)
			}
		}, w)

	d.Show()
}
