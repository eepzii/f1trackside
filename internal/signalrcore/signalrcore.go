package signalrcore

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"
)

// New initializes and returns a new SignalR Core client using the provided endpoint and config.
//
// The provided config must not be nil and must include both an HTTP client and a WebSocket dialer.
// If omitted, default values are applied for the logger and idle timeout.
func New(httpURL string, config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil: %w", errInvalidInput)
	}

	httpParsedURL, err := url.Parse(httpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	// Parse the URL a second time to create a completely independent instance.
	// Copying the url.URL struct directly would result in a shallow copy of its internal pointers.
	wsParsedURL, err := url.Parse(httpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	switch wsParsedURL.Scheme {
	case "http":
		wsParsedURL.Scheme = "ws"
	case "https":
		wsParsedURL.Scheme = "wss"
	default:
		return nil, fmt.Errorf("invalid URL protocol: %w", errInvalidInput)
	}

	if config.Client == nil {
		return nil, fmt.Errorf("http client cannot be nil: %w", errInvalidConfig)
	}

	if config.Dialer == nil {
		return nil, fmt.Errorf("websocket dialer cannot be nil: %w", errInvalidConfig)
	}

	var logger *slog.Logger
	if config.Logger != nil {
		logger = config.Logger
	} else {
		logger = slog.New(slog.DiscardHandler)
	}

	var idleTimeout time.Duration
	if config.IdleTimeout != 0 {
		idleTimeout = config.IdleTimeout
	} else {
		idleTimeout = 45 * time.Second
	}

	client := &Client{
		baseURL:      httpParsedURL,
		websocketURL: wsParsedURL,

		eventChan:   make(map[string]chan Event),
		pendingChan: make(map[string]chan Message),
		doneChan:    make(chan struct{}),

		client: config.Client,
		dialer: config.Dialer,
		logger: logger,

		idleTimeout: idleTimeout,
		token:       config.Token,
	}

	return client, nil
}
