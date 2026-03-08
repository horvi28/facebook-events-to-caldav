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
	vevent.SetURL(event.URL)

	// Add the URL to the end of the description as well
	vevent.SetDescription(event.Description + "\n\n" + event.URL)

	eventStart := time.Unix(event.StartTimestamp, 0)
	vevent.SetStartAt(eventStart)

	if event.EndTimestamp != nil {
		eventEnd := time.Unix(*event.EndTimestamp, 0)
		vevent.SetEndAt(eventEnd)
	}

	vevent.SetLocation(event.Location.Name)

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
