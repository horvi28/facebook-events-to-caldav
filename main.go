package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func scrapeEventsForHost(hostUrl string) []*fbevents.EventData {
	// Scrape full event details but only for upcoming events
	options := &fbevents.Options{
		IncludePastEvents: false,
	}
	events, err := fbevents.ScrapeFbEventListFull(hostUrl, options)
	if err != nil {
		log.Fatalf("Scrape failed for host %s: %v", hostUrl, err)
	}
	return events
}

func scrapeEventsForHosts(hostUrls []string) []*fbevents.EventData {
	allEvents := make([]*fbevents.EventData, 0, len(hostUrls)*2) // Init the slice whith at least hostUrl * 2 capacity
	for _, hostUrl := range hostUrls {
		events := scrapeEventsForHost(hostUrl)
		allEvents = append(allEvents, events...)
	}
	return allEvents
}

func getHostList() []string {
	const hostsFilePath = "hosts.txt"

	file, err := os.Open(hostsFilePath)
	if err != nil {
		log.Fatalf("Error opening %s file: %v", hostsFilePath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	return lines
}

func main() {
	const loadFromJson = false

	var (
		events []*fbevents.EventData
		err    error
	)

	if loadFromJson {
		events, err = loadEventsFromJsonFile()
		if err != nil {
			log.Fatalf("Event load from json failed: %v", err)
		}
	} else {
		events = scrapeEventsForHosts(getHostList())

		if err := dumpEventsToJsonFile(events); err != nil {
			log.Printf("Event dump to json failed: %v", err)
		}
	}

	// Crete calendar
	calendar := createCalendar()

	// Add events to the calendar
	// Only process an event once
	seen := make(map[string]bool)
	for _, event := range events {
		if !seen[event.ID] {
			calendar.AddVEvent(createICalEventFromEventData(event))
			seen[event.ID] = true
		}
		break
	}
	fmt.Printf("%s", calendar.Serialize())

	writeCalendarToFile(calendar, "events.ics")
}
