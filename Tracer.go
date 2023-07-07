// this application tracking and logging the user activity
// periodically poll the keyboard and mouse activity status and detect idle periods and window focus changes
// if the user is active periodically make a blurred screenshot and stores this and the activity data in a local folder

package main

// go:generate protoc --go_out=. ir_record.proto

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/asvany/InspectoryRespector/geo"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/asvany/ChannelWithConcurrentSenders/cc"

	"github.com/asvany/InspectoryRespector/ir_protocol"
	"github.com/joho/godotenv"
)

type InputEventCollector struct {
	window ir_protocol.WindowChange
}

func (c *InputEventCollector) processLocation(loc_chan chan *ir_protocol.Location) {
	for loc := range loc_chan {
		fmt.Println("New Location:", loc.String())
		c.window.Location = loc
	}
}

func (c *InputEventCollector) processInputEvents(input_events InputEventsChannelType) {
	for event := range input_events.ROChannel() {
		fmt.Println("Event:", event.String())
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
			fmt.Printf("ButtonEvent with button: %d\n", v.Button)
			dev.ButtonEvents = append(dev.ButtonEvents, &ir_protocol.ButtonEvent{
				Timestamp:  timestamppb.New(v.timestamp),
				ButtonCode: uint32(v.Button),
				IsDown:     v.EventType&1 == 0,
			})
		case *KeyEventData:
			fmt.Printf("KeyEvent with keycode: %d\n", v.KeyCode)
			dev.KeyEvents = append(dev.KeyEvents, &ir_protocol.KeyEvent{
				Timestamp: timestamppb.New(v.timestamp),
				KeyCode:   uint32(v.KeyCode),
				IsDown:    v.EventType&1 == 0,
			})
		case *MotionEvent:
			fmt.Printf("MotionEvent with axis position: %v\n", v.AxisPosition)
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

func (c *InputEventCollector) WriteToFile() error {
	data, err := proto.Marshal(&c.window)
	if err != nil {
		return fmt.Errorf("marshalling error %v", err)
	}
	err = os.WriteFile("data", data, 0600)
	if err != nil {
		return fmt.Errorf("writing error %v", err)
	}
	options := &protojson.MarshalOptions{
		UseProtoNames:   true,  // Használja a Proto neveket a mezők esetén
		UseEnumNumbers:  false, // Használja az Enum számértékeit az enum mezők esetén
		EmitUnpopulated: true,  // A beállítatlan mezők is legyenek jelen a JSON-ben
	}

	jsonString, err := options.Marshal(&c.window)
	if err != nil {
		return err
	}

	err = os.WriteFile("data.json", []byte(jsonString), 0600)
	if err != nil {
		return fmt.Errorf("writing error %v", err)
	}

	return nil

}

// this is the main file and this is the start function of the application
func main() {

	loc_chan := make(chan *ir_protocol.Location)

	ticker := time.NewTicker(100 * time.Second)

	go func() {
		go geo.GetLocation(loc_chan)
		for range ticker.C {
			fmt.Println("Function runs every 100 seconds")
			go geo.GetLocation(loc_chan)
			// Add your code here that you want to run
		}
	}()

	var wg sync.WaitGroup
	input_events := cc.NewChannelWithConcurrentSenders[InputEvent](10)
	// input_events := InputEventsChannelType(10)

	InputEventCollector := &InputEventCollector{
		// create WindowChange struct with an empty event list
		window: ir_protocol.WindowChange{
			Timestamp: timestamppb.New(time.Now()),

			Events: make(map[uint64]*ir_protocol.DeviceEvents),
		},
	}

	fmt.Println("Hello World")
	err := godotenv.Load("secret.env", "unsecret.env")
	if err != nil {
		log.Println("WARNING: error while loading all env files: ", err)
	}

	stopChan := make(chan os.Signal, 1)

	signal.Notify(stopChan, os.Interrupt)

	//InputList()
	go InputEventCollector.processInputEvents(input_events)
	go InputEventCollector.processLocation(loc_chan)

	SetupInput(stopChan, &wg, input_events)

	fmt.Print("WAIT")

	wg.Wait()
	input_events.Wait()
	fmt.Println("WRITE")
	err = InputEventCollector.WriteToFile()
	if err != nil {
		log.Println("ERROR: ", err)
	}

	//sleep 1sec before exit
	fmt.Println("SLEEP")
	time.Sleep(time.Second * 1)
	fmt.Println("EXIT")
}
