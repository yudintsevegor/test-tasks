package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BotToken string `envconfig:"BOT_TOKEN"`
	WebHook  string `envconfig:"WEB_HOOK"`

	WeatherKeyAPI string `envconfig:"WEATHER_KEY_API"`

	IsDebug bool `envconfig:"DEBUG"`
}

func Load(path string) (*Config, error) {
	err := godotenv.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load .env config file: %w", err)
	}

	config := new(Config)
	err = envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("get config from env: %w", err)
	}

	return config, nil
}
