package signalrcore

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

func (c *Client) listen() {
	defer func() {
		close(c.doneChan)
		c.cleanUp()
	}()

	for {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.idleTimeout)); err != nil {
			c.errorMu.Lock()
			c.err = errors.New("cannot set read deadline")
			c.errorMu.Unlock()
			return
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.errorMu.Lock()
			c.err = err
			c.errorMu.Unlock()
			return
		}

		for sub := range bytes.SplitSeq(data, []byte{0x1e}) {
			if len(sub) <= 0 {
				continue
			}

			var msg Message
			if err := json.Unmarshal(sub, &msg); err != nil {
				c.logger.Error("failed to parse message", "raw_message", msg)
				continue
			}

			switch msg.Type {
			case 1:
				err := c.handleInvocation(msg)

				if errors.Is(err, errChannelUnavailable) || errors.Is(err, errBufferOverflow) {
					c.logger.Warn("invocation dropped",
						"target", msg.Target,
						"reason", err,
					)
					continue
				}

				if err != nil {
					c.logger.Error("unexpected error", "reason", err)
				}
			case 3:
				err := c.handleCompletion(msg)

				if errors.Is(err, errChannelUnavailable) || errors.Is(err, errBufferOverflow) {
					c.logger.Warn("completion dropped",
						"invocation_id", msg.InvocationID,
						"reason", err,
					)
					continue
				}

				if err != nil {
					c.logger.Error("unexpected error", "reason", err)
				}
			case 6:
				c.logger.Debug("ping received")
			case 7:
				c.handleClose(msg)
				return
			default:
				c.logger.Error("message type unsupported",
					"type", msg.Type,
					"target", msg.Target,
					"raw_message", msg,
				)
			}
		}
	}
}
