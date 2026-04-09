package signalrcore

import (
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Client struct {
	negotiateUrl *url.URL
	websocketUrl *url.URL

	config *Config
	conn   *websocket.Conn
}

type SignalRMessage struct {
	Type         int    `json:"type"`
	Target       string `json:"target"`
	InvocationId string `json:"invocationId"`
	Arguments    []any  `json:"arguments"`
}

type Config struct {
	Token  string
	Dialer *websocket.Dialer
}

type negotiationRes struct {
	body struct {
		ConnectionID        string `json:"connectionId"`
		AvailableTransports []struct {
			Transport string   `json:"transport"`
			Formats   []string `json:"transferFormats"`
		} `json:"availableTransports"`
		NegotiateVersion int `json:"negotiateVersion"`
	}
	cookies []*http.Cookie
}
