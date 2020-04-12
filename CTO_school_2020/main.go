package main

import (
	"log"
	"net/http"
	"os"

	"weather_informer/config"
)

const (
	locationList = "/locations"
	citiesPath   = "./city_list/city.list.json"
	configPath   = "./config/config.env"
)

func main() {
	locations, supportedLocations, err := createLocations(citiesPath)
	if err != nil {
		log.Fatalf("create locations: path: %v; err: %v", citiesPath, err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load configs: path: %v; err: %v", configPath, err)
	}

	bot, err := createBot(cfg)
	if err != nil {
		log.Fatalf("bot createon: %v", err)
	}

	weatherInformer := newWeatherInformer(bot, cfg, locations, supportedLocations)
	go weatherInformer.startBot()

	http.HandleFunc(locationList, weatherInformer.handleSupportedLocations)

	port := os.Getenv("PORT")

	log.Printf("start server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
