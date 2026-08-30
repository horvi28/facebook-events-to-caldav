package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"time"
)

// scrapeState remembers when each host was scraped successfully the last time.
//
// This is what lets a run pick the hosts that have been waiting the longest
// instead of going through all of them every time. A host is only recorded
// after it succeeded, so a failed one stays overdue and is picked again by the
// next run instead of being treated as done.
type scrapeState map[string]time.Time

// loadScrapeState reads the state file. A missing file is an empty state
// instead of an error, which is the expected situation on the very first run.
func loadScrapeState(statePath string) (scrapeState, error) {
	log.Printf("Loading state file: %s", statePath)

	stateJSON, err := os.ReadFile(statePath)
	if errors.Is(err, os.ErrNotExist) {
		// This is not an error, it means no recorded hosts yet
		log.Printf("No state file at %s yet, no host has been scraped so far", statePath)
		return scrapeState{}, nil
	}
	// Catch different kind of errors than the 'os.ErrNotExist' as well
	if err != nil {
		return nil, fmt.Errorf("read state file %s: %w", statePath, err)
	}

	var state scrapeState
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return nil, fmt.Errorf("decode state file %s: %w", statePath, err)
	}

	log.Printf("Loaded the last scrape time of %d hosts from %s", len(state), statePath)
	return state, nil
}

func saveScrapeState(state scrapeState, statePath string) error {
	stateJSON, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state to JSON: %w", err)
	}

	log.Printf("Saving the last scrape time of %d hosts to %s", len(state), statePath)
	if err := os.WriteFile(statePath, stateJSON, fileMode); err != nil {
		return fmt.Errorf("write state file %s: %w", statePath, err)
	}

	return nil
}

// selectHostsToScrape picks the hosts that have been waiting the longest.
//
// Scraping every host in one run is what makes the traffic look like a crawl,
// so a run only takes the most overdue ones and the rest follow on the next
// runs. A host that is missing from the state has never been scraped and comes
// first, and hosts that are equally overdue keep the order of the hosts file.
func selectHostsToScrape(hosts []string, state scrapeState, hostsPerRun int) []string {
	// Sorted on a copy, as the caller keeps using its own order
	overdueFirst := slices.Clone(hosts)
	slices.SortStableFunc(overdueFirst, func(firstHost, secondHost string) int {
		// A host that is not in the state gets the zero time, which sorts first
		return state[firstHost].Compare(state[secondHost])
	})

	// Slice it to the number of hostsPerRun or if the list is smaller then the whole list
	return overdueFirst[:min(hostsPerRun, len(overdueFirst))]
}
