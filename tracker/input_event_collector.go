package tracker

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/asvany/InspectoryRespector/common"
	"github.com/asvany/InspectoryRespector/ir_protocol"
	"github.com/asvany/InspectoryRespector/xinputhandler"
	"github.com/asvany/InspectoryRespector/xwindow"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InputEventCollector struct {
	Window   *ir_protocol.WindowChange // current window input data
	Last_fp  string                    // last fingerprint
	Hostname string                    // hostname
	Location *ir_protocol.Location     // location
	Out_dir  string                    // output directory
	Tmp_dir  string                    // temporary directory
	Enabled  bool                      // enabled state
}

func (c *InputEventCollector) ProcessLocation(loc_chan chan *ir_protocol.Location) {
	for loc := range loc_chan {
		fmt.Println("New Location:", loc.String())
		c.Location = loc
		c.Window.Location = c.Location
	}
}

func (c *InputEventCollector) ProcessInputEvents(input_events xinputhandler.InputEventsChannelType) {
	for event := range input_events.ROChannel() {
		if c.Enabled {
			// fmt.Println("Event:", event.String())
			//switch by event type
			dev, ok := c.Window.Events[event.GetDeviceId()]
			if !ok {
				dev = &ir_protocol.DeviceEvents{
					DeviceName: event.GetDeviceName(),
				}
				c.Window.Events[event.GetDeviceId()] = dev
			}
			// var event interface{} = &ButtonEventData{InputEventData: InputEventData{DeviceId: 1, DeviceName: "Device1", EventType: 1, timestamp: time.Now()}, Button: 2}

			switch v := event.(type) {
			case *xinputhandler.ButtonEventData:
				// fmt.Printf("ButtonEvent with button: %d\n", v.Button)
				dev.ButtonEvents = append(dev.ButtonEvents, &ir_protocol.ButtonEvent{
					Timestamp:  timestamppb.New(v.Timestamp),
					ButtonCode: uint32(v.Button),
					IsDown:     v.EventType&1 == 0,
				})
			case *xinputhandler.KeyEventData:
				// fmt.Printf("KeyEvent with keycode: %d\n", v.KeyCode)
				dev.KeyEvents = append(dev.KeyEvents, &ir_protocol.KeyEvent{
					Timestamp: timestamppb.New(v.Timestamp),
					KeyCode:   uint32(v.KeyCode),
					IsDown:    v.EventType&1 == 0,
				})
			case *xinputhandler.MotionEvent:
				// fmt.Printf("MotionEvent with axis position: %v\n", v.AxisPosition)
				motion_event := &ir_protocol.MotionEvent{
					Timestamp:     timestamppb.New(v.Timestamp),
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

}

// this function is the most important function of the application, it is periodically check the window focus changes and when the focus changed the function logs the current record and start a new one
func (c *InputEventCollector) ProcessWindowFocusChanges() {
	if !c.Enabled {
		return
	}
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

	if fp != c.Last_fp {
		fmt.Printf("new fp:%v\n", fp)
		fmt.Printf("old fp:%v\n", c.Last_fp)
		// for k, v := range propValues {
		// log.Printf("k:%v v:%v\n", k, v)
		// }
		if len(c.Window.Properties) != 0 {
			fmt.Println("write to file")
			c.WriteToFile()
		}
		c.Window = c.GetNewWindow(&propValues)
		c.Last_fp = fp
	}
}

var mu sync.Mutex

// this method move the temporally dumps to the output directory if it necessary and returns the current filename
func (c *InputEventCollector) GetFileNameAndOrganizeFiles(force bool) string {
	mu.Lock()
	defer mu.Unlock()

	currentTS := time.Now().Format("2006_01_02-15")
	currentDay := time.Now().Format("2006_01_02")
	out_dir := c.Out_dir + "/" + c.Hostname + "/" + currentDay

	filename_base := "data_" + c.Hostname + "_" + currentTS + ".protodump"

	filename := c.Tmp_dir + "/" + filename_base
	os.MkdirAll(c.Tmp_dir, 0755)

	//check that filename_base is exist in the tmp_dir
	if _, err := os.Stat(filename); os.IsNotExist(err) || force {
		// check that any file exist in the tmp_dir
		files, err := os.ReadDir(c.Tmp_dir)
		if err != nil {
			log.Fatal(err)
		}
		if len(files) != 0 {
			// move all files to the out_dir
			os.MkdirAll(out_dir, 0755)
			for _, file := range files {

				if !force {
					os.Rename(c.Tmp_dir+"/"+file.Name(), out_dir+"/"+file.Name())
					fmt.Printf("move file:%v\n", file.Name())
				} else {
					common.Copy(c.Tmp_dir+"/"+file.Name(), out_dir+"/"+file.Name())
					fmt.Printf("copy file:%v\n", file.Name())
				}
			}
		}
	}

	return filename

}

func (c *InputEventCollector) WriteToFile() error {

	filename := c.GetFileNameAndOrganizeFiles(false)
	data, err := proto.Marshal(c.Window)
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

func (c *InputEventCollector) GetNewWindow(props *xwindow.WinProps) *ir_protocol.WindowChange {
	window := ir_protocol.WindowChange{
		Timestamp:  timestamppb.New(time.Now()),
		Properties: *props,
		Events:     make(map[uint64]*ir_protocol.DeviceEvents),
		Location:   c.Location,
		Hostname:   c.Hostname,
	}
	return &window
}
