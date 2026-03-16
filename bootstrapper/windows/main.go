package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "QuraMate Setup",
		Width:            1060,
		Height:           680,
		MinWidth:         980,
		MinHeight:        640,
		DisableResize:    false,
		Frameless:        false,
		WindowStartState: options.Normal,
		BackgroundColour: &options.RGBA{R: 248, G: 242, B: 235, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
