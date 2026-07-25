// Package main scrapes Facebook events and publishes them as iCalendar data to a CalDAV server.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
	fbevents "github.com/grucha/facebook-events-scraper"
)

const defaultCalendarFilePath = "events.ics"

type Args struct {
	HostsFile     string `arg:"--hosts"          default:"hosts.txt"   help:"Path to the file that contains Facebook host URLs, for which events should be scraped"`
	FromCache     bool   `arg:"--from-cache"                           help:"Load events from the cache file instead of scraping"`
	CacheFile     string `arg:"--cache-file"     default:"events.json" help:"Path to the cache file"`
	WriteCalendar bool   `arg:"--write-calendar"                       help:"Write events into a local calendar file"`
	CalendarFile  string `arg:"--calendar-file"  default:"events.ics"  help:"Path to the calendar file"`
	UploadCalDAV  bool   `arg:"--upload-caldav"                        help:"Upload events to the CalDAV server"`
}

func run() error {
	var args Args
	arg.MustParse(&args)

	// If the user has added a separate calendar file, but --write-calendar is no given, throw an error
	if !args.WriteCalendar && args.CalendarFile != defaultCalendarFilePath {
		return errors.New("--calendar-file requires --write-calendar")
	}

	hostConfig, err := loadHostConfig(args.HostsFile)
	if err != nil {
		return err
	}

	allEvents, err := loadEvents(hostConfig, args.FromCache, args.CacheFile)
	if err != nil {
		return err
	}

	if args.WriteCalendar {
		if err := writeEventsCalendar(allEvents, args.CalendarFile, hostConfig.ColorOf); err != nil {
			return err
		}
	}

	if args.UploadCalDAV {
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("load configuration: %w", err)
		}

		if err := uploadEventsToCalDAV(allEvents, config); err != nil {
			return fmt.Errorf("upload events to CalDAV: %w", err)
		}
	}

	logSummary(len(allEvents), args)
	return nil
}

func loadEvents(hostConfig *HostConfig, fromCache bool, cacheFilePath string) (map[string]*fbevents.EventData, error) {
	if fromCache {
		allEvents, err := loadEventsFromJSONFile(cacheFilePath)
		if err != nil {
			return nil, fmt.Errorf("load events from JSON file: %w", err)
		}
		return allEvents, nil
	}

	allEvents, err := scrapeEventsForHosts(hostConfig)
	if err != nil {
		return nil, fmt.Errorf("scrape events for hosts: %w", err)
	}

	if err := dumpEventsToJSONFile(allEvents, cacheFilePath); err != nil {
		return nil, fmt.Errorf("dump events to JSON file: %w", err)
	}

	return allEvents, nil
}

func writeEventsCalendar(events map[string]*fbevents.EventData, eventsCalendarPath string, colorByCategory map[string]string) error {
	// Convert the events to an iCal calendar and write it to a file
	calendar, err := convertEventsToICalCalendar(events, colorByCategory)
	if err != nil {
		return fmt.Errorf("convert events to iCal calendar: %w", err)
	}

	if err := writeCalendarToFile(calendar, eventsCalendarPath); err != nil {
		return fmt.Errorf("write calendar file: %w", err)
	}

	return nil
}

func logSummary(eventsCount int, args Args) {
	log.Printf("Summary:")
	if !args.FromCache {
		log.Printf("Scraped %d events", eventsCount)
	} else {
		log.Printf("Loaded %d events from JSON file: %s", eventsCount, args.CacheFile)
	}

	if args.WriteCalendar {
		log.Printf("Wrote events into %s", args.CalendarFile)
	} else {
		log.Printf("Skipped writing events into local file")
	}

	if args.UploadCalDAV {
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
