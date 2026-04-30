package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowSkillSelectionDialog exibe uma lista de skills instaladas para seleção
func ShowSkillSelectionDialog(engine *backend.Engine, onSelect func(string)) {
	cfg := engine.GetAdaConfig()
	
	installed, err := engine.GetInstalledSkills()
	if err != nil {
		w := fyne.CurrentApp().Driver().AllWindows()[0]
		dialog.ShowError(err, w)
		return
	}

	// Filtra skills que já não estão no workspace
	availableSkills := []string{}
	activeMap := make(map[string]bool)
	for _, name := range cfg.Skills {
		activeMap[name] = true
	}

	for _, s := range installed {
		if !activeMap[s] {
			availableSkills = append(availableSkills, s)
		}
	}

	if len(availableSkills) == 0 {
		w := fyne.CurrentApp().Driver().AllWindows()[0]
		dialog.ShowInformation("Aviso", "Todas as skills instaladas já estão no workspace ou não há skills instaladas.", w)
		return
	}

	var selectedName string
	list := widget.NewList(
		func() int { return len(availableSkills) },
		func() fyne.CanvasObject {
			icon := adaTheme.NewIcon(adaTheme.IconTools, adaTheme.SizeCardSmall)
			name := widget.NewLabel("Nome da Skill")
			return container.NewBorder(nil, nil, icon, nil, name)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			s := availableSkills[id]
			box := obj.(*fyne.Container)
			
			// No NewBorder, o conteúdo central (name) é o primeiro objeto na lista
			// enquanto as bordas (icon) vêm depois.
			nameLabel := box.Objects[0].(*widget.Label)
			nameLabel.SetText(s)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedName = availableSkills[id]
	}

	w := fyne.CurrentApp().Driver().AllWindows()[0]
	d := dialog.NewCustomConfirm("Adicionar Skill ao Workspace", "Adicionar", "Cancelar",
		container.NewGridWrap(fyne.NewSize(500, 400), list),
		func(ok bool) {
			if ok && selectedName != "" {
				onSelect(selectedName)
			}
		}, w)

	d.Show()
}
