package signalrcore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func (c *Client) Close(ctx context.Context) error {
	for {
		current := c.state.Load()

		if current == StateNew || current == StateClosed {
			return fmt.Errorf("connection already closed: %w", errInvalidState)
		}

		if current == StateClosing {
			return fmt.Errorf("connection actively closing: %w", errInvalidState)
		}

		if current == StateConnecting {
			return fmt.Errorf("cannot close while connecting: %w", errInvalidState)
		}

		if c.state.CompareAndSwap(current, StateClosing) {
			break
		}
	}

	defer func() {
		select {
		case <-c.doneChan:
		case <-ctx.Done():
			if c.conn != nil {
				c.conn.Close()
			}
		}
	}()

	msg, err := json.Marshal(
		Message{
			Type: 7,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal closing message: %w", err)
	}

	msg = append(msg, 0x1e)

	if err := c.write(ctx, websocket.TextMessage, msg); err != nil {
		return err
	}

	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	if err := c.write(ctx, websocket.CloseMessage, closeMsg); err != nil {
		return err
	}

	return nil
}
