package signalrcore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
)

func TestInvoke_NotConnected(t *testing.T) {

	tests := []struct {
		name         string
		initialState uint32
	}{
		{
			name:         "rejects state new",
			initialState: StateNew,
		},
		{
			name:         "rejects state connecting",
			initialState: StateConnecting,
		},
		{
			name:         "rejects state closing",
			initialState: StateClosing,
		},
		{
			name:         "rejects state closed",
			initialState: StateClosed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			client := &Client{}
			client.state.Store(test.initialState)

			res, err := client.Invoke(t.Context(), "Test", "test")

			if res != nil {
				t.Errorf("expected nil response, got %q", string(res))
			}

			if !errors.Is(err, errNotConnected) {
				t.Fatalf("expected %v, got %v", errNotConnected, err)
			}
		})
	}
}

func TestInvoke_Marshal(t *testing.T) {
	client := &Client{}
	client.state.Store(StateConnected)

	res, err := client.Invoke(t.Context(), "Test", func() {})

	if res != nil {
		t.Errorf("expected nil response, got %q", string(res))
	}

	var jsonErr *json.UnsupportedTypeError
	if !errors.As(err, &jsonErr) || jsonErr.Type.Kind() != reflect.Func {
		t.Fatalf("expected JSON unsupported type error, got %v", err)
	}
}

func TestInvoke_IncrementID(t *testing.T) {
	client := &Client{
		pendingChan: make(map[string]chan Message),
	}

	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
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

	client.conn = clientConn
	client.state.Store(StateConnected)

	routines := 100
	start := make(chan struct{})
	var wg sync.WaitGroup

	for range routines {
		wg.Go(func() {
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			_, err := client.Invoke(ctx, "Test", "test")

			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}

	close(start)
	wg.Wait()

	if id := client.invocationID.Load(); id != uint64(routines) {
		t.Fatalf("expected next ID to be %d, got %d", routines, id)
	}
}

func TestInvoke_PayloadIntegrity(t *testing.T) {

	tests := []struct {
		name      string
		arguments []any
		wantMsg   Message
	}{
		{
			name:      "no arguments",
			arguments: nil,
			wantMsg: Message{
				Type:         1,
				InvocationID: "1",
				Target:       "Test",
				Arguments:    []byte("[]"),
			},
		},
		{
			name:      "some arguments",
			arguments: []any{"arg1", "arg2"},
			wantMsg: Message{
				Type:         1,
				InvocationID: "1",
				Target:       "Test",
				Arguments:    []byte(`["arg1","arg2"]`),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			type result struct {
				hasSuffix bool
				diff      string
			}
			resultChan := make(chan result, 1)

			client := &Client{
				pendingChan: make(map[string]chan Message),
			}

			server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
				_, readMsg, err := conn.ReadMessage()
				if err != nil {
					return
				}

				hasSuffix := bytes.HasSuffix(readMsg, []byte{0x1e})

				cleanMsg := bytes.TrimSuffix(readMsg, []byte{0x1e})
				var msg Message
				if err := json.Unmarshal(cleanMsg, &msg); err != nil {
					t.Errorf("unable to unmarshal message %v", err)
					return
				}

				resultChan <- result{
					hasSuffix: hasSuffix,
					diff:      cmp.Diff(test.wantMsg, msg),
				}

				client.pendingMu.Lock()
				ch, ok := client.pendingChan[msg.InvocationID]
				client.pendingMu.Unlock()
				if !ok {
					t.Errorf("channel not ok")
					return
				}

				if ch != nil {
					ch <- Message{
						Type:         3,
						InvocationID: msg.InvocationID,
						Result:       []byte("test"),
					}
				}
			})
			defer server.Close()

			clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("client failed to dial websocket: %v", err)
			}
			defer clientConn.Close()

			client.conn = clientConn
			client.state.Store(StateConnected)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			_, err = client.Invoke(ctx, "Test", test.arguments...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			select {
			case res := <-resultChan:
				if !res.hasSuffix {
					t.Errorf(`separator "0x1e" not found at the end of the payload`)
				}
				if res.diff != "" {
					t.Errorf("payload mismatch (-want +got):\n%s", res.diff)
				}
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for server to validate payload")
			}
		})
	}
}

func TestInvoke_Write(t *testing.T) {
	client := &Client{
		pendingChan: make(map[string]chan Message),
	}

	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}
	defer clientConn.Close()

	client.conn = clientConn
	client.state.Store(StateConnected)

	clientConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Invoke(ctx, "Test", "test")

	if res != nil {
		t.Errorf("expected nil response, got %q", string(res))
	}

	if !errors.Is(err, net.ErrClosed) {
		t.Fatalf("expected %v, got %v", net.ErrClosed, err)
	}
}

func TestInvoke_Select(t *testing.T) {

	tests := []struct {
		name     string
		want     json.RawMessage
		task     func(ch chan Message, req Message)
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "good invoke",
			want: []byte("test"),
			task: func(ch chan Message, req Message) {
				ch <- Message{
					Type:         3,
					InvocationID: req.InvocationID,
					Result:       []byte("test"),
				}
			},
			checkErr: nil,
		},
		{
			name: "context exceeded",
			want: nil,
			task: func(ch chan Message, req Message) {},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected %v, got %v", context.DeadlineExceeded, err)
				}
			},
		},
		{
			name: "closed channel",
			want: nil,
			task: func(ch chan Message, req Message) {
				close(ch)
			},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, errChannelUnavailable) {
					t.Fatalf("expected %v, got %v", errChannelUnavailable, err)
				}
			},
		},
		{
			name: "message error",
			want: nil,
			task: func(ch chan Message, req Message) {
				ch <- Message{
					Error: "error",
				}
			},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, errInvocation) {
					t.Fatalf("expected %v, got %v", errInvocation, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{
				pendingChan: make(map[string]chan Message),
			}

			server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
				_, readMsg, err := conn.ReadMessage()
				if err != nil {
					return
				}

				for sub := range bytes.SplitSeq(readMsg, []byte{0x1e}) {
					if len(sub) <= 0 {
						continue
					}

					var msg Message
					if err := json.Unmarshal(sub, &msg); err != nil {
						t.Errorf("unable to unmarshal message %v", err)
						return
					}

					client.pendingMu.Lock()
					ch, ok := client.pendingChan[msg.InvocationID]
					client.pendingMu.Unlock()
					if !ok {
						t.Errorf("channel not ok")
						return
					}

					test.task(ch, msg)
				}
			})
			defer server.Close()

			clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("client failed to dial websocket: %v", err)
			}
			defer clientConn.Close()

			client.conn = clientConn
			client.state.Store(StateConnected)

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			res, err := client.Invoke(ctx, "Test", "test")

			if test.want != nil && !bytes.Equal(res, test.want) {
				t.Errorf("expected %q response, got %q", string(test.want), string(res))
			}

			if len(client.pendingChan) != 0 {
				t.Errorf("map entry of pending channel not cleared")
			}

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
