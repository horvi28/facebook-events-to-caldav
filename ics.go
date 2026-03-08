package main

import (
	"errors"
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
	vevent.SetDescription(event.Description)
	vevent.SetURL(event.URL)

	eventStart := time.Unix(event.StartTimestamp, 0)
	vevent.SetStartAt(eventStart)

	if event.EndTimestamp != nil {
		eventEnd := time.Unix(*event.EndTimestamp, 0)
		vevent.SetEndAt(eventEnd)
	}

	vevent.SetLastModifiedAt(time.Now())

	return vevent
}

func writeCalendarToFile(calendar *ics.Calendar, filePath string) error {
	if calendar == nil {
		return errors.New("Calendar cannot be nil")
	}

	calendarContent := calendar.Serialize()
	return os.WriteFile(filePath, []byte(calendarContent), 0o644)
}
