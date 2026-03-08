package main

import (
	"encoding/json"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func dumpEventsToJsonFile(filePath string, events []*fbevents.EventData) error {
	eventsJson, marshalError := json.MarshalIndent(events, "", "  ")
	if marshalError != nil {
		return marshalError
	}

	return os.WriteFile(filePath, eventsJson, 0o644)
}

func loadEventsFromJsonFile(filePath string) ([]*fbevents.EventData, error) {
	eventsJson, readError := os.ReadFile(filePath)
	if readError != nil {
		return nil, readError
	}

	var events []*fbevents.EventData
	if unmarshalError := json.Unmarshal(eventsJson, &events); unmarshalError != nil {
		return nil, unmarshalError
	}

	return events, nil
}
