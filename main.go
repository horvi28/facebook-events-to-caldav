package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	eventUrl := "https://www.facebook.com/bolygo.hu/upcoming_hosted_events"

	// Launch browser
	u := launcher.New().
		Headless(true).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(eventUrl)

	// Wait for dynamic content
	page.MustWaitLoad()
	time.Sleep(5 * time.Second)

	elements, err := page.Elements("a[href*='/events/']")
	if err != nil {
		log.Fatal(err)
	}

	seen := make(map[string]bool)

	for _, el := range elements {
		href, _ := el.Attribute("href")
		if href != nil && !seen[*href] {
			fmt.Println(*href)
			seen[*href] = true
		}
	}
}
