package signalrcore

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func (c *Client) Start() error {
	res, err := negotiate(*c.baseURL)
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

	handshakeMessage := []byte(`{"protocol":"json","version":1}`)
	handshakeMessage = append(handshakeMessage, 0x1e)
	if err := c.conn.WriteMessage(websocket.TextMessage, handshakeMessage); err != nil {
		return err
	}

	return nil
}
