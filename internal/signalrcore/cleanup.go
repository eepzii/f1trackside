package signalrcore

func (c *Client) cleanUp() {
	current := c.state.Load()

	if current != StateConnected && current != StateClosing {
		return
	}

	if !c.state.CompareAndSwap(current, StateClosed) {
		return
	}

	if c.conn != nil {
		c.conn.Close()
	}

	c.eventMu.Lock()
	for target, stream := range c.eventChan {
		close(stream)
		delete(c.eventChan, target)
	}
	c.eventMu.Unlock()

	c.pendingMu.Lock()
	for id, replyChan := range c.pendingChan {
		close(replyChan)
		delete(c.pendingChan, id)
	}
	c.pendingMu.Unlock()

	c.logger.Info("cleanup complete")
}
