package calendar

import (
	"fmt"
	"log/slog"
	"net/http"
)

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
