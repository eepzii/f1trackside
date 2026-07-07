package calendar

import (
	"fmt"
	"log/slog"
	"net/http"
)

// NewClient initializes a Client to fetch and parse the F1 ICS calendar.
//
// It requires a valid HTTP client, but will safely fall back to a slog.New(slog.DiscardHandler)
// if the provided logger is nil.
func NewClient(httpClient *http.Client, logger *slog.Logger) (*Client, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("http client cannot be nil: %w", errInvalidInput)
	}

	clientLogger := slog.New(slog.DiscardHandler)
	if logger != nil {
		clientLogger = logger
	}

	return &Client{
		httpClient: httpClient,
		logger:     clientLogger,
	}, nil
}
