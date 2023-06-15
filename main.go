// this application tracking and logging the user activity
// periodically poll the keyboard and mouse activity status and detect idle periods and window focus changes
// if the user is active periodically make a blurred screenshot and stores this and the activity data in a local folder

package main

import "fmt"
import "log"
import "github.com/joho/godotenv"


// this is the main file and this is the start function of the application
func main() {
	fmt.Println("Hello World")
	err := godotenv.Load("secret.env", "unsecret.env")
	if err != nil {
		log.Println("WARNING: error while loading all env files: ", err)
	}

	
}