package main

import (
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

func scrapeEventsForHosts(hostUrls []string) (map[string]*fbevents.EventData, error) {
	log.Printf("Start scraping events for %d hosts", len(hostUrls))

	// Use a map to automatically deduplicate by event ID
	eventsMap := make(map[string]*fbevents.EventData)

	// Random generator to sleep between scrapes of different hosts to avoid hitting Facebook too hard
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hardBlockCounter := 0
	const maxHardBlockCount = 5
	for i, hostURL := range hostUrls {
		log.Printf("[%d/%d] Scraping events for %s", i+1, len(hostUrls), hostURL)
		events, err := scrapeEventsForHost(hostURL)
		if err != nil {
			// Hard block detected, sleep for 1-2 hours
			log.Printf("Hard block detected for host: %s, sleeping for 1-2 hours", hostURL)
			hardBlockCounter++
			if hardBlockCounter > maxHardBlockCount {
				return eventsMap, fmt.Errorf("stop scraping after %d hard blocks: %w", hardBlockCounter, err)
			}
			sleepRandom(r, 1*time.Hour, 2*time.Hour)
			continue // Skip this host and continue with next
		}

		// Add events to map (duplicates automatically overwritten)
		for _, event := range events {
			eventsMap[event.ID] = event
		}
		log.Printf("Added %d events from %s (total unique: %d)", len(events), hostURL, len(eventsMap))

		// Big pause after every 15 profiles, 60-120s
		if (i+1)%15 == 0 {
			log.Printf("[%d/%d] Scraping events, sleeping for 60-120 seconds", i+1, len(hostUrls))
			sleepRandom(r, 60*time.Second, 120*time.Second)
		} else {
			// Small sleep between every profile: 2-5s
			sleepRandom(r, 2*time.Second, 5*time.Second)
		}
	}
	log.Printf("Finished scraping, found %d unique events", len(eventsMap))
	return eventsMap, nil
}
