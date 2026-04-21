package signalrcore

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) handshake(ctx context.Context) error {

	handshakeMessage := []byte(`{"protocol":"json","version":1}`)
	handshakeMessage = append(handshakeMessage, 0x1e)

	if err := c.write(ctx, websocket.TextMessage, handshakeMessage); err != nil {
		return err
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}

	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return err
	}

	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}

	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	successMsg := []byte("{}")
	successMsg = append(successMsg, 0x1e)

	if !bytes.Equal(msg, successMsg) {
		return fmt.Errorf("handshake failed: unexpected response %q", msg)
	}

	return nil
}
