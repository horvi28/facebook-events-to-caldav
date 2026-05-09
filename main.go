// Package main scrapes Facebook events and publishes them as iCalendar data to a CalDAV server.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"
	fbevents "github.com/grucha/facebook-events-scraper"
)

type Args struct {
	SkipScraping bool `arg:"--skip-scraping" help:"skip scraping and load events from JSON file instead (./events.json)"`
}

func getHostList() (hostList []string, err error) {
	const hostsFilePath = "hosts.txt"

	file, err := os.Open(hostsFilePath)
	if err != nil {
		return nil, fmt.Errorf("open hosts file %s: %w", hostsFilePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close hosts file %s: %w", hostsFilePath, closeErr)
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read hosts file %s: %w", hostsFilePath, err)
	}

	return lines, nil
}

func run() error {
	var args Args
	arg.MustParse(&args)

	const eventsJSONPath = "events.json"

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	var allEvents map[string]*fbevents.EventData

	if args.SkipScraping {
		eventsFromJSON, err := loadEventsFromJSONFile(eventsJSONPath)
		if err != nil {
			return fmt.Errorf("load events from JSON file: %w", err)
		}
		// Deduplicate events from JSON and convert to a map
		allEvents = make(map[string]*fbevents.EventData)
		for _, event := range eventsFromJSON {
			allEvents[event.ID] = event
		}
	} else {
		hostList, err := getHostList()
		if err != nil {
			return fmt.Errorf("load host list: %w", err)
		}

		allEvents, err = scrapeEventsForHosts(hostList)
		if err != nil {
			return fmt.Errorf("scrape events for hosts: %w", err)
		}

		// Dump to JSON for caching, but first convert it to a slice
		eventsSlice := make([]*fbevents.EventData, 0, len(allEvents))
		for _, event := range allEvents {
			eventsSlice = append(eventsSlice, event)
		}
		if err := dumpEventsToJSONFile(eventsSlice, eventsJSONPath); err != nil {
			return fmt.Errorf("dump events to JSON file: %w", err)
		}
	}

	// Convert the events to an iCal calendar and write it to a file
	calendar, err := convertEventsToICalCalendar(allEvents)
	if err != nil {
		return fmt.Errorf("convert events to iCal calendar: %w", err)
	}

	if err := writeCalendarToFile(calendar, "events.ics"); err != nil {
		return fmt.Errorf("write calendar file: %w", err)
	}

	// Upload to CalDAV server
	if err := uploadEventsToCalDAV(allEvents, config); err != nil {
		return fmt.Errorf("upload events to CalDAV: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
