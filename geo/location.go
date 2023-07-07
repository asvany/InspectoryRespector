package geo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asvany/InspectoryRespector/ir_protocol"
)

type location struct {
	IP        string  `json:"ip"`
	Country   string  `json:"country_name"`
	Region    string  `json:"region_name"`
	City      string  `json:"city"`
	Zipcode   string  `json:"zip_code"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func GetLocation(loc_chan chan * ir_protocol.Location) {
	resp, err := http.Get("https://freegeoip.app/json/")
	if err != nil {
		fmt.Println(err)
		return
	}
	var loc location

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&loc); err != nil {
		fmt.Println(err)
		return
	}

	loc_message := ir_protocol.Location{
		Ip:        loc.IP,
		Country:   loc.Country,
		Region:    loc.Region,
		City:      loc.City,
		Zipcode:   loc.Zipcode,
		Latitude:  float32(loc.Latitude),
		Longitude: float32(loc.Longitude),
	}
	loc_chan <- &loc_message
	

	// fmt.Printf("IP: %s, Country: %s, Region: %s, City: %s, Zipcode: %s, Latitude: %f, Longitude: %f\n",
	// 	loc.IP, loc.Country, loc.Region, loc.City, loc.Zipcode, loc.Latitude, loc.Longitude)
}
