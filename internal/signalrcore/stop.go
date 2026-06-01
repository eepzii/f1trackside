package signalrcore

func (c *Client) Stop(target string) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()

	if stream, exists := c.eventChan[target]; exists {
		close(stream)
		delete(c.eventChan, target)
	}
}
