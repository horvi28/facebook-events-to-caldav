# Facebook Event Scraper

> [!WARNING]
> WIP: This project is in an early stage and a work in progress. I started it as a way to learn the Go language.

A facebook event scraper, that scrapes upcoming events from a list of Facebook hosts, writes them to an iCalendar file, and can upload each event to a CalDAV server.

## Features

- Scrapes upcoming events from Facebook host URLs listed in `hosts.txt`.
- Deduplicates events by Facebook event ID.
- Can write a local `events.ics` calendar file.
- Caches scraped event data in `events.json` for faster local development.
- Can upload individual `.ics` events to a CalDAV server.
- Supports `.env` based configuration for CalDAV credentials.

## Configuration

Create a local `.env` file using `.env.example` as the template:

```env
CALDAV_URL=http://localhost:5232
CALDAV_USERNAME=your-username
CALDAV_PASSWORD=your-password
CALDAV_CALENDAR_PATH=/your-user/your-calendar-id/
```

The `.env` file is ignored by Git so credentials stay local. CalDAV configuration is only needed when running with `--upload-caldav`.

## Usage

Create a `hosts.txt` file and add the Facebook profiles for which you would like to scrape events. See `hosts.example.txt` as an example. Lines starting with `#` are ignored.

### Arguments

| Argument           | Default       | Description                                                    |
| ------------------ | ------------- | -------------------------------------------------------------- |
| `--hosts`          | `hosts.txt`   | Path to the file containing Facebook host URLs.                |
| `--from-cache`     | `false`       | Load events from the cache file instead of scraping Facebook.  |
| `--cache-file`     | `events.json` | Path to the JSON cache file.                                   |
| `--write-calendar` | `false`       | Write events to a local iCalendar file.                        |
| `--calendar-file`  | `events.ics`  | Path to the local iCalendar file. Requires `--write-calendar`. |
| `--upload-caldav`  | `false`       | Upload events to the configured CalDAV server.                 |

### Examples

```sh
# By default, the scraper reads `hosts.txt`, scrapes events, and writes the events to only a cache (events.json):
go run .

# Use custom input and output paths:
go run . --hosts private-hosts.txt --cache-file path-to-events.json --write-calendar --calendar-file path-to-events.ics

# Write a local calendar file (this will create an `events.ics` file in the current directory):
go run . --write-calendar

# Write a local calendar file to another path:
go run . --write-calendar --calendar-file path-to-events.ics

# Scrape events and upload them to CalDAV:
go run . --upload-caldav

# Load events from the cache (events.json) instead of scraping:
go run . --from-cache
```
