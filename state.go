package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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
	if err := os.WriteFile(statePath, stateJSON, 0o644); err != nil {
		return fmt.Errorf("write state file %s: %w", statePath, err)
	}

	return nil
}
