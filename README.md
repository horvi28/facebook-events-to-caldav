# Facebook Event Scraper

A Facebook event scraper that scrapes upcoming events from a list of Facebook hosts. It can also write events to an iCalendar file or upload each event to a CalDAV server.

> [!WARNING]
> WIP: This project is in an early stage and a work in progress. I started it as a way to learn the Go language.
>
> **Disclaimer:** Facebook's Terms of Service explicitly prohibit scraping data from their website. Use this project at your own risk.

## Features

- Scrapes upcoming events from Facebook host URLs grouped by category in `hosts.yaml`.
- Deduplicates events by Facebook event ID.
- Can write a local `events.ics` calendar file, colored by your own per-category colors.
- Keeps the scraped events in `events.json`, which doubles as the record of what has already been fetched, so the details of an event are only ever requested once. See [The event cache](#the-event-cache) section.
- Drops events from the cache once they are over, so it does not grow without end.
- Can upload individual `.ics` events to a CalDAV server.
- Supports `.env` based configuration for CalDAV credentials.

## Usage

Create a `hosts.yaml` file grouping the Facebook profiles you want to scrape under your own categories.
Each category may set an optional `color` (hex) used for its events in the calendar.
Facebook rarely tags events with a category, so the one you assign drives the color. See `hosts.example.yaml` for the format.

### Arguments

| Argument           | Default       | Description                                                     |
| ------------------ | ------------- | --------------------------------------------------------------- |
| `--hosts`          | `hosts.yaml`  | Path to the YAML file mapping categories to Facebook host URLs. |
| `--from-cache`     | `false`       | Load events from the cache file instead of scraping Facebook.   |
| `--cache-file`     | `events.json` | Path to the JSON cache file.                                    |
| `--write-calendar` | `false`       | Write events to a local iCalendar file.                         |
| `--calendar-file`  | `events.ics`  | Path to the local iCalendar file. Requires `--write-calendar`.  |
| `--upload-caldav`  | `false`       | Upload events to the configured CalDAV server.                  |

### Examples

```sh
# By default, the scraper reads `hosts.yaml`, scrapes events, and writes only the cache (`events.json`):
go run .

# Use custom input and output paths:
go run . --hosts private-hosts.yaml --cache-file path-to-events.json --write-calendar --calendar-file path-to-events.ics

# Write a local calendar file. This creates `events.ics` in the current directory:
go run . --write-calendar

# Write a local calendar file to another path:
go run . --write-calendar --calendar-file path-to-events.ics

# Scrape events and upload them to CalDAV:
go run . --upload-caldav

# Load events from the cache (events.json) instead of scraping:
go run . --from-cache
```

## Configuration

Create a local `.env` file using `.env.example` as a template:

```env
CALDAV_URL=http://localhost:5232
CALDAV_USERNAME=your-username
CALDAV_PASSWORD=your-password
CALDAV_CALENDAR_PATH=/your-user/your-calendar-id/
```

The `.env` file is ignored by Git so credentials stay local. CalDAV configuration is only needed when running with `--upload-caldav`.

## Scraping behavior

This project uses [`grucha/facebook-events-scraper`](https://github.com/grucha/facebook-events-scraper) to scrape Facebook event data. All credit for the scraper package goes to its original author. This project is a wrapper around it.

The scraper only requests upcoming events. Past events are not included.

### How many requests a scrape costs

Scraping a host is not a single request. It happens in two phases:

1. **Listing** the host's upcoming events. One request, and it returns only a summary of each event.
2. **Fetching the details** of an event. One request *per event*, for the details of the event.

So a host with 20 upcoming events would cost 21 requests, and the detail requests are by far the larger part of the total. To try to keep that number down and event cache is used.

### The event cache

`events.json` is not only a convenience for local development, it is what makes the scraping cheap. It holds every event known so far, and the IDs in it are what the scraper checks to decide whether a detail request can be skipped.

The flow is the following:

- Each run loads the cache and drops the events that are already passed. -> This way it will not grow forever.
- Scrapes the new events, but only those, which is not in the `events.json`.
- Merges the new events into the cache.

In practice this means the first run is expensive and later runs are cheaper, as they only pay for the events that are genuinely new.

Two consequences worth knowing:

- **Event details are never refreshed.** If a location or a start time changes after the event was first cached, the change is not picked up. Delete the event from `events.json` (or the whole file) to force a re-fetch. -> This might change with a heuristic in the future, but for now, this can save a lot of request.
- **Moving a profile to another category does not recolor its existing events.** The category is stamped onto an event when its details are scraped, so only events scraped after the change get the new color.

> [!WARNING]
> Deleting `events.json` means the next run fetches the details of **every** listed event again. For a large `hosts.yaml` that is several hundred extra requests in one go, which is exactly the kind of traffic that gets noticed. Only delete it when you actually want a full re-scrape.

Events are kept for 48 hours after they end, so a scrape running in the evening does not empty out the current day. An event without an end time, which is roughly half of them, counts as ending when it starts.

### Pacing

To avoid hitting Facebook too hard, the scraper waits:

- 5-20 seconds between the detail requests of two events
- 10-30 seconds after each host
- 10-20 minutes after every 10 hosts
- 1-2 hours after a failed host before continuing, and it gives up entirely once more than 5 hosts have failed

The host list is processed one host at a time. Events are deduplicated by Facebook event ID, so if the same event appears for multiple hosts, it is only kept once.

## My setup

I run this project regularly on my server and upload the scraped events to a [Radicale](https://radicale.org/) CalDAV server.

On Android, I use [DAVx5](https://www.davx5.com/) to sync the CalDAV calendar to my phone and [KashCal](https://kashcal.onekash.org/) to view the events. It also works with **Google Calendar**, but I wanted to keep these events in a separate app. [Etar Calendar](https://github.com/Etar-Group/Etar-Calendar) works well too.

On the web, I use [InfCloud](https://inf-it.com/open-source/clients/infcloud/) as a CalDAV client, mostly because I have not found another lightweight standalone web client that fits this setup yet.

## Credits

This project directly uses these open source packages:

- [`grucha/facebook-events-scraper`](https://github.com/grucha/facebook-events-scraper) for scraping Facebook event data.
- [`arran4/golang-ical`](https://github.com/arran4/golang-ical) for creating iCalendar files.
- [`alexflint/go-arg`](https://github.com/alexflint/go-arg) for command line argument parsing.
- [`caarlos0/env`](https://github.com/caarlos0/env) for parsing environment variables.
- [`joho/godotenv`](https://github.com/joho/godotenv) for loading local `.env` files.
