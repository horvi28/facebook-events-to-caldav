package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func sleepRandom(r *rand.Rand, minimumDuration, maximumDuration time.Duration) {
	duration := minimumDuration + time.Duration(r.Int63n(int64(maximumDuration-minimumDuration)))
	time.Sleep(duration)
}

func scrapeEventsForHost(hostURL string) ([]*fbevents.EventData, error) {
	// Scrape full event details but only for upcoming events
	options := &fbevents.Options{
		IncludePastEvents: false,
	}
	events, err := fbevents.ScrapeFbEventListFull(hostURL, options)
	if err != nil {
		return nil, fmt.Errorf("scrape Facebook events for host %s: %w", hostURL, err)
	}
	log.Printf("Scraped %d events for %s", len(events), hostURL)
	return events, nil
}

func scrapeEventsForHosts(config *HostConfig) (map[string]*fbevents.EventData, error) {
	hostUrls := config.Hosts
	log.Printf("Start scraping events for %d hosts", len(hostUrls))

	// Use a map to automatically deduplicate by event ID
	eventsMap := make(map[string]*fbevents.EventData)

	// Random generator to sleep between scrapes of different hosts to avoid hitting Facebook too hard
	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
	var scrapeErrors []error
	const maxFailedHostCount = 5
	for index, hostURL := range hostUrls {
		eventNumber := index + 1
		log.Printf("[%d/%d] Scraping events for %s", eventNumber, len(hostUrls), hostURL)
		events, err := scrapeEventsForHost(hostURL)
		if err != nil {
			log.Printf("Skipping host after scrape failure: %s: %v", hostURL, err)
			scrapeErrors = append(scrapeErrors, err)
			if len(scrapeErrors) > maxFailedHostCount {
				return eventsMap, fmt.Errorf("stop scraping after %d failed hosts: %w", len(scrapeErrors), errors.Join(scrapeErrors...))
			}
			// If a host fails, sleep for 1-2 hours before continuing, just to be safe and avoid hitting Facebook too hard
			sleepRandom(randomGenerator, 1*time.Hour, 2*time.Hour)
			continue // Skip this host and continue with next
		}

		// Add events to map (duplicates automatically overwritten)
		for _, event := range events {
			eventsMap[event.ID] = event

			// Facebook rarely sets categories for events, so fall back to the host's own
			// category from the config to not have everything with the same color.
			// It is appended to the front as the first category will be used for the color.
			fallbackCategory := fbevents.EventCategory{Label: config.CategoryOf[hostURL]}
			event.Categories = append([]fbevents.EventCategory{fallbackCategory}, event.Categories...)
		}
		log.Printf("Added %d events from %s (total unique: %d)", len(events), hostURL, len(eventsMap))

		// Big pause after every 10 profiles, 10-20 min
		if eventNumber%10 == 0 {
			log.Printf("[%d/%d] Scraping events, sleeping for 10-20 minutes", eventNumber, len(hostUrls))
			sleepRandom(randomGenerator, 10*time.Minute, 20*time.Minute)
		} else {
			// Small sleep between every profile: 10-30s
			sleepRandom(randomGenerator, 10*time.Second, 30*time.Second)
		}
	}
	log.Printf("Finished scraping, found %d unique events", len(eventsMap))
	return eventsMap, nil
}
