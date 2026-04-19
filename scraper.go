package main

import (
	"log"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func scrapeEventsForHost(hostUrl string) []*fbevents.EventData {
	// Scrape full event details but only for upcoming events
	options := &fbevents.Options{
		IncludePastEvents: false,
	}
	log.Printf("Scrape events for %s", hostUrl)
	events, err := fbevents.ScrapeFbEventListFull(hostUrl, options)
	if err != nil {
		log.Fatalf("Scrape failed for host %s: %v", hostUrl, err)
	}
	log.Printf("Scraped %d events for %s", len(events), hostUrl)
	return events
}

func scrapeEventsForHosts(hostUrls []string) []*fbevents.EventData {
	log.Printf("Start scraping events for %d hosts", len(hostUrls))
	// Init the slice whith at least hostUrl * 2 capacity
	allEvents := make([]*fbevents.EventData, 0, len(hostUrls)*2)
	for _, hostUrl := range hostUrls {
		events := scrapeEventsForHost(hostUrl)
		allEvents = append(allEvents, events...)
	}
	return allEvents
}
