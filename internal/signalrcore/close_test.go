package signalrcore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
)

func TestClose_Done(t *testing.T) {
	client := &Client{
		doneChan: make(chan struct{}),
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

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Close(ctx)
	}()

	select {
	case <-errChan:
		t.Fatalf("defer block in .Close() is missing or not blocking")
	case <-time.NewTimer(20 * time.Millisecond).C:
	}

	close(client.doneChan)

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := client.conn.Close(); err != nil {
			t.Fatalf("expected websocket connection not to be closed by .Close()")
		}
	case <-time.NewTimer(100 * time.Millisecond).C:
		t.Fatalf("timed out waiting for done channel closure")
	}

	if finalState := client.state.Load(); finalState != StateClosing {
		t.Fatalf("expected state %d, got %d", StateClosing, finalState)
	}
}

func TestClose_Context(t *testing.T) {
	client := &Client{
		doneChan: make(chan struct{}),
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

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Close(ctx)
	}()

	select {
	case <-errChan:
		t.Fatalf("defer block in .Close() is missing or not blocking")
	case <-time.NewTimer(20 * time.Millisecond).C:
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := client.conn.Close(); err == nil {
			t.Fatalf("expected websocket connection to be closed by .Close()")
		}
	case <-time.NewTimer(100 * time.Millisecond).C:
		t.Fatalf("timed out waiting for context closure")
	}

	if finalState := client.state.Load(); finalState != StateClosing {
		t.Fatalf("expected state %d, got %d", StateClosing, finalState)
	}
}

func TestClose_WriteError(t *testing.T) {
	client := &Client{
		doneChan: make(chan struct{}),
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

	client.conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Close(ctx)
	}()

	select {
	case err := <-errChan:
		if !errors.Is(err, net.ErrClosed) {
			t.Fatalf("expected %v, got %v", net.ErrClosed, err)
		}
	case <-time.NewTimer(100 * time.Millisecond).C:
		t.Fatalf("timed out waiting for write error")
	}
}

func TestClose_InvalidStates(t *testing.T) {

	tests := []struct {
		name         string
		invalidState uint32
	}{
		{
			name:         "rejects state new",
			invalidState: StateNew,
		},
		{
			name:         "rejects state connecting",
			invalidState: StateConnecting,
		},
		{
			name:         "rejects state closing",
			invalidState: StateClosing,
		},
		{
			name:         "rejects state closed",
			invalidState: StateClosed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{}
			client.state.Store(test.invalidState)

			errChan := make(chan error, 1)
			go func() {
				errChan <- client.Close(context.Background())
			}()

			select {
			case err := <-errChan:
				if !errors.Is(err, errInvalidState) {
					t.Errorf("expected %v, got %v", errInvalidState, err)
				}
			case <-time.NewTimer(100 * time.Millisecond).C:
				t.Fatal("timed out waiting to be rejected by invalid state")
			}
		})
	}
}

func TestClose_PayloadIntegrity(t *testing.T) {
	wantMsg := Message{
		Type: 7,
	}

	type result struct {
		hasSuffix bool
		diff      string
	}
	resultChan := make(chan result, 1)

	client := &Client{
		doneChan: make(chan struct{}),
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
			diff:      cmp.Diff(wantMsg, msg),
		}

		close(client.doneChan)
	})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}
	defer clientConn.Close()

	client.conn = clientConn
	client.state.Store(StateConnected)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Close(ctx)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.NewTimer(50 * time.Millisecond).C:
		t.Fatalf("timed out during close execution")
	}

	select {
	case res := <-resultChan:
		if !res.hasSuffix {
			t.Errorf(`separator "0x1e" not found at the end of the payload`)
		}
		if res.diff != "" {
			t.Errorf("payload mismatch (-want +got):\n%s", res.diff)
		}
	case <-time.NewTimer(200 * time.Millisecond).C:
		t.Fatal("timed out waiting for server to validate payload")
	}
}

func TestClose_Concurrency(t *testing.T) {
	client := &Client{
		doneChan: make(chan struct{}),
	}

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

	client.conn = clientConn
	client.state.Store(StateConnected)

	routines := 20
	start := make(chan struct{})
	var wg sync.WaitGroup

	var successCount atomic.Int32
	var stateErrCount atomic.Int32

	for range routines {
		wg.Go(func() {
			<-start

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			err := client.Close(ctx)

			if err == nil {
				successCount.Add(1)
			} else if errors.Is(err, errInvalidState) {
				stateErrCount.Add(1)
			} else {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}

	close(start)
	wg.Wait()

	if success := successCount.Load(); success != 1 {
		t.Errorf("expected exactly 1 successful .Close() execution, got %v", success)
	}

	if stateErrs := stateErrCount.Load(); stateErrs != int32(routines-1) {
		t.Errorf("expected exactly %d state errors, got %d", routines-1, stateErrs)
	}

}
