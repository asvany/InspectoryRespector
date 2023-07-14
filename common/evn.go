package common

import (
	"log"
	"fmt"
	"os"
	"github.com/joho/godotenv"

)
func InitEnv()  {
	err := godotenv.Load("secret.env")
	if err != nil {
		log.Println("WARNING: error while loading secret env files: ", err)
	}

	err = godotenv.Load("unsecret.env")
	if err != nil {
		log.Println("WARNING: error while loading unsecret env files: ", err)
	}
	InitDISPLAY()
}

func InitDISPLAY() {
	display := os.Getenv("DISPLAY")
	if display == "" {
		log.Fatal("ERROR: DISPLAY environment variable not set")
	} else {
		fmt.Printf("DISPLAY is set to:%v\n", display)
	}
}