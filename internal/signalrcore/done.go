package signalrcore

// Done returns a channel that is closed when the client connection has fully terminated.
func (c *Client) Done() <-chan struct{} {
	return c.doneChan
}
