package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ScrapeConfig holds the pacing values that decide how hard Facebook is hit.
// Every value comes from the config file, there are no defaults in the code.
type ScrapeConfig struct {
	PauseBetweenEventsMin time.Duration `yaml:"pause_between_events_min"`
	PauseBetweenEventsMax time.Duration `yaml:"pause_between_events_max"`
}

// configFile is the top level of the config file. The pacing values live under
// their own key to leave room for the batch and daemon settings later on.
type configFile struct {
	Scrape ScrapeConfig `yaml:"scrape"`
}

// validateRange checks a min/max pair of the config file. A maximum that is not
// above its minimum would panic in sleepRandom, and a missing key parses as a
// zero, which is why both values have to be positive.
func validateRange(minimumKey string, minimumValue time.Duration, maximumKey string, maximumValue time.Duration) error {
	if minimumValue <= 0 {
		return fmt.Errorf("%s must be a positive duration, got %v", minimumKey, minimumValue)
	}
	if maximumValue <= minimumValue {
		return fmt.Errorf("%s (%v) must be greater than %s (%v)", maximumKey, maximumValue, minimumKey, minimumValue)
	}

	return nil
}

// validateScrapeConfig rejects a config that would make the scraper panic or
// that is missing a key.
func validateScrapeConfig(config ScrapeConfig) error {
	return validateRange(
		"pause_between_events_min", config.PauseBetweenEventsMin,
		"pause_between_events_max", config.PauseBetweenEventsMax,
	)
}

// loadScrapeConfig reads the pacing values from the config file. The file is
// required, as the scraper has no built-in values to fall back to.
func loadScrapeConfig(configFilePath string) (ScrapeConfig, error) {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return ScrapeConfig{}, fmt.Errorf("read config file %s: %w", configFilePath, err)
	}

	var file configFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return ScrapeConfig{}, fmt.Errorf("parse config file %s: %w", configFilePath, err)
	}

	if err := validateScrapeConfig(file.Scrape); err != nil {
		return ScrapeConfig{}, fmt.Errorf("invalid config file %s: %w", configFilePath, err)
	}

	return file.Scrape, nil
}
