package main // import "weather_bot"

import (
	"log"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const(
	BotToken = "1145010162:AAEtIpRRCxod21Gm161uz599P7c_FVyCl0U"
	WebHook = "https://weatherinformer.herokuapp.com/"
)

func main() {
	bot,err := createBot()
	if err != nil{
		log.Fatal(err)
	}

	log.Printf("%+v", bot)
}

func createBot() (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return nil, err
	}

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(WebHook)); err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	return bot, nil
}