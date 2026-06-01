package signalrcore

import (
	"sync"
	"testing"
)

func TestStop(t *testing.T) {
	client := &Client{
		eventChan: make(map[string]chan Event),
	}

	target := "target"

	stream := make(chan Event, 100)
	client.eventChan[target] = stream

	client.Stop(target)

	if _, exists := client.eventChan[target]; exists {
		t.Fatalf("want target %q removed in eventChan map, got exists", target)
	}

	select {
	case _, ok := <-stream:
		if ok {
			t.Fatalf("expected channel to be closed")
		}
	default:
		t.Fatalf(`want closed channel, got "<-stream" channel blocked after .Stop()`)
	}

	client.Stop(target)
}

func TestStop_Concurrency(t *testing.T) {
	client := &Client{
		eventChan: make(map[string]chan Event),
	}

	target := "target"

	stream := make(chan Event, 100)
	client.eventChan[target] = stream

	routines := 20
	start := make(chan struct{})
	var wg sync.WaitGroup

	for range routines {
		wg.Go(func() {
			<-start
			client.Stop(target)
		})
	}

	close(start)
	wg.Wait()

	select {
	case _, ok := <-stream:
		if ok {
			t.Fatalf("expected channel to be closed")
		}
	default:
		t.Fatalf(`want closed channel, got "<-stream" channel blocked after .Stop()`)
	}

	if _, exists := client.eventChan[target]; exists {
		t.Fatalf("expected target to be removed from map after concurrent stops")
	}
}
