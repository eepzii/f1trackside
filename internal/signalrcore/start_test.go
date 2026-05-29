package signalrcore

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
)

func TestStart_Process(t *testing.T) {

	tests := []struct {
		name        string
		serverSetup func(t *testing.T) (*httptest.Server, string)
		checkErr    func(t *testing.T, err error)
	}{
		{
			name: "good start",
			serverSetup: func(t *testing.T) (*httptest.Server, string) {
				var upgrader websocket.Upgrader
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/negotiate" {
						w.Write([]byte(`{"connectionId":"abc1234"}`))
						return
					}

					conn, err := upgrader.Upgrade(w, r, http.Header{})
					if err != nil {
						t.Fatalf("could not upgrade: %v", err)
					}
					defer conn.Close()

					for {
						_, _, err := conn.ReadMessage()
						if err != nil {
							return
						}
						conn.WriteMessage(websocket.TextMessage, []byte("{}\x1e"))
					}
				}))
				return server, server.URL
			},
		},
		{
			name: "negotiation error",
			serverSetup: func(t *testing.T) (*httptest.Server, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				}))
				return server, server.URL
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errNegotiation)
				}

				if !errors.Is(err, errNegotiation) {
					t.Fatalf("expected %v, got %v", errNegotiation, err)
				}
			},
		},
		{
			name: "dial error",
			serverSetup: func(t *testing.T) (*httptest.Server, string) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/negotiate" {
						w.Write([]byte(`{"connectionId":"abc1234"}`))
					}
				}))
				return server, server.URL
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", websocket.ErrBadHandshake)
				}

				if !errors.Is(err, websocket.ErrBadHandshake) {
					t.Fatalf("expected %v, got %v", websocket.ErrBadHandshake, err)
				}
			},
		},
		{
			name: "handshake error",
			serverSetup: func(t *testing.T) (*httptest.Server, string) {
				var upgrader websocket.Upgrader
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/negotiate" {
						w.Write([]byte(`{"connectionId":"abc1234"}`))
						return
					}

					conn, err := upgrader.Upgrade(w, r, http.Header{})
					if err != nil {
						t.Fatalf("could not upgrade: %v", err)
					}
					defer conn.Close()

					for {
						if _, _, err := conn.ReadMessage(); err != nil {
							return
						}
						conn.WriteMessage(websocket.TextMessage, append([]byte(`{"error":"unsupported protocol"}`), 0x1e))
					}
				}))
				return server, server.URL
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errHandshake)
				}

				if !errors.Is(err, errHandshake) {
					t.Fatalf("expected %v, got %v", errHandshake, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server, baseURL := test.serverSetup(t)
			defer server.Close()

			httpURL, err := url.Parse(baseURL)
			if err != nil {
				t.Fatalf("test aborted: could not parse url %v", err)
			}

			wsURL, err := url.Parse(baseURL)
			if err != nil {
				t.Fatalf("test aborted: could not parse url %v", err)
			}
			wsURL.Scheme = "ws"

			client := &Client{
				baseURL:      httpURL,
				websocketURL: wsURL,

				doneChan: make(chan struct{}),

				client:      server.Client(),
				dialer:      websocket.DefaultDialer,
				logger:      slog.New(slog.DiscardHandler),
				idleTimeout: 45 * time.Second,
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err = client.Start(ctx)

			if test.checkErr != nil {
				test.checkErr(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			finalState := client.state.Load()

			if err != nil {
				if finalState != StateNew {
					t.Errorf("expected state %d on error, got %d", StateNew, finalState)
				}
				if client.conn != nil {
					if closeErr := client.conn.Close(); closeErr == nil {
						t.Errorf("expected websocket connection to be closed on error by the defer block")
					}
				}
			} else {
				if finalState != StateConnected {
					t.Errorf("expected state %d on success, got %d", StateConnected, finalState)
				}

				client.conn.Close()

				select {
				case <-client.doneChan:
				case <-time.NewTimer(3 * time.Second).C:
					t.Fatalf("expected to call client's .listen() method")
				}
			}
		})
	}
}

func TestStart_InvalidStates(t *testing.T) {

	tests := []struct {
		name         string
		invalidState uint32
	}{
		{
			name:         "rejects state connecting",
			invalidState: StateConnecting,
		},
		{
			name:         "rejects state connected",
			invalidState: StateConnected,
		},
		{
			name:         "rejects state closing",
			invalidState: StateClosing,
		},
		{
			name:         "rejects state closed",
			invalidState: StateClosed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{

				doneChan: make(chan struct{}),

				client: http.DefaultClient,
				dialer: websocket.DefaultDialer,
				logger: slog.New(slog.DiscardHandler),
			}
			client.state.Store(test.invalidState)

			err := client.Start(context.Background())

			if !errors.Is(err, errInvalidState) {
				t.Fatalf("expected %v, got %v", errInvalidState, err)
			}
		})
	}
}

func TestStart_Params(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/negotiate" {
			http.SetCookie(w, &http.Cookie{Name: "Cookie1", Value: "coo/kie/1"})
			http.SetCookie(w, &http.Cookie{Name: "Cookie2", Value: "coo/kie/2"})

			response := []byte(`{"negotiateVersion":0,"connectionId":"abc1234","availableTransports":[{"transport":"WebSockets","transferFormats":["Text","Binary"]},{"transport":"ServerSentEvents","transferFormats":["Text"]},{"transport":"LongPolling","transferFormats":["Text","Binary"]}]}`)
			w.Write(response)
			return
		}

		wantParams := url.Values{
			"id":        {"abc1234"},
			"transport": {"webSockets"},
		}
		if diff := cmp.Diff(wantParams, r.URL.Query()); diff != "" {
			t.Errorf("query params mismatch (-want +got):\n%s", diff)
		}

		if c, err := r.Cookie("Cookie1"); err != nil || c.Value != "coo/kie/1" {
			t.Errorf("expected Cookie1=coo/kie/1, got err/mismatch: %v", err)
		}
		if c, err := r.Cookie("Cookie2"); err != nil || c.Value != "coo/kie/2" {
			t.Errorf("expected Cookie2=coo/kie/2, got err/mismatch: %v", err)
		}

		if auth := r.Header.Get("Authorization"); auth != "Bearer j.w.t" {
			t.Errorf("expected Bearer j.w.t, got %q", auth)
		}
	}))
	defer server.Close()

	httpURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}

	wsURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}
	wsURL.Scheme = "ws"

	client := &Client{
		baseURL:      httpURL,
		websocketURL: wsURL,

		doneChan: make(chan struct{}),

		client:      server.Client(),
		dialer:      websocket.DefaultDialer,
		logger:      slog.New(slog.DiscardHandler),
		idleTimeout: 45 * time.Second,
		token:       "j.w.t",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = client.Start(ctx)

	if err == nil {
		t.Fatal("expected .Start() to fail")
	}

	if !errors.Is(err, websocket.ErrBadHandshake) {
		t.Fatalf("expected ErrBadHandshake, got %v", err)
	}
}

func TestStart_IsStateConnecting(t *testing.T) {
	var client *Client

	signal := make(chan struct{})
	var closeOnce sync.Once

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/negotiate" {
			if client.state.Load() == StateConnecting {
				closeOnce.Do(func() {
					close(signal)
				})
			}
			w.Write([]byte(`{"connectionId":"abc1234"}`))
		}
	}))
	defer server.Close()

	httpURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}

	wsURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}
	wsURL.Scheme = "ws"

	client = &Client{
		baseURL:      httpURL,
		websocketURL: wsURL,

		doneChan: make(chan struct{}),

		client:      server.Client(),
		dialer:      websocket.DefaultDialer,
		logger:      slog.New(slog.DiscardHandler),
		idleTimeout: 45 * time.Second,
	}

	go client.Start(t.Context())

	select {
	case <-signal:
	case <-time.NewTimer(3 * time.Second).C:
		t.Fatal("could not detect StateConnecting during .Start() execution")
	}
}

func TestStart_Concurrency(t *testing.T) {
	var upgrader websocket.Upgrader
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/negotiate" {
			w.Write([]byte(`{"connectionId":"abc1234"}`))
			return
		}
		conn, err := upgrader.Upgrade(w, r, http.Header{})
		if err != nil {
			t.Fatalf("could not upgrade: %v", err)
		}
		defer conn.Close()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}

			conn.WriteMessage(websocket.TextMessage, append([]byte(`{}`), 0x1e))
		}
	}))
	defer server.Close()

	httpURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}

	wsURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}
	wsURL.Scheme = "ws"

	client := &Client{
		baseURL:      httpURL,
		websocketURL: wsURL,

		doneChan: make(chan struct{}),

		client:      server.Client(),
		dialer:      websocket.DefaultDialer,
		logger:      slog.New(slog.DiscardHandler),
		idleTimeout: 45 * time.Second,
	}

	routines := 20
	start := make(chan struct{})
	var wg sync.WaitGroup

	var successCount atomic.Int32
	var stateErrCount atomic.Int32

	for range routines {
		wg.Go(func() {
			<-start

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := client.Start(ctx)

			if err == nil {
				successCount.Add(1)
			} else if errors.Is(err, errInvalidState) {
				stateErrCount.Add(1)
			} else {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}

	close(start)
	wg.Wait()

	if success := successCount.Load(); success != 1 {
		t.Errorf("expected exactly 1 successful .Start() execution, got %v", success)
	}

	if stateErrs := stateErrCount.Load(); stateErrs != int32(routines-1) {
		t.Errorf("expected exactly %d state errors, got %d", routines-1, stateErrs)
	}
}
