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
	OnToggle  func(bool)
}

// Toggle reconstrói o conteúdo para aplicar os tamanhos corretamente
func (s *SidebarMenu) Toggle() {
	s.Collapsed = !s.Collapsed
	s.refresh()
	if s.OnToggle != nil {
		s.OnToggle(s.Collapsed)
	}
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
		icon := adaTheme.NewIcon(s.Icons[i], adaTheme.SizeMenuBig)

		iconSlotBase := canvas.NewRectangle(color.Transparent)
		iconSlotBase.SetMinSize(fyne.NewSize(64, 48))
		iconSlot := container.NewStack(iconSlotBase, container.NewCenter(icon))

		var item fyne.CanvasObject
		if s.Collapsed {
			// Apenas ícone centralizado, sem margens extras ou labels
			item = iconSlot
		} else {
			margin := canvas.NewRectangle(color.Transparent)
			margin.SetMinSize(fyne.NewSize(12, 0))
			label := widget.NewLabel(s.Labels[i])
			item = container.NewHBox(margin, iconSlot, label)
		}

		btn := adaTheme.NewClickableButton(func() {
			if s.OnSelect != nil {
				s.OnSelect(s.Labels[i])
			}
		})
		styledBtn := container.NewThemeOverride(btn, &adaTheme.GhostTheme{})
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
		Labels:    []string{"Workspace", "Chat", "Agentes", "Skills", "Ferramentas", "Configurações"},
		Icons:     []string{adaTheme.IconStorage, adaTheme.IconChat, adaTheme.IconRobot, adaTheme.IconTools, adaTheme.IconHammer, adaTheme.IconSettings},
	}
	sm.refresh()
	return sm
}

func NewSidebarTitle() fyne.CanvasObject {
	// Ícone Logo Tech - Usando o componente padrão para garantir alinhamento
	logo := adaTheme.NewIcon(adaTheme.IconLogo, 32, adaTheme.AccentColor)

	logoSlotBase := canvas.NewRectangle(color.Transparent)
	logoSlotBase.SetMinSize(fyne.NewSize(64, 48))
	logoSlot := container.NewStack(logoSlotBase, container.NewCenter(logo))

	name := widget.NewLabel("Ada Love AI")
	name.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewHBox(logoSlot, name)
}
