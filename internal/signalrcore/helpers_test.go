package signalrcore

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func websocketServer(t *testing.T, handler func(conn *websocket.Conn)) (*httptest.Server, string) {
	t.Helper()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverConn, err := upgrader.Upgrade(w, r, http.Header{})
		if err != nil {
			t.Fatalf("could not upgrade: %v", err)
		}
		defer serverConn.Close()

		handler(serverConn)
	}))

	wsURL := strings.Replace(server.URL, "http", "ws", 1)
	return server, wsURL
}
