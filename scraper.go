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

// shouldScrapeEventDetails tells whether the details of a listed event have to
// be requested. Every false here saves one request.
func shouldScrapeEventDetails(shortEvent fbevents.ShortEventData, cachedEvents map[string]*fbevents.EventData) bool {
	// A cancelled event is not needed
	if shortEvent.IsCanceled {
		return false
	}

	// The details of an event are only ever scraped once. The merge with the
	// cache is what keeps an already known event in the calendar.
	_, isAlreadyInCache := cachedEvents[shortEvent.ID]

	return !isAlreadyInCache
}

// scrapeEventsForHost fetches the host's event list and then the full details of
// each listed event that is not in the cache yet.
//
// This is what fbevents.ScrapeFbEventListFull does internally, but driving the
// loop here lets us sleep between the event detail requests and skip the ones
// we already know. Listing a host costs a single request while its details cost
// one request per event, so skipping the known ones is what keeps the number of
// requests low.
func scrapeEventsForHost(hostURL string, cachedEvents map[string]*fbevents.EventData, randomGenerator *rand.Rand) ([]*fbevents.EventData, error) {
	// Only list the upcoming events, past ones are not interested
	upcoming := fbevents.EventTypeUpcoming
	shortEvents, err := fbevents.ScrapeFbEventList(hostURL, &upcoming, nil)
	if err != nil {
		return nil, fmt.Errorf("scrape Facebook event list for host %s: %w", hostURL, err)
	}
	log.Printf("Listed %d events for %s", len(shortEvents), hostURL)

	events := make([]*fbevents.EventData, 0, len(shortEvents))
	skippedCount := 0
	for _, shortEvent := range shortEvents {
		if !shouldScrapeEventDetails(shortEvent, cachedEvents) {
			skippedCount++
			continue
		}

		// Sleep 5-20s between every event detail scrape
		sleepRandom(randomGenerator, 5*time.Second, 20*time.Second)

		event, err := fbevents.ScrapeFbEvent(shortEvent.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("scrape Facebook event %s (%s) for host %s: %w", shortEvent.ID, shortEvent.Name, hostURL, err)
		}
		events = append(events, event)
	}

	log.Printf("Scraped %d new events for %s, skipped %d cancelled or already cached ones", len(events), hostURL, skippedCount)
	return events, nil
}

// scrapeEventsForHosts scrapes every host in the config. The cached events are
// only read, to tell which event details can be skipped.
func scrapeEventsForHosts(config *HostConfig, cachedEvents map[string]*fbevents.EventData) (map[string]*fbevents.EventData, error) {
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
		events, err := scrapeEventsForHost(hostURL, cachedEvents, randomGenerator)
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
			// Skip the cancelled events
			if event.IsCanceled {
				continue
			}

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
