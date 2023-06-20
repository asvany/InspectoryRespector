// this application tracking and logging the user activity
// periodically poll the keyboard and mouse activity status and detect idle periods and window focus changes
// if the user is active periodically make a blurred screenshot and stores this and the activity data in a local folder

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

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

func (c *InputEventCollector) processInputEvents(input_events InputEventsChannelType) {
	for event := range input_events.ROChannel() {
		fmt.Println("Event:", event.String())
		c.window.Events = append(c.window.Events, event)
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
	var wg sync.WaitGroup
	input_events := cc.NewChannelWithConcurrentSenders[*ir_protocol.InputEvent](10)

	InputEventCollector := &InputEventCollector{
		// create WindowChange struct with an empty event list
		window: ir_protocol.WindowChange{
			Timestamp: timestamppb.New(time.Now()),

			Events: []*ir_protocol.InputEvent{},
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
