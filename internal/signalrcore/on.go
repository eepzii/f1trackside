package signalrcore

import "fmt"

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
