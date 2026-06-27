package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	ics "github.com/arran4/golang-ical"
	fbevents "github.com/grucha/facebook-events-scraper"
)

func createCalendar() *ics.Calendar {
	calendar := ics.NewCalendar()
	calendar.SetMethod(ics.MethodPublish)
	calendar.SetCalscale("GREGORIAN")

	return calendar
}

func createICalEventFromEventData(event *fbevents.EventData) (*ics.VEvent, error) {
	if event == nil {
		return nil, errors.New("event cannot be nil")
	}
	if event.ID == "" {
		return nil, errors.New("event ID cannot be empty")
	}

	vevent := ics.NewEvent(event.ID)
	vevent.SetSummary(event.Name)
	vevent.SetURL(event.URL)

	// Add the URL to the end of the description as well
	vevent.SetDescription(event.Description + "\n\n" + event.URL)

	eventStart := time.Unix(event.StartTimestamp, 0)
	vevent.SetStartAt(eventStart)

	// Event end is optional
	if event.EndTimestamp != nil {
		eventEnd := time.Unix(*event.EndTimestamp, 0)
		vevent.SetEndAt(eventEnd)
	}

	// Set location if available
	if event.Location != nil && event.Location.Name != "" {
		vevent.SetLocation(event.Location.Name)
	}
	if event.Location != nil && event.Location.Coordinates != nil {
		vevent.SetGeo(event.Location.Coordinates.Latitude, event.Location.Coordinates.Longitude)
	}

	if event.IsCanceled {
		vevent.SetStatus(ics.ObjectStatusCancelled)
	}

	vevent.SetColor("#FF0000") // Default color (red)
	// Set color based on event category if available (using the first category as a simple heuristic)
	if len(event.Categories) > 0 {
		vevent.SetColor(getEventColorBasedOnCategory(event.Categories[0]))
	}
	for _, category := range event.Categories {
		vevent.AddCategory(category.Label)
	}
	vevent.SetDtStampTime(time.Now())
	vevent.SetLastModifiedAt(time.Now())

	return vevent, nil
}

func getEventColorBasedOnCategory(category fbevents.EventCategory) string {
	// Unfortunately, Facebook does not give a category for every event
	switch category.Label {
	case "Music":
		return "#1DB954" // Spotify green
	case "Sports":
		return "#40E0D0" // Turquoise
	case "Party":
		return "#FFD700" // Gold
	case "Film":
		return "#0000FF" // Blue
	case "Gardening":
		return "#006400" // Dark green
	case "Theater":
		return "#800080" // Purple
	case "Games":
		return "#FF69B4" // Hot pink
	default:
		return "#FF0000" // Default color (red)
	}
}

func convertSingleEventToICal(event *fbevents.EventData) (*ics.Calendar, error) {
	vevent, err := createICalEventFromEventData(event)
	if err != nil {
		return nil, fmt.Errorf("create iCal event: %w", err)
	}

	log.Printf("Converting single event %s to iCal calendar", event.ID)
	calendar := createCalendar()
	calendar.AddVEvent(vevent)
	return calendar, nil
}

func convertEventsToICalCalendar(events map[string]*fbevents.EventData) (*ics.Calendar, error) {
	log.Printf("Converting %d events to iCal calendar", len(events))
	// Create calendar
	calendar := createCalendar()

	// Add events to the calendar
	labels := make(map[string]int) // To track unique labels for logging
	for _, event := range events {
		vevent, err := createICalEventFromEventData(event)
		if err != nil {
			return nil, fmt.Errorf("create iCal event: %w", err)
		}

		// Count labels for logging purposes
		if event.Categories != nil {
			for _, category := range event.Categories {
				labels[category.Label]++
			}
		}

		calendar.AddVEvent(vevent)
	}

	log.Printf("%d unique labels have been found in the events", len(labels))
	for label := range labels {
		log.Printf("Label: %s, Count: %d", label, labels[label])
	}

	log.Printf("Finished converting events to iCal calendar with %d events", len(calendar.Events()))
	return calendar, nil
}

func writeCalendarToFile(calendar *ics.Calendar, filePath string) error {
	log.Printf("Writing calendar to file: %s", filePath)
	if calendar == nil {
		return errors.New("calendar cannot be nil")
	}

	calendarContent := calendar.Serialize()
	if err := os.WriteFile(filePath, []byte(calendarContent), 0o644); err != nil {
		return fmt.Errorf("write calendar file %s: %w", filePath, err)
	}

	return nil
}
