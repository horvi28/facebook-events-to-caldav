// Package main scrapes Facebook events and publishes them as iCalendar data to a CalDAV server.
package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
	fbevents "github.com/grucha/facebook-events-scraper"
)

type Args struct {
	SkipScraping bool `arg:"--skip-scraping" help:"skip scraping and load events from JSON file instead (./events.json)"`
}

func run() error {
	var args Args
	arg.MustParse(&args)

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	allEvents, err := loadEvents(args.SkipScraping)
	if err != nil {
		return err
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

func loadEvents(skipScraping bool) (map[string]*fbevents.EventData, error) {
	const eventsJSONPath = "events.json"

	if skipScraping {
		allEvents, err := loadEventsFromJSONFile(eventsJSONPath)
		if err != nil {
			return nil, fmt.Errorf("load events from JSON file: %w", err)
		}
		return allEvents, nil
	}

	hostList, err := getHostList()
	if err != nil {
		return nil, fmt.Errorf("load host list: %w", err)
	}

	allEvents, err := scrapeEventsForHosts(hostList)
	if err != nil {
		return nil, fmt.Errorf("scrape events for hosts: %w", err)
	}

	if err := dumpEventsToJSONFile(allEvents, eventsJSONPath); err != nil {
		return nil, fmt.Errorf("dump events to JSON file: %w", err)
	}

	return allEvents, nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
