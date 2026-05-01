package components

import (
	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	WorkspaceLabel *widget.Label
}

func NewDetailsPanel() *DetailsPanel {
	dp := &DetailsPanel{
		SearchEntry:    widget.NewEntry(),
		ChatList:       container.NewVBox(),
		WorkspaceLabel: widget.NewLabelWithStyle("WORKSPACE", fyne.TextAlignLeading, fyne.TextStyle{Italic: true, Bold: true}),
	}
	dp.WorkspaceLabel.Importance = widget.LowImportance
	dp.SearchEntry.SetPlaceHolder("Buscar conversas...")

	// --- ABA CONVERSAS ---
	convList := container.NewBorder(
		container.NewVBox(
			dp.WorkspaceLabel,
			container.NewPadded(dp.SearchEntry),
		),
		nil, nil, nil,
		container.NewVScroll(dp.ChatList),
	)

	dp.CanvasObject = container.NewPadded(convList)
	return dp
}

func (dp *DetailsPanel) SetWorkspaceName(name string) {
	if name == "" {
		dp.WorkspaceLabel.SetText("WORKSPACE")
	} else {
		dp.WorkspaceLabel.SetText(strings.ToUpper(name))
	}
}

func (dp *DetailsPanel) UpdateSessions(sessions []*backend.ChatSession, activeID string) {
	dp.ChatList.Objects = nil
	for _, sess := range sessions {
		sess := sess // Garantir cópia local para os closures
		id := sess.ID
		titleStr := sess.Title
		if titleStr == "" {
			titleStr = "Nova Conversa"
		}

		// Botão de menu com tema transparente
		menuBtn := adaTheme.NewClickableButton(nil)
		menuBtn.Text = "⋮"
		styledMenuBtn := container.NewThemeOverride(menuBtn, &adaTheme.GhostTheme{TextSize: 18})

		var pinIcon fyne.CanvasObject
		if sess.Pinned {
			pinIcon = adaTheme.NewIcon(adaTheme.MenuPinIcon, adaTheme.SizeIconTiny)
		} else {
			pinIcon = layout.NewSpacer()
		}

		titleLabel := widget.NewLabel(titleStr)
		titleLabel.Truncation = fyne.TextTruncateEllipsis

		bg := newTappableBG(func() {
			fmt.Printf("[DetailsPanel] Selecionando sessão: %s (Título: %s)\n", id, titleStr)
			if dp.OnChatSelect != nil {
				dp.OnChatSelect(id)
			}
		})

		// Se estiver ativo, destaca
		if id == activeID {
			bg.fill = color.NRGBA{R: 100, G: 100, B: 255, A: 30}
			titleLabel.TextStyle = fyne.TextStyle{Bold: true}
		}

		// Adiciona um pequeno recuo para o ícone não ficar colado na borda esquerda
		leftMargin := canvas.NewRectangle(color.Transparent)
		leftMargin.SetMinSize(fyne.NewSize(8, 0))
		leftContent := container.NewHBox(leftMargin, styledMenuBtn)

		visualContent := container.NewBorder(
			nil, nil,
			leftContent, // Left (com margem)
			pinIcon,     // Right
			titleLabel,  // Center
		)

		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(0, 40))

		row := container.NewMax(bg, spacer, container.NewPadded(visualContent))

		menuBtn.OnTapped = func() {
			pinnedText := "Fixar"
			if sess.Pinned {
				pinnedText = "Desafixar"
			}
			items := []*fyne.MenuItem{
				fyne.NewMenuItem(adaTheme.MenuShareIcon+" Compartilhar conversa", func() {
					if dp.OnShare != nil {
						dp.OnShare(id)
					}
				}),
				fyne.NewMenuItem(adaTheme.MenuPinIcon+" "+pinnedText, func() {
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
