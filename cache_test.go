package main

import (
	"testing"
	"time"

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

// createEventAt builds an event running between the two given offsets from the
// reference time.
func createEventAt(id string, referenceTime time.Time, startOffset, endOffset time.Duration) *fbevents.EventData {
	event := createEventWithoutEndAt(id, referenceTime, startOffset)

	endTimestamp := referenceTime.Add(endOffset).Unix()
	event.EndTimestamp = &endTimestamp

	return event
}

// createEventWithoutEndAt builds an event that only has a start time, which is
// the case for roughly half of the events Facebook returns.
func createEventWithoutEndAt(id string, referenceTime time.Time, startOffset time.Duration) *fbevents.EventData {
	return &fbevents.EventData{
		ID:             id,
		StartTimestamp: referenceTime.Add(startOffset).Unix(),
	}
}

func TestPruneOutdatedEvents(t *testing.T) {
	now := time.Date(2026, time.August, 23, 20, 0, 0, 0, time.UTC)
	const month = 30 * 24 * time.Hour

	testCases := []struct {
		name         string
		event        *fbevents.EventData
		shouldBeKept bool
	}{
		{
			name:         "upcoming event is kept",
			event:        createEventAt("upcoming", now, 48*time.Hour, 52*time.Hour),
			shouldBeKept: true,
		},
		{
			name:         "multi day event that is still running is kept",
			event:        createEventAt("running", now, -72*time.Hour, 72*time.Hour),
			shouldBeKept: true,
		},
		{
			name:         "event that ended earlier today is kept for the retention time",
			event:        createEventAt("ended-earlier-today", now, -6*time.Hour, -4*time.Hour),
			shouldBeKept: true,
		},
		{
			name:         "event that ended long ago is pruned",
			event:        createEventAt("old", now, -month, -month+4*time.Hour),
			shouldBeKept: false,
		},
		{
			name:         "event without an end time falls back to its start and is kept while recent",
			event:        createEventWithoutEndAt("no-end-recent", now, -2*time.Hour),
			shouldBeKept: true,
		},
		{
			name:         "event without an end time is pruned once its start is old",
			event:        createEventWithoutEndAt("no-end-old", now, -month),
			shouldBeKept: false,
		},
		{
			name:         "event ending the moment it starts behaves like one without an end time",
			event:        createEventAt("zero-length-old", now, -month, -month),
			shouldBeKept: false,
		},
		{
			name:         "upcoming event is kept even if its end time is before its start",
			event:        createEventAt("broken-end", now, 48*time.Hour, -month),
			shouldBeKept: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			events := map[string]*fbevents.EventData{testCase.event.ID: testCase.event}

			keptEvents := pruneOutdatedEvents(events, now)

			if _, kept := keptEvents[testCase.event.ID]; kept != testCase.shouldBeKept {
				t.Errorf("event %q kept = %v, want %v", testCase.event.ID, kept, testCase.shouldBeKept)
			}
		})
	}
}

func TestPruneOutdatedEventsDoesNotChangeTheInput(t *testing.T) {
	now := time.Date(2026, time.August, 23, 20, 0, 0, 0, time.UTC)
	events := map[string]*fbevents.EventData{
		"old": createEventWithoutEndAt("old", now, -30*24*time.Hour),
	}

	pruneOutdatedEvents(events, now)

	if len(events) != 1 {
		t.Errorf("the given map should be left alone, got %d events", len(events))
	}
	if events["old"].ID != "old" {
		t.Errorf("the given map attributes should be not changed")
	}
}
