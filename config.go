package main

import (
	"fmt"

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
		return Config{}, fmt.Errorf("load .env file: %w", err)
	}

	var config Config
	if err := env.Parse(&config); err != nil {
		return Config{}, fmt.Errorf("parse environment variables: %w", err)
	}

	return config, nil
}
