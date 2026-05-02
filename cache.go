package main

import (
	"encoding/json"
	"log"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

const eventsJSONPath = "events.json"

func dumpEventsToJSONFile(events []*fbevents.EventData) error {
	eventsJSON, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("Dumping %d events to JSON file: %s", len(events), eventsJSONPath)
	return os.WriteFile(eventsJSONPath, eventsJSON, 0o644)
}

func loadEventsFromJSONFile() ([]*fbevents.EventData, error) {
	log.Printf("Loading events from JSON file: %s", eventsJSONPath)
	// Check if file exists before trying to read it
	if _, err := os.Stat(eventsJSONPath); os.IsNotExist(err) {
		return nil, err
	}
	eventsJSON, err := os.ReadFile(eventsJSONPath)
	if err != nil {
		return nil, err
	}

	var events []*fbevents.EventData
	if err := json.Unmarshal(eventsJSON, &events); err != nil {
		return nil, err
	}

	log.Printf("Loaded %d events from JSON file: %s", len(events), eventsJSONPath)
	return events, nil
}
