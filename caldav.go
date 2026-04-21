package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func doRequest(httpClient *http.Client, method, url string, body *strings.Reader) (*http.Response, error) {
	// Use a per-request timeout so one slow request doesn't hang the whole run
	requestCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use a direct HTTP PUT so we can set Content-Type explicitly
	request, err := http.NewRequestWithContext(requestCtx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("Failed to create %s request for %s: %v", method, url, err)
	}
	//Basic auth using configured credentials
	request.SetBasicAuth(config.username, config.password)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Request failed: %v", err)
	}
	defer response.Body.Close()
	return response, nil
}

// uploadEventToCalDAV uploads a single event to the CalDAV server
func uploadEventToCalDAV(httpClient *http.Client, event *fbevents.EventData) error {
	if event == nil {
		return fmt.Errorf("Event cannot be nil")
	}
	log.Printf("Uploading event: '%s' ...", event.Name)

	// Create a single-event calendar
	calendar := convertSingleEventToICal(event)

	// Serialize to ICS format
	icsData := calendar.Serialize()

	// Path on Radicale
	remoteUrl := config.radicaleURL + config.calendarPath + event.ID + ".ics"

	response, err := doRequest(httpClient, http.MethodPut, remoteUrl, strings.NewReader(icsData))
	log.Printf("Upload response for event '%s'. Status code: %s", event.Name, response.Status)

	if err != nil {
		return fmt.Errorf("Failed to upload event %s: %v", event.Name, err)
	}

	return nil
}

// uploadEventsToCalDAV uploads all events to the CalDAV server
func uploadEventsToCalDAV(events map[string]*fbevents.EventData) error {
	// Create a HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Upload each event
	for _, event := range events {
		err := uploadEventToCalDAV(httpClient, event)
		if err != nil {
			log.Printf("Failed to upload event %s: %v", event.Name, err)
		}
	}

	log.Printf("Finished uploading %d events to CalDAV server", len(events))
	return nil
}
