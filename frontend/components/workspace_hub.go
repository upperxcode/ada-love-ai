package components

import (
	adaTheme "ada-love-ai/frontend/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewWorkspaceHub() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("WORKSPACE OPERACIONAL", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabelWithStyle("Gerencie o ecossistema técnico e os agentes deste espaço.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// Função auxiliar para criar cards de seção
	createSectionCard := func(title, desc string, icon string) fyne.CanvasObject {
		titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		descLbl := widget.NewLabel(desc)
		descLbl.Wrapping = fyne.TextWrapWord

		btnStyled := adaTheme.NewTextIconButton(icon, "Gerenciar", adaTheme.SizeMenuSmall, func() {})

		content := container.NewVBox(titleLbl, descLbl, layout.NewSpacer(), btnStyled)
		bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
		bg.CornerRadius = 8

		return container.NewStack(bg, container.NewPadded(content))
	}

	grid := container.NewGridWithColumns(2,
		createSectionCard("📂 Working Dirs", "Diretórios físicos no Linux onde estão seus projetos e códigos.", adaTheme.IconFolder),
		createSectionCard("🧠 Knowledge Base (RAG)", "Documentos, PDFs e manuais que servem de fonte de verdade.", adaTheme.IconInfo),
		createSectionCard("📜 Histórico e Artefatos", "Logs de interação, diagramas e relatórios gerados.", adaTheme.IconDocument),
		createSectionCard("🤖 Agentes", "Configure múltiplos agentes com papéis e modelos distintos.", adaTheme.IconRobot),
		createSectionCard("🛠️ Skills", "Habilidades que os agentes invocam (ex: ler arquivos, rodar testes).", adaTheme.IconTools),
	)

	return container.NewBorder(
		container.NewVBox(title, subtitle, widget.NewSeparator()),
		nil, nil, nil,
		container.NewPadded(container.NewPadded(grid)),
	)
}
