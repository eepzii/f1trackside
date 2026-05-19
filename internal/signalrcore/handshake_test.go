package signalrcore

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHandshake_Table(t *testing.T) {

	tests := []struct {
		name                 string
		handshakeRes         []byte
		closeServer          bool
		ignoreReadMessageErr bool
		ctxSetup             func() (context.Context, context.CancelFunc)
		checkErr             func(t *testing.T, err error)
	}{
		{
			name:         "good handshake",
			handshakeRes: append([]byte(`{}`), 0x1e),
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
		},
		{
			name:         "canceled context",
			handshakeRes: append([]byte(`{}`), 0x1e),
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", context.Canceled)
				}

				if !errors.Is(err, context.Canceled) {
					t.Fatalf("expected %v, got %v", context.Canceled, err)
				}
			},
		},
		{
			name: "abnormal closure (1006)",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Millisecond)
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected abnormal closure (1006), got nil")
				}

				var closeErr *websocket.CloseError
				if !errors.As(err, &closeErr) || closeErr.Code != 1006 {
					t.Fatalf("expected abnormal closure (1006), got: %v", err)
				}
			},
		},
		{
			name:         "server returns error message",
			handshakeRes: append([]byte(`{"error":"unsupported protocol"}`), 0x1e),
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Second)
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errHandshake)
				}

				if !errors.Is(err, errHandshake) {
					t.Fatalf("expected %v, got %v", errHandshake, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
				if test.closeServer {
					return
				}

				if _, _, err := conn.ReadMessage(); err != nil {
					if test.ignoreReadMessageErr {
						t.Errorf("server failed to read handshake from client: %v", err)
					}
					return
				}

				if len(test.handshakeRes) > 0 {
					if err := conn.WriteMessage(websocket.TextMessage, test.handshakeRes); err != nil {
						return
					}
				}
			})
			defer server.Close()

			clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("client failed to dial websocket: %v", err)
			}
			defer clientConn.Close()

			client := &Client{
				conn: clientConn,
			}

			ctx, cancel := test.ctxSetup()
			defer cancel()

			err = client.handshake(ctx)

			if test.checkErr != nil {
				test.checkErr(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestHandshake_DeadlineReset(t *testing.T) {
	handshakeRes := append([]byte(`{}`), 0x1e)

	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, handshakeRes); err != nil {
				return
			}

			time.Sleep(30 * time.Millisecond)
		}
	})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}
	defer clientConn.Close()

	client := &Client{
		conn: clientConn,
	}
	data := append([]byte(`{"protocol":"json","version":1}`), 0x1e)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = client.handshake(ctx)
	if err != nil {
		t.Fatalf("first handshake failed: %v", err)
	}

	err = client.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		t.Fatalf("could not write second handshake message: %v", err)
	}

	_, _, err = client.conn.ReadMessage()
	if err != nil {
		t.Fatalf("second handshake failed. deadline was not reset: %v", err)
	}
}

func TestHandshake_HasDeadline(t *testing.T) {
	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}

		time.Sleep(3 * time.Second)
	})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}
	defer clientConn.Close()

	client := &Client{
		conn: clientConn,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.handshake(ctx)
	}()

	select {
	case err := <-errCh:
		var netErr net.Error
		if !(errors.As(err, &netErr) && netErr.Timeout()) {
			t.Fatalf("expected network timeout error (i/o timeout), got %v", err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("deadline was not applied to the connection")
	}
}
