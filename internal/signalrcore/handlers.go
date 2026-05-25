package signalrcore

import (
	"errors"
	"fmt"
)

func (c *Client) handleInvocation(msg Message) error {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()

	ch, ok := c.eventChan[msg.Target]
	if !ok {
		return errChannelUnavailable
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
		c.logger.Debug("event handled", "target", msg.Target, "buffer_size", len(ch))
	default:
		return fmt.Errorf("%w (capacity: %d)", errBufferOverflow, cap(ch))
	}

	return nil
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
