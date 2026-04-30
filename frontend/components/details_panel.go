package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type DetailsPanel struct {
	CanvasObject fyne.CanvasObject
	SearchEntry  *widget.Entry
	ChatList     *fyne.Container
	OnChatSelect func(string)

	OnPin          func(string)
	OnDelete       func(string)
	OnRename       func(string)
	OnAddToProject func(string)
	OnShare        func(string)
}

func NewDetailsPanel() *DetailsPanel {
	dp := &DetailsPanel{
		SearchEntry: widget.NewEntry(),
		ChatList:    container.NewVBox(),
	}
	dp.SearchEntry.SetPlaceHolder("Buscar conversas...")

	// --- ABA WORKSPACES (Usando o padrão ConfigSection) ---
	wsSection := ConfigSection{
		Title: "SEUS WORKSPACES",
		Items: []string{"🌌 Workspace Padrão", "💜 Ada-Love-Ai", "📱 App Mobile", "💾 Backend Go"},
		HeaderActions: []fyne.CanvasObject{
			adaTheme.NewIconButton(adaTheme.IconAdd, adaTheme.SizeMenuSmall, func() { fmt.Println("Novo WS") }),
		},
		OnDel: func(i int) { fmt.Println("Del WS", i) },
	}

	workspacesBox := container.NewBorder(
		nil, nil, nil, nil,
		container.NewVScroll(wsSection.Render()),
	)

	// --- ABA CONVERSAS ---
	convList := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("HISTÓRICO DO WORKSPACE", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
			container.NewPadded(dp.SearchEntry),
		),
		nil, nil, nil,
		container.NewVScroll(dp.ChatList),
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("🏢 Workspaces", container.NewPadded(workspacesBox)),
		container.NewTabItem("💬 Conversas", container.NewPadded(convList)),
	)

	dp.CanvasObject = container.NewPadded(tabs)
	return dp
}

func (dp *DetailsPanel) UpdateSessions(sessions []*backend.ChatSession) {
	dp.ChatList.Objects = nil
	for _, sess := range sessions {
		id := sess.ID
		titleStr := sess.Title
		if titleStr == "" {
			titleStr = "Nova Conversa"
		}

		menuBtn := adaTheme.NewClickableButton(nil)
		menuBtn.Text = "⋮"

		var pinIcon fyne.CanvasObject
		if sess.Pinned {
			pinIcon = adaTheme.NewIcon(adaTheme.MenuPinIcon, 16)
		} else {
			pinIcon = layout.NewSpacer()
		}

		titleLabel := widget.NewLabel(titleStr)
		titleLabel.Truncation = fyne.TextTruncateEllipsis

		row := container.NewStack()
		selectBtn := adaTheme.NewClickableButton(func() {
			if dp.OnChatSelect != nil {
				dp.OnChatSelect(id)
			}
		})

		visualContent := container.NewBorder(
			nil, nil,
			menuBtn,    // Left
			pinIcon,    // Right
			titleLabel, // Center
		)

		row.Add(selectBtn)
		row.Add(container.NewPadded(visualContent))

		menuBtn.OnTapped = func() {
			items := []*fyne.MenuItem{
				fyne.NewMenuItem(adaTheme.MenuShareIcon+" Compartilhar conversa", func() {
					if dp.OnShare != nil {
						dp.OnShare(id)
					}
				}),
				fyne.NewMenuItem(map[bool]string{true: adaTheme.MenuPinIcon + " Desafixar", false: adaTheme.MenuPinIcon + " Fixar"}[sess.Pinned], func() {
					if dp.OnPin != nil {
						dp.OnPin(id)
					}
				}),
				fyne.NewMenuItem(adaTheme.MenuEditIcon+" Renomear", func() {
					if dp.OnRename != nil {
						dp.OnRename(id)
					}
				}),
				fyne.NewMenuItem(adaTheme.MenuProjectIcon+" Adicionar ao Workspace", func() {
					if dp.OnAddToProject != nil {
						dp.OnAddToProject(id)
					}
				}),
				fyne.NewMenuItemSeparator(),
				fyne.NewMenuItem(adaTheme.MenuDeleteIcon+" Excluir", func() {
					if dp.OnDelete != nil {
						dp.OnDelete(id)
					}
				}),
			}

			menu := fyne.NewMenu("", items...)
			popUp := widget.NewPopUpMenu(menu, fyne.CurrentApp().Driver().CanvasForObject(menuBtn))
			popUp.ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(menuBtn))
		}

		dp.ChatList.Add(row)
	}
	dp.ChatList.Refresh()
}
