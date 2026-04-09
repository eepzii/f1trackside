package signalrcore

import (
	"errors"
	"net/url"
)

func New(negotiationUrl string, config *Config) (*Client, error) {
	var c = &Client{}
	var err error

	c.negotiateUrl, err = url.Parse(negotiationUrl)
	if err != nil {
		return nil, err
	}
	c.negotiateUrl.JoinPath("negotiate")

	c.websocketUrl, err = url.Parse(negotiationUrl)
	if err != nil {
		return nil, err
	}

	switch c.websocketUrl.Scheme {
	case "http":
		c.websocketUrl.Scheme = "ws"
	case "https":
		c.websocketUrl.Scheme = "wss"
	default:
		return nil, errors.New("invalid url protocol")
	}

	c.config = config

	return c, nil
}
