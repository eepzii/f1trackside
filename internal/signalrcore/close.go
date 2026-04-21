package signalrcore

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) Close(ctx context.Context) error {
	_, ok := ctx.Deadline()
	if !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	for {
		current := c.state.Load()

		if current == StateNew || current == StateClosed {
			return errors.New("connection already closed")
		}

		if current == StateClosing {
			return errors.New("connection actively closing")
		}

		if current == StateConnecting {
			return errors.New("cannot close while connecting")
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
		return err
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
