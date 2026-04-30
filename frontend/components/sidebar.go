package components

import (
	"image/color"

	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SidebarMenu encapsula o estado e o container da barra lateral
type SidebarMenu struct {
	Container *fyne.Container
	TitleObj  *fyne.Container
	Collapsed bool
	Labels    []string
	Icons     []string
	OnSelect  func(string)
}

// Toggle reconstrói o conteúdo para aplicar os tamanhos corretamente
func (s *SidebarMenu) Toggle() {
	s.Collapsed = !s.Collapsed
	s.refresh()
}

func (s *SidebarMenu) refresh() {
	if s.TitleObj != nil && len(s.TitleObj.Objects) > 2 {
		if s.Collapsed {
			s.TitleObj.Objects[2].Hide()
		} else {
			s.TitleObj.Objects[2].Show()
		}
		s.TitleObj.Refresh()
	}

	s.Container.Objects = nil
	for i := range s.Labels {
		icon := canvas.NewText(s.Icons[i], adaTheme.TextColor)
		icon.TextSize = adaTheme.SizeMenuBig

		margin := canvas.NewRectangle(color.Transparent)
		margin.SetMinSize(fyne.NewSize(12, 0))

		iconSlotBase := canvas.NewRectangle(color.Transparent)
		iconSlotBase.SetMinSize(fyne.NewSize(40, 0))
		iconSlot := container.NewStack(iconSlotBase, container.NewCenter(icon))

		var item fyne.CanvasObject
		if s.Collapsed {
			item = container.NewHBox(margin, iconSlot)
		} else {
			label := widget.NewLabel(s.Labels[i])
			item = container.NewHBox(margin, iconSlot, label)
		}

		btn := adaTheme.NewClickableButton(func() {
			if s.OnSelect != nil {
				s.OnSelect(s.Labels[i])
			}
		})
		styledBtn := container.NewThemeOverride(btn, adaTheme.GhostTheme{})
		s.Container.Add(container.NewStack(styledBtn, item))
	}
	s.Container.Refresh()
}

func NewNavMenu(titleObj fyne.CanvasObject) *SidebarMenu {
	var tContainer *fyne.Container
	if tc, ok := titleObj.(*fyne.Container); ok {
		tContainer = tc
	}

	sm := &SidebarMenu{
		Container: container.NewVBox(),
		TitleObj:  tContainer,
		Collapsed: false,
		Labels:    []string{"Workspaces", "Chat", "Agentes", "Skills", "Configurações"},
		Icons:     []string{adaTheme.IconStorage, "󰭻", adaTheme.IconRobot, adaTheme.IconTools, adaTheme.IconSettings},
	}
	sm.refresh()
	return sm
}

func NewSidebarTitle() fyne.CanvasObject {
	// Margem esquerda de 16px
	margin := canvas.NewRectangle(color.Transparent)
	margin.SetMinSize(fyne.NewSize(16, 0))

	// Ícone Logo Tech
	logo := canvas.NewText("󰚩", adaTheme.AccentColor)
	logo.TextSize = 32
	logo.TextStyle = fyne.TextStyle{Bold: true}

	logoSlotBase := canvas.NewRectangle(color.Transparent)
	logoSlotBase.SetMinSize(fyne.NewSize(40, 0))
	logoSlot := container.NewStack(logoSlotBase, container.NewCenter(logo))

	name := widget.NewLabel("Ada Love AI")
	name.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewHBox(margin, logoSlot, name)
}
