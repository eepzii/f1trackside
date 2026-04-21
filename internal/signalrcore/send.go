package signalrcore

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
)

func (c *Client) Send(signalRMsg Message) error {
	if c.state.Load() != StateConnected {
		return errors.New("client not connected")
	}

	msg, err := json.Marshal(signalRMsg)
	if err != nil {
		return err
	}
	msg = append(msg, 0x1e)
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *Client) On() (<-chan []byte, chan error) {
	msg := make(chan []byte, 1)
	errChan := make(chan error, 1)

	if c.state.Load() != StateConnected {
		errChan <- errors.New("client not connected")
		return msg, errChan
	}

	go func() {
		for {
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				errChan <- err
				msg <- data
				return
			}
			msg <- data
		}
	}()

	return msg, errChan
}
