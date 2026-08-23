package main

import (
	"testing"

	fbevents "github.com/grucha/facebook-events-scraper"
)

func TestGetOrganizersAsString(t *testing.T) {
	hosts := []fbevents.EventHost{
		{ID: "1", Name: "foo", URL: "foo.com", Type: fbevents.HostTypeUser},
		{ID: "2", Name: "bar"},
		{ID: "3", Name: "baz", Type: fbevents.HostTypeGroup},
	}
	organizers := getOrganizersAsString(hosts)
	if organizers != "foo, bar, baz" {
		t.Errorf("Host names should be concaneted in order separated by a comma, got %q", organizers)
	}
}

func TestGetEventColorBasedOnCategory(t *testing.T) {
	colorByCategory := map[string]string{
		"Sports":  "#1DB954",
		"Concert": "#800020",
	}

	tests := []struct {
		name          string
		categories    []fbevents.EventCategory
		expectedColor string
	}{
		{
			name: "first matching category wins",
			categories: []fbevents.EventCategory{
				{Label: "Music", URL: "foo.com"},
				{Label: "Sports", URL: "foo.com"},
				{Label: "Concert"},
			},
			expectedColor: "#1DB954",
		},
		{
			name: "no category found in map falls back to default",
			categories: []fbevents.EventCategory{
				{Label: "Music"},
			},
			expectedColor: "#FF0000",
		},
		{
			name:          "no categories at all falls back to default",
			categories:    nil,
			expectedColor: "#FF0000",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if color := getEventColorBasedOnCategory(testCase.categories, colorByCategory); color != testCase.expectedColor {
				t.Errorf("getEventColorBasedOnCategory() = %q, want %q", color, testCase.expectedColor)
			}
		})
	}
}
