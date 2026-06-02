package signalrcore

import (
	"context"
	"fmt"
	"net/http"
)

// Start establishes the underlying WebSocket connection to the SignalR server.
//
// The context governs the timeout and cancellation of the initial setup requests.
func (c *Client) Start(ctx context.Context) error {
	if !c.state.CompareAndSwap(StateNew, StateConnecting) {
		return fmt.Errorf("client must be in a new state to start: %w", errInvalidState)
	}

	var success bool
	defer func() {
		if !success {
			c.state.Store(StateNew)

			if c.conn != nil {
				c.conn.Close()
			}
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
		return fmt.Errorf("failed to dial to %s: %w", wsURL.String(), err)
	}

	err = c.handshake(ctx)
	if err != nil {
		return err
	}

	success = true
	c.state.Store(StateConnected)

	go c.listen()

	c.logger.Info("connected to signalR server", "url", wsURL.Host)

	return nil
}
