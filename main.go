package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	// USE SEMPRE O NOME DO MÓDULO DEFINIDO NO GO.MOD
	"ada-love-ai/backend"
	"ada-love-ai/frontend/theme"
	"ada-love-ai/frontend/ui"
)

func main() {
	os.Setenv("FYNE_VIDEO", "wayland")
	myApp := app.New()
	myApp.Settings().SetTheme(&theme.MyTheme{})

	// Inicializa o motor Picoclaw
	engine, err := backend.NewEngine()
	if err != nil {
		fmt.Printf("Erro ao inicializar o motor: %v\n", err)
		os.Exit(1)
	}
	defer engine.Close()

	window := myApp.NewWindow("Ada-Love-Ai")
	window.SetContent(ui.CreateMainLayout(engine))

	window.Resize(fyne.NewSize(1400, 900))

	window.ShowAndRun()
}
