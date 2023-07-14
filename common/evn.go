package common

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func InitEnv(pathPrefix string) {
	err := godotenv.Load(pathPrefix + "secret.env")
	if err != nil {
		log.Println("WARNING: error while loading secret env files: ", err)
	}

	err = godotenv.Load(pathPrefix + "unsecret.env")
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
