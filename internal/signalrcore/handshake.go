package signalrcore

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) handshake(ctx context.Context) error {

	handshakeMessage := append([]byte(`{"protocol":"json","version":1}`), 0x1e)
	if err := c.write(ctx, websocket.TextMessage, handshakeMessage); err != nil {
		return err
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}
		defer c.conn.SetReadDeadline(time.Time{})
	}

	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read handshake message: %w", err)
	}

	successMsg := append([]byte("{}"), 0x1e)
	if !bytes.Equal(msg, successMsg) {
		return fmt.Errorf("unexpected response %q: %w", msg, errHandshake)
	}

	return nil
}
