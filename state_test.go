package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestScrapeStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	// Truncated to a second, as the state file only keeps that much precision
	scrapedAt := time.Now().Truncate(time.Second)

	// Create a state
	state := scrapeState{
		"https://fb.com/a": scrapedAt,
		"https://fb.com/b": scrapedAt.Add(-24 * time.Hour),
	}

	// Check the roundtrip
	if err := saveScrapeState(state, path); err != nil {
		t.Fatalf("saveScrapeState: %v", err)
	}

	loadedState, err := loadScrapeState(path)
	if err != nil {
		t.Fatalf("loadScrapeState: %v", err)
	}

	if len(loadedState) != len(state) {
		t.Fatalf("loaded %d hosts, want %d", len(loadedState), len(state))
	}
	for hostURL, want := range state {
		if got := loadedState[hostURL]; !got.Equal(want) {
			t.Errorf("last scrape of %s = %v, want %v", hostURL, got, want)
		}
	}
}

// TestLoadScrapeStateWithoutFile covers the very first run, where nothing has
// been scraped yet. That is not an error, unlike a missing config file.
func TestLoadScrapeStateWithoutFile(t *testing.T) {
	state, err := loadScrapeState(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("a missing state file has to be an empty state: %v", err)
	}
	if len(state) != 0 {
		t.Errorf("state has %d hosts, want an empty one", len(state))
	}
}

func TestLoadScrapeStateWithBrokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	if _, err := loadScrapeState(path); err == nil {
		t.Fatal("a broken state file has to be an error, silently forgetting every host would re-scrape all of them")
	}
}

// scrapedAgo builds a state where each host was scraped the given time ago
// now is taken as a parameter, to have it consistent for multiple calls and let the caller decide
func scrapedAgo(now time.Time, scrapedAgoByHost map[string]time.Duration) scrapeState {
	state := scrapeState{}
	for hostURL, ago := range scrapedAgoByHost {
		state[hostURL] = now.Add(-ago)
	}

	return state
}

func TestSelectHostsToScrape(t *testing.T) {
	now := time.Now()
	// Listed in the order of the hosts file, which is not the order of the last
	// scrape, so a selection that ignores the state shows up as a failure
	hosts := []string{"recent", "oldest", "never", "middle"}
	state := scrapedAgo(now, map[string]time.Duration{
		"recent": 1 * time.Hour,
		"oldest": 30 * 24 * time.Hour,
		"middle": 7 * 24 * time.Hour,
		// "never" is missing on purpose, it has never been scraped
	})

	testCases := []struct {
		name        string
		hostsPerRun int
		wantHosts   []string
	}{
		{
			name:        "a never scraped host comes before every scraped one",
			hostsPerRun: 1,
			wantHosts:   []string{"never"},
		},
		{
			name:        "the rest follow from the most overdue to the freshest",
			hostsPerRun: 4,
			wantHosts:   []string{"never", "oldest", "middle", "recent"},
		},
		{
			name:        "only the given number of hosts is picked",
			hostsPerRun: 2,
			wantHosts:   []string{"never", "oldest"},
		},
		{
			name:        "asking for more hosts than there are gives all of them",
			hostsPerRun: 100,
			wantHosts:   []string{"never", "oldest", "middle", "recent"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			selectedHosts := selectHostsToScrape(hosts, state, testCase.hostsPerRun)
			if !slices.Equal(selectedHosts, testCase.wantHosts) {
				t.Errorf("selectHostsToScrape = %v, want %v", selectedHosts, testCase.wantHosts)
			}
		})
	}
}

// TestSelectHostsToScrapeOnTheFirstRun covers the empty state, where no host has
// ever been scraped, so the hosts file decides the order.
func TestSelectHostsToScrapeOnTheFirstRun(t *testing.T) {
	hosts := []string{"c", "a", "b"}

	selectedHosts := selectHostsToScrape(hosts, scrapeState{}, 2)

	if !slices.Equal(selectedHosts, []string{"c", "a"}) {
		t.Errorf("selectHostsToScrape = %v, want the first two of the hosts file", selectedHosts)
	}
}

// TestSelectHostsToScrapeKeepsTheInput makes sure the hosts of the config are
// not reordered, as the rest of the app relies on their order.
func TestSelectHostsToScrapeKeepsTheInput(t *testing.T) {
	now := time.Now()
	hosts := []string{"recent", "oldest"}
	state := scrapedAgo(now, map[string]time.Duration{
		"recent": 1 * time.Hour,
		"oldest": 30 * 24 * time.Hour,
	})

	selectHostsToScrape(hosts, state, 2)

	// Check if 'hosts' is not modified
	if !slices.Equal(hosts, []string{"recent", "oldest"}) {
		t.Errorf("the hosts were reordered to %v", hosts)
	}
}
