# Facebook Event Scraper

> [!WARNING]
> WIP: This project is in an early stage and a work in progress. I started it as a way to learn the Go language.

A facebook event scraper, that scrapes upcoming events from a list of Facebook hosts, writes them to an iCalendar file, and can upload each event to a CalDAV server.

## Features

- Scrapes upcoming events from Facebook host URLs listed in `hosts.txt`.
- Deduplicates events by Facebook event ID.
- Writes a local `events.ics` calendar file.
- Caches scraped event data in `events.json` for faster local development.
- Uploads individual `.ics` events to a CalDAV server.
- Supports `.env` based configuration for CalDAV credentials.

## Configuration

Create a local `.env` file using `.env.example` as the template:

```env
CALDAV_URL=http://localhost:5232
CALDAV_USERNAME=your-username
CALDAV_PASSWORD=your-password
CALDAV_CALENDAR_PATH=/your-user/your-calendar-id/
```

The `.env` file is ignored by Git so credentials stay local.

## Usage

Create a `hosts.txt` file and add the Facebook profiles for which you would like to scrape events. See `hosts.example.txt` as an example. Then pass this file file with `--hosts`.

Then run it. Without a CalDAV server, it will probably fail at the end, but it will still scrape the events and saves into a `events.ics` and `events.json` file.

```sh
go run . --hosts hosts.txt
```

To upload only and skip writing the local `events.ics` file:

```sh
go run . --hosts hosts.txt --skip-calendar-file
```
