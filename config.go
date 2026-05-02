package main

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	CalDAVURL    string `env:"CALDAV_URL,notEmpty,required"`
	Username     string `env:"CALDAV_USERNAME,notEmpty,required"`
	Password     string `env:"CALDAV_PASSWORD,notEmpty,required"`
	CalendarPath string `env:"CALDAV_CALENDAR_PATH,notEmpty,required"`
}

func loadConfig() (Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("Unable to parse environment variables: %v", err)
	}

	return config, nil
}
