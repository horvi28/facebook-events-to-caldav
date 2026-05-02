package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

const eventsJSONPath = "events.json"

func dumpEventsToJSONFile(events []*fbevents.EventData) error {
	eventsJSON, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal events to JSON: %w", err)
	}

	log.Printf("Dumping %d events to JSON file: %s", len(events), eventsJSONPath)
	if err := os.WriteFile(eventsJSONPath, eventsJSON, 0o644); err != nil {
		return fmt.Errorf("write events JSON file %s: %w", eventsJSONPath, err)
	}

	return nil
}

func loadEventsFromJSONFile() ([]*fbevents.EventData, error) {
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
	return events, nil
}
