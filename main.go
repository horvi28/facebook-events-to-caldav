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
	const loadFromJson = true

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

	// Convert the events to an iCal calendar and write it to a file
	calendar := convertEventsToICalCalendar(events)
	writeCalendarToFile(calendar, "events.ics")
}
