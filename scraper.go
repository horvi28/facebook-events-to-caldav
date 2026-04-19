package main

import (
	"log"
	"math/rand"
	"time"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func sleepRandom(r *rand.Rand, min, max time.Duration) {
	duration := min + time.Duration(r.Int63n(int64(max-min)))
	time.Sleep(duration)
}

func scrapeEventsForHost(hostUrl string) ([]*fbevents.EventData, error) {
	// Scrape full event details but only for upcoming events
	options := &fbevents.Options{
		IncludePastEvents: false,
	}
	log.Printf("Scrape events for %s", hostUrl)
	events, err := fbevents.ScrapeFbEventListFull(hostUrl, options)
	if err != nil {
		log.Printf("Scrape failed for host %s: %v", hostUrl, err)
		return nil, err
	}
	log.Printf("Scraped %d events for %s", len(events), hostUrl)
	return events, nil
}

func scrapeEventsForHosts(hostUrls []string) []*fbevents.EventData {
	log.Printf("Start scraping events for %d hosts", len(hostUrls))

	// Init the slice whith at least hostUrl * 2 capacity
	allEvents := make([]*fbevents.EventData, 0, len(hostUrls)*2)

	// Random generator to sleep between scrapes of different hosts to avoid hitting Facebook too hard
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hardBlockCounter := 0
	for i, hostUrl := range hostUrls {
		events, err := scrapeEventsForHost(hostUrl)
		if err != nil {
			// Hard block detected, sleep for at least 1 hour (1-2 hours randomly)
			log.Printf("Hard block detected for host: %s, sleeping for 1-2 hours", hostUrl)
			hardBlockCounter++
			if hardBlockCounter > 5 {
				log.Fatalf("Too many hard blocks detected, stopping scraper")
				break
			}
			sleepRandom(r, 1*time.Hour, 2*time.Hour)
			continue // Skip this host and continue with next
		}
		allEvents = append(allEvents, events...)

		// Big pause after every 15 profiles, 60-120s
		if (i+1)%15 == 0 {
			log.Printf("[%d/%d] Scraping events, sleeping for 60-120 seconds", i+1, len(hostUrls))
			sleepRandom(r, 60*time.Second, 120*time.Second)
		} else {
			// Small sleep between every profile: 2-5s
			sleepRandom(r, 2*time.Second, 5*time.Second)
		}
	}
	return allEvents
}
