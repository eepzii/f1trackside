package signalrcore

func (c *Client) Done() <-chan struct{} {
	return c.doneChan
}
