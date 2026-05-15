# Facebook Event Scraper

A Facebook event scraper that scrapes upcoming events from a list of Facebook hosts. It can also write events to an iCalendar file or upload each event to a CalDAV server.

> [!WARNING]
> WIP: This project is in an early stage and a work in progress. I started it as a way to learn the Go language.
>
> **Disclaimer:** Facebook's Terms of Service explicitly prohibit scraping data from their website. Use this project at your own risk.

## Features

- Scrapes upcoming events from Facebook host URLs listed in `hosts.txt`.
- Deduplicates events by Facebook event ID.
- Can write a local `events.ics` calendar file.
- Caches scraped event data in `events.json` for faster local development.
- Can upload individual `.ics` events to a CalDAV server.
- Supports `.env` based configuration for CalDAV credentials.

## Usage

Create a `hosts.txt` file and add the Facebook profiles you want to scrape events from. See `hosts.example.txt` as an example. Empty lines and lines starting with `#` are ignored.

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
# By default, the scraper reads `hosts.txt`, scrapes events, and writes only the cache (`events.json`):
go run .

# Use custom input and output paths:
go run . --hosts private-hosts.txt --cache-file path-to-events.json --write-calendar --calendar-file path-to-events.ics

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

To avoid hitting Facebook too hard, the scraper waits between hosts:

- 2-5 seconds after each host
- 60-120 seconds after every 15 hosts
- 30-60 minutes after a failed host before continuing

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
