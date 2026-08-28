package main

import (
	"testing"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func TestShouldScrapeEventDetails(t *testing.T) {
	cachedEvents := eventsByID{
		"known": {ID: "known"},
	}

	testCases := []struct {
		name            string
		shortEvent      fbevents.ShortEventData
		shouldBeScraped bool
	}{
		{
			name:            "an unknown event needs its details",
			shortEvent:      fbevents.ShortEventData{ID: "new"},
			shouldBeScraped: true,
		},
		{
			name:            "an event that is already in the cache is skipped",
			shortEvent:      fbevents.ShortEventData{ID: "known"},
			shouldBeScraped: false,
		},
		{
			name:            "a cancelled event is skipped",
			shortEvent:      fbevents.ShortEventData{ID: "new", IsCanceled: true},
			shouldBeScraped: false,
		},
		{
			name:            "a cancelled event that is cached is skipped as well",
			shortEvent:      fbevents.ShortEventData{ID: "known", IsCanceled: true},
			shouldBeScraped: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if scraped := shouldScrapeEventDetails(testCase.shortEvent, cachedEvents); scraped != testCase.shouldBeScraped {
				t.Errorf("shouldScrapeEventDetails = %v, want %v", scraped, testCase.shouldBeScraped)
			}
		})
	}
}

func TestShouldScrapeEventDetailsWithEmptyCache(t *testing.T) {
	shortEvent := fbevents.ShortEventData{ID: "new"}
	emptyCache := eventsByID{}

	if !shouldScrapeEventDetails(shortEvent, emptyCache) {
		t.Error("every event needs its details on the very first run")
	}
}
