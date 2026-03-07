package main

import (
	"fmt"
	"log"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func main() {
	eventUrl := "https://www.facebook.com/bolygo.hu/upcoming_hosted_events"

	// Scrape full event details but only for upcoming events
	opts := &fbevents.Options{
		IncludePastEvents: false,
	}
	events, err := fbevents.ScrapeFbEventListFull(eventUrl, opts)
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
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
