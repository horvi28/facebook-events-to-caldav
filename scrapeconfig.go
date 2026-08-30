package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// pauseRange is the bounds of a random pause.
type pauseRange struct {
	Min time.Duration `yaml:"min"`
	Max time.Duration `yaml:"max"`
}

// ScrapeConfig holds the pacing values that decide how hard Facebook is hit.
// Every value comes from the config file, there are no defaults in the code.
type ScrapeConfig struct {
	PauseBetweenEvents  pauseRange `yaml:"pause_between_events"`
	PauseBetweenHosts   pauseRange `yaml:"pause_between_hosts"`
	LongPauseAfterHosts int        `yaml:"long_pause_after_hosts"`
	LongPause           pauseRange `yaml:"long_pause"`
	PauseAfterFailure   pauseRange `yaml:"pause_after_failure"`
	MaxFailedHosts      int        `yaml:"max_failed_hosts"`
	HostsPerRun         int        `yaml:"hosts_per_run"`
}

// configFile is the top level of the config file. The pacing values live under
// their own key to leave room for the batch and daemon settings later on.
type configFile struct {
	Scrape ScrapeConfig `yaml:"scrape"`
}

// validate checks a pause of the config file. A maximum that is not above its
// minimum would panic in sleepRandom, and a missing key parses as a zero, which
// is why both values have to be positive.
func (pause pauseRange) validate(key string) error {
	if pause.Min <= 0 {
		return fmt.Errorf("%s.min must be a positive duration, got %v", key, pause.Min)
	}
	if pause.Max <= pause.Min {
		return fmt.Errorf("%s.max (%v) must be greater than %s.min (%v)", key, pause.Max, key, pause.Min)
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
	// A slice and not a map, as a map would report a random one of the broken
	// pauses instead of always the first one
	pauses := []struct {
		key        string
		pauseRange pauseRange
	}{
		{"pause_between_events", config.PauseBetweenEvents},
		{"pause_between_hosts", config.PauseBetweenHosts},
		{"long_pause", config.LongPause},
		{"pause_after_failure", config.PauseAfterFailure},
	}
	for _, pause := range pauses {
		if err := pause.pauseRange.validate(pause.key); err != nil {
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
