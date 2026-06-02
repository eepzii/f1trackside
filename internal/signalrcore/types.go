package signalrcore

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Client manages the SignalR connection lifecycle, internal state, and the concurrent routing of messages.
type Client struct {
	baseURL      *url.URL
	websocketURL *url.URL

	eventChan   map[string]chan Event
	eventMu     sync.Mutex
	pendingChan map[string]chan Message
	pendingMu   sync.Mutex
	doneChan    chan struct{}
	err         error
	errorMu     sync.Mutex

	client      *http.Client
	dialer      *websocket.Dialer
	logger      *slog.Logger
	idleTimeout time.Duration
	token       string

	conn    *websocket.Conn
	writeMu sync.Mutex

	invocationID atomic.Uint64
	state        atomic.Uint32
}

// Message represents a generalized SignalR Hub Protocol payload used for communicating with the server.
type Message struct {
	Type         int             `json:"type"`
	Target       string          `json:"target,omitempty"`
	InvocationID string          `json:"invocationId,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	Arguments    json.RawMessage `json:"arguments,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// Config defines the initialization options for creating a new SignalR client.
type Config struct {
	Client      *http.Client
	Dialer      *websocket.Dialer
	Logger      *slog.Logger
	IdleTimeout time.Duration
	Token       string
}

// Event encapsulates the raw data or error extracted from an incoming server message.
type Event struct {
	Data []byte
	Err  error
}

type negotiation struct {
	body    negotiationBody
	cookies []*http.Cookie
}

type negotiationBody struct {
	ConnectionID        string `json:"connectionId"`
	AvailableTransports []struct {
		Transport       string   `json:"transport"`
		TransferFormats []string `json:"transferFormats"`
	} `json:"availableTransports"`
	NegotiateVersion int `json:"negotiateVersion"`
}
