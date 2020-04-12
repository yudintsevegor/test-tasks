package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"weather_informer/config"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type weatherInformer struct {
	bot                *tgbotapi.BotAPI
	apiKey             string
	locations          map[string]coordinates
	supportedLocations []string

	url     string
	mu      sync.Mutex
	isDebug bool
}

func newWeatherInformer(
	bot *tgbotapi.BotAPI, cfg *config.Config, locations map[string]coordinates, supportedLocations []string,
) weatherInformer {
	return weatherInformer{
		bot:                bot,
		apiKey:             cfg.WeatherKeyAPI,
		locations:          locations,
		supportedLocations: supportedLocations,

		url:     cfg.WebHook,
		isDebug: cfg.IsDebug,
	}
}

func (w weatherInformer) handleSupportedLocations(wr http.ResponseWriter, req *http.Request) {
	text := fmt.Sprintf(
		"Supported locations: %v\nFormat: Country-City [for countries with state format is Country-State-City]\n",
		len(w.supportedLocations))

	_, _ = wr.Write([]byte(text + strings.Join(w.supportedLocations, "\n")))
}

const (
	startMessage = "/start"
	helloMessage = `
	Hello,
	i am a weather bot. You can send me a location in format country-city
	[or country-state-city for countries with states (ex: us)].
	And i send to you a current weather, hourly weather for the next 12 hours
	and recommendation!

	Enjoy!
	`
)

func (w weatherInformer) startBot() {
	log.Println("bot started")

	updates := w.bot.ListenForWebhook("/")
	for update := range updates {
		if update.Message.Text == startMessage {
			w.sendMessage(update.Message.Chat, helloMessage)
			continue
		}

		message, err := w.handleMessage(update.Message)
		if err != nil {
			log.Printf("handle message error: %v", err)

			if len(message) == 0 {
				message = "error during request"
			}
			w.sendMessage(update.Message.Chat, message)

			continue
		}

		w.sendMessage(update.Message.Chat, message)
	}

}

const hoursInHalfDay = 12

var isCurrentWeatherDay = true

func (w weatherInformer) handleMessage(message *tgbotapi.Message) (string, error) {
	coordinate, err := w.getCoordinates(message.Text)
	if err != nil {
		text := fmt.Sprintf(`unknown location: '%v'; use list '%v'`, message.Text, w.url+locationList)

		return text, err
	}

	response, err := w.getWeatherByAPI(w.apiKey, coordinate)
	if err != nil {
		return "", fmt.Errorf("get weather info by day text: %w", err)
	}

	text, err := createWeatherText(response.Current, response.Timezone, isCurrentWeatherDay)
	if err != nil {
		return "", fmt.Errorf("create weather text: %w", err)
	}

	limit := hoursInHalfDay
	if len(response.Hourly) < hoursInHalfDay {
		limit = len(response.Hourly)
	}

	weatherTypes := make(map[WeatherType]struct{})

	for _, hourlyWeather := range response.Hourly[:limit] {
		hourlyText, err := createWeatherText(hourlyWeather, response.Timezone, !isCurrentWeatherDay)
		if err != nil {
			log.Println(fmt.Errorf("create hourly weather text: %w", err))
			continue
		}

		text += "\n" + hourlyText
		for _, info := range hourlyWeather.WeatherInfo {
			weatherTypes[info.Main] = struct{}{}
		}
	}

	text += "\n\n\n" + createRecommendations(weatherTypes)

	return text, nil
}

func (w weatherInformer) sendMessage(chat *tgbotapi.Chat, message string) {
	_, err := w.bot.Send(tgbotapi.NewMessage(chat.ID, message))
	if err != nil {
		log.Printf("sending to %v message; error: %v", chat.LastName, err)
	}
}
