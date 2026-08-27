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
	PauseBetweenHostsMin  time.Duration `yaml:"pause_between_hosts_min"`
	PauseBetweenHostsMax  time.Duration `yaml:"pause_between_hosts_max"`
	LongPauseAfterHosts   int           `yaml:"long_pause_after_hosts"`
	LongPauseMin          time.Duration `yaml:"long_pause_min"`
	LongPauseMax          time.Duration `yaml:"long_pause_max"`
	PauseAfterFailureMin  time.Duration `yaml:"pause_after_failure_min"`
	PauseAfterFailureMax  time.Duration `yaml:"pause_after_failure_max"`
	MaxFailedHosts        int           `yaml:"max_failed_hosts"`
	HostsPerRun           int           `yaml:"hosts_per_run"`
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

// validateCount checks one of the plain numbers of the config file. A missing
// key parses as a zero, so requiring at least one catches that as well.
func validateCount(key string, value int) error {
	if value < 1 {
		return fmt.Errorf("%s must be at least 1, got %d", key, value)
	}

	return nil
}

// validateScrapeConfig rejects a config that would make the scraper panic
func validateScrapeConfig(config ScrapeConfig) error {
	ranges := []struct {
		minimumKey   string
		minimumValue time.Duration
		maximumKey   string
		maximumValue time.Duration
	}{
		{"pause_between_events_min", config.PauseBetweenEventsMin, "pause_between_events_max", config.PauseBetweenEventsMax},
		{"pause_between_hosts_min", config.PauseBetweenHostsMin, "pause_between_hosts_max", config.PauseBetweenHostsMax},
		{"long_pause_min", config.LongPauseMin, "long_pause_max", config.LongPauseMax},
		{"pause_after_failure_min", config.PauseAfterFailureMin, "pause_after_failure_max", config.PauseAfterFailureMax},
	}
	for _, valueRange := range ranges {
		if err := validateRange(valueRange.minimumKey, valueRange.minimumValue, valueRange.maximumKey, valueRange.maximumValue); err != nil {
			return err
		}
	}

	// A zero would panic as the number of hosts is taken modulo this one
	if err := validateCount("long_pause_after_hosts", config.LongPauseAfterHosts); err != nil {
		return err
	}

	if err := validateCount("max_failed_hosts", config.MaxFailedHosts); err != nil {
		return err
	}

	return validateCount("hosts_per_run", config.HostsPerRun)
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
