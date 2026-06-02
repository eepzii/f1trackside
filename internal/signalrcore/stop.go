package signalrcore

// Stop deregisters the listener for the specified target and closes its associated event channel.
//
// If no listener is currently registered for the target, Stop is a no-op.
func (c *Client) Stop(target string) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()

	if stream, exists := c.eventChan[target]; exists {
		close(stream)
		delete(c.eventChan, target)
	}
}
