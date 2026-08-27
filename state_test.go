package main

import (
	"os"
	"path/filepath"
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
