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

	"github.com/joho/godotenv"
)

// this is the main file and this is the start function of the application
func main() {
	var wg sync.WaitGroup

	fmt.Println("Hello World")
	err := godotenv.Load("secret.env", "unsecret.env")
	if err != nil {
		log.Println("WARNING: error while loading all env files: ", err)
	}

	stopChan := make(chan os.Signal, 1)

	signal.Notify(stopChan, os.Interrupt)

	//InputList()
	SetupInput(stopChan, &wg)

	fmt.Print("WAIT")

	wg.Wait()
	//sleep 1sec before exit
	fmt.Println("SLEEP")
	time.Sleep(time.Second * 1)
	fmt.Println("EXIT")
}
