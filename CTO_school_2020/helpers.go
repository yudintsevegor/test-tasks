package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"weather_informer/config"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type location struct {
	City       string     `json:"name"`
	Country    string     `json:"country"`
	State      string     `json:"state"`
	Coordinate coordinate `json:"coord"`
}

type coordinate struct {
	Longitude float64 `json:"lon"`
	Latitude  float64 `json:"lat"`
}

func createLocations(path string) (map[string]coordinate, []string, error) {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, errors.New("read path")
	}

	locations := make([]location, 0, 1)
	if err := json.Unmarshal(body, &locations); err != nil {
		return nil, nil, errors.New("unmarshal locations")
	}

	var (
		cityCoordinate  = make(map[string]coordinate)
		supportedCities = make([]string, 0, len(locations))
	)

	for _, location := range locations {
		key := fmt.Sprintf("%v-%v", location.Country, location.City)
		if len(location.State) != 0 {
			key = fmt.Sprintf("%v-%v-%v", location.Country, location.State, location.City)
		}

		supportedCities = append(supportedCities, key)
		cityCoordinate[strings.ToLower(key)] = location.Coordinate
	}

	return cityCoordinate, supportedCities, nil
}

func createBot(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("new bot token: %v; err: %v", cfg.BotToken, err)
	}

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(cfg.WebHook)); err != nil {
		return nil, fmt.Errorf("set web hook: %v; err: %v", cfg.WebHook, err)
	}

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	return bot, nil
}

func createRecommendations(weatherInfo map[WeatherType]struct{}) string {
	out := "Recommendations: \n"
	for weatherType, _ := range weatherInfo {
		switch weatherType {
		case Thunderstorm:
			out += "It's better to stay at home. But if you want, you can take something from Thunderstorm."
		case Drizzle:
			out += "Take coat or umbrella (or both). It's Drizzle outside."
		case Rain:
			out += "Take coat and umbrella. It's Rain outside."
		case Snow:
			out += "Take down jacket and valenki. It's Snow outside."
		case Clouds:
			out += "Take trousers and jacket. Clouds are outside."
		case Clear:
			out += "Clear sky minimum an hour the next 24h. Cool! take shorts and t_shirt."
		}
	}

	return out
}
