package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"time"

	fbevents "github.com/grucha/facebook-events-scraper"
)

// eventRetention keeps events in the cache for a while after they ended, so
// that a scrape running in the evening does not empty out the current day and
// the calendar still shows what was on earlier.
const eventRetention = 48 * time.Hour

func eventsMapToSlice(eventsMap eventsByID) []*fbevents.EventData {
	eventsSlice := make([]*fbevents.EventData, 0, len(eventsMap))
	for _, event := range eventsMap {
		eventsSlice = append(eventsSlice, event)
	}
	return eventsSlice
}

func dumpEventsToJSONFile(events eventsByID, eventsJSONPath string) error {
	eventsSlice := eventsMapToSlice(events)
	eventsJSON, err := json.MarshalIndent(eventsSlice, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal events to JSON: %w", err)
	}

	log.Printf("Dumping %d events to JSON file: %s", len(events), eventsJSONPath)
	if err := os.WriteFile(eventsJSONPath, eventsJSON, 0o644); err != nil {
		return fmt.Errorf("write events JSON file %s: %w", eventsJSONPath, err)
	}

	return nil
}

func eventsSliceToMap(events []*fbevents.EventData) eventsByID {
	// This also deduplicates events by converting them to a map
	eventsMap := make(eventsByID)
	for _, event := range events {
		eventsMap[event.ID] = event
	}
	return eventsMap
}

func loadEventsFromJSONFile(eventsJSONPath string) (eventsByID, error) {
	log.Printf("Loading events from JSON file: %s", eventsJSONPath)

	eventsJSON, err := os.ReadFile(eventsJSONPath)
	if err != nil {
		return nil, fmt.Errorf("read events JSON file %s: %w", eventsJSONPath, err)
	}

	var events []*fbevents.EventData
	if err := json.Unmarshal(eventsJSON, &events); err != nil {
		return nil, fmt.Errorf("decode events JSON file %s: %w", eventsJSONPath, err)
	}

	log.Printf("Loaded %d events from JSON file: %s", len(events), eventsJSONPath)
	return eventsSliceToMap(events), nil
}

// loadCachedEvents loads the cache file like loadEventsFromJSONFile, but treats a
// missing file as an empty cache instead of an error. That is the expected state
// on a first run, while --from-cache should still fail on a missing file.
func loadCachedEvents(eventsJSONPath string) (eventsByID, error) {
	events, err := loadEventsFromJSONFile(eventsJSONPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("No cache file at %s yet, starting with an empty cache", eventsJSONPath)
		return eventsByID{}, nil
	}
	if err != nil {
		return nil, err
	}

	return events, nil
}

// mergeEvents returns the union of the cached and the freshly scraped events.
// Scraped events win on conflict, as they are the newer data.
func mergeEvents(cachedEvents, scrapedEvents eventsByID) eventsByID {
	mergedEvents := make(eventsByID, len(cachedEvents)+len(scrapedEvents))
	maps.Copy(mergedEvents, cachedEvents)
	// This will overwrite the existing ones
	maps.Copy(mergedEvents, scrapedEvents)

	log.Printf("Merged %d scraped events into %d cached events, %d in total", len(scrapedEvents), len(cachedEvents), len(mergedEvents))
	return mergedEvents
}

// getEventEndTime returns the time the event is over. Facebook does not always
// give an end time, roughly half of the scraped events only have a start one,
// so the start time is used as the fallback.
func getEventEndTime(event *fbevents.EventData) time.Time {
	endTimestamp := event.StartTimestamp
	if event.EndTimestamp != nil {
		// In the rare case where the endTimestamp would be earlier than
		// the startTime (No such event has come back from Facebook so far),
		// take whichever timestamp is later.
		endTimestamp = max(endTimestamp, *event.EndTimestamp)
	}

	return time.Unix(endTimestamp, 0)
}

// pruneOutdatedEvents drops the events that ended before the retention cutoff.
//
// Without this a past event would stay in the cache forever.
// The current time is a parameter to keep this testable.
func pruneOutdatedEvents(events eventsByID, now time.Time) eventsByID {
	cutoff := now.Add(-eventRetention)

	keptEvents := make(eventsByID, len(events))
	for eventID, event := range events {
		// A multi day event that is still running has to be kept, so the end
		// time decides, not the start one
		if getEventEndTime(event).After(cutoff) {
			keptEvents[eventID] = event
		}
	}

	log.Printf("Pruned %d outdated events, %d are kept", len(events)-len(keptEvents), len(keptEvents))
	return keptEvents
}
