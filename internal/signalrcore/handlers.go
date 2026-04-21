package signalrcore

import "errors"

func (c *Client) handleInvocation(msg Message) {
	c.eventMu.Lock()
	ch, ok := c.eventChan[msg.Target]
	c.eventMu.Unlock()

	if !ok {
		return
	}

	var err error
	if msg.Error != "" {
		err = errors.New(msg.Error)
	}

	select {
	case ch <- Event{
		Data: msg.Arguments,
		Err:  err,
	}:
		c.logger.Debug("event sent", "target", msg.Target, "buffer_size", len(ch))
	default:
		c.logger.Warn("dropping server invocation: subscriber buffer overflow",
			"target", msg.Target,
			"buffer_capacity", cap(ch),
		)
	}
}

func (c *Client) handleCompletion(msg Message) {
	c.pendingMu.Lock()
	ch, ok := c.pendingChan[msg.InvocationID]
	c.pendingMu.Unlock()
	if !ok {
		c.logger.Warn("received completion for untracked invocation",
			"invocation_id", msg.InvocationID,
			"buffer_size", len(ch),
			"note", "request may have timed out or been cancelled",
		)
		return
	}
	ch <- msg
}

func (c *Client) handleClose(msg Message) {
	var err error
	if msg.Error != "" {
		err = errors.New(msg.Error)
	}

	c.errorMu.Lock()
	c.err = err
	c.errorMu.Unlock()
}
