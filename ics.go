package main

import (
	"errors"
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

func createICalEventFromEventData(event *fbevents.EventData) *ics.VEvent {
	if event == nil {
		return nil
	}

	vevent := ics.NewEvent(event.ID)
	vevent.SetSummary(event.Name)
	vevent.SetURL(event.URL)

	// Add the URL to the end of the description as well
	vevent.SetDescription(event.Description + "\n\n" + event.URL)

	eventStart := time.Unix(event.StartTimestamp, 0)
	vevent.SetStartAt(eventStart)

	if event.EndTimestamp != nil {
		eventEnd := time.Unix(*event.EndTimestamp, 0)
		vevent.SetEndAt(eventEnd)
	}

	if event.Location != nil && event.Location.Name != "" {
		vevent.SetLocation(event.Location.Name)
	}

	vevent.SetLastModifiedAt(time.Now())

	return vevent
}

func convertSingleEventToICal(event *fbevents.EventData) *ics.Calendar {
	log.Printf("Converting single event %s to iCal calendar", event.ID)
	calendar := createCalendar()
	vevent := createICalEventFromEventData(event)
	if vevent != nil {
		calendar.AddVEvent(vevent)
	}
	return calendar
}

func convertEventsToICalCalendar(events map[string]*fbevents.EventData) *ics.Calendar {
	log.Printf("Converting %d events to iCal calendar", len(events))
	// Create calendar
	calendar := createCalendar()

	// Add events to the calendar
	for _, event := range events {
		calendar.AddVEvent(createICalEventFromEventData(event))
	}
	log.Printf("Finished converting events to iCal calendar with %d events", len(calendar.Events()))
	return calendar
}

func writeCalendarToFile(calendar *ics.Calendar, filePath string) error {
	log.Printf("Writing calendar to file: %s", filePath)
	if calendar == nil {
		return errors.New("Calendar cannot be nil")
	}

	calendarContent := calendar.Serialize()
	return os.WriteFile(filePath, []byte(calendarContent), 0o644)
}
