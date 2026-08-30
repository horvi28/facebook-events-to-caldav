// Package main scrapes Facebook events and publishes them as iCalendar data to a CalDAV server.
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alexflint/go-arg"
	fbevents "github.com/grucha/facebook-events-scraper"
)

const defaultCalendarFilePath = "events.ics"

// eventsByID is every known event, keyed by the Facebook event ID
type eventsByID map[string]*fbevents.EventData

type Args struct {
	HostsFile     string `arg:"--hosts"          default:"hosts.yaml"  help:"File contains host's urls, for which events should be scraped"`
	ConfigFile    string `arg:"--config"         default:"config.yaml" help:"File that contains the scraping configuration"`
	FromCache     bool   `arg:"--from-cache"                           help:"Load events from the cache file instead of scraping"`
	CacheFile     string `arg:"--cache-file"     default:"events.json" help:"Events cache file"`
	StateFile     string `arg:"--state-file"     default:"state.json"  help:"File that remembers when each host was scraped the last time"`
	WriteCalendar bool   `arg:"--write-calendar"                       help:"Write events into a local calendar file"`
	CalendarFile  string `arg:"--calendar-file"  default:"events.ics"  help:"Output calendar file if --write-calendar enabled"`
	UploadCalDAV  bool   `arg:"--upload-caldav"                        help:"Upload events to the configured CalDAV server"`
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

	scrapeConfig, err := loadScrapeConfig(args.ConfigFile)
	if err != nil {
		return err
	}

	// Load the events either from the cache or scraping them
	var allEvents eventsByID
	var scrapedEventsCount int
	if args.FromCache {
		allEvents, err = loadEventsFromJSONFile(args.CacheFile)
	} else {
		allEvents, scrapedEventsCount, err = scrapeAndCacheEvents(hostConfig, scrapeConfig, args.CacheFile, args.StateFile)
	}
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

		if err := uploadEventsToCalDAV(allEvents, config, hostConfig.ColorOf); err != nil {
			return fmt.Errorf("upload events to CalDAV: %w", err)
		}
	}

	logSummary(len(allEvents), scrapedEventsCount, args)
	return nil
}

// scrapeAndCacheEvents returns every event that scraped freshly + events from the cache
// and how many of them were scraped by this run.
func scrapeAndCacheEvents(hostConfig *HostConfig, scrapeConfig ScrapeConfig, cacheFilePath, stateFilePath string) (eventsByID, int, error) {
	// Keep those events of the hosts that are not scraped this time, but already in the cache
	cachedEvents, err := loadCachedEvents(cacheFilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("load cached events: %w", err)
	}

	// Drop the events that are over, otherwise the cache would only grow
	cachedEvents = pruneOutdatedEvents(cachedEvents, time.Now())

	// Remembers when each host was scraped the last time
	state, err := loadScrapeState(stateFilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("load scrape state: %w", err)
	}

	scrapedEvents, err := scrapeEventsForHosts(hostConfig, cachedEvents, scrapeConfig, state)
	if err != nil {
		return nil, 0, fmt.Errorf("scrape events for hosts: %w", err)
	}

	if err := saveScrapeState(state, stateFilePath); err != nil {
		return nil, 0, fmt.Errorf("save scrape state: %w", err)
	}

	// Merge the events in the cache and the newly scraped ones
	allEvents := mergeEvents(cachedEvents, scrapedEvents)

	if err := dumpEventsToJSONFile(allEvents, cacheFilePath); err != nil {
		return nil, 0, fmt.Errorf("dump events to JSON file: %w", err)
	}

	return allEvents, len(scrapedEvents), nil
}

func writeEventsCalendar(events eventsByID, eventsCalendarPath string, colorByCategory map[string]string) error {
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

func logSummary(eventsCount, scrapedEventsCount int, args Args) {
	log.Printf("Summary:")
	if !args.FromCache {
		// Only the most overdue hosts are scraped, so most of the events come
		// from the cache and the total says nothing about this run
		log.Printf("Scraped %d new events, %d in total with the cached ones", scrapedEventsCount, eventsCount)
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
