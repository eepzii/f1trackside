package signalrcore

import (
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestHandleInvocation_Message(t *testing.T) {

	tests := []struct {
		name string
		msg  Message
		want Event
	}{
		{
			name: "good message",
			msg: Message{
				Target:    "good_message",
				Arguments: []byte(`{"good":"message"}`),
			},
			want: Event{
				Data: []byte(`{"good":"message"}`),
			},
		},
		{
			name: "error message",
			msg: Message{
				Target:    "error_message",
				Arguments: []byte(`{"error":"message"}`),
				Error:     "error",
			},
			want: Event{
				Data: []byte(`{"error":"message"}`),
				Err:  errors.New("error"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{
				eventChan: make(map[string]chan Event),
				logger:    slog.New(slog.DiscardHandler),
			}

			target := test.msg.Target
			client.eventChan[target] = make(chan Event, 1)

			err := client.handleInvocation(test.msg)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			event := <-client.eventChan[target]

			errComparer := cmp.Comparer(func(x, y error) bool {
				if x == nil || y == nil {
					return x == nil && y == nil
				}

				return x.Error() == y.Error()
			})

			if diff := cmp.Diff(test.want, event, errComparer); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}

}

func TestHandleInvocation_Mutex(t *testing.T) {
	client := &Client{
		eventChan: make(map[string]chan Event),
		logger:    slog.New(slog.DiscardHandler),
	}

	client.eventChan["target1"] = make(chan Event, 1)

	start := make(chan struct{})
	var wg sync.WaitGroup

	wg.Go(func() {
		<-start
		client.handleInvocation(Message{
			Target: "target1",
		})
	})

	wg.Go(func() {
		<-start
		client.eventMu.Lock()
		if ch, ok := client.eventChan["target1"]; ok {
			delete(client.eventChan, "target1")
			close(ch)
		}
		client.eventMu.Unlock()
	})

	close(start)
	wg.Wait()
}

func TestHandleInvocation_Returns(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		client := &Client{
			eventChan: make(map[string]chan Event),
			logger:    slog.New(slog.DiscardHandler),
		}

		target := "test"
		client.eventChan[target] = make(chan Event, 1)

		err := client.handleInvocation(Message{
			Target: target,
		})

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("channel unavailable", func(t *testing.T) {
		client := &Client{}

		err := client.handleInvocation(Message{})

		if !errors.Is(err, errChannelUnavailable) {
			t.Fatalf("expected %v, got %v", errChannelUnavailable, err)
		}
	})

	t.Run("buffer overflow", func(t *testing.T) {
		client := &Client{
			eventChan: make(map[string]chan Event),
			logger:    slog.New(slog.DiscardHandler),
		}

		target := "test"
		client.eventChan[target] = make(chan Event)

		errChan := make(chan error)
		go func() {
			errChan <- client.handleInvocation(Message{
				Target: target,
			})
		}()

		select {
		case err := <-errChan:
			if !errors.Is(err, errBufferOverflow) {
				t.Fatalf("expected %v, got %v", errBufferOverflow, err)
			}
		case <-time.NewTimer(time.Second).C:
			t.Fatalf("expected %v, got timed out function", errBufferOverflow)
		}
	})
}

func TestHandleCompletion_Mutex(t *testing.T) {
	client := &Client{
		pendingChan: make(map[string]chan Message),
		logger:      slog.New(slog.DiscardHandler),
	}

	client.pendingChan["id1"] = make(chan Message, 1)

	start := make(chan struct{})
	var wg sync.WaitGroup

	wg.Go(func() {
		<-start
		client.handleCompletion(Message{
			InvocationID: "id1",
		})
	})

	wg.Go(func() {
		<-start
		client.pendingMu.Lock()
		if ch, ok := client.pendingChan["id1"]; ok {
			delete(client.pendingChan, "id1")
			close(ch)
		}
		client.pendingMu.Unlock()
	})

	close(start)
	wg.Wait()
}

func TestHandleCompletion_Returns(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		client := &Client{
			pendingChan: make(map[string]chan Message),
			logger:      slog.New(slog.DiscardHandler),
		}

		id := "test"
		client.pendingChan[id] = make(chan Message, 1)

		err := client.handleCompletion(Message{
			InvocationID: id,
		})

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("channel unavailable", func(t *testing.T) {
		client := &Client{}

		err := client.handleCompletion(Message{})

		if !errors.Is(err, errChannelUnavailable) {
			t.Fatalf("expected %v, got %v", errChannelUnavailable, err)
		}
	})

	t.Run("buffer overflow", func(t *testing.T) {
		client := &Client{
			pendingChan: make(map[string]chan Message),
			logger:      slog.New(slog.DiscardHandler),
		}

		id := "test"
		client.pendingChan[id] = make(chan Message)

		errChan := make(chan error)
		go func() {
			errChan <- client.handleCompletion(Message{
				InvocationID: id,
			})
		}()

		select {
		case err := <-errChan:
			if !errors.Is(err, errBufferOverflow) {
				t.Fatalf("expected %v, got %v", errBufferOverflow, err)
			}
		case <-time.NewTimer(time.Second).C:
			t.Fatalf("expected %v, got timed out function", errBufferOverflow)
		}
	})
}
