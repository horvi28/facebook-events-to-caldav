package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func doRequest(httpClient *http.Client, method, url, username, password string, body *strings.Reader) (*http.Response, error) {
	// Use a per-request timeout so one slow request doesn't hang the whole run
	requestCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use a direct HTTP PUT so we can set Content-Type explicitly
	request, err := http.NewRequestWithContext(requestCtx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create %s request for %s: %w", method, url, err)
	}
	request.Header.Set("Content-Type", "text/calendar; charset=utf-8")
	request.SetBasicAuth(username, password)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send %s request to %s: %w", method, url, err)
	}

	return response, nil
}

// uploadEventToCalDAV uploads a single event to the CalDAV server
func uploadEventToCalDAV(httpClient *http.Client, event *fbevents.EventData, config Config) (err error) {
	if event == nil {
		return errors.New("event cannot be nil")
	}
	log.Printf("Uploading event: '%s' ...", event.Name)

	// Create a single-event calendar
	calendar, err := convertSingleEventToICal(event)
	if err != nil {
		return fmt.Errorf("convert event %q to iCal: %w", event.Name, err)
	}

	// Serialize to ICS format
	icsData := calendar.Serialize()

	// Path on Radicale
	remoteURL := config.CalDAVURL + config.CalendarPath + event.ID + ".ics"

	response, err := doRequest(httpClient, http.MethodPut, remoteURL, config.Username, config.Password, strings.NewReader(icsData))
	if err != nil {
		return fmt.Errorf("upload event %q: %w", event.Name, err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close upload response for event %q: %w", event.Name, closeErr)
		}
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
		if readErr != nil {
			return fmt.Errorf("read failed upload response for event %q: %w", event.Name, readErr)
		}

		return fmt.Errorf("upload event %q: unexpected status %s: %s", event.Name, response.Status, strings.TrimSpace(string(responseBody)))
	}

	log.Printf("Upload response for event '%s'. Status code: %s", event.Name, response.Status)

	return nil
}

// uploadEventsToCalDAV uploads all events to the CalDAV server
func uploadEventsToCalDAV(events map[string]*fbevents.EventData, config Config) error {
	// Create a HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Upload each event
	var uploadErrors []error
	for _, event := range events {
		err := uploadEventToCalDAV(httpClient, event, config)
		if err != nil {
			uploadErrors = append(uploadErrors, err)
		}
	}

	if len(uploadErrors) > 0 {
		return fmt.Errorf("upload %d of %d events: %w", len(uploadErrors), len(events), errors.Join(uploadErrors...))
	}

	log.Printf("Finished uploading %d events to CalDAV server", len(events))
	return nil
}
