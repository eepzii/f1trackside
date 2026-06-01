package signalrcore

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestErr(t *testing.T) {
	client := &Client{}

	if err := client.Err(); err != nil {
		t.Fatalf("expected error to be nil, got %v", err)
	}

	client.err = errors.New("error")
	if err := client.Err(); err == nil {
		t.Fatalf("expected error message, got %v", err)
	}
}

func TestErr_Concurrency(t *testing.T) {
	client := &Client{}

	routines := 100
	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := range routines {
		wg.Go(func() {
			<-start

			errMsg := fmt.Sprintf("error%d", i)
			client.errorMu.Lock()
			client.err = errors.New(errMsg)
			client.errorMu.Unlock()

			client.Err()
		})
	}

	close(start)
	wg.Wait()
}
