package signalrcore

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) write(ctx context.Context, msgType int, data []byte) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("write aborted: %w", err)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set write deadline: %w", err)
		}
		defer c.conn.SetWriteDeadline(time.Time{})
	}

	if err := c.conn.WriteMessage(msgType, data); err != nil {
		return fmt.Errorf("write failed (size %d bytes): %w", len(data), err)
	}

	return nil
}
