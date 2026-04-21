package signalrcore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func (c *Client) Start(ctx context.Context) error {
	if !c.state.CompareAndSwap(StateNew, StateConnecting) {
		return errors.New("connection already in progress or closed")
	}

	var success bool
	defer func() {
		if !success {
			c.state.Store(StateNew)
		}
	}()

	res, err := c.negotiate(ctx, *c.baseURL)
	if err != nil {
		return err
	}

	wsURL := *c.websocketURL
	params := wsURL.Query()

	params.Set("id", res.body.ConnectionID)
	params.Set("transport", "webSockets")
	wsURL.RawQuery = params.Encode()

	headers := make(http.Header)
	var cookieStr string
	for i, cookie := range res.cookies {
		if i > 0 {
			cookieStr += "; "
		}
		cookieStr += fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
	}

	if cookieStr != "" {
		headers.Set("Cookie", cookieStr)
	}
	if c.token != "" {
		headers.Set("Authorization", "Bearer "+c.token)
	}

	c.conn, _, err = c.dialer.DialContext(ctx, wsURL.String(), headers)
	if err != nil {
		return err
	}

	err = c.handshake(ctx)
	if err != nil {
		return err
	}

	go c.listen()

	success = true
	c.state.Store(StateConnected)

	c.logger.Info("connected to signalR server", "url", wsURL.Host)

	return nil
}
