package signalrcore

import (
	"testing"
)

func TestDone(t *testing.T) {
	client := &Client{
		doneChan: make(chan struct{}),
	}

	got := client.Done()

	if got == nil {
		t.Fatalf("expected .Done() to return an initialized channel, got nil")
	}
}
