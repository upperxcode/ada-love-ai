package components

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"ada-love-ai/backend"
	adaTheme "ada-love-ai/frontend/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SkillsHub struct {
	engine *backend.Engine
	tabs   *container.AppTabs

	// Stats
	totalLabel      *widget.Label
	thirdPartyLabel *widget.Label

	// Management View
	mgmtGrid   *fyne.Container
	mgmtSearch *widget.Entry
	mgmtSplit  *container.Split

	// Details Panel
	detailsBox     *fyne.Container
	detailsIcon    *widget.Label
	detailsTitle   *widget.Label
	detailsSummary *widget.Label
	detailsReg     *widget.Label
	detailsVer     *widget.Label
	detailsURL     *widget.Label
	detailsPreview *fyne.Container // Container para renderização híbrida (RichText + Tables)
	detailsScroll  *container.Scroll
	detailsRaw     *widget.Entry
	detailsMeta    *fyne.Container
	detailsStack   *fyne.Container // Para alternar entre Preview, Raw e Metadata
	detailsTabs    *fyne.Container // Seletor de modo customizado

	// Hub View
	hubGrid   *fyne.Container
	hubSearch *widget.Entry
	loading   *widget.ProgressBarInfinite

	CanvasObject fyne.CanvasObject

	mu sync.Mutex
}

func NewSkillsHub(engine *backend.Engine) *SkillsHub {
	h := &SkillsHub{
		engine: engine,
	}

	// 1. Setup Stats
	h.totalLabel = widget.NewLabel("0")
	h.thirdPartyLabel = widget.NewLabel("0")

	// 2. Setup Loading
	h.loading = widget.NewProgressBarInfinite()
	h.loading.Hide()

	// 3. Setup Details UI
	h.detailsIcon = widget.NewLabel(adaTheme.IconInfo)
	h.detailsTitle = widget.NewLabelWithStyle("Detalhes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	h.detailsSummary = widget.NewLabel("Carregando resumo...")
	h.detailsSummary.Wrapping = fyne.TextWrapWord
	h.detailsSummary.Importance = widget.LowImportance

	// Cards de Informação
	var regCard, verCard, urlCard fyne.CanvasObject
	h.detailsReg, regCard = h.createInfoCard("REGISTRY", "clawhub")
	h.detailsVer, verCard = h.createInfoCard("INSTALLED VERSION", "1.0.0")
	h.detailsURL, urlCard = h.createInfoCard("URL", "https://...")

	infoCards := container.NewGridWithColumns(2,
		regCard,
		verCard,
	)
	fullURLCard := container.NewVBox(urlCard)

	// Modos de Visualização
	h.detailsPreview = container.NewVBox()
	h.detailsScroll = container.NewVScroll(container.NewPadded(h.detailsPreview))

	h.detailsRaw = widget.NewMultiLineEntry()
	h.detailsRaw.Disable()
	h.detailsRaw.TextStyle = fyne.TextStyle{Monospace: true}

	h.detailsMeta = container.NewVBox()

	h.detailsStack = container.NewStack(h.detailsScroll)

	// Seletor de Modos (Tabs customizadas)
	btnPreview := widget.NewButton("Preview", nil)
	btnRaw := widget.NewButton("Raw", nil)
	btnMeta := widget.NewButton("Metadata", nil)

	updateTabStyles := func(active string) {
		btnPreview.Importance = widget.LowImportance
		btnRaw.Importance = widget.LowImportance
		btnMeta.Importance = widget.LowImportance
		switch active {
		case "p":
			btnPreview.Importance = widget.HighImportance
		case "r":
			btnRaw.Importance = widget.HighImportance
		case "m":
			btnMeta.Importance = widget.HighImportance
		}
		btnPreview.Refresh()
		btnRaw.Refresh()
		btnMeta.Refresh()
	}

	btnPreview.OnTapped = func() {
		h.detailsStack.Objects = []fyne.CanvasObject{h.detailsScroll}
		h.detailsStack.Refresh()
		updateTabStyles("p")
	}
	btnRaw.OnTapped = func() {
		h.detailsStack.Objects = []fyne.CanvasObject{h.detailsRaw}
		h.detailsStack.Refresh()
		updateTabStyles("r")
	}
	btnMeta.OnTapped = func() {
		h.detailsStack.Objects = []fyne.CanvasObject{h.detailsMeta}
		h.detailsStack.Refresh()
		updateTabStyles("m")
	}
	updateTabStyles("p")

	h.detailsTabs = container.NewHBox(btnPreview, btnRaw, btnMeta)

	closeBtn := widget.NewButton(adaTheme.IconClose, func() {
		h.detailsBox.Hide()
		h.mgmtSplit.Offset = 1.0
		h.mgmtSplit.Refresh()
	})
	closeBtn.Importance = widget.LowImportance

	// Assemble Header
	detailsHeader := container.NewVBox(
		container.NewBorder(nil, nil, h.detailsIcon, closeBtn, h.detailsTitle),
		h.detailsSummary,
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(infoCards, fullURLCard)),
		container.NewPadded(h.detailsTabs),
	)

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 20}
	bg.StrokeWidth = 1
	// Removido SetMinSize fixo para permitir o fechamento total do Split

	h.detailsBox = container.NewStack(bg, container.NewBorder(detailsHeader, nil, nil, nil, container.NewPadded(h.detailsStack)))

	// 4. Build Management View
	h.mgmtGrid = container.New(layout.NewGridWrapLayout(fyne.NewSize(320, 140)))
	h.mgmtSearch = widget.NewEntry()
	h.mgmtSearch.SetPlaceHolder("Filtrar skills instaladas...")
	h.mgmtSearch.OnChanged = func(s string) {
		h.RefreshLocal(s)
	}

	mgmtContent := h.buildManagementView()

	// 5. Build Hub View
	h.hubGrid = container.New(layout.NewGridWrapLayout(fyne.NewSize(320, 240)))
	h.hubSearch = widget.NewEntry()
	h.hubSearch.SetPlaceHolder("Buscar novas skills na hub...")
	h.hubSearch.OnSubmitted = func(s string) {
		h.RefreshHub(s)
	}

	hubContent := h.buildHubView()

	// 6. Setup Tabs
	h.tabs = container.NewAppTabs(
		container.NewTabItem(adaTheme.IconStorage+" Gerenciar", mgmtContent),
		container.NewTabItem(adaTheme.IconSearch+" Hub", hubContent),
	)
	h.tabs.SetTabLocation(container.TabLocationTop)

	// 7. Header
	lblTitle := widget.NewLabelWithStyle("Skills", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	iconTitle := widget.NewLabel(adaTheme.IconSettings)
	left := container.NewHBox(iconTitle, lblTitle)

	installBtn := widget.NewButton(adaTheme.IconAdd+" Import Skill", func() {
		h.tabs.SelectIndex(1)
		h.hubSearch.FocusGained()
	})
	installBtn.Importance = widget.HighImportance

	header := NewHeaderBar(left, nil, installBtn)

	h.CanvasObject = container.NewBorder(
		header.CanvasObject,
		nil, nil, nil,
		h.tabs,
	)

	// Inicia a carga de dados com segurança
	go func() {
		// Espera o app inicializar completamente
		for fyne.CurrentApp() == nil {
			time.Sleep(100 * time.Millisecond)
		}
		time.Sleep(500 * time.Millisecond) // Margem extra
		h.RefreshLocal("")
	}()

	return h
}

func (h *SkillsHub) buildManagementView() fyne.CanvasObject {
	totalCard := h.createStatCard("TOTAL SKILLS", h.totalLabel, adaTheme.IconDocument)
	thirdCard := h.createStatCard("THIRD-PARTY", h.thirdPartyLabel, adaTheme.IconInfo)
	statsRow := container.NewGridWithColumns(2, totalCard, thirdCard)

	// Input de busca com largura expansível
	h.mgmtSearch.SetPlaceHolder("Filtrar meus agentes...")
	h.mgmtSearch.OnChanged = func(s string) {
		h.RefreshLocal(s)
	}

	searchRow := container.NewPadded(container.NewBorder(
		nil, nil,
		adaTheme.NewIcon(adaTheme.IconSearch, adaTheme.SizeControlBig),
		widget.NewSelect([]string{"All Types", "System", "Third-party"}, func(s string) {}),
		h.mgmtSearch,
	))

	topContent := container.NewVBox(
		container.NewPadded(statsRow),
		searchRow,
		widget.NewSeparator(),
	)

	gridScroll := container.NewVScroll(container.NewPadded(h.mgmtGrid))

	h.mgmtSplit = container.NewHSplit(gridScroll, h.detailsBox)
	h.mgmtSplit.Offset = 1.0

	return container.NewBorder(
		topContent,
		nil, nil, nil,
		h.mgmtSplit,
	)
}

func (h *SkillsHub) showDetails(name string) {
	fyne.Do(func() { h.loading.Show() })
	go func() {
		info, err := h.engine.GetSkillFullInfo(name)
		fyne.Do(func() { h.loading.Hide() })

		fyne.Do(func() {
			h.mu.Lock()
			h.detailsTitle.SetText(strings.ToUpper(name))
			if info != nil {
				h.detailsSummary.SetText(info.Description)
				h.detailsReg.SetText(info.Registry)
				h.detailsVer.SetText(info.Version)
				h.detailsURL.SetText(info.URL)
				h.renderMarkdownHybrid(info.Markdown)
				h.detailsRaw.SetText(info.Raw)

				// Metadata Cards
				h.detailsMeta.Objects = []fyne.CanvasObject{
					container.NewGridWithColumns(2,
						h.createMetaCard("NAME", info.Name),
						h.createMetaCard("LINE COUNT", fmt.Sprintf("%d", info.LineCount)),
						h.createMetaCard("CHAR COUNT", fmt.Sprintf("%d", info.CharCount)),
					),
					h.createMetaCard("DESCRIPTION", info.Description),
				}
				h.detailsMeta.Refresh()
			}

			if err != nil {
				h.detailsSummary.SetText("Erro ao carregar detalhes da skill.")
				errText := widget.NewRichText()
				errText.ParseMarkdown(fmt.Sprintf("> ⚠️ **Erro**: %v", err))
				h.detailsPreview.Objects = []fyne.CanvasObject{errText}
			}
			h.mu.Unlock()

			h.detailsPreview.Refresh()
			h.detailsScroll.Offset = fyne.NewPos(0, 0)
			h.detailsScroll.Refresh()

			h.detailsBox.Show()
			h.mgmtSplit.Offset = 0.6 // 60% para a esquerda, 40% para os detalhes
			h.mgmtSplit.Refresh()
		})
	}()
}

func (h *SkillsHub) buildHubView() fyne.CanvasObject {
	title := canvas.NewText("Discover Skills", color.White)
	title.TextSize = 36
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabelWithStyle("Encontre novas capacidades na hub oficial", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	subtitle.Importance = widget.LowImportance

	// Busca da Hub agora responsiva
	h.hubSearch.SetPlaceHolder("Buscar novas skills na hub...")
	searchContainer := container.NewBorder(nil, nil, nil,
		widget.NewButton(adaTheme.IconSearch+" Search", func() {
			h.RefreshHub(h.hubSearch.Text)
		}),
		h.hubSearch,
	)

	searchRow := container.NewPadded(container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(500, 40)), searchContainer)))

	header := container.NewVBox(
		layout.NewSpacer(),
		title,
		subtitle,
		container.NewPadded(searchRow),
		container.NewCenter(h.loading),
		layout.NewSpacer(),
	)

	return container.NewBorder(
		header,
		nil, nil, nil,
		container.NewVScroll(container.NewPadded(h.hubGrid)),
	)
}

func (h *SkillsHub) createStatCard(label string, value *widget.Label, iconStr string) fyne.CanvasObject {
	l := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Importance = widget.LowImportance
	value.TextStyle = fyne.TextStyle{Bold: true}

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 12
	bg.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 15}
	bg.StrokeWidth = 1

	iconView := adaTheme.NewIcon(iconStr, adaTheme.SizeCardBig)

	top := container.NewHBox(l, layout.NewSpacer())
	bottom := container.NewHBox(value, layout.NewSpacer(), container.NewPadded(iconView))

	content := container.NewVBox(top, bottom)

	return container.NewStack(bg, container.NewPadded(content))
}

func (h *SkillsHub) RefreshLocal(filter string) {
	go func() {
		installed, _ := h.engine.GetInstalledSkills()
		var objects []fyne.CanvasObject
		for _, name := range installed {
			if filter != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
				continue
			}
			objects = append(objects, h.createMgmtCard(name))
		}

		fyne.Do(func() {
			h.mu.Lock()
			h.totalLabel.SetText(fmt.Sprintf("%d", len(installed)))
			h.thirdPartyLabel.SetText(fmt.Sprintf("%d", len(installed)))
			h.mgmtGrid.Objects = objects
			h.mu.Unlock()

			h.mgmtGrid.Refresh()
		})
	}()
}

func (h *SkillsHub) RefreshHub(query string) {
	if query == "" {
		return
	}
	fyne.Do(func() { h.loading.Show() })

	go func() {
		installed, _ := h.engine.GetInstalledSkills()
		installedMap := make(map[string]bool)
		for _, s := range installed {
			installedMap[strings.ToLower(s)] = true
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		results, err := h.engine.SearchSkills(ctx, query)
		var objects []fyne.CanvasObject
		if err != nil {
			objects = append(objects, widget.NewLabel(fmt.Sprintf("Erro: %v", err)))
		} else if len(results) == 0 {
			objects = append(objects, widget.NewLabel("Nenhuma skill encontrada."))
		} else {
			for _, res := range results {
				objects = append(objects, h.createRemoteCard(res, installedMap))
			}
		}

		fyne.Do(func() {
			h.mu.Lock()
			h.hubGrid.Objects = objects
			h.mu.Unlock()

			h.loading.Hide()
			h.hubGrid.Refresh()
		})
	}()
}

func (h *SkillsHub) createMgmtCard(name string) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Truncation = fyne.TextTruncateEllipsis

	// Botão de Informações/Ver Detalhes
	infoBtnStyled := adaTheme.NewIconButton(adaTheme.IconView, adaTheme.SizeControlBig, func() {
		h.showDetails(name)
	})

	// Botão de Apagar (apenas ícone)
	deleteBtnStyled := adaTheme.NewIconButton(adaTheme.IconDelete, adaTheme.SizeControlBig, func() {
		var win fyne.Window
		if app := fyne.CurrentApp(); app != nil && len(app.Driver().AllWindows()) > 0 {
			win = app.Driver().AllWindows()[0]
		}
		dialog.ShowConfirm("Remover Skill", "Deseja realmente remover esta skill?", func(ok bool) {
			if ok {
				h.engine.UninstallSkill(name)
				h.RefreshLocal(h.mgmtSearch.Text)
			}
		}, win)
	})

	actions := container.NewHBox(infoBtnStyled, deleteBtnStyled)

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 12
	bg.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 10}
	bg.StrokeWidth = 1

	// BorderLayout garante que o título ocupe o espaço e os ícones fiquem à direita
	header := container.NewBorder(nil, nil, nil, actions, title)
	content := container.NewVBox(header, widget.NewLabelWithStyle("Instalada localmente", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}))

	return container.NewStack(bg, container.NewPadded(content))
}

func (h *SkillsHub) createRemoteCard(res backend.SearchResult, installed map[string]bool) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(res.DisplayName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	regBadge := widget.NewLabelWithStyle(strings.ToUpper(res.RegistryName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	regBadge.Importance = widget.WarningImportance
	summary := widget.NewLabel(truncateString(res.Summary, 150))
	summary.Wrapping = fyne.TextWrapWord
	summary.Importance = widget.LowImportance

	installBtn := widget.NewButton(adaTheme.IconAdd, func() {
		h.installSkill(res)
	})

	if installed[strings.ToLower(res.DisplayName)] || installed[strings.ToLower(res.Slug)] {
		installBtn.SetText(adaTheme.IconCheck)
		installBtn.Disable()
	} else {
		installBtn.Importance = widget.HighImportance
	}

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 12
	bg.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 25}
	bg.StrokeWidth = 1

	top := container.NewHBox(title, regBadge, layout.NewSpacer(), installBtn)
	content := container.NewVBox(top, summary)

	return container.NewStack(bg, container.NewPadded(content))
}

func (h *SkillsHub) installSkill(res backend.SearchResult) {
	fyne.Do(func() { h.loading.Show() })
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		err := h.engine.InstallSkill(ctx, res.RegistryName, res.Slug, res.Version)

		fyne.Do(func() {
			h.loading.Hide()
			if err == nil {
				h.RefreshLocal("")
				h.RefreshHub(h.hubSearch.Text)
			}
		})
	}()
}

func (h *SkillsHub) createInfoCard(label, initialValue string) (*widget.Label, fyne.CanvasObject) {
	l := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Importance = widget.LowImportance

	val := widget.NewLabel(initialValue)
	val.TextStyle = fyne.TextStyle{Bold: true}

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 8
	bg.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 15}
	bg.StrokeWidth = 1

	return val, container.NewStack(bg, container.NewPadded(container.NewVBox(l, val)))
}

func (h *SkillsHub) createMetaCard(label, value string) fyne.CanvasObject {
	l := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Importance = widget.LowImportance

	val := widget.NewLabel(value)
	val.TextStyle = fyne.TextStyle{Bold: true}
	val.Wrapping = fyne.TextWrapWord

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 12
	bg.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 10}
	bg.StrokeWidth = 1

	return container.NewStack(bg, container.NewPadded(container.NewVBox(l, val)))
}

func (h *SkillsHub) renderMarkdownHybrid(md string) {
	h.detailsPreview.Objects = nil

	// 1. Limpeza básica de frontmatter (YAML)
	if strings.HasPrefix(md, "---") {
		parts := strings.SplitN(md, "---", 3)
		if len(parts) >= 3 {
			md = parts[2]
		}
	}
	md = strings.ReplaceAll(md, "\r\n", "\n")

	lines := strings.Split(md, "\n")
	var currentText []string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		// Detecta início de tabela (| Col |)
		if strings.HasPrefix(line, "|") && strings.Count(line, "|") > 1 {
			// Renderiza texto acumulado anterior
			if len(currentText) > 0 {
				rt := widget.NewRichText()
				rt.Wrapping = fyne.TextWrapWord
				rt.ParseMarkdown(strings.Join(currentText, "\n"))
				h.detailsPreview.Add(rt)
				currentText = nil
			}

			// Coleta todas as linhas da tabela
			var tableLines []string
			for ; i < len(lines); i++ {
				l := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(l, "|") && l != "" {
					break
				}
				if l == "" && len(tableLines) > 0 {
					break
				}
				tableLines = append(tableLines, l)
			}
			i--

			// Cria o widget de tabela robusto
			h.detailsPreview.Add(h.createTableWidget(tableLines))
		} else {
			currentText = append(currentText, lines[i])
		}
	}

	// Renderiza texto restante
	if len(currentText) > 0 {
		rt := widget.NewRichText()
		rt.Wrapping = fyne.TextWrapWord
		rt.ParseMarkdown(strings.Join(currentText, "\n"))
		h.detailsPreview.Add(rt)
	}
}

func (h *SkillsHub) createTableWidget(lines []string) fyne.CanvasObject {
	var data [][]string
	for _, line := range lines {
		if strings.Contains(line, "---") || strings.Contains(line, "===") || line == "" {
			continue
		}
		cells := strings.Split(line, "|")
		var row []string
		for _, c := range cells {
			trimmed := strings.TrimSpace(c)
			row = append(row, trimmed)
		}
		if len(row) > 0 && row[0] == "" {
			row = row[1:]
		}
		if len(row) > 0 && row[len(row)-1] == "" {
			row = row[:len(row)-1]
		}
		if len(row) > 0 {
			data = append(data, row)
		}
	}

	if len(data) == 0 {
		return widget.NewLabel("Tabela inválida")
	}

	rows, cols := len(data), len(data[0])
	table := widget.NewTable(
		func() (int, int) { return rows, cols },
		func() fyne.CanvasObject {
			l := widget.NewLabel("Cell Content")
			l.Wrapping = fyne.TextWrapWord
			return l
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if id.Row < len(data) && id.Col < len(data[id.Row]) {
				label.SetText(data[id.Row][id.Col])
			}
			if id.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)

	// Ajuste de colunas
	for i := 0; i < cols; i++ {
		table.SetColumnWidth(i, 220)
	}

	// ALTURA CONTROLADA: Crucial para não travar o layout
	// 40px por linha + respiro. Limitamos a 400px para tabelas gigantes.
	height := float32(rows)*40 + 10
	if height > 400 {
		height = 400
	}

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 20}
	bg.StrokeWidth = 1
	bg.CornerRadius = 8

	// Usamos um container com tamanho fixo para a tabela
	// Isso garante que o VScroll pai saiba exatamente quanto de espaço reservar
	return container.NewStack(
		bg,
		container.NewPadded(container.New(layout.NewGridWrapLayout(fyne.NewSize(500, height)), table)),
	)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
