// this application tracking and logging the user activity
// periodically poll the keyboard and mouse activity status and detect idle periods and window focus changes
// if the user is active periodically make a blurred screenshot and stores this and the activity data in a local folder

package main

//go:generate protoc --go_out=. --python_out=ir_protocol_py ir_record.proto

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/asvany/InspectoryRespector/xwindow"

	"github.com/asvany/InspectoryRespector/common"
	"github.com/asvany/InspectoryRespector/geo"

	// "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/asvany/ChannelWithConcurrentSenders/cc"

	"github.com/asvany/InspectoryRespector/ir_protocol"
)

type InputEventCollector struct {
	window   *ir_protocol.WindowChange // current window input data
	last_fp  string                    // last fingerprint
	hostname string                    // hostname
	location *ir_protocol.Location     // location
	out_dir  string                    // output directory
	tmp_dir  string                    // temporary directory
}

func (c *InputEventCollector) processLocation(loc_chan chan *ir_protocol.Location) {
	for loc := range loc_chan {
		fmt.Println("New Location:", loc.String())
		c.location = loc
		c.window.Location = c.location
	}
}

func (c *InputEventCollector) processInputEvents(input_events InputEventsChannelType) {
	for event := range input_events.ROChannel() {
		// fmt.Println("Event:", event.String())
		//switch by event type
		dev, ok := c.window.Events[event.GetDeviceId()]
		if !ok {
			dev = &ir_protocol.DeviceEvents{
				DeviceName: event.GetDeviceName(),
			}
			c.window.Events[event.GetDeviceId()] = dev
		}
		// var event interface{} = &ButtonEventData{InputEventData: InputEventData{DeviceId: 1, DeviceName: "Device1", EventType: 1, timestamp: time.Now()}, Button: 2}

		switch v := event.(type) {
		case *ButtonEventData:
			// fmt.Printf("ButtonEvent with button: %d\n", v.Button)
			dev.ButtonEvents = append(dev.ButtonEvents, &ir_protocol.ButtonEvent{
				Timestamp:  timestamppb.New(v.timestamp),
				ButtonCode: uint32(v.Button),
				IsDown:     v.EventType&1 == 0,
			})
		case *KeyEventData:
			// fmt.Printf("KeyEvent with keycode: %d\n", v.KeyCode)
			dev.KeyEvents = append(dev.KeyEvents, &ir_protocol.KeyEvent{
				Timestamp: timestamppb.New(v.timestamp),
				KeyCode:   uint32(v.KeyCode),
				IsDown:    v.EventType&1 == 0,
			})
		case *MotionEvent:
			// fmt.Printf("MotionEvent with axis position: %v\n", v.AxisPosition)
			motion_event := &ir_protocol.MotionEvent{
				Timestamp:     timestamppb.New(v.timestamp),
				AxisPositions: make(map[uint32]int32),
			}
			for axis, position := range v.AxisPosition {
				motion_event.AxisPositions[axis] = position
			}
			dev.MotionEvents = append(dev.MotionEvents, motion_event)

		default:
			fmt.Printf("Unknown event type: %T\n", v)
		}
	}

}

// this function is the most importan function of the application, it is periodicaly check the window focus changes and when the focus changed the function logs the current record and start a new one
func (c *InputEventCollector) processWindowFocusChanges() {
	xi, err := xwindow.NewXInfo()
	if err != nil {
		log.Printf("ERROR:%v\n", err)
		return
	}
	defer xi.Close()
	propValues := make(xwindow.WinProps)

	fp, err := xi.GetFullKey(&propValues)
	if err != nil {
		log.Printf("ERROR:%v\n", err)
		return
	}

	if fp != c.last_fp {
		fmt.Printf("new fp:%v\n", fp)
		fmt.Printf("old fp:%v\n", c.last_fp)
		// for k, v := range propValues {
		// log.Printf("k:%v v:%v\n", k, v)
		// }
		if len(c.window.Properties) != 0 {
			fmt.Println("write to file")
			c.WriteToFile()
		}
		c.window = c.getNewWindow(&propValues)
		c.last_fp = fp
	}
}
var mu sync.Mutex

// this method move the temporaly dumps to the output directory if it neccessary and returns the current filename
func (c * InputEventCollector) GetFileNameAndOrganizeFiles() string{
	mu.Lock()
	defer mu.Unlock()
	
	currentTS := time.Now().Format("2006_01_02-15")
	currentDay := time.Now().Format("2006_01_02")
	out_dir := c.out_dir + "/" + c.hostname + "/" + currentDay

	filename_base := "data_" + c.hostname + "_" + currentTS + ".protodump"

	filename := c.tmp_dir + "/" + filename_base
	os.MkdirAll(c.tmp_dir, 0755)


	//check that filenname_base is exist in the tmp_dir
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// check that any file exist in the tmp_dir
		files, err := os.ReadDir(c.tmp_dir)
		if err != nil {
			log.Fatal(err)
		}
		if len(files) != 0 {
			// move all files to the out_dir
			os.MkdirAll(out_dir, 0755)
			for _, file := range files {
				fmt.Printf("move file:%v\n", file.Name())
				os.Rename(c.tmp_dir+"/"+file.Name(), out_dir+"/"+file.Name())
			}
		}
	}

	return filename




}

func (c *InputEventCollector) WriteToFile() error {
	

	filename := c.GetFileNameAndOrganizeFiles()
	data, err := proto.Marshal(c.window)
	if err != nil {
		return fmt.Errorf("marshalling error %v", err)
	} // Open file for appending
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening file  error %v", err)
	}
	defer f.Close()

	//calculate the size of the data
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	// Write size to file
	if _, err := f.Write(size); err != nil {
		return fmt.Errorf("writing error %v", err)
	}
	// Write data to file
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing error %v", err)
	}

	

	return nil

}

func (c *InputEventCollector) getNewWindow(props *xwindow.WinProps) *ir_protocol.WindowChange {
	window := ir_protocol.WindowChange{
		Timestamp:  timestamppb.New(time.Now()),
		Properties: *props,
		Events:     make(map[uint64]*ir_protocol.DeviceEvents),
		Location:   c.location,
		Hostname:   c.hostname,
	}
	return &window
}

// this is the main file and this is the start function of the application
func main() {
	fmt.Println("Hello World")
	common.InitEnv("")


	out_dir := os.Getenv("DUMP_DIR")
	HOME :=os.Getenv("HOME")
	if out_dir == "" {
		out_dir = HOME+"/InspecptryRespector_dumps"
	}
	out_dir, err := filepath.Abs(out_dir)
	if err != nil {
		log.Println("ERROR: ", err)
	}

	tmp_dir := os.Getenv("IR_TMP_DIR")
	if tmp_dir == "" {
		tmp_dir = HOME+"/.cache/InspecptryRespector"
	}
	tmp_dir, err = filepath.Abs(tmp_dir)
	if err != nil {
		log.Println("ERROR: ", err)
	}


	// check the DISPLAY environment variable

	loc_chan := locationHandling()

	var wg sync.WaitGroup
	input_events := cc.NewChannelWithConcurrentSenders[InputEvent](10)

	hostname := getHostname()

	InputEventCollector := &InputEventCollector{
		last_fp:  "",
		hostname: hostname,
		location: nil,
		out_dir:  out_dir,
		tmp_dir: tmp_dir,
	}
	fmt.Printf("out_dir:%v\n", out_dir)
	fmt.Printf("tmp_dir:%v\n", tmp_dir)


	InputEventCollector.window = InputEventCollector.getNewWindow(&xwindow.WinProps{})

	wc_ticker := time.NewTicker(100 * time.Millisecond)

	go func() {
		for range wc_ticker.C {
			go InputEventCollector.processWindowFocusChanges()
		}
	}()


	cache_updater_ticker := time.NewTicker(60 * time.Second)

	go func() {
		for range cache_updater_ticker.C {
			go InputEventCollector.GetFileNameAndOrganizeFiles()
		}
	}()

	stopChan := make(chan os.Signal, 1)

	signal.Notify(stopChan, os.Interrupt)

	go InputEventCollector.processInputEvents(input_events)
	go InputEventCollector.processLocation(loc_chan)

	SetupInput(stopChan, &wg, input_events)

	fmt.Print("WAIT")

	wg.Wait()
	input_events.Wait()
	fmt.Println("LAST WRITE")
	err = InputEventCollector.WriteToFile()
	if err != nil {
		log.Println("ERROR: ", err)
	}

	//sleep 1sec before exit
	fmt.Println("SLEEP")
	time.Sleep(time.Second * 1)
	fmt.Println("EXIT")
}


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
