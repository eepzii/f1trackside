package signalrcore

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

func TestCleanUp_State(t *testing.T) {
	tests := []struct {
		name          string
		initialState  uint32
		expectedState uint32
	}{
		{
			name:          "state new",
			initialState:  StateNew,
			expectedState: StateNew,
		},
		{
			name:          "state connecting",
			initialState:  StateConnecting,
			expectedState: StateConnecting,
		},
		{
			name:          "state connected",
			initialState:  StateConnected,
			expectedState: StateClosed,
		},
		{
			name:          "state closing",
			initialState:  StateClosing,
			expectedState: StateClosed,
		},
		{
			name:          "state closed",
			initialState:  StateClosed,
			expectedState: StateClosed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{
				eventChan:   make(map[string]chan Event),
				pendingChan: make(map[string]chan Message),
				logger:      slog.New(slog.DiscardHandler),
			}
			client.state.Store(test.initialState)

			client.cleanUp()

			if got := client.state.Load(); got != test.expectedState {
				t.Errorf("expected state %d, got %d", test.expectedState, got)
			}
		})
	}
}

func TestCleanUp_Connection(t *testing.T) {
	server, wsURL := websocketServer(t, func(conn *websocket.Conn) {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	})
	defer server.Close()

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client failed to dial websocket: %v", err)
	}

	client := &Client{
		conn:   clientConn,
		logger: slog.New(slog.DiscardHandler),
	}
	client.state.Store(StateConnected)

	client.cleanUp()

	err = client.conn.Close()
	if err == nil {
		t.Fatal("expected close to fail, but it succeeded")
	}

	if !errors.Is(err, net.ErrClosed) {
		t.Errorf("expected net.ErrClosed, got %v", err)
	}
}

func TestCleanUp_Maps(t *testing.T) {
	client := &Client{
		eventChan:   make(map[string]chan Event),
		pendingChan: make(map[string]chan Message),
		logger:      slog.New(slog.DiscardHandler),
	}
	client.state.Store(StateClosing)

	eventChanKeys := []string{"target1", "target2"}
	pendingChanKeys := []string{"id1", "id2", "id3"}

	for _, key := range eventChanKeys {
		client.eventChan[key] = make(chan Event)
	}

	for _, key := range pendingChanKeys {
		client.pendingChan[key] = make(chan Message)
	}

	client.cleanUp()

	if size := len(client.eventChan); size != 0 {
		t.Errorf("expected eventChan to be empty, got %d entries on eventChan", size)
	}

	if size := len(client.pendingChan); size != 0 {
		t.Errorf("expected pendingChan to be empty, got %d entries on pendingChan", size)
	}
}

func TestCleanUp_Concurrency(t *testing.T) {
	client := &Client{
		eventChan:   make(map[string]chan Event),
		pendingChan: make(map[string]chan Message),
		logger:      slog.New(slog.DiscardHandler),
	}
	client.state.Store(StateClosing)

	var wg sync.WaitGroup
	routines := 100
	start := make(chan struct{})

	for i := range routines {
		wg.Go(func() {
			<-start

			key := fmt.Sprintf("key%d", i)

			client.eventMu.Lock()
			client.eventChan[key] = make(chan Event)
			client.eventMu.Unlock()

			client.pendingMu.Lock()
			client.pendingChan[key] = make(chan Message)
			client.pendingMu.Unlock()
		})
	}

	close(start)

	client.cleanUp()
	wg.Wait()
}

func TestCleanUp_Channels(t *testing.T) {
	client := &Client{
		eventChan:   make(map[string]chan Event),
		pendingChan: make(map[string]chan Message),
		logger:      slog.New(slog.DiscardHandler),
	}
	client.state.Store(StateClosing)

	eventChans := make([]chan Event, 3)
	for i := range eventChans {
		ch := make(chan Event)
		eventChans[i] = ch
		client.eventChan[fmt.Sprintf("event%d", i)] = ch
	}

	pendingChans := make([]chan Message, 3)
	for i := range pendingChans {
		ch := make(chan Message)
		pendingChans[i] = ch
		client.pendingChan[fmt.Sprintf("pending%d", i)] = ch
	}

	client.cleanUp()

	for i, ch := range eventChans {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("event channel %d was not closed", i)
			}
		default:
			t.Errorf("event channel %d is still open", i)
		}
	}

	for i, ch := range pendingChans {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("pending channel %d was not closed", i)
			}
		default:
			t.Errorf("pending channel %d is still open", i)
		}
	}
}
