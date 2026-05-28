package signalrcore

import (
	"bytes"
	"errors"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestListen_Defer(t *testing.T) {
	t.Run("cleanup executes before done channel closure", func(t *testing.T) {
		server, wsURL := websocketServer(t, func(conn *websocket.Conn) {})
		defer server.Close()

		clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("client failed to dial websocket: %v", err)
		}
		defer clientConn.Close()

		client := &Client{
			conn:        clientConn,
			eventChan:   make(map[string]chan Event),
			pendingChan: make(map[string]chan Message),
			doneChan:    make(chan struct{}),
			logger:      slog.New(slog.DiscardHandler),
		}

		client.state.Store(StateClosing)
		ch := make(chan Message, 1)
		client.pendingChan["id1"] = ch

		go client.listen()

		select {
		case <-client.doneChan:
			t.Fatal("done channel closed before .cleanUp()")
		case <-ch:
		case <-time.NewTimer(500 * time.Millisecond).C:
			t.Fatalf("timed out")
		}
	})

	t.Run("clean up function called", func(t *testing.T) {
		server, wsURL := websocketServer(t, func(conn *websocket.Conn) {})
		defer server.Close()

		clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("client failed to dial websocket: %v", err)
		}
		defer clientConn.Close()

		client := &Client{
			conn:        clientConn,
			eventChan:   make(map[string]chan Event),
			pendingChan: make(map[string]chan Message),
			doneChan:    make(chan struct{}),
			logger:      slog.New(slog.DiscardHandler),
		}

		client.state.Store(StateClosing)
		client.eventChan["target1"] = make(chan Event, 1)
		client.pendingChan["id1"] = make(chan Message, 1)

		client.listen()

		if len(client.eventChan) != 0 || len(client.pendingChan) != 0 {
			t.Fatalf("maps were not cleaned up; eventChan: %d, pendingChan: %d",
				len(client.eventChan), len(client.pendingChan))
		}
	})

	t.Run("done channel", func(t *testing.T) {
		server, wsURL := websocketServer(t, func(conn *websocket.Conn) {})
		defer server.Close()

		clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("client failed to dial websocket: %v", err)
		}
		defer clientConn.Close()

		client := &Client{
			conn:        clientConn,
			eventChan:   make(map[string]chan Event),
			pendingChan: make(map[string]chan Message),
			doneChan:    make(chan struct{}),
			logger:      slog.New(slog.DiscardHandler),
		}

		client.listen()

		select {
		case <-client.doneChan:
		case <-time.NewTimer(500 * time.Millisecond).C:
			t.Fatalf("channel not closed")
		}
	})
}

func TestListen_ExitConditions(t *testing.T) {

	tests := []struct {
		name     string
		task     func(t *testing.T, conn *websocket.Conn)
		client   func(conn *websocket.Conn) *Client
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "set read deadline",
			task: func(t *testing.T, conn *websocket.Conn) {},
			client: func(conn *websocket.Conn) *Client {
				conn.Close()
				return &Client{
					conn:     conn,
					doneChan: make(chan struct{}),
				}
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", net.ErrClosed)
				}

				if !errors.Is(err, net.ErrClosed) {
					t.Fatalf("expected %v, got %v", net.ErrClosed, err)
				}
			},
		},
		{
			name: "read websocket message",
			task: func(t *testing.T, conn *websocket.Conn) {
				if err := conn.WriteMessage(websocket.TextMessage, append([]byte(`{"type":7}`), 0x1e)); err != nil {
					t.Errorf("server failed to write: %v", err)
					return
				}
				closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
				if err := conn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
					t.Errorf("server failed to write close message: %v", err)
					return
				}
			},
			client: func(conn *websocket.Conn) *Client {
				return &Client{
					conn:        conn,
					doneChan:    make(chan struct{}),
					idleTimeout: 3 * time.Second,
					logger:      slog.New(slog.DiscardHandler),
				}
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatal("expected normal closure (1000), got nil")
				}

				var netErr *websocket.CloseError
				if !errors.As(err, &netErr) || netErr.Code != 1000 {
					t.Fatalf("expected normal closure (1000), got %v", err)
				}
			},
		},
		{
			name: "read deadline exceeded",
			task: func(t *testing.T, conn *websocket.Conn) {
				time.Sleep(500 * time.Millisecond)
			},
			client: func(conn *websocket.Conn) *Client {
				return &Client{
					conn:        conn,
					doneChan:    make(chan struct{}),
					idleTimeout: 100 * time.Millisecond,
					logger:      slog.New(slog.DiscardHandler),
				}
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected network timeout error (i/o timeout), got nil")
				}

				var netErr net.Error
				if !(errors.As(err, &netErr) && netErr.Timeout()) {
					t.Fatalf("expected network timeout error (i/o timeout), got %v", err)
				}
			},
		},
		{
			name: "signalr message",
			task: func(t *testing.T, conn *websocket.Conn) {
				if err := conn.WriteMessage(websocket.TextMessage, append([]byte(`{"type":7,"error":"Internal server error"}`), 0x1e)); err != nil {
					t.Errorf("server failed to write: %v", err)
					return
				}
				conn.ReadMessage()
			},
			client: func(conn *websocket.Conn) *Client {
				return &Client{
					conn:        conn,
					doneChan:    make(chan struct{}),
					idleTimeout: 3 * time.Second,
					logger:      slog.New(slog.DiscardHandler),
				}
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errServerClosed)
				}

				if !errors.Is(err, errServerClosed) {
					t.Fatalf("expected %v, got %v", errServerClosed, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
				if test.task != nil {
					test.task(t, conn)
				}
			})
			defer server.Close()

			clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("client failed to dial websocket: %v", err)
			}
			defer clientConn.Close()

			if test.client == nil {
				t.Fatalf("cannot run test with nil client")
			}

			client := test.client(clientConn)

			var wg sync.WaitGroup
			routines := 100
			start := make(chan struct{})

			for range routines {
				wg.Go(func() {
					<-start
					client.errorMu.Lock()
					_ = client.err
					client.errorMu.Unlock()
				})
			}

			close(start)
			client.listen()
			wg.Wait()

			client.errorMu.Lock()
			clientErr := client.err
			client.errorMu.Unlock()

			if test.checkErr != nil {
				test.checkErr(t, clientErr)
				return
			}

			if clientErr != nil {
				t.Fatalf("unexpected error: %v", clientErr)
			}
		})
	}

}

func TestListen_HandlesInvocation(t *testing.T) {
	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		err := conn.WriteMessage(websocket.TextMessage, append([]byte(`{"type":1,"target":"target1","arguments":["target1Arg"]}`), 0x1e))
		if err != nil {
			t.Errorf("server failed to write: %v", err)
			return
		}
	})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}
	defer clientConn.Close()

	client := &Client{
		conn:        clientConn,
		eventChan:   make(map[string]chan Event),
		doneChan:    make(chan struct{}),
		idleTimeout: 3 * time.Second,
		logger:      slog.New(slog.DiscardHandler),
	}

	testChan := make(chan Event, 1)
	target := "target1"
	client.eventChan[target] = testChan

	go client.listen()

	select {
	case msg := <-testChan:
		if !bytes.Equal(msg.Data, []byte(`["target1Arg"]`)) {
			t.Fatalf("expected message on %s, got %s", target, string(msg.Data))
		}
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timed out waiting for invocation")
	}
}

func TestListen_HandlesCompletion(t *testing.T) {
	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		err := conn.WriteMessage(websocket.TextMessage, append([]byte(`{"type":3,"invocationId":"id1","result":["id1Result"]}`), 0x1e))
		if err != nil {
			t.Errorf("server failed to write: %v", err)
			return
		}
	})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}
	defer clientConn.Close()

	client := &Client{
		conn:        clientConn,
		pendingChan: make(map[string]chan Message),
		doneChan:    make(chan struct{}),
		idleTimeout: 3 * time.Second,
		logger:      slog.New(slog.DiscardHandler),
	}

	testChan := make(chan Message, 1)
	id := "id1"
	client.pendingChan[id] = testChan

	go client.listen()

	select {
	case msg := <-testChan:
		if !bytes.Equal(msg.Result, []byte(`["id1Result"]`)) {
			t.Fatalf("expected message on %s, got %s", id, string(msg.Result))
		}
	case <-time.NewTimer(time.Second).C:
		t.Fatal("timed out waiting for completion")
	}
}

func TestListen_Resilience(t *testing.T) {
	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		payloads := []string{
			`{malformed JSON}`,
			`{"type":1,"target":"target1"}`,
			`{"type":3,"invocationId":"id1"}`,
			`{"type":6}`,
			`{"type":99}`,
			`{"type":7}`,
		}

		for _, payload := range payloads {
			msg := append([]byte(payload), 0x1e)
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				t.Errorf("server failed to write payload %q: %v", payload, err)
				return
			}
		}

		closeMsg := websocket.FormatCloseMessage(1000, "")
		if err := conn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			t.Errorf("server failed to write close message: %v", err)
			return
		}
	})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}
	defer clientConn.Close()

	client := &Client{
		conn:        clientConn,
		doneChan:    make(chan struct{}),
		idleTimeout: 3 * time.Second,
		logger:      slog.New(slog.DiscardHandler),
	}

	client.listen()

	client.errorMu.Lock()
	clientErr := client.err
	client.errorMu.Unlock()

	var closeErr *websocket.CloseError
	if !errors.As(clientErr, &closeErr) || closeErr.Code != 1000 {
		t.Fatalf("expected normal closure (1000), got %v", clientErr)
	}
}
