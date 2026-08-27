package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// validScrapeConfig is a config with every key set, for the cases that are only
// interested in one of them.
func validScrapeConfig() ScrapeConfig {
	return ScrapeConfig{
		PauseBetweenEventsMin: 5 * time.Second,
		PauseBetweenEventsMax: 20 * time.Second,
		PauseBetweenHostsMin:  10 * time.Second,
		PauseBetweenHostsMax:  30 * time.Second,
		LongPauseAfterHosts:   10,
		LongPauseMin:          10 * time.Minute,
		LongPauseMax:          20 * time.Minute,
		PauseAfterFailureMin:  1 * time.Hour,
		PauseAfterFailureMax:  2 * time.Hour,
		MaxFailedHosts:        5,
	}
}

// writeConfigFile writes the given YAML into a temporary config file and
// returns its path.
func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	return path
}

func TestLoadScrapeConfig(t *testing.T) {
	// Every value differs from the shipped config.yaml, and the units are mixed
	// on purpose, to show that any of them is understood
	const validConfig = `
scrape:
  pause_between_events_min: 3s
  pause_between_events_max: 7m
  pause_between_hosts_min: 1500ms
  pause_between_hosts_max: 2s
  long_pause_after_hosts: 4
  long_pause_min: 1m
  long_pause_max: 1h30m
  pause_after_failure_min: 2h
  pause_after_failure_max: 3h
  max_failed_hosts: 2
`
	config, err := loadScrapeConfig(writeConfigFile(t, validConfig))
	if err != nil {
		t.Fatalf("loadScrapeConfig: %v", err)
	}

	want := ScrapeConfig{
		PauseBetweenEventsMin: 3 * time.Second,
		PauseBetweenEventsMax: 7 * time.Minute,
		PauseBetweenHostsMin:  1500 * time.Millisecond,
		PauseBetweenHostsMax:  2 * time.Second,
		LongPauseAfterHosts:   4,
		LongPauseMin:          1 * time.Minute,
		LongPauseMax:          90 * time.Minute,
		PauseAfterFailureMin:  2 * time.Hour,
		PauseAfterFailureMax:  3 * time.Hour,
		MaxFailedHosts:        2,
	}
	if config != want {
		t.Errorf("loadScrapeConfig = %+v, want %+v", config, want)
	}
}

// TestLoadScrapeConfigFromRepositoryFile checks that the config file shipped
// with the project is a valid one, as nothing else would notice a typo in it.
func TestLoadScrapeConfigFromRepositoryFile(t *testing.T) {
	if _, err := loadScrapeConfig("config.yaml"); err != nil {
		t.Fatalf("the config.yaml of the repository is not valid: %v", err)
	}
}

func TestLoadScrapeConfigWithoutFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := loadScrapeConfig(path)
	if err == nil {
		t.Fatal("a missing config file has to be an error, the scraper has no defaults to fall back to")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not name the config file path", err)
	}
}

// TestLoadScrapeConfigRejectsUnparsableDurations covers the values that never
// reach the validation, because the YAML library itself refuses to turn them
// into a duration. Those errors point at the line instead of naming the key.
func TestLoadScrapeConfigRejectsUnparsableDurations(t *testing.T) {
	testCases := []struct {
		name        string
		value       string
		wantInError string
	}{
		{
			name:        "a number without a unit",
			value:       "5",
			wantInError: "cannot unmarshal !!int `5` into time.Duration",
		},
		{
			name:        "an unknown unit",
			value:       "5x",
			wantInError: "cannot unmarshal !!str `5x` into time.Duration",
		},
		{
			name:        "not a duration at all",
			value:       "soon",
			wantInError: "cannot unmarshal !!str `soon` into time.Duration",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			content := "scrape:\n  pause_between_events_min: " + testCase.value + "\n  pause_between_events_max: 20s\n"

			_, err := loadScrapeConfig(writeConfigFile(t, content))
			if err == nil {
				t.Fatal("expected an error, got none")
			}
			if !strings.Contains(err.Error(), testCase.wantInError) {
				t.Errorf("error %q does not mention %s", err, testCase.wantInError)
			}
		})
	}
}

// TestValidateScrapeConfig covers the values that parse fine but would panic in
// sleepRandom, as rand.Int63n panics unless the range is positive, or in the
// modulo of the long pause. A key that is missing from the file parses as a
// zero, so it lands here as well.
//
// Every case starts from a valid config and breaks a single value, so that a
// range that is never checked shows up as a failure.
func TestValidateScrapeConfig(t *testing.T) {
	testCases := []struct {
		name        string
		breakConfig func(*ScrapeConfig)
		valid       bool
	}{
		{
			name:  "a valid config",
			valid: true,
		},
		// For the invalid test cases, breakConfig will change a value in the validScrapeConfig() to make it invalid and test it
		{
			name:        "the pause between events is inverted",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEventsMax = config.PauseBetweenEventsMin - time.Second },
			valid:       false,
		},
		{
			name:        "the pause between hosts is inverted",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenHostsMax = config.PauseBetweenHostsMin - time.Second },
			valid:       false,
		},
		{
			name:        "the long pause is inverted",
			breakConfig: func(config *ScrapeConfig) { config.LongPauseMax = config.LongPauseMin - time.Second },
			valid:       false,
		},
		{
			name:        "the pause after a failure is inverted",
			breakConfig: func(config *ScrapeConfig) { config.PauseAfterFailureMax = config.PauseAfterFailureMin - time.Second },
			valid:       false,
		},
		{
			name:        "a maximum equal to its minimum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEventsMax = config.PauseBetweenEventsMin },
			valid:       false,
		},
		{
			name:        "a missing minimum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEventsMin = 0 },
			valid:       false,
		},
		{
			name:        "a missing maximum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEventsMax = 0 },
			valid:       false,
		},
		{
			name:        "a negative minimum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEventsMin = -5 * time.Second },
			valid:       false,
		},
		{
			name:        "a missing long pause interval",
			breakConfig: func(config *ScrapeConfig) { config.LongPauseAfterHosts = 0 },
			valid:       false,
		},
		{
			name:        "a missing maximum of failed hosts",
			breakConfig: func(config *ScrapeConfig) { config.MaxFailedHosts = 0 },
			valid:       false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Get a valid config and modify some of the values with breakConfig
			config := validScrapeConfig()
			if testCase.breakConfig != nil {
				testCase.breakConfig(&config)
			}

			err := validateScrapeConfig(config)

			// The valid case
			if testCase.valid {
				if err != nil {
					t.Fatalf("validateScrapeConfig: %v", err)
				}
				return
			}

			// All the others here expects an error
			if err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}

// TestLoadScrapeConfigValidatesTheValues makes sure the loader actually runs the
// validation, as the cases above call it directly.
func TestLoadScrapeConfigValidatesTheValues(t *testing.T) {
	const invalidConfig = `
scrape:
  pause_between_events_min: 20s
  pause_between_events_max: 5s
`
	path := writeConfigFile(t, invalidConfig)

	_, err := loadScrapeConfig(path)
	if err == nil {
		t.Fatal("expected the loader to reject a maximum below its minimum")
	}
}
