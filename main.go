package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"niumer/internal/config"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := config.Load(embeddedAppConfig); err != nil {
		log.Fatalf("config: %v", err)
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "niumer",
		Width:  1280,
		Height: 800,
		MinWidth:  800,
		MinHeight: 480,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
