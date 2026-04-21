package signalrcore

import (
	"context"
	"time"
)

func (c *Client) write(ctx context.Context, msgType int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}

	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		return err
	}

	defer c.conn.SetWriteDeadline(time.Time{})

	return c.conn.WriteMessage(msgType, data)
}
