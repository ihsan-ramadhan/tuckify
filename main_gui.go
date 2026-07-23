//go:build desktop

package main

import (
	"embed"
	"os"

	"github.com/ihsan-ramadhan/tuckify/cmd"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailslinux "github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:gui/frontend/dist
var assets embed.FS

//go:embed gui/frontend/src/assets/images/logo.png
var appIcon []byte

// knownCLICommands is the set of subcommands that should be dispatched
// to the CLI cobra handler instead of launching the GUI.
var knownCLICommands = map[string]bool{
	"run": true, "schedule": true, "list": true,
	"start": true, "stop": true, "restart": true,
	"logs": true, "delete": true, "edit": true,
	"init": true, "uninstall": true, "validate": true,
	"help": true, "--help": true, "-h": true,
	"--version": true, "-v": true,
}

func main() {
	// Dual-mode: if invoked with a CLI subcommand, run as CLI
	if len(os.Args) > 1 && knownCLICommands[os.Args[1]] {
		cmd.Execute()
		return
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "tuckify-gui",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Linux: &wailslinux.Options{
			Icon:        appIcon,
			ProgramName: "tuckify",
		},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
