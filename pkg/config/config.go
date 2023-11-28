package config

import (
	"fmt"
	"sync"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	TelegramToken string `env:"TELEGRAM_TOKEN"`
	WeatherToken  string `env:"WEATHER_TOKEN"`
	Port          string `env:"PORT"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
	MongoAddr     string `env:"MONGO_ADDR"`
}

var once sync.Once

var configInstance *Config

func GetConfig() *Config {
	if configInstance == nil {
		once.Do(func() {
			fmt.Println("Creating config instance now.")

			var cfg Config

			err := godotenv.Load()
			if err != nil {
				log.Warn("No .env file")
			}

			if err = env.Parse(&cfg); err != nil {
				log.Fatal(err)
			}

			configInstance = &cfg
		})

	}

	return configInstance
}
