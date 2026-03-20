package main

import (
	"embed"

	"QuraMate/internal/app"
	updatepkg "QuraMate/internal/updater"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var AppVersion = "dev"

func main() {
	updatepkg.AppVersion = AppVersion

	// Create an instance of the app structure
	appService := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "QuraMate",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        appService.Startup,
		Bind: []interface{}{
			appService,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "quramate-8d3b-4f22",
			OnSecondInstanceLaunch: appService.OnSecondInstanceLaunch,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
