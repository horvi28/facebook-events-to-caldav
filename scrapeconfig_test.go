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
		PauseBetweenEvents:  pauseRange{Min: 5 * time.Second, Max: 20 * time.Second},
		PauseBetweenHosts:   pauseRange{Min: 10 * time.Second, Max: 30 * time.Second},
		LongPauseAfterHosts: 10,
		LongPause:           pauseRange{Min: 10 * time.Minute, Max: 20 * time.Minute},
		PauseAfterFailure:   pauseRange{Min: 1 * time.Hour, Max: 2 * time.Hour},
		MaxFailedHosts:      5,
		HostsPerRun:         20,
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

func TestLoadAppConfig(t *testing.T) {
	// Every value differs from the shipped config.yaml, and the units are mixed
	// on purpose, to show that any of them is understood
	const validConfig = `
scrape:
  pause_between_events:
    min: 3s
    max: 7m
  pause_between_hosts:
    min: 1500ms
    max: 2s
  long_pause_after_hosts: 4
  long_pause:
    min: 1m
    max: 1h30m
  pause_after_failure:
    min: 2h
    max: 3h
  max_failed_hosts: 2
  hosts_per_run: 7
daemon:
  interval: 6h
`
	config, err := loadAppConfig(writeConfigFile(t, validConfig))
	if err != nil {
		t.Fatalf("loadAppConfig: %v", err)
	}

	want := ScrapeConfig{
		PauseBetweenEvents:  pauseRange{Min: 3 * time.Second, Max: 7 * time.Minute},
		PauseBetweenHosts:   pauseRange{Min: 1500 * time.Millisecond, Max: 2 * time.Second},
		LongPauseAfterHosts: 4,
		LongPause:           pauseRange{Min: 1 * time.Minute, Max: 90 * time.Minute},
		PauseAfterFailure:   pauseRange{Min: 2 * time.Hour, Max: 3 * time.Hour},
		MaxFailedHosts:      2,
		HostsPerRun:         7,
	}
	if config.Scrape != want {
		t.Errorf("loadAppConfig = %+v, want %+v", config.Scrape, want)
	}

	if config.Daemon.Interval != 6*time.Hour {
		t.Errorf("daemon interval = %v, want 6h", config.Daemon.Interval)
	}
}

// TestLoadAppConfigFromRepositoryFile checks that the config file shipped
// with the project is a valid one, as nothing else would notice a typo in it.
func TestLoadAppConfigFromRepositoryFile(t *testing.T) {
	if _, err := loadAppConfig("config.yaml"); err != nil {
		t.Fatalf("the config.yaml of the repository is not valid: %v", err)
	}
}

func TestLoadAppConfigWithoutFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := loadAppConfig(path)
	if err == nil {
		t.Fatal("a missing config file has to be an error, the scraper has no defaults to fall back to")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not name the config file path", err)
	}
}

// TestLoadAppConfigRejectsUnparsableDurations covers the values that never
// reach the validation, because the YAML library itself refuses to turn them
// into a duration. Those errors point at the line instead of naming the key.
func TestLoadAppConfigRejectsUnparsableDurations(t *testing.T) {
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
			content := "scrape:\n  pause_between_events:\n    min: " + testCase.value + "\n    max: 20s\n"

			_, err := loadAppConfig(writeConfigFile(t, content))
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
			name: "the pause between events is inverted",
			breakConfig: func(config *ScrapeConfig) {
				config.PauseBetweenEvents.Max = config.PauseBetweenEvents.Min - time.Second
			},
			valid: false,
		},
		{
			name:        "the pause between hosts is inverted",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenHosts.Max = config.PauseBetweenHosts.Min - time.Second },
			valid:       false,
		},
		{
			name:        "the long pause is inverted",
			breakConfig: func(config *ScrapeConfig) { config.LongPause.Max = config.LongPause.Min - time.Second },
			valid:       false,
		},
		{
			name:        "the pause after a failure is inverted",
			breakConfig: func(config *ScrapeConfig) { config.PauseAfterFailure.Max = config.PauseAfterFailure.Min - time.Second },
			valid:       false,
		},
		{
			name:        "a maximum equal to its minimum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEvents.Max = config.PauseBetweenEvents.Min },
			valid:       false,
		},
		{
			name:        "a missing minimum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEvents.Min = 0 },
			valid:       false,
		},
		{
			name:        "a missing maximum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEvents.Max = 0 },
			valid:       false,
		},
		{
			name:        "a negative minimum",
			breakConfig: func(config *ScrapeConfig) { config.PauseBetweenEvents.Min = -5 * time.Second },
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
		{
			name:        "a missing number of hosts per run",
			breakConfig: func(config *ScrapeConfig) { config.HostsPerRun = 0 },
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

// TestLoadAppConfigValidatesTheValues makes sure the loader actually runs the
// validation, as the cases above call it directly.
func TestLoadAppConfigValidatesTheValues(t *testing.T) {
	const invalidConfig = `
scrape:
  pause_between_events:
    min: 20s
    max: 5s
`
	path := writeConfigFile(t, invalidConfig)

	_, err := loadAppConfig(path)
	if err == nil {
		t.Fatal("expected the loader to reject a maximum below its minimum")
	}
	// The error has to be about the inverted pause. An unknown key is silently
	// ignored by the YAML library, so without this the test would also pass on a
	// config that says nothing at all.
	if !strings.Contains(err.Error(), "pause_between_events.max") {
		t.Errorf("error %q is not about the inverted pause", err)
	}
}

// TestValidateDaemonConfig covers the interval that the daemon waits between two
// cycles. A missing key parses as a zero, and a daemon that does not wait would
// scrape back to back, which we want to avoid.
func TestValidateDaemonConfig(t *testing.T) {
	testCases := []struct {
		name     string
		interval time.Duration
		valid    bool
	}{
		{name: "a valid interval", interval: 24 * time.Hour, valid: true},
		{name: "a missing interval", interval: 0, valid: false},
		{name: "a negative interval", interval: -1 * time.Hour, valid: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateDaemonConfig(DaemonConfig{Interval: testCase.interval})

			if testCase.valid {
				if err != nil {
					t.Fatalf("validateDaemonConfig: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}

// TestLoadAppConfigValidatesTheDaemonSection makes sure the loader runs the
// daemon validation as well, not only the one of the scrape section.
func TestLoadAppConfigValidatesTheDaemonSection(t *testing.T) {
	// A config with a valid scrape section but no daemon section at all
	validScrape := `
scrape:
  pause_between_events:
    min: 5s
    max: 20s
  pause_between_hosts:
    min: 10s
    max: 30s
  long_pause_after_hosts: 10
  long_pause:
    min: 10m
    max: 20m
  pause_after_failure:
    min: 1h
    max: 2h
  max_failed_hosts: 5
  hosts_per_run: 20
`
	_, err := loadAppConfig(writeConfigFile(t, validScrape))
	if err == nil {
		t.Fatal("expected the loader to reject a config without a daemon section")
	}
	if !strings.Contains(err.Error(), "daemon.interval") {
		t.Errorf("error %q is not about the missing interval", err)
	}
}
