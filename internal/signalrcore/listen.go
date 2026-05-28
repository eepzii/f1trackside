package signalrcore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (c *Client) listen() {
	defer func() {
		c.cleanUp()
		close(c.doneChan)
	}()

	for {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.idleTimeout)); err != nil {
			c.errorMu.Lock()
			c.err = fmt.Errorf("failed to set read deadline: %w", err)
			c.errorMu.Unlock()
			return
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.errorMu.Lock()
			c.err = fmt.Errorf("read loop closed: %w", err)
			c.errorMu.Unlock()
			return
		}

		for sub := range bytes.SplitSeq(data, []byte{0x1e}) {
			if len(sub) <= 0 {
				continue
			}

			var msg Message
			if err := json.Unmarshal(sub, &msg); err != nil {
				c.logger.Error("failed to parse message", "error", err, "size", len(sub))
				c.logger.Debug("malformed message payload", "raw_message", string(sub))
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
					c.logger.Error("failed to handle invocation", "reason", err)
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
					c.logger.Error("failed to handle completion", "reason", err)
				}
			case 6:
				c.logger.Debug("ping received")
			case 7:
				if msg.Error == "" {
					time.AfterFunc(3*time.Second, func() {
						c.conn.Close()
					})

					c.logger.Debug("graceful disconnect")
					continue
				}

				c.errorMu.Lock()
				c.err = fmt.Errorf("%w (%s)", errServerClosed, msg.Error)
				c.errorMu.Unlock()

				c.logger.Warn("server terminated connection", "reason", msg.Error)
				return
			default:
				c.logger.Error("unsupported message type",
					"type", msg.Type,
					"target", msg.Target,
					"size", len(sub),
				)

				c.logger.Debug("unsupported message payload", "raw_message", string(sub))
			}
		}
	}
}
