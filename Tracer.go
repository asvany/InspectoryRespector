// this application tracking and logging the user activity
// periodically poll the keyboard and mouse activity status and detect idle periods and window focus changes
// if the user is active periodically make a blurred screenshot and stores this and the activity data in a local folder

package main

//go:generate protoc --go_out=. --python_out=ir_protocol_py ir_record.proto

import (
	// "encoding/binary"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/asvany/InspectoryRespector/tracker"
	"github.com/asvany/InspectoryRespector/xwindow"

	"github.com/asvany/InspectoryRespector/xinputhandler"

	"github.com/asvany/InspectoryRespector/common"
	"github.com/asvany/InspectoryRespector/geo"
	"github.com/asvany/InspectoryRespector/tray"

	// "google.golang.org/protobuf/encoding/protojson"
	// "google.golang.org/protobuf/proto"

	"github.com/asvany/ChannelWithConcurrentSenders/cc"

	"github.com/asvany/InspectoryRespector/ir_protocol"
)

func getHostname() string {
	hostname, err := os.Hostname()

	if err != nil {
		fmt.Printf("Error retrieving hostname: %v\n", err)
	} else {
		fmt.Println("Hostname:", hostname)
	}
	return hostname
}

func locationHandling() chan *ir_protocol.Location {
	loc_chan := make(chan *ir_protocol.Location)

	ticker := time.NewTicker(100 * time.Second)

	go func() {
		go geo.GetLocation(loc_chan)
		for range ticker.C {
			fmt.Println("Function runs every 100 seconds")
			go geo.GetLocation(loc_chan)

		}
	}()
	return loc_chan
}

// this is the main file and this is the start function of the application
func main() {
	fmt.Println("Hello World: v1.2")
	common.InitEnv("")

	out_dir := os.Getenv("DUMP_DIR")
	HOME := os.Getenv("HOME")
	if out_dir == "" {
		out_dir = HOME + "/InspecptryRespector_dumps"
	}
	out_dir, err := filepath.Abs(out_dir)
	if err != nil {
		log.Println("ERROR: ", err)
	}

	tmp_dir := os.Getenv("IR_TMP_DIR")
	if tmp_dir == "" {
		tmp_dir = HOME + "/.cache/InspecptryRespector"
	}
	tmp_dir, err = filepath.Abs(tmp_dir)
	if err != nil {
		log.Println("ERROR: ", err)
	}

	// check the DISPLAY environment variable

	loc_chan := locationHandling()

	var wg sync.WaitGroup
	input_events := cc.NewChannelWithConcurrentSenders[xinputhandler.InputEvent](10)

	hostname := getHostname()

	inputEventCollector := &tracker.InputEventCollector{
		Last_fp:  "",
		Hostname: hostname,
		Location: nil,
		Out_dir:  out_dir,
		Tmp_dir:  tmp_dir,
	}
	fmt.Printf("out_dir:%v\n", out_dir)
	fmt.Printf("tmp_dir:%v\n", tmp_dir)

	inputEventCollector.Window = inputEventCollector.GetNewWindow(&xwindow.WinProps{})

	wc_ticker := time.NewTicker(100 * time.Millisecond)

	go func() {
		for range wc_ticker.C {
			go inputEventCollector.ProcessWindowFocusChanges()
		}
	}()

	cache_updater_ticker := time.NewTicker(60 * time.Second)

	go func() {
		for range cache_updater_ticker.C {
			go inputEventCollector.GetFileNameAndOrganizeFiles(false)
		}
	}()

	quitChan := make(chan bool)

	go inputEventCollector.ProcessInputEvents(input_events)
	go inputEventCollector.ProcessLocation(loc_chan)

	xinputhandler.SetupInput(quitChan, &wg, input_events)

	wg.Add(1)
	tray.InitTray(quitChan, &wg, inputEventCollector)

	ctx_ctrl_c, cancel_wait_ctrl_c := context.WithCancel(context.Background())

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	// signal.Notify(stopChan, os.Interrupt)
	go func() {
		for {
			select {
			case <-stopChan:
				{
					fmt.Println("received stop signal")
					close(quitChan)
				}
			case <-ctx_ctrl_c.Done():
				{
					log.Printf("Conext Done\n")
					return
				}
			}
		}
	}()

	fmt.Println("WAIT")

	wg.Wait()
	fmt.Println("WAIT END")
	cancel_wait_ctrl_c()
	input_events.Wait()
	fmt.Println("LAST WRITE")
	err = inputEventCollector.WriteToFile()
	inputEventCollector.GetFileNameAndOrganizeFiles(true)
	if err != nil {
		log.Println("ERROR: ", err)
	}

	//sleep 1sec before exit
	fmt.Println("SLEEP")
	time.Sleep(time.Second * 1)
	fmt.Println("EXIT")
}
