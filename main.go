// Package main scrapes Facebook events and publishes them as iCalendar data to a CalDAV server.
package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
	fbevents "github.com/grucha/facebook-events-scraper"
)

type Args struct {
	SkipScraping     bool   `arg:"--skip-scraping" help:"skip scraping and load events from JSON file instead (./events.json)"`
	SkipCalendarFile bool   `arg:"--skip-calendar-file" help:"skip writing events to the local events.ics file"`
	SkipUpload       bool   `arg:"--skip-upload" help:"skip uploading events to the CalDAV server"`
	HostsFile        string `arg:"--hosts" default:"hosts.txt" help:"path to the file that contains Facebook host URLs, for which events should be scraped"`
}

func run() error {
	var args Args
	arg.MustParse(&args)

	allEvents, err := loadEvents(args.SkipScraping, args.HostsFile)
	if err != nil {
		return err
	}

	if !args.SkipCalendarFile {
		if err := writeEventsCalendar(allEvents); err != nil {
			return err
		}
	}

	if !args.SkipUpload {
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("load configuration: %w", err)
		}

		if err := uploadEventsToCalDAV(allEvents, config); err != nil {
			return fmt.Errorf("upload events to CalDAV: %w", err)
		}
	}

	logSummary(len(allEvents), !args.SkipScraping, !args.SkipCalendarFile, !args.SkipUpload)
	return nil
}

func loadEvents(skipScraping bool, hostsFilePath string) (map[string]*fbevents.EventData, error) {
	const eventsJSONPath = "events.json"

	if skipScraping {
		allEvents, err := loadEventsFromJSONFile(eventsJSONPath)
		if err != nil {
			return nil, fmt.Errorf("load events from JSON file: %w", err)
		}
		return allEvents, nil
	}

	hostList, err := getHostList(hostsFilePath)
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

func writeEventsCalendar(events map[string]*fbevents.EventData) error {
	// Convert the events to an iCal calendar and write it to a file
	const eventsCalendarPath = "events.ics"

	calendar, err := convertEventsToICalCalendar(events)
	if err != nil {
		return fmt.Errorf("convert events to iCal calendar: %w", err)
	}

	if err := writeCalendarToFile(calendar, eventsCalendarPath); err != nil {
		return fmt.Errorf("write calendar file: %w", err)
	}

	return nil
}

func logSummary(eventsCount int, scrapingRun bool, calendarFileWritten bool, eventsUploaded bool) {
	log.Printf("Summary:")
	if scrapingRun {
		log.Printf("Scraped %d events", eventsCount)
	} else {
		log.Printf("Loaded %d events from JSON file", eventsCount)
	}

	if calendarFileWritten {
		log.Printf("Wrote events.ics")
	} else {
		log.Printf("Skipped events.ics")
	}

	if eventsUploaded {
		log.Printf("Uploaded %d events to CalDAV", eventsCount)
	} else {
		log.Printf("Skipped CalDAV upload")
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
