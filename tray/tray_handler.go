package tray

// #cgo pkg-config: gtk+-3.0

// https://developer.fyne.io/explore/canvas

import (
	// "os"

	"log"
	"sync"

	"github.com/asvany/InspectoryRespector/tracker"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	// "fyne.io/fyne/v2/widget"
)

var icon_file_enabled = "binary/icon_base_2.ico"
var icon_file_disabled = "binary/icon_base_2_bw.ico"

// var image = trayhost.NewImageFromPath(icon_file)

func InitTray(quitChan chan bool, wg *sync.WaitGroup, inputEventCollector *tracker.InputEventCollector) {

	icon_enabled_resource, err := fyne.LoadResourceFromPath(icon_file_enabled)
	if err != nil {
		log.Printf("Error loading icon_enabled: %v", err)
	}
	icon_disabled_resource, err := fyne.LoadResourceFromPath(icon_file_disabled)
	if err != nil {
		log.Printf("Error loading icon_disabled: %v", err)
	}

	// icon_enabled := widget.NewIcon(icon_disabled_resource)
	// icon_disabled := widget.NewIcon(icon_enabled_resource)

	log.Printf("InitTray\n")
	a := app.New()

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Enable/Disable", func() {
				if inputEventCollector.Enabled {
					inputEventCollector.Enabled = false
					desk.SetSystemTrayIcon(icon_disabled_resource)
					a.SetIcon(icon_disabled_resource)

				} else {
					inputEventCollector.Enabled = true
					desk.SetSystemTrayIcon(icon_enabled_resource)
					a.SetIcon(icon_enabled_resource)

				}

			}))
		desk.SetSystemTrayMenu(m)
		desk.SetSystemTrayIcon(icon_enabled_resource)
		a.SetIcon(icon_enabled_resource)

	}

	// w.ShowAndRun()
	a.Run()
	// application := glib.ApplicationNew("com.example.app", glib.APPLICATION_FLAGS_NONE)

	go func() {
		<-quitChan

		log.Printf("Stop signal received in GTK quit handler\n")
		a.Quit()
		wg.Done()
		// call quit when the main loop is idle
		// glib.IdleAdd(func() {
		// 	log.Printf("GTK main quit calling from IdleAdd\n")
		// 	gtk.MainQuit()

		// })
	}()

	log.Printf("Indicator created")
	go func() {
		log.Printf("GTK main started\n")

		log.Printf("GTK main finished\n")

	}()

}
