package signalrcore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gorilla/websocket"
)

func (c *Client) Invoke(ctx context.Context, target string, arguments ...any) (json.RawMessage, error) {
	if c.state.Load() != StateConnected {
		return nil, fmt.Errorf("cannot invoke %q: %w", target, errNotConnected)
	}

	args, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments on %q: %w", target, err)
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
		return nil, fmt.Errorf("failed to marshal invocation on %q: %w", target, err)
	}

	msg = append(msg, 0x1e)

	if err := c.write(ctx, websocket.TextMessage, msg); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("invocation %q aborted by context: %w", target, ctx.Err())
	case reply, ok := <-receiverChan:
		if !ok {
			return nil, fmt.Errorf("pending channel closed unexpectedly: %w", errChannelUnavailable)
		}

		if reply.Error != "" {
			return nil, fmt.Errorf("invocation rejected with %q: %w", reply.Error, errInvocation)
		}

		return reply.Result, nil
	}
}
