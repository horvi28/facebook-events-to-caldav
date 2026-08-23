package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func eventsMapToSlice(eventsMap map[string]*fbevents.EventData) []*fbevents.EventData {
	eventsSlice := make([]*fbevents.EventData, 0, len(eventsMap))
	for _, event := range eventsMap {
		eventsSlice = append(eventsSlice, event)
	}
	return eventsSlice
}

func dumpEventsToJSONFile(events map[string]*fbevents.EventData, eventsJSONPath string) error {
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

func eventsSliceToMap(events []*fbevents.EventData) map[string]*fbevents.EventData {
	// This also deduplicates events by converting them to a map
	eventsMap := make(map[string]*fbevents.EventData)
	for _, event := range events {
		eventsMap[event.ID] = event
	}
	return eventsMap
}

func loadEventsFromJSONFile(eventsJSONPath string) (map[string]*fbevents.EventData, error) {
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
func loadCachedEvents(eventsJSONPath string) (map[string]*fbevents.EventData, error) {
	events, err := loadEventsFromJSONFile(eventsJSONPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("No cache file at %s yet, starting with an empty cache", eventsJSONPath)
		return map[string]*fbevents.EventData{}, nil
	}
	if err != nil {
		return nil, err
	}

	return events, nil
}

// mergeEvents returns the union of the cached and the freshly scraped events.
// Scraped events win on conflict, as they are the newer data.
func mergeEvents(cachedEvents, scrapedEvents map[string]*fbevents.EventData) map[string]*fbevents.EventData {
	mergedEvents := make(map[string]*fbevents.EventData, len(cachedEvents)+len(scrapedEvents))
	maps.Copy(mergedEvents, cachedEvents)
	// This will overwrite the existing ones
	maps.Copy(mergedEvents, scrapedEvents)

	log.Printf("Merged %d scraped events into %d cached events, %d in total", len(scrapedEvents), len(cachedEvents), len(mergedEvents))
	return mergedEvents
}
