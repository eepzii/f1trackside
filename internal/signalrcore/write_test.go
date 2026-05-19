package signalrcore

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWrite_Table(t *testing.T) {

	tests := []struct {
		name     string
		data     []byte
		setupCtx func() (context.Context, context.CancelFunc)
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "good write",
			data: append([]byte("hello world"), 0x1e),
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
		},
		{
			name: "canceled context",
			data: append([]byte("hello world"), 0x1e),
			setupCtx: func() (context.Context, context.CancelFunc) {
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
			name: "context reached deadline",
			data: make([]byte, 10*1024*1024),
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Millisecond)
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
				conn.ReadMessage()
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

			ctx, cancel := test.setupCtx()
			defer cancel()

			err = client.write(ctx, websocket.TextMessage, test.data)

			if test.checkErr != nil {
				test.checkErr(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error, got %v", err)
			}
		})
	}

}

func TestWrite_Concurrency(t *testing.T) {
	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
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
	data := append([]byte(`{}`), 0x1e)

	var wg sync.WaitGroup
	routines := 1000
	errs := make(chan error, routines)

	start := make(chan struct{})

	for range routines {
		wg.Go(func() {
			<-start

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := client.write(ctx, websocket.TextMessage, data); err != nil {
				errs <- err
			}
		})
	}

	close(start)

	wg.Wait()
	close(errs)

	if len(errs) > 0 {
		totalErrs := len(errs)
		firstErr := <-errs
		t.Fatalf("%d/%d routines failed. first error: %v", totalErrs, routines, firstErr)
	}
}

func TestWrite_DeadlineReset(t *testing.T) {
	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
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
	data := append([]byte(`{}`), 0x1e)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = client.write(ctx, websocket.TextMessage, data)
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	err = client.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		t.Fatalf("second write failed. deadline was not reset: %v", err)
	}
}
