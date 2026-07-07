# calendar

An ICS event parser built specifically for F1 sessions. It fetches the calendar, respects the server's cache duration, parses the raw `VEVENT` entries into clean `Session` structs and sorts the sessions chronologically.

## Quickstart

To initialize the client and safely fetch the calendar:

```go
package main

import (
    "context"
    "log"
    "log/slog"
    "net/http"
    "os"
    "time"

    "github.com/eepzii/f1trackside/internal/calendar"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout,
        &slog.HandlerOptions{
            Level: slog.LevelError,
        },
    ))

    client, err := calendar.NewClient(http.DefaultClient, logger)
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    sessions, err := client.FetchCalendar(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // ...
}
```

Or, if you prefer to use your own calendar endpoint (a great way to ensure availability), you can inject a custom `RoundTripper` like this:

```go
package main

import (
    "context"
    "log"
    "log/slog"
    "net/http"
    "net/url"
    "os"
    "time"

    "github.com/eepzii/f1trackside/internal/calendar"
)

type roundTripper struct {
     localURL string
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    override, err := url.Parse(r.localURL)
     if err != nil {
        return nil, err
    }

    req.URL.Scheme = override.Scheme
    req.URL.Host = override.Host
    req.URL.Path = override.Path

    return http.DefaultTransport.RoundTrip(req)
}

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout,
        &slog.HandlerOptions{
            Level: slog.LevelError,
        },
    ))

    client, err := calendar.NewClient(&http.Client{
        Transport: &roundTripper{
            localURL: os.Getenv("CALENDAR_URL"),
        },
    }, logger)
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    sessions, err := client.FetchCalendar(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // ...
}
```


## API Reference

### Methods

#### `func NewClient(httpClient *http.Client, logger *slog.Logger) (*Client, error)`

Creates a new ICS calendar client configured with a provided HTTP client and logger.

#### `func (c *Client) FetchCalendar(ctx context.Context) ([]Session, error)`

Fetches and parses the F1 ICS calendar. It handles all the caching and mutex synchronization under the hood, guaranteeing safe concurrent access and preventing redundant network calls. 

### Types & Data Structures

#### `type Client struct`

Holds the HTTP client, the cached sessions and the next expiration timestamp of the ICS calendar.

#### `type Session struct`

Represents a single F1 session (like a Race or Qualifying). It contains the clean, parsed start and end times, plus the original `UID` from the raw `VEVENT`.

```go
type Session struct {
    UID       string
    Title     string
    Type      SessionType
    Location  string
    StartTime time.Time
    EndTime   time.Time
}
```

#### `type SessionType int`

An enum representing the different types of F1 sessions.

```go
const (
    Unknown SessionType = iota
    Practice1
    Practice2
    Practice3
    SprintQualification
    Qualifying
    SprintRace
    Race
)
```

## Architecture & Design Philosophy

#### Why Chronological Parsing over Weekend Grouping?

Instead of trying to group sessions into F1 "Weekends" or "Seasons", this package intentionally parses raw `VEVENT` entries into a flat, chronologically sorted list of `Session` structs.

This defensive design choice was made due to the limitations of the upstream ICS data:

- **No Unique IDs**: The calendar lacks specific identifiers tying a session to a specific Grand Prix weekend.

- **Brittle String Matching**: Relying on exact session names makes the backend vulnerable to typos or naming changes (e.g., ecal changing "Sprint Qualification" to "Sprint Qualifying").

- **Unpredictable Time Grouping**: Grouping by time gaps is risky. If the upstream provider suddenly adds "Driver Parades" or "Press Conferences" to the calendar, it breaks the time boundaries of a standard weekend.

#### Frontend vs. Backend Responsibilities:

This design ensures that backend processes always get a reliable and sorted timeline based purely on the `DTSTART` property.

Future UI implementations that require "Weekend" structures will be handled by separate helper functions. Those helpers can safely use strict string-matching to filter out `Unknown` sessions without ever breaking the core backend event loops.