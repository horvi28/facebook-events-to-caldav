package main

import (
	"testing"
	"time"
)

func TestNextPause(t *testing.T) {
	const daemonInterval = 24 * time.Hour

	testCases := []struct {
		name          string
		cycleDuration time.Duration
		want          time.Duration
	}{
		{
			name:          "a short scrape waits the rest of the interval",
			cycleDuration: 30 * time.Minute,
			want:          daemonInterval - 30*time.Minute,
		},
		{
			name:          "an instant scrape waits the whole interval",
			cycleDuration: 0,
			want:          daemonInterval,
		},
		{
			// Otherwise the daemon would scrape without a break
			name:          "a scrape as long as the interval waits a whole interval",
			cycleDuration: daemonInterval,
			want:          daemonInterval,
		},
		{
			name:          "a scrape longer than the interval waits a whole interval",
			cycleDuration: daemonInterval + time.Hour,
			want:          daemonInterval,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if pause := nextPause(testCase.cycleDuration, daemonInterval); pause != testCase.want {
				t.Errorf("nextPause(%v, %v) = %v, want %v", testCase.cycleDuration, daemonInterval, pause, testCase.want)
			}
		})
	}
}
