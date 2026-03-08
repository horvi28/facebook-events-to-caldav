package main

import (
	"fmt"
	"log"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func main() {
	eventUrl := "https://www.facebook.com/bolygo.hu/upcoming_hosted_events"
	eventsJsonPath := "events.json"

	var events []*fbevents.EventData

	if _, err := os.Stat(eventsJsonPath); err == nil {
		events, err = loadEventsFromJsonFile(eventsJsonPath)
		if err != nil {
			log.Fatalf("Event load failed: %v", err)
		}
	} else {
		// Scrape full event details but only for upcoming events
		options := &fbevents.Options{
			IncludePastEvents: false,
		}
		events, err := fbevents.ScrapeFbEventListFull(eventUrl, options)
		if err != nil {
			log.Fatalf("Scrape failed: %v", err)
		}

		if err := dumpEventsToJsonFile(eventsJsonPath, events); err != nil {
			log.Fatalf("Event dump failed: %v", err)
		}
	}

	// Print unique event URLs
	seen := make(map[string]bool)
	for _, event := range events {
		if !seen[event.URL] {
			fmt.Println(event.URL)
			seen[event.URL] = true
		}
	}
}
