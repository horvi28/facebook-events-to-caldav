package main

import (
	"encoding/json"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

const eventsJsonPath = "events.json"

func dumpEventsToJsonFile(events []*fbevents.EventData) error {
	eventsJson, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(eventsJsonPath, eventsJson, 0o644)
}

func loadEventsFromJsonFile() ([]*fbevents.EventData, error) {
	// Check if file exists before trying to read it
	if _, err := os.Stat(eventsJsonPath); os.IsNotExist(err) {
		return nil, err
	}

	eventsJson, err := os.ReadFile(eventsJsonPath)
	if err != nil {
		return nil, err
	}

	var events []*fbevents.EventData
	if err := json.Unmarshal(eventsJson, &events); err != nil {
		return nil, err
	}

	return events, nil
}
