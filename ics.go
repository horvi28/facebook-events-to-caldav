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

func createICalEventFromEventData(event *fbevents.EventData, colorByCategory map[string]string) (*ics.VEvent, error) {
	if event == nil {
		return nil, errors.New("event cannot be nil")
	}
	if event.ID == "" {
		return nil, errors.New("event ID cannot be empty")
	}

	vevent := ics.NewEvent(event.ID)
	vevent.SetSummary(event.Name)
	vevent.SetURL(event.URL)

	// Add the URL to the start of the description as well
	vevent.SetDescription(event.URL + "\n\n" + event.Description)

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

	// Set color based on event category if available (The first match will win)
	vevent.SetColor(getEventColorBasedOnCategory(event.Categories, colorByCategory))

	for _, category := range event.Categories {
		vevent.AddCategory(category.Label)
	}

	vevent.SetDtStampTime(time.Now())
	vevent.SetLastModifiedAt(time.Now())

	return vevent, nil
}

func getEventColorBasedOnCategory(categories []fbevents.EventCategory, colorByCategory map[string]string) string {
	// The first category which is in the map will win
	for _, category := range categories {
		if color := colorByCategory[category.Label]; color != "" {
			return color
		}
	}

	// Then fall back to the default color (Red)
	return "#FF0000"
}

func convertSingleEventToICal(event *fbevents.EventData, colorByCategory map[string]string) (*ics.Calendar, error) {
	vevent, err := createICalEventFromEventData(event, colorByCategory)
	if err != nil {
		return nil, fmt.Errorf("create iCal event: %w", err)
	}

	log.Printf("Converting single event %s to iCal calendar", event.ID)
	calendar := createCalendar()
	calendar.AddVEvent(vevent)
	return calendar, nil
}

func convertEventsToICalCalendar(events map[string]*fbevents.EventData, colorByCategory map[string]string) (*ics.Calendar, error) {
	log.Printf("Converting %d events to iCal calendar", len(events))
	// Create calendar
	calendar := createCalendar()

	// Add events to the calendar
	labels := make(map[string]int) // To track unique labels for logging
	for _, event := range events {
		vevent, err := createICalEventFromEventData(event, colorByCategory)
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
