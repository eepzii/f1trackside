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

func (c *Client) handleCompletion(msg Message) error {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()

	ch, ok := c.pendingChan[msg.InvocationID]
	if !ok {
		return errChannelUnavailable
	}

	select {
	case ch <- msg:
		c.logger.Debug("completion handled",
			"invocation_id", msg.InvocationID,
			"buffer_size", len(ch),
		)
	default:
		return fmt.Errorf("%w (capacity: %d)", errBufferOverflow, cap(ch))
	}

	return nil
}
