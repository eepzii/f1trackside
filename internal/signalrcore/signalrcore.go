package signalrcore

import (
	"errors"
	"net/url"
)

func New(negotiationUrl string, config *Config) (*Client, error) {
	var c = &Client{}
	var err error

	c.baseURL, err = url.Parse(negotiationUrl)
	if err != nil {
		return nil, err
	}
	c.baseURL.JoinPath("negotiate")

	c.websocketURL, err = url.Parse(negotiationUrl)
	if err != nil {
		return nil, err
	}

	switch c.websocketURL.Scheme {
	case "http":
		c.websocketURL.Scheme = "ws"
	case "https":
		c.websocketURL.Scheme = "wss"
	default:
		return nil, errors.New("invalid url protocol")
	}

	c.token = config.Token

	if config.Dialer == nil {
		return nil, errors.New("dialer cannot be nil")
	}
	c.dialer = config.Dialer

	return c, nil
}
