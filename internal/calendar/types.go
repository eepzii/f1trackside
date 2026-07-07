package calendar

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Client holds the external dependencies and the internal cache state.
type Client struct {
	mu         sync.RWMutex
	expiresAt  time.Time
	sessions   []Session
	httpClient *http.Client
	logger     *slog.Logger
}

// Session represents a parsed F1 session.
//
// It contains the core metadata of the VEVENT from the original ICS calendar.
type Session struct {
	UID       string
	Title     string
	Type      SessionType
	Location  string
	StartTime time.Time
	EndTime   time.Time
}
