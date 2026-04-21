package signalrcore

func (c *Client) Stop(target string) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()

	delete(c.eventChan, target)
}
