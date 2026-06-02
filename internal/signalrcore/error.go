package signalrcore

// Err returns the underlying error that caused the client connection to terminate.
//
// It returns nil if called before the Done channel closes.
func (c *Client) Err() error {
	c.errorMu.Lock()
	defer c.errorMu.Unlock()

	return c.err
}
