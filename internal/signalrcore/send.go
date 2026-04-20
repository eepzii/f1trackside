package signalrcore

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

func (c *Client) Send(signalRMsg Message) error {
	msg, err := json.Marshal(signalRMsg)
	if err != nil {
		return err
	}
	msg = append(msg, 0x1e)
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *Client) On() (<-chan []byte, chan error) {
	msg := make(chan []byte, 1)
	errChan := make(chan error)

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

func (c *Client) Stop() error {
	var closeMsg = []byte("{\"type\":7}\u001e")
	if err := c.conn.WriteMessage(websocket.TextMessage, closeMsg); err != nil {
		return err
	}
	return nil
}
