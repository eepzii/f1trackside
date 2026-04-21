package signalrcore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gorilla/websocket"
)

func (c *Client) Invoke(ctx context.Context, target string, arguments ...any) (json.RawMessage, error) {
	if c.state.Load() != StateConnected {
		return nil, errors.New("client not connected")
	}

	args, err := json.Marshal(arguments)
	if err != nil {
		return nil, err
	}

	nextID := c.invocationID.Add(1)
	invocationID := strconv.FormatUint(nextID, 10)
	receiverChan := make(chan Message, 1)

	c.pendingMu.Lock()
	c.pendingChan[invocationID] = receiverChan
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pendingChan, invocationID)
		c.pendingMu.Unlock()
	}()

	msg, err := json.Marshal(
		Message{
			Type:         1,
			Target:       target,
			InvocationID: invocationID,
			Arguments:    args,
		},
	)
	if err != nil {
		return nil, err
	}

	msg = append(msg, 0x1e)

	if err := c.write(ctx, websocket.TextMessage, msg); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case reply, ok := <-receiverChan:
		if !ok {
			return nil, errors.New("pending channel for invocation closed")
		}

		if reply.Error != "" {
			return nil, fmt.Errorf("invocation failed with %s", reply.Error)
		}

		return reply.Result, nil
	}
}
