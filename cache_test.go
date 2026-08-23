package main

import (
	"testing"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func TestMergeEventsKeepsCachedAndPrefersScraped(t *testing.T) {
	cachedEvents := map[string]*fbevents.EventData{
		"1": {ID: "1", Name: "Only in cache"},
		"2": {ID: "2", Name: "Old name"},
	}
	scrapedEvents := map[string]*fbevents.EventData{
		"2": {ID: "2", Name: "New name"},
		"3": {ID: "3", Name: "Only scraped"},
	}

	mergedEvents := mergeEvents(cachedEvents, scrapedEvents)

	if len(mergedEvents) != 3 {
		t.Fatalf("expected 3 merged events, got %d", len(mergedEvents))
	}
	// An event of a host that was not scraped this time must survive
	if mergedEvents["1"] == nil {
		t.Error("cached event 1 was dropped")
	}
	// The freshly scraped data wins over the cached one
	if name := mergedEvents["2"].Name; name != "New name" {
		t.Errorf("scraped event '2' details should win in case of a merge, got name %q", name)
	}
	// The newly scarpes should be also there
	if mergedEvents["3"] == nil {
		t.Error("newly scraped event was dropped")
	}
}

func TestMergeEventsWithEmptyCache(t *testing.T) {
	cachedEvents := map[string]*fbevents.EventData{}
	scrapedEvents := map[string]*fbevents.EventData{
		"1": {ID: "1", Name: "First run"},
	}

	mergedEvents := mergeEvents(cachedEvents, scrapedEvents)

	if len(mergedEvents) != 1 {
		t.Fatalf("expected 1 merged event, got %d", len(mergedEvents))
	}
}

func TestMergeEventsWithEmptyScraped(t *testing.T) {
	cachedEvents := map[string]*fbevents.EventData{
		"1": {ID: "1", Name: "Only in cache"},
	}
	scrapedEvents := map[string]*fbevents.EventData{}

	mergedEvents := mergeEvents(cachedEvents, scrapedEvents)

	if len(mergedEvents) != 1 {
		t.Fatalf("expected 1 merged event, got %d", len(mergedEvents))
	}
}

func TestLoadCachedEventsWithMissingFileReturnsEmpty(t *testing.T) {
	events, err := loadCachedEvents(t.TempDir() + "/does-not-exist.json")
	if err != nil {
		t.Fatalf("expected no error for a missing cache file, got %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected an empty cache, got %d events", len(events))
	}
}
