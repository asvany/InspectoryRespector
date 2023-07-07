package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Location struct {
	IP        string  `json:"ip"`
	Country   string  `json:"country_name"`
	Region    string  `json:"region_name"`
	City      string  `json:"city"`
	Zipcode   string  `json:"zip_code"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func GetLocation(loc * Location ) {
	resp, err := http.Get("https://freegeoip.app/json/")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	

	if err := json.NewDecoder(resp.Body).Decode(loc); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("IP: %s, Country: %s, Region: %s, City: %s, Zipcode: %s, Latitude: %f, Longitude: %f\n",
		loc.IP, loc.Country, loc.Region, loc.City, loc.Zipcode, loc.Latitude, loc.Longitude)
}
