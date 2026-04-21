package signalrcore

import (
	"errors"
	"log/slog"
	"net/url"
	"time"
)

func New(httpURL string, config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	httpParsedURL, err := url.Parse(httpURL)
	if err != nil {
		return nil, err
	}

	// Parse the URL a second time to create a completely independent instance.
	// Copying the url.URL struct directly would result in a shallow copy of its internal pointers.
	wsParsedURL, err := url.Parse(httpURL)
	if err != nil {
		return nil, err
	}

	switch wsParsedURL.Scheme {
	case "http":
		wsParsedURL.Scheme = "ws"
	case "https":
		wsParsedURL.Scheme = "wss"
	default:
		return nil, errors.New("invalid URL protocol")
	}

	if config.Client == nil {
		return nil, errors.New("http client cannot be nil")
	}

	if config.Dialer == nil {
		return nil, errors.New("websocket dialer cannot be nil")
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
