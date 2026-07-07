package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"ada-love-ai/backend"
	"ada-love-ai/pkg/envutil"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var frontendAssets embed.FS

func main() {
	// Load project-local .env files so API keys stored as env-var references
	// (e.g. OPENROUTER_API_KEY) resolve even without a real env entry.
	envutil.LoadEnvFiles()

	engine, err := backend.NewEngine()
	if err != nil {
		fmt.Printf("Erro ao inicializar o motor: %v\n", err)
		os.Exit(1)
	}
	defer engine.Close()

	app := NewApp(engine)

	assets, _ := fs.Sub(frontendAssets, "frontend/dist")

	err = wails.Run(&options.App{
		Title:  "Ada Love AI",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fmt.Printf("Erro ao iniciar o app: %v\n", err)
		os.Exit(1)
	}
}
