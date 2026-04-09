package signalrcore

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func (c *Client) Start() error {
	res, err := negotiate(*c.negotiateUrl)
	if err != nil {
		return err
	}

	params := c.websocketUrl.Query()
	params.Set("id", res.body.ConnectionID)
	params.Set("transport", "webSockets")
	c.websocketUrl.RawQuery = params.Encode()

	headers := make(http.Header)
	for i, cookie := range res.cookies {
		if i == 0 {
			headers.Set("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
			continue
		}
		headers.Add("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	headers.Set("Authorization", "Bearer "+c.config.Token)

	c.conn, _, err = c.config.Dialer.Dial(c.websocketUrl.String(), headers)
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
