# signalrcore

A minimalistic, concurrent SignalR Core client built for low-latency streaming pipelines.

## Motivation

**TL;DR:** I wanted it, so I built it.

While there are excellent, full-featured SignalR libraries available in the Go ecosystem, they are generally designed to cover every possible feature of the protocol. 

This client was built from the ground up to focus strictly on what is needed for high-throughput, low-latency streaming. By keeping it minimalistic and purpose-built, it avoids unnecessary architectural overhead and gives you direct, concurrent access to the raw JSON streams exactly as you need them.

## Quickstart

To initialize the client, subscribe to a data hub, and safely consume a real-time stream:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/eepzii/f1trackside/internal/signalrcore"
	"github.com/gorilla/websocket"
)

func main() {
	config := signalrcore.Config{
		Dialer: websocket.DefaultDialer,
		Client: http.DefaultClient,
		Logger: slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		)),
		IdleTimeout: 45 * time.Second,
		Token:       "your-auth-token-here", // Replace with your actual JWT token if endpoint is protected
	}

	// NOTE: Replace localhost:8080/hub with your target SignalR hub endpoint
	client, err := signalrcore.New("http://localhost:8080/hub", &config)
	if err != nil {
		log.Fatal(err)
	}

	feed, err := client.On("feed")
	if err != nil {
		log.Fatal(err)
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()

	if err := client.Start(startCtx); err != nil {
		log.Fatal(err)
	}

	invokeCtx, invokeCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer invokeCancel()

	invokeMsg, err := client.Invoke(invokeCtx, "Subscribe", "Arg1", "Arg2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Subscription confirmation:", string(invokeMsg))

	timer := time.NewTimer(3 * time.Hour)
	defer timer.Stop()

	for {
		select {
		case event, ok := <-feed:
			if !ok {
				// Set to nil to disable this select case and prevent infinite looping.
				// The loop will now rely on <-client.Done() to cleanly terminate.
				// (If you prefer to exit immediately when the stream drops, use 'return' here instead)
				feed = nil 
				continue
			}
			if event.Err != nil {
				fmt.Println("Stream error:", event.Err)
				continue
			}
			fmt.Println("Payload received:", string(event.Data))
		case <-timer.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := client.Close(ctx); err != nil {
				fmt.Println("Close error:", err)
			}
			cancel()
		case <-client.Done():
			fmt.Println("Client connection terminated:", client.Err())
			return
		}
	}
}

 ```
## API Reference

### Methods

#### `func New(httpURL string, config *Config) (*Client, error)`

Constructor for the SignalR Core client. Initializes internal state using the target endpoint and configuration settings.

#### `func (c *Client) On(target string) (<-chan Event, error)`

Registers a synchronous message listener for the specified target event. Returns a read-only channel where incoming payloads are routed. Returns an error if the target is already registered. **Must be called before** `Start()`.

#### `func (c *Client) Stop(target string)`

Deregisters an active listener. Closes the internal channel associated with the target. If no listener matches, this operation is a no-op.

#### `func (c *Client) Start(ctx context.Context) error`

Establishes the connection lifecycle. Executes the HTTP handshake negotiation, upgrades the connection to WebSockets, completes the protocol initialization sequence, and spawns the background stream reading goroutines.

#### `func (c *Client) Invoke(ctx context.Context, target string, arguments ...any) (json.RawMessage, error)`

Calls a server-side hub method with the provided arguments and blocks until the server returns a response payload.

#### `func (c *Client) Close(ctx context.Context) error`

Gracefully disconnects by transmitting a SignalR closure message followed by a standard WebSocket close frame (Code 1000). Use a bounded context to prevent execution stalls if the remote host is unresponsive.

#### `func (c *Client) Done() <-chan struct{}`

Returns a channel that closes when the client's internal loop has fully torn down and disconnected.

#### `func (c *Client) Err() error`

Returns the underlying cause of a connection failure. Always returns `nil` if checked before `Done()` completes.<br><br>

### Types & Data Structures

#### `type Client struct`

Manages the SignalR connection lifecycle, internal state, and the concurrent routing of messages.

#### `type Message struct`

A generalized structural mapping for the SignalR Hub Protocol JSON layout.

```go
type Message struct {
	Type         int             `json:"type"`
	Target       string          `json:"target,omitempty"`
	InvocationID string          `json:"invocationId,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	Arguments    json.RawMessage `json:"arguments,omitempty"`
	Error        string          `json:"error,omitempty"`
}
```

#### `type Config struct`

Initialization requirements and optional parameters.

```go
type Config struct {
	Client      *http.Client        // required for negotiation
	Dialer      *websocket.Dialer   // required for WebSocket upgrade
	Logger      *slog.Logger        // optional: defaults to slog.New(slog.DiscardHandler)
	IdleTimeout time.Duration       // optional: defaults to 45s
	Token       string              // optional: bearer authentication token
}
```

#### `type Event struct`

Encapsulates raw data fields or errors delivered by an active background server stream.

```go
type Event struct {
	Data []byte
	Err  error
}
```