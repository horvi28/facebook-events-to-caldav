package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
	// A config with two different units, to show that any unit is understood
	const validConfig = `
scrape:
  pause_between_events_min: 3s
  pause_between_events_max: 7m
`
	config, err := loadScrapeConfig(writeConfigFile(t, validConfig))
	if err != nil {
		t.Fatalf("loadScrapeConfig: %v", err)
	}

	if config.PauseBetweenEventsMin != 3*time.Second {
		t.Errorf("PauseBetweenEventsMin = %v, want 3s", config.PauseBetweenEventsMin)
	}
	if config.PauseBetweenEventsMax != 7*time.Minute {
		t.Errorf("PauseBetweenEventsMax = %v, want 7m", config.PauseBetweenEventsMax)
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
// sleepRandom, as rand.Int63n panics unless the range is positive. A key that is
// missing from the file parses as a zero duration, so it lands here as well.
func TestValidateScrapeConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config ScrapeConfig
		valid  bool
	}{
		{
			name:   "a valid config",
			config: ScrapeConfig{PauseBetweenEventsMin: 5 * time.Second, PauseBetweenEventsMax: 20 * time.Second},
			valid:  true,
		},
		{
			name:   "a maximum below its minimum",
			config: ScrapeConfig{PauseBetweenEventsMin: 20 * time.Second, PauseBetweenEventsMax: 5 * time.Second},
			valid:  false,
		},
		{
			name:   "a maximum equal to its minimum",
			config: ScrapeConfig{PauseBetweenEventsMin: 5 * time.Second, PauseBetweenEventsMax: 5 * time.Second},
			valid:  false,
		},
		{
			name:   "a missing minimum",
			config: ScrapeConfig{PauseBetweenEventsMax: 20 * time.Second},
			valid:  false,
		},
		{
			name:   "a missing maximum",
			config: ScrapeConfig{PauseBetweenEventsMin: 5 * time.Second},
			valid:  false,
		},
		{
			name:   "a negative minimum",
			config: ScrapeConfig{PauseBetweenEventsMin: -5 * time.Second, PauseBetweenEventsMax: 20 * time.Second},
			valid:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateScrapeConfig(testCase.config)

			// The valid cases
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
