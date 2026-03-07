package main

import (
	"fmt"
	"log"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func main() {
	eventUrl := "https://www.facebook.com/bolygo.hu/upcoming_hosted_events"

	// Scrape a short list of events from the URL
	// opts := &fbevents.Options{
	// 	IncludePastEvents: false,
	// }
	// events, err := fbevents.ScrapeFbEventListFull(eventUrl, opts)
	upcoming := fbevents.EventTypeUpcoming
	events, err := fbevents.ScrapeFbEventListFromPage(eventUrl, &upcoming, nil)
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}

	// Print unique event URLs
	seen := make(map[string]bool)
	for _, e := range events {
		if !seen[e.URL] {
			fmt.Println(e.URL)
			seen[e.URL] = true
		}
	}
}
