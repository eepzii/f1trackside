package calendar

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	mu         sync.RWMutex
	expiresAt  time.Time
	sessions   []Session
	httpClient *http.Client
	logger     *slog.Logger
}

type Session struct {
	UID       string
	Title     string
	Type      SessionType
	Location  string
	StartTime time.Time
	EndTime   time.Time
}
