// Package main scrapes Facebook events and publishes them as iCalendar data to a CalDAV server.
package main

import (
	"bufio"
	"log"
	"os"

	fbevents "github.com/grucha/facebook-events-scraper"
)

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
	const loadFromJSON = true

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	var allEvents map[string]*fbevents.EventData

	if loadFromJSON {
		eventsFromJSON, err := loadEventsFromJSONFile()
		if err != nil {
			log.Fatalf("Event load from json failed: %v", err)
		}
		// Deduplicate events from JSON and convert to a map
		allEvents = make(map[string]*fbevents.EventData)
		for _, event := range eventsFromJSON {
			allEvents[event.ID] = event
		}
	} else {
		allEvents = scrapeEventsForHosts(getHostList())
		// Dump to JSON for caching, but first convert it to a slice
		eventsSlice := make([]*fbevents.EventData, 0, len(allEvents))
		for _, event := range allEvents {
			eventsSlice = append(eventsSlice, event)
		}
		if err := dumpEventsToJSONFile(eventsSlice); err != nil {
			log.Printf("Event dump to json failed: %v", err)
		}
	}

	// Convert the events to an iCal calendar and write it to a file
	calendar := convertEventsToICalCalendar(allEvents)
	writeCalendarToFile(calendar, "events.ics")

	// Upload to CalDAV server
	uploadEventsToCalDAV(allEvents, config)
}
