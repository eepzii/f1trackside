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

	params := c.websocketURL.Query()
	params.Set("id", res.body.ConnectionID)
	params.Set("transport", "webSockets")
	c.websocketURL.RawQuery = params.Encode()

	headers := make(http.Header)
	for i, cookie := range res.cookies {
		if i == 0 {
			headers.Set("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
			continue
		}
		headers.Add("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	headers.Set("Authorization", "Bearer "+c.token)

	c.conn, _, err = c.dialer.Dial(c.websocketURL.String(), headers)
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

	return nil
}
