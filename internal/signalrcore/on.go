package signalrcore

import "fmt"

// On registers a listener for incoming server messages matching the specified target, returning a channel to receive the messages as extracted events.
//
// It returns an error if a listener is already registered for the given target.
func (c *Client) On(target string) (<-chan Event, error) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()

	_, exists := c.eventChan[target]
	if exists {
		return nil, fmt.Errorf("cannot listen on %q: %w", target, errDuplicateListener)
	}

	stream := make(chan Event, 100)
	c.eventChan[target] = stream

	return stream, nil
}
