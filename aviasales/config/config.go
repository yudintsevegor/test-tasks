package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// only for local testing )
	DropData bool   `envconfig:"DROP_DATA" default:"false"`
	Port     string `envconfig:"PORT" default:"9090"`

	Redis Redis
	Mongo Mongo
}

type Redis struct {
	Port string `envconfig:"REDIS_PORT" default:"6379"`
	Host string `envconfig:"REDIS_HOST" default:"localhost"`

	CacheRetries int `envconfig:"CACHE_RETRIES" default:"5"`
	CacheTTL     int `envconfig:"CACHE_TTL" default:"60"`
}

type Mongo struct {
	Name     string `envconfig:"MONGO_DBNAME" required:"true"`
	Username string `envconfig:"MONGO_USERNAME" required:"true"`
	Password string `envconfig:"MONGO_PASSWORD" required:"true"`
	Host     string `envconfig:"MONGO_HOST" default:"localhost"`
	Port     string `envconfig:"MONGO_PORT" default:"27017"`

	CollectionName string `envconfig:"MONGO_COLLECTION_NAME" required:"true"`
}

func Load(path string) (*Config, error) {
	err := godotenv.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load %s config file: %w", path, err)
	}

	config := new(Config)
	err = envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("get config from env: %w", err)
	}

	return config, nil
}
