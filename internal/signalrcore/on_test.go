package signalrcore

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestOn(t *testing.T) {
	client := &Client{
		eventChan: make(map[string]chan Event),
	}

	target := "target"
	stream, err := client.On(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, exists := client.eventChan[target]; !exists {
		t.Fatalf("expected channel to be added to client's eventChan map")
	}

	if cap(stream) != 100 {
		t.Fatalf("expected channel capacty of 100, got %d", cap(stream))
	}
}

func TestOn_DuplicateListener(t *testing.T) {
	client := &Client{
		eventChan: make(map[string]chan Event),
	}

	target := "target"

	_, err := client.On(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.On(target)
	if !errors.Is(err, errDuplicateListener) {
		t.Fatalf("expected %v, got %v", errDuplicateListener, err)
	}
}

func TestOn_Concurrency(t *testing.T) {
	client := &Client{
		eventChan: make(map[string]chan Event),
	}

	routines := 20
	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := range routines {
		wg.Go(func() {
			<-start

			target := fmt.Sprintf("target%d", i)
			_, err := client.On(target)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	close(start)
	wg.Wait()

	if len(client.eventChan) != routines {
		t.Fatalf("expected %d map entries in eventChan, got %d", routines, len(client.eventChan))
	}
}
