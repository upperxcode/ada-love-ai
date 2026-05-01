package theme

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/fonts/JetBrainsMonoNerdFont-Regular.ttf
var fontData []byte

var JetBrainsNerdFont = fyne.NewStaticResource("JetBrainsMonoNerdFont-Regular.ttf", fontData)
