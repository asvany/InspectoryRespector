package tray

import (
	// "os"

	"log"
	"os"
	"runtime"
	"sync"

	"github.com/dawidd6/go-appindicator"
	// "github.com/gotk3/gotk3/glib"
	"github.com/asvany/InspectoryRespector/tracker"
	"github.com/gotk3/gotk3/gtk"
)

var enabled bool

var icon_file_enabled = "binary/icon_base_2.png"
var icon_file_disabled = "binary/icon_base_2_bw.png"

// change appindicator icon
func SetIcon(indicator *appindicator.Indicator, enabled bool) {
	if enabled {
		indicator.SetIcon(icon_file_enabled)
	} else {
		indicator.SetIcon(icon_file_disabled)
	}
}

// var image = trayhost.NewImageFromPath(icon_file)

func InitTray(stopChan chan os.Signal, wg *sync.WaitGroup, inputEventCollector *tracker.InputEventCollector) {
	defer func() {

		wg.Done()

		log.Println("InitTray Done\n")
	}()
	enabled = true
	log.Printf("InitTray\n")
	runtime.LockOSThread()

	gtk.Init(nil)

	// Create an instance of glib.Application
	// application := glib.ApplicationNew("com.example.app", glib.APPLICATION_FLAGS_NONE)

	indicator := appindicator.New("Tracker", "indicator-messages", appindicator.CategoryApplicationStatus)
	indicator.SetStatus(appindicator.StatusActive)

	menu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal("Unable to create menu:", err)
	}

	item, err := gtk.MenuItemNewWithLabel("enable/disable")
	if err != nil {
		log.Fatal("Unable create menu item")
	}
	item.Connect("activate", func() {
		inputEventCollector.Enabled = !enabled
		if inputEventCollector.Enabled {
			inputEventCollector.Enabled = false
			log.Printf("InputEventCollector disabled\n")
		} else {
			inputEventCollector.Enabled = true
			log.Printf("InputEventCollector enabled\n")
		}
		SetIcon(indicator, inputEventCollector.Enabled)
	})
	menu.Append(item)

	item, err = gtk.MenuItemNewWithLabel("Quit")
	if err != nil {
		log.Fatal("Unable create menu item")
	}
	item.Connect("activate", func() {
		log.Printf("Quit Application Menuitem pressed, sent stop\n")
		stopChan <- os.Interrupt
	})
	menu.Append(item)

	menu.ShowAll()
	indicator.SetMenu(menu)
	SetIcon(indicator, inputEventCollector.Enabled)

	log.Printf("Indicator created")
	go func() {
		log.Printf("GTK main started\n")
		gtk.Main()
		log.Printf("GTK main finished\n")
	}()

	<-stopChan
	gtk.MainQuit()
	log.Printf("GTK main quit called\n")
	wg.Done()

}
