package signalrcore

import "fmt"

func (c *Client) On(target string) (<-chan Event, error) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()

	_, exists := c.eventChan[target]
	if exists {
		return nil, fmt.Errorf("target %q already has a listener", target)
	}

	stream := make(chan Event, 100)
	c.eventChan[target] = stream

	return stream, nil
}
