package main

import (
	"context"
	"sync"

	// "github.com/oxzi/go-xinput"
	"github.com/geistesk/go-xinput"

	"os"

	"fmt"
	"os/exec"
)
// TODO alphanumeric characters only encoding
// TODO protobuf
// TODO Systray
// TODO Window focus switch
// TODO timestamp control

// struct InputEntry(
// 	Type string
// )

// xinput-list is a limited reimplementation of `xinput list`.
func InputList() []xinput.XDeviceInfo {
	display := xinput.XOpenDisplay(nil)
	defer xinput.XCloseDisplay(display)

	devices := []xinput.XDeviceInfo{}

	for _, device := range xinput.GetXDeviceInfos(display) {
		cmd := exec.Command("xinput", "query-state", fmt.Sprintf("%d", device.Id))

		_, err := cmd.Output()
		if err != nil {
			fmt.Printf("invalid device:%v %v\n", device.Name, err)

		} else {
			devices = append(devices, device)
		}
	}
	return devices

}

func validEvent(value xinput.Event) bool {
	for _, item := range []xinput.EventType{
		xinput.ButtonPressEvent,
		xinput.ButtonReleaseEvent,
		xinput.KeyPressEvent,
		xinput.KeyReleaseEvent,
		xinput.MotionEvent,
	} {
		if item == value.Type {
			return true
		}
	}
	return false
}

func EventLogNG(valid_devices []xinput.XDeviceInfo, stopChan chan os.Signal, wg *sync.WaitGroup) {

	// counter := 0

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()

		wg.Done()

		fmt.Println("EventLogNG Done")
	}()

	// Start a goroutine for each device
	for _, device := range valid_devices {
		wg.Add(1)
		// counter += 1
		// fmt.Printf("Starting Event Handler %v\n", counter)

		go func(device xinput.XDeviceInfo) {
			fmt.Printf("Event Handler started: id:%v name:%v\n", device.Id, device.Name)
			display := xinput.XOpenDisplay(nil)

			eventMap, err := xinput.NewEventMap(display, device)
			if err != nil {
				fmt.Printf("Failed to create EventMap: %v\n", err)

			}

			defer func() {
				wg.Done()

				// counter -= 1
				// fmt.Printf("Stopping Event Handler %v\n", counter)

				fmt.Printf("Close EventMap for device %v\n", device.Name)
				eventMap.Close()
				fmt.Printf("Closed EventMap for device %v\n", device.Name)
				xinput.XCloseDisplay(display)
				fmt.Printf("Closed display for device %v\n", device.Name)

			}()

			for {
				select {
				case event := <-eventMap.Events():
					if validEvent(event) {
						fmt.Printf("TS:%v event: device:%v device.Id:%v event.type:%v event.Field:%v event.Axes:%v \n",
						 TimeStamp(),device.Name, device.Id, event.Type, event.Field, event.Axes)
					}
				case <-ctx.Done():
					return

				}
			}
		}(device)
	}

	// Wait for a signal
	<-stopChan
	fmt.Println("STOP")

	// Cancel the context, which will stop all goroutines

	fmt.Println("CANCEL")

}

func SetupInput(stopChan chan os.Signal, wg *sync.WaitGroup) {
	valid_devices := InputList()
	for _, device := range valid_devices {
		fmt.Printf("%-40s\tid=%d\t[%v]\n", device.Name, device.Id, device.Use)
	}
	wg.Add(1)
	go EventLogNG(valid_devices, stopChan, wg)
}
