package signalrcore

func (c *Client) Err() error {
	c.errorMu.Lock()
	defer c.errorMu.Unlock()

	return c.err
}
